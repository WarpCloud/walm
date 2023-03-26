package impl

import (
	"WarpCloud/walm/pkg/helm/impl/plugins"
	"WarpCloud/walm/pkg/k8s"
	k8sHelm "WarpCloud/walm/pkg/k8s/client/helm"
	"WarpCloud/walm/pkg/k8s/utils"
	"WarpCloud/walm/pkg/models/common"
	k8sModel "WarpCloud/walm/pkg/models/k8s"
	"WarpCloud/walm/pkg/models/release"
	"WarpCloud/walm/pkg/redis"
	"WarpCloud/walm/pkg/setting"
	"WarpCloud/walm/pkg/util"
	"WarpCloud/walm/pkg/util/transwarpjsonnet"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/ghodss/yaml"
	"github.com/go-resty/resty"
	"github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
	"helm.sh/helm/pkg/action"
	"helm.sh/helm/pkg/chart"
	"helm.sh/helm/pkg/chart/loader"
	"helm.sh/helm/pkg/chartutil"
	"helm.sh/helm/pkg/kube"
	"helm.sh/helm/pkg/registry"
	helmRelease "helm.sh/helm/pkg/release"
	"helm.sh/helm/pkg/storage"
	"helm.sh/helm/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/klog"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	compatibleNamespace    = "kube-system"
	releaseMaxHistory      = 10
	defaultDownloadTimeout = 5 * time.Second
)

type ChartRepository struct {
	Name     string
	URL      string
	Username string
	Password string
}

type Helm struct {
	chartRepoMap   map[string]*ChartRepository
	registryClient *registry.Client
	k8sCache       k8s.Cache
	list           *action.List
	kubeClients    *k8sHelm.Client
	restyClient    *resty.Client
	actionConfigs  *lru.Cache
}

func (helmImpl *Helm) getActionConfig(namespace string) (*action.Configuration, error) {
	if actionConfig, ok := helmImpl.actionConfigs.Get(namespace); ok {
		return actionConfig.(*action.Configuration), nil
	} else {
		kubeConfig, kubeClient := helmImpl.kubeClients.GetKubeClient(namespace)
		clientset, err := kubeClient.Factory.KubernetesClientSet()
		if err != nil {
			klog.Errorf("failed to get clientset: %s", err.Error())
			return nil, err
		}

		d := driver.NewConfigMapsEx(clientset.CoreV1().ConfigMaps(namespace), clientset.CoreV1().ConfigMaps(compatibleNamespace), namespace)
		d.Log = klog.Infof
		store := storage.Init(d)
		config := &action.Configuration{
			KubeClient:       kubeClient,
			Releases:         store,
			RESTClientGetter: kubeConfig,
			Log:              klog.Infof,
		}
		helmImpl.actionConfigs.Add(namespace, config)
		return config, nil
	}
}

func (helmImpl *Helm) ListAllReleases() (releaseCaches []*release.ReleaseCache, err error) {
	helmReleases, err := helmImpl.list.Run()
	if err != nil {
		klog.Errorf("failed to list helm releases: %s\n", err.Error())
		return nil, err
	}

	filteredHelmReleases := filterHelmReleases(helmReleases)
	for _, helmRelease := range filteredHelmReleases {
		releaseCache, err := helmImpl.convertHelmRelease(helmRelease)
		if err != nil {
			klog.Errorf("failed to convert helm release %s/%s : %s", helmRelease.Namespace, helmRelease.Name, err.Error())
			return nil, err
		}
		releaseCaches = append(releaseCaches, releaseCache)
	}
	return
}

// keep latest deployed one. if there is no deployed one ,keep the latest version.
func filterHelmReleases(releases []*helmRelease.Release) (filteredReleases map[string]*helmRelease.Release) {
	filteredReleases = map[string]*helmRelease.Release{}
	for _, release := range releases {
		filedName := redis.BuildFieldName(release.Namespace, release.Name)
		if existedRelease, ok := filteredReleases[filedName]; ok {
			if existedRelease.Info != nil && existedRelease.Info.Status == helmRelease.StatusDeployed {
				if release.Info != nil && release.Info.Status == helmRelease.StatusDeployed &&
					existedRelease.Version < release.Version {
					filteredReleases[filedName] = release
				}
			} else {
				if release.Info != nil && release.Info.Status == helmRelease.StatusDeployed {
					filteredReleases[filedName] = release
				} else if existedRelease.Version < release.Version {
					filteredReleases[filedName] = release
				}
			}
		} else {
			filteredReleases[filedName] = release
		}
	}
	return
}

func (helmImpl *Helm) DeleteRelease(namespace string, name string) error {
	action, err := helmImpl.getDeleteAction(namespace)
	if err != nil {
		klog.Errorf("failed to get current helm client : %s", err.Error())
		return err
	}

	_, err = action.Run(name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			klog.Warningf("release %s is not found from helm", name)
		} else {
			klog.Errorf("failed to delete release from helm : %s", err.Error())
			return err
		}
	}
	return nil
}

func (helmImpl *Helm) loadChart(chartFiles []*common.BufferedFile, releaseRequest *release.ReleaseRequestV2) (
	rawChart *chart.Chart, err error) {
	// priority: chartFiles > chartImage > chartName
	if chartFiles != nil {
		rawChart, err = loader.LoadFiles(convertBufferFiles(chartFiles))
	} else if releaseRequest.ChartImage != "" {
		rawChart, err = helmImpl.getRawChartByImage(releaseRequest.ChartImage)
	} else {
		rawChart, err = helmImpl.getRawChartFromRepo(releaseRequest.RepoName, releaseRequest.ChartName, releaseRequest.ChartVersion)
	}
	return
}

// 优先级
// 1. metaInfoParams
// 2. prettyParams converted to metaInfoParams
// 3. prettyParams
func processAdvancedParams(chartInfo *release.ChartDetailInfo, releaseRequest *release.ReleaseRequestV2) error {
	if releaseRequest.MetaInfoParams != nil {
		return processMetaInfoParams(chartInfo.MetaInfo, releaseRequest.MetaInfoParams, releaseRequest.ConfigValues)
	} else if releaseRequest.ReleasePrettyParams != nil {
		if chartInfo.MetaInfo != nil {
			metaInfoParams, err := convertPrettyParamsToMetainfoParams(chartInfo.MetaInfo, releaseRequest.ReleasePrettyParams)
			if err != nil {
				klog.Errorf("failed to convert pretty params to meta info params")
				return err
			}
			return processMetaInfoParams(chartInfo.MetaInfo, metaInfoParams, releaseRequest.ConfigValues)
		} else {
			processPrettyParams(&releaseRequest.ReleaseRequest)
		}
	}
	return nil
}

func processMetaInfoParams(metaInfo *release.ChartMetaInfo, metaInfoParams *release.MetaInfoParams, configValues map[string]interface{}) error {
	klog.Info("processing meta info params")
	metaInfoConfigs, err := metaInfoParams.BuildConfigValues(metaInfo)
	if err != nil {
		klog.Errorf("failed to get meta info parameters : %s", err.Error())
		return err
	}
	util.MergeValues(configValues, metaInfoConfigs, false)
	return nil
}

func (helmImpl *Helm) InstallOrCreateRelease(namespace string, releaseRequest *release.ReleaseRequestV2, chartFiles []*common.BufferedFile,
	dryRun bool, update bool, oldReleaseInfo *release.ReleaseInfoV2) (*release.ReleaseCache, error) {
	return helmImpl.InstallOrCreateReleaseWithStrict(namespace, releaseRequest, chartFiles, dryRun, update, oldReleaseInfo, false,true)
}

func (helmImpl *Helm) InstallOrCreateReleaseWithStrict(namespace string, releaseRequest *release.ReleaseRequestV2, chartFiles []*common.BufferedFile,
	dryRun bool, update bool, oldReleaseInfo *release.ReleaseInfoV2, fullUpdate bool, strict bool) (*release.ReleaseCache, error) {
	rawChart, err := helmImpl.loadChart(chartFiles, releaseRequest)
	if err != nil {
		klog.Errorf("failed to load chart : %s", err.Error())
		return nil, err
	}

	chartInfo, err := buildChartInfo(rawChart)
	if err != nil {
		klog.Errorf("failed to build chart info : %s", err.Error())
		return nil, err
	}

	if releaseRequest.ConfigValues == nil {
		releaseRequest.ConfigValues = map[string]interface{}{}
	}

	err = processAdvancedParams(chartInfo, releaseRequest)
		if err != nil {
		klog.Errorf("failed to process advanced params : %s", err.Error())
			return nil, err
		}

	dependencies := releaseRequest.Dependencies
	releaseLabels := releaseRequest.ReleaseLabels
	releasePlugins := releaseRequest.Plugins
	configValues := releaseRequest.ConfigValues
	isomateConfig := releaseRequest.IsomateConfig
	if update {
		// reuse config values, dependencies, release labels, walm plugins
		configValues, dependencies, releaseLabels, releasePlugins, err = reuseReleaseRequest(oldReleaseInfo, releaseRequest, fullUpdate)
		if err != nil {
			klog.Errorf("failed to reuse release request : %s", err.Error())
			return nil, err
		}

		err = mergeIsomateConfig(isomateConfig, oldReleaseInfo.IsomateConfig)
		if err != nil {
			klog.Errorf("failed to merge old isomate config: %s", err.Error())
			return nil, err
		}
	}

	// merge chart default plugins
	if chartInfo.MetaInfo != nil {
		releasePlugins, err = mergeReleasePlugins(releasePlugins, chartInfo.MetaInfo.Plugins)
		if err != nil {
			klog.Errorf("failed to merge chart default plugins : %s", err.Error())
			return nil, err
		}
	}

	// get all the dependency releases' output configs from ReleaseConfig or dummy service(for compatible)
	dependencyConfigs, err := helmImpl.GetDependencyOutputConfigs(namespace, dependencies, chartInfo, strict)
	if err != nil {
		klog.Errorf("failed to get all the dependency releases' output configs : %s", err.Error())
		return nil, err
	}

	// add default plugin
	releasePlugins = addDefaultPlugins(releasePlugins)

	// compatible
	if chartInfo.WalmVersion == common.WalmVersionV1 {
		value, ok := configValues[transwarpjsonnet.TranswarpInstallIDKey]
		if !ok || (ok && value == "") {
			configValues[transwarpjsonnet.TranswarpInstallIDKey] = rand.String(5)
		}
	}

	valueOverride := map[string]interface{}{}
	util.MergeValues(valueOverride, dependencyConfigs, false)
	util.MergeValues(valueOverride, configValues, false)

	valueOverride[plugins.WalmPluginConfigKey] = releasePlugins

	if releaseRequest.IsomateConfig != nil && len(releaseRequest.IsomateConfig.Isomates) > 0 {
		err = helmImpl.processChartWithIsomate(chartInfo, releaseRequest,
			rawChart, namespace, configValues, dependencyConfigs, dependencies,
			releaseLabels, releasePlugins, valueOverride, update)
	if err != nil {
			klog.Errorf("failed to process chart with isomate config : %s", err.Error())
		return nil, err
	}
	} else {
		err = transwarpjsonnet.ProcessChart(chartInfo, releaseRequest, rawChart, namespace, configValues, dependencyConfigs, dependencies, releaseLabels, nil)
		if err != nil {
			return nil, err
		}
	}

	releaseCache, err := helmImpl.doInstallUpgradeReleaseFromChart(namespace, releaseRequest.Name, rawChart, valueOverride, update, dryRun, releasePlugins)
	if err != nil {
		klog.Errorf("failed to create or update release from chart : %s", err.Error())
		return nil, err
	}

	return releaseCache, nil
}

func addDefaultPlugins(releasePlugins []*k8sModel.ReleasePlugin) []*k8sModel.ReleasePlugin {
	releasePlugins = append(releasePlugins, &k8sModel.ReleasePlugin{
		Name: plugins.ValidateReleaseConfigPluginName,
	}, &k8sModel.ReleasePlugin{
		Name: plugins.IsomateSetConverterPluginName,
	})
	return releasePlugins
}

func mergePausePlugin(paused bool, releasePlugins []*k8sModel.ReleasePlugin) (mergedPlugins []*k8sModel.ReleasePlugin, err error) {
	if paused {
		mergedPlugins, err = mergeReleasePlugins([]*k8sModel.ReleasePlugin{
				{
					Name:    plugins.PauseReleasePluginName,
					Version: "1.0",
				},
			}, releasePlugins)
		} else {
		mergedPlugins, err = mergeReleasePlugins([]*k8sModel.ReleasePlugin{
				{
					Name:    plugins.PauseReleasePluginName,
					Version: "1.0",
					Disable: true,
				},
			}, releasePlugins)
		}
	return
}

func (helmImpl *Helm) processChartWithIsomate(chartInfo *release.ChartDetailInfo, releaseRequest *release.ReleaseRequestV2,
	rawChart *chart.Chart, namespace string, configValues, dependencyConfigs map[string]interface{}, dependencies,
	releaseLabels map[string]string, releasePlugins []*k8sModel.ReleasePlugin, valueOverride map[string]interface{}, update bool) (err error) {
	rawChartTemplates := []*chart.File{}
	for _, template := range rawChart.Templates {
		rawChartTemplates = append(rawChartTemplates, template)
	}
	rawChartFiles := []*chart.File{}
	for _, file := range rawChart.Files {
		rawChartFiles = append(rawChartFiles, file)
	}

	manifests := map[string]string{}
	var chartFiles []*chart.File
	for _, isomate := range releaseRequest.IsomateConfig.Isomates {
		err = transwarpjsonnet.ProcessChart(chartInfo, releaseRequest, rawChart, namespace, configValues, dependencyConfigs, dependencies, releaseLabels, isomate.ConfigValues)
		if err != nil {
			return err
		}

		isomateReleasePlugins, err := mergeReleasePlugins(isomate.Plugins, releasePlugins)
		if err != nil {
			klog.Errorf("failed to merge release plugins : %s", err.Error())
			return err
		}

		args := plugins.IsomateNameArgs{
			Name:           isomate.Name,
			DefaultIsomate: isomate.Name == releaseRequest.IsomateConfig.DefaultIsomateName,
		}

		argsBytes, err := json.Marshal(args)
		if err != nil {
			klog.Errorf("failed to marshal isomate name args : %s", err.Error())
			return err
		}
		isomateReleasePlugins = append(isomateReleasePlugins, &k8sModel.ReleasePlugin{
			Name: plugins.IsomateNamePluginName,
			Args: string(argsBytes),
		})

		isomateValueOverride := map[string]interface{}{}
		util.MergeValues(isomateValueOverride, valueOverride, false)
		util.MergeValues(isomateValueOverride, isomate.ConfigValues, false)

		releaseCache, err := helmImpl.doInstallUpgradeReleaseFromChart(namespace, releaseRequest.Name, rawChart, isomateValueOverride, update, true, isomateReleasePlugins)
		if err != nil {
			klog.Errorf("failed to create or update release from chart : %s", err.Error())
			return err
		}

		manifests[isomate.Name] = releaseCache.Manifest

		rawChart.Templates = rawChartTemplates
		if chartFiles == nil {
			chartFiles = rawChart.Files
		}
		rawChart.Files = rawChartFiles
	}

	defaultIsomateName := releaseRequest.IsomateConfig.DefaultIsomateName
	if defaultIsomateName == "" {
		defaultIsomateName = releaseRequest.IsomateConfig.Isomates[0].Name
	}
	rawChart.Templates, err = helmImpl.mergeIsomateResources(namespace, manifests, defaultIsomateName)
	if err != nil {
		klog.Errorf("failed to merge isomate resources : %s", err.Error())
		return err
	}
	rawChart.Files = chartFiles
	return nil
}

func (helmImpl *Helm) PauseOrRecoverRelease(paused bool, oldReleaseInfo *release.ReleaseInfoV2) (*release.ReleaseCache, error) {
	getAction, err := helmImpl.getGetAction(oldReleaseInfo.Namespace)
	if err != nil {
		klog.Errorf("failed to get GetReleaseAction : %s", err.Error())
		return nil, err
	}

	helmRel, err := getAction.Run(oldReleaseInfo.Name)
	if err != nil {
		klog.Errorf("failed to get release %s/%s from helm : %s", oldReleaseInfo.Namespace, oldReleaseInfo.Name, err.Error())
		return nil, err
	}

	rawChart := helmRel.Chart

	// merge pause release plugin
	releasePlugins, err := mergePausePlugin(paused, oldReleaseInfo.Plugins)
	if err != nil {
		klog.Errorf("failed to merge pause plugin : %s", err.Error())
		return nil, err
	}

	// add default plugin
	releasePlugins = addDefaultPlugins(releasePlugins)

	valueOverride := helmRel.Config
	valueOverride[plugins.WalmPluginConfigKey] = releasePlugins

	releaseCache, err := helmImpl.doInstallUpgradeReleaseFromChart(oldReleaseInfo.Namespace, oldReleaseInfo.Name, rawChart, valueOverride, true, false, releasePlugins)
	if err != nil {
		klog.Errorf("failed to update release from chart : %s", err.Error())
		return nil, err
	}
	return releaseCache, nil
}

func (helmImpl *Helm) doInstallUpgradeReleaseFromChart(namespace, name string, rawChart *chart.Chart, valueOverride map[string]interface{},
	update bool, dryRun bool, releasePlugins []*k8sModel.ReleasePlugin) (releaseCache *release.ReleaseCache, err error) {

	releaseChan := make(chan *helmRelease.Release, 1)
	releaseErrChan := make(chan error, 1)

	expChan := make(chan struct{})
	_, kubeClient := helmImpl.kubeClients.GetKubeClient(namespace)

	// execute pre_install plugins
	go func() {
		select {
		case release := <-releaseChan:
			defer func() {
				if err := recover(); err != nil {
					releaseErrChan <- errors.New(fmt.Sprintf("panic happend: %v", err))
				}
			}()
			context, err := buildContext(kubeClient, release)
			if err != nil {
				releaseErrChan <- err
				return
	}

			err = runPlugins(releasePlugins, context, plugins.Pre_Install)
			if err != nil {
				releaseErrChan <- err
				return
			}

			manifest, err := buildManifest(context.Resources)
	if err != nil {
				klog.Errorf("failed to build manifest : %s", err.Error())
				releaseErrChan <- err
				return
			}
			release.Manifest = manifest
			releaseChan <- release
		case <-expChan:
			klog.Warning("failed to execute pre_install plugins with exception")
	}

	}()
	defer close(expChan)

	var helmRelease *helmRelease.Release
	if update {
		action, err := helmImpl.getUpgradeAction(namespace)
		if err != nil {
			return nil, err
		}
		action.DryRun = dryRun
		action.Namespace = namespace
		action.MaxHistory = releaseMaxHistory
		action.ReleaseChan = releaseChan
		action.ReleaseErrChan = releaseErrChan
		helmRelease, err = action.Run(name, rawChart, valueOverride)
		if err != nil {
			klog.Errorf("failed to upgrade release %s/%s from chart : %s", namespace, name, err.Error())
			return nil, err
		}
	} else {
		action, err := helmImpl.getInstallAction(namespace)
		if err != nil {
			return nil, err
		}
		action.DryRun = dryRun
		action.Namespace = namespace
		action.ReleaseName = name
		action.ReleaseChan = releaseChan
		action.ReleaseErrChan = releaseErrChan
		helmRelease, err = action.Run(rawChart, valueOverride)
		if err != nil {
			klog.Errorf("failed to install release %s/%s from chart : %s", namespace, name, err.Error())
			if !dryRun {
				action1, err1 := helmImpl.getDeleteAction(namespace)
				if err1 != nil {
					klog.Errorf("failed to get helm delete action : %s", err.Error())
				} else {
					_, err1 = action1.Run(name)
					if err1 != nil {
						klog.Errorf("failed to rollback to delete release %s/%s : %s", namespace, name, err1.Error())
					}
				}
				}
			return nil, err
				}
			}

	// execute post_install plugins
	context, err := buildContext(kubeClient, helmRelease)
	if err != nil {
			return nil, err
		}

	err = runPlugins(releasePlugins, context, plugins.Post_Install)
	if err != nil {
		return nil, err
	}
	return helmImpl.convertHelmRelease(helmRelease)
}

func buildContext(kubeClient *kube.Client, release *helmRelease.Release) (*plugins.PluginContext, error) {
	resources, err := kubeClient.Build(bytes.NewBufferString(release.Manifest))
	if err != nil {
		klog.Errorf("failed to build k8s resources : %s", err.Error())
		return nil, err
	}
	context := &plugins.PluginContext{
		R:         release,
		Resources: []runtime.Object{},
	}
	for _, resource := range resources {
		context.Resources = append(context.Resources, resource.Object)
	}
	return context, nil
}

func runPlugins(releasePlugins []*k8sModel.ReleasePlugin, context *plugins.PluginContext, runnerType plugins.RunnerType) error {
	sort.Sort(utils.SortablePlugins(releasePlugins))
	for _, plugin := range releasePlugins {
		if plugin.Disable {
			continue
		}
		runner := plugins.GetRunner(plugin)
		if runner != nil && runner.Type == runnerType {
			klog.Infof("start to exec %s plugin %s", runnerType, plugin.Name)
			err := runner.Run(context, plugin.Args)
			if err != nil {
				klog.Errorf("failed to exec %s plugin %s : %s", runnerType, plugin.Name, err.Error())
				return err
			}
			klog.Infof("succeed to exec %s plugin %s", runnerType, plugin.Name)
		}
	}
	return nil
}

func buildManifest(resources []runtime.Object) (string, error) {
	var sb strings.Builder
	for _, resource := range resources {
		resourceBytes, err := yaml.Marshal(resource)
		if err != nil {
			return "", err
		}
		sb.WriteString("\n---\n")
		sb.Write(resourceBytes)
	}
	return sb.String(), nil
}

func (helmImpl *Helm) convertHelmRelease(helmRelease *helmRelease.Release) (releaseCache *release.ReleaseCache, err error) {
	releaseSpec := release.ReleaseSpec{}
	releaseSpec.Name = helmRelease.Name
	releaseSpec.Namespace = helmRelease.Namespace
	releaseSpec.Dependencies = make(map[string]string)
	releaseSpec.Version = int32(helmRelease.Version)
	releaseSpec.ChartVersion = helmRelease.Chart.Metadata.Version
	releaseSpec.ChartName = helmRelease.Chart.Metadata.Name
	releaseSpec.ChartAppVersion = helmRelease.Chart.Metadata.AppVersion
	releaseSpec.ConfigValues = map[string]interface{}{}
	util.MergeValues(releaseSpec.ConfigValues, helmRelease.Config, false)
	releaseCache = &release.ReleaseCache{
		ReleaseSpec: releaseSpec,
	}

	releaseCache.ComputedValues, err = chartutil.CoalesceValues(helmRelease.Chart, helmRelease.Config)
	if err != nil {
		klog.Errorf("failed to get computed values : %s", err.Error())
		return nil, err
	}

	releaseCache.MetaInfoValues, releaseCache.PrettyParams, _ = buildMetaInfoValues(helmRelease.Chart, releaseCache.ComputedValues)
	releaseCache.ReleaseResourceMetas, err = helmImpl.getReleaseResourceMetas(helmRelease)
	if err != nil {
		return nil, err
	}
	releaseCache.Manifest = helmRelease.Manifest
	releaseCache.HelmVersion = helmRelease.HelmVersion
	return
}

func (helmImpl *Helm) getReleaseResourceMetas(helmRelease *helmRelease.Release) (resources []release.ReleaseResourceMeta, err error) {
	resources = []release.ReleaseResourceMeta{}
	if helmRelease.Name == "app-manager" {
		return nil, nil
	}
	_, kubeClient := helmImpl.kubeClients.GetKubeClient(helmRelease.Namespace)
	results, err := kubeClient.Build(bytes.NewBufferString(helmRelease.Manifest))
	if err != nil {
		klog.Errorf("failed to get release resource metas of %s", helmRelease.Name)
		return resources, err
	}
	for _, result := range results {
		resource := release.ReleaseResourceMeta{
			Kind:      k8sModel.ResourceKind(result.Object.GetObjectKind().GroupVersionKind().Kind),
			Namespace: result.Namespace,
			Name:      result.Name,
		}
		resources = append(resources, resource)
	}
	return
}

func buildMetaInfoValues(chart *chart.Chart, computedValues map[string]interface{}) (*release.MetaInfoParams, *release.PrettyChartParams, error) {
	chartMetaInfo, err := getChartMetaInfo(chart)
	if err != nil {
		return nil, nil, err
	}
	if chartMetaInfo != nil {
		metaInfoParams, err := chartMetaInfo.BuildMetaInfoParams(computedValues)
		if err != nil {
			return nil, nil, err
		}
		prettyParams := convertMetaInfoParamsToPrettyParams(chartMetaInfo, metaInfoParams)
		return metaInfoParams, prettyParams, nil
	}

	return nil, nil, nil
}

func (helmImpl *Helm) getGetAction(namespace string) (*action.Get, error) {
	config, err := helmImpl.getActionConfig(namespace)
	if err != nil {
			return nil, err
		}
	return action.NewGet(config), nil
}

func (helmImpl *Helm) getInstallAction(namespace string) (*action.Install, error) {
	config, err := helmImpl.getActionConfig(namespace)
	if err != nil {
		return nil, err
	}
	return action.NewInstall(config), nil
}

func (helmImpl *Helm) getUpgradeAction(namespace string) (*action.Upgrade, error) {
	config, err := helmImpl.getActionConfig(namespace)
	if err != nil {
		return nil, err
	}
	return action.NewUpgrade(config), nil
}

func (helmImpl *Helm) getDeleteAction(namespace string) (*action.Uninstall, error) {
	config, err := helmImpl.getActionConfig(namespace)
	if err != nil {
		return nil, err
	}
	return action.NewUninstall(config), nil
}

func (helm *Helm) mergeIsomateResources(namespace string, manifests map[string]string, defaultIsomateName string) ([]*chart.File, error) {
	_, kubeClient := helm.kubeClients.GetKubeClient(namespace)
	resourceMap := map[string]*resource.Info{}
	for isomateName, manifest := range manifests {
		resources, err := kubeClient.Build(bytes.NewBufferString(manifest))
		if err != nil {
			klog.Errorf("failed to build resources : %s", err.Error())
			return nil, err
		}
		isDefaultIsomate := false
		if isomateName == defaultIsomateName {
			isDefaultIsomate = true
		}

		for _, resource := range resources {
			resourceKey := buildResourceKey(resource)
			if existedResource, ok := resourceMap[resourceKey]; ok {
				if resource.Object.GetObjectKind().GroupVersionKind().Kind == string(k8sModel.IsomateSetKind) {
					mergedIsomateSet, err := plugins.MergeIsomateSets(existedResource.Object, resource.Object)
					if err != nil {
						klog.Errorf("failed to merge isomate sets : %s", err.Error())
						return nil, err
					}
					existedResource.Object = mergedIsomateSet
				} else if isDefaultIsomate {
					resourceMap[resourceKey] = resource
				}
	} else {
				resourceMap[resourceKey] = resource
			}
		}
	}

	templates := []*chart.File{}
	for _, resource := range resourceMap {
		resourceBytes, err := yaml.Marshal(resource.Object)
		if err != nil {
			klog.Errorf("failed to marshal resource : %s", err.Error())
			return nil, err
		}
		templates = append(templates, &chart.File{
			Name: transwarpjsonnet.BuildNotRenderedFileName(buildResourceKey(resource)),
			Data: resourceBytes,
		})
	}

	return templates, nil
}

func buildResourceKey(resource *resource.Info) string {
	return resource.Object.GetObjectKind().GroupVersionKind().Kind + "-" + resource.Namespace + "-" + resource.Name
}

func reuseReleaseRequest(releaseInfo *release.ReleaseInfoV2, releaseRequest *release.ReleaseRequestV2, fullUpdate bool) (
	configValues map[string]interface{}, dependencies map[string]string, releaseLabels map[string]string,
	walmPlugins []*k8sModel.ReleasePlugin, err error,
) {
	configValues = map[string]interface{}{}

	// compatible
	if _, ok := releaseRequest.ConfigValues[transwarpjsonnet.TranswarpInstallIDKey]; ok {
		if transwarpInstallID, ok2 := releaseInfo.ConfigValues[transwarpjsonnet.TranswarpInstallIDKey]; ok2 {
			if transwarpInstallID != "" {
				klog.Warningf("update request config value has key %s: %s. ignore it", transwarpjsonnet.TranswarpInstallIDKey, transwarpInstallID)
				releaseRequest.ConfigValues[transwarpjsonnet.TranswarpInstallIDKey] = transwarpInstallID
			}
		}
	}

	if fullUpdate {
		util.MergeValues(configValues, releaseRequest.ConfigValues, true)
	} else {
	util.MergeValues(configValues, releaseInfo.ConfigValues, false)
	util.MergeValues(configValues, releaseRequest.ConfigValues, false)
	}

	dependencies = map[string]string{}
	for key, value := range releaseInfo.Dependencies {
		dependencies[key] = value
	}
	for key, value := range releaseRequest.Dependencies {
		if value == "" {
			if _, ok := dependencies[key]; ok {
				delete(dependencies, key)
			}
		} else {
			dependencies[key] = value
		}
	}

	releaseLabels = map[string]string{}
	for key, value := range releaseInfo.ReleaseLabels {
		releaseLabels[key] = value
	}
	for key, value := range releaseRequest.ReleaseLabels {
		if value == "" {
			if _, ok := releaseLabels[key]; ok {
				delete(releaseLabels, key)
			}
		} else {
			releaseLabels[key] = value
		}
	}

	walmPlugins, err = mergeReleasePlugins(releaseRequest.Plugins, releaseInfo.Plugins)
	if err != nil {
		return
	}
	return
}

func mergeIsomateConfig(isomateConfig, oldIsomateConfig *k8sModel.IsomateConfig) (err error) {
	if isomateConfig == nil || oldIsomateConfig == nil || len(oldIsomateConfig.Isomates) == 0 {
		return nil
	}
	oldIsomatesMap := map[string]*k8sModel.Isomate{}
	for _, isomate := range oldIsomateConfig.Isomates {
		oldIsomatesMap[isomate.Name] = isomate
	}

	for _, isomate := range isomateConfig.Isomates {
		if oldIsomate, ok := oldIsomatesMap[isomate.Name]; ok {
			isomate.ConfigValues = util.MergeValues(oldIsomate.ConfigValues, isomate.ConfigValues, false)
			isomate.Plugins, err = mergeReleasePlugins(isomate.Plugins, oldIsomate.Plugins)
			if err != nil {
				klog.Errorf("failed to merge release plugins : %s", err.Error())
			return
		}
		}
	}
	return
}

func mergeReleasePlugins(plugins, defaultPlugins []*k8sModel.ReleasePlugin) (mergedPlugins []*k8sModel.ReleasePlugin, err error) {
	releasePluginsMap := map[string]*k8sModel.ReleasePlugin{}
	for _, plugin := range plugins {
		if _, ok := releasePluginsMap[plugin.Name]; ok {
			return nil, buildDuplicatedPluginError(plugin.Name)
		} else {
			releasePluginsMap[plugin.Name] = plugin
		}
	}
	for _, plugin := range defaultPlugins {
		if _, ok := releasePluginsMap[plugin.Name]; !ok {
			releasePluginsMap[plugin.Name] = plugin
		}
	}
	for _, plugin := range releasePluginsMap {
		mergedPlugins = append(mergedPlugins, plugin)
	}
	return
}

func buildDuplicatedPluginError(pluginName string) error {
	return fmt.Errorf("more than one plugin %s is not allowed", pluginName)
}

func convertBufferFiles(chartFiles []*common.BufferedFile) []*loader.BufferedFile {
	result := []*loader.BufferedFile{}
	for _, file := range chartFiles {
		result = append(result, &loader.BufferedFile{
			Name: file.Name,
			Data: file.Data,
		})
	}
	return result
}

func NewHelm(repoList []*setting.ChartRepo, registryClient *registry.Client, k8sCache k8s.Cache, kubeClients *k8sHelm.Client) (*Helm, error) {
	chartRepoMap := make(map[string]*ChartRepository)

	for _, chartRepo := range repoList {
		chartRepository := ChartRepository{
			Name:     chartRepo.Name,
			URL:      chartRepo.URL,
			Username: "",
			Password: "",
		}
		chartRepoMap[chartRepo.Name] = &chartRepository
	}

	actionConfigs, _ := lru.New(100)
	restyClient := resty.New()
	restyClient.SetTimeout(defaultDownloadTimeout)
	helm := &Helm{
		k8sCache:       k8sCache,
		kubeClients:    kubeClients,
		registryClient: registryClient,
		chartRepoMap:   chartRepoMap,
		actionConfigs:  actionConfigs,
		restyClient:    restyClient,
	}

	actionConfig, err := helm.getActionConfig("")
	if err != nil {
		return nil, err
	}
	list := action.NewList(actionConfig)
	list.AllNamespaces = true
	list.All = true
	list.StateMask = action.ListDeployed | action.ListFailed | action.ListPendingInstall | action.ListPendingRollback |
		action.ListPendingUpgrade | action.ListUninstalled | action.ListUninstalling | action.ListUnknown

	helm.list = list

	return helm, nil

}

func NewRegistryClient(chartImageConfig *setting.ChartImageConfig) (*registry.Client, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	option := &registry.ClientOptions{
		Out: os.Stdout,
			Resolver: docker.NewResolver(docker.ResolverOptions{
				Client: client,
			}),
	}

	return registry.NewClient(option)
}
