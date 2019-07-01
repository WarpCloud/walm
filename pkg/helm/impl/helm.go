package impl

import (
	"WarpCloud/walm/pkg/models/release"
	"WarpCloud/walm/pkg/models/common"
	"k8s.io/helm/pkg/walm"
	"k8s.io/helm/pkg/walm/plugins"
	"github.com/sirupsen/logrus"
	"WarpCloud/walm/pkg/util"
	"WarpCloud/walm/pkg/util/transwarpjsonnet"
	"k8s.io/helm/pkg/chart"
	"k8s.io/helm/pkg/chart/loader"
	"k8s.io/helm/pkg/registry"
	"fmt"
	"strings"
	"WarpCloud/walm/pkg/k8s"
	k8sModel "WarpCloud/walm/pkg/models/k8s"
	"WarpCloud/walm/pkg/k8s/client"
	"k8s.io/helm/pkg/helm"
	"github.com/hashicorp/golang-lru"
	"k8s.io/helm/pkg/storage/driver"
	helmRelease "k8s.io/helm/pkg/hapi/release"
	"k8s.io/helm/pkg/chartutil"
	"bytes"
	errorModel "WarpCloud/walm/pkg/models/error"
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
	helmClients    *lru.Cache
}

func (helmImpl *Helm) DeleteRelease(namespace string, name string) error {
	currentHelmClient, err := helmImpl.getCurrentHelmClient(namespace)
	if err != nil {
		logrus.Errorf("failed to get current helm client : %s", err.Error())
		return err
	}

	opts := []helm.UninstallOption{
		helm.UninstallPurge(true),
	}
	_, err = currentHelmClient.UninstallRelease(
		name, opts...,
	)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logrus.Warnf("release %s is not found from helm", name)
		} else {
			logrus.Errorf("failed to delete release from helm : %s", err.Error())
			return err
		}
	}
	return nil
}

func (helmImpl *Helm) InstallOrCreateRelease(namespace string, releaseRequest *release.ReleaseRequestV2, chartFiles []*common.BufferedFile,
	dryRun bool, update bool, oldReleaseInfo *release.ReleaseInfoV2, paused *bool) (*release.ReleaseCache, error) {
	var rawChart *chart.Chart
	var chartErr error
	// priority: chartFiles > chartImage > chartName
	if chartFiles != nil {
		rawChart, chartErr = loader.LoadFiles(convertBufferFiles(chartFiles))
	} else if releaseRequest.ChartImage != "" {
		rawChart, chartErr = helmImpl.getRawChartByImage(releaseRequest.ChartImage)
	} else {
		rawChart, chartErr = helmImpl.getRawChartFromRepo(releaseRequest.RepoName, releaseRequest.ChartName, releaseRequest.ChartVersion)
	}
	if chartErr != nil {
		logrus.Errorf("failed to get raw chart : %s", chartErr.Error())
		return nil, chartErr
	}

	chartInfo, err := buildChartInfo(rawChart)
	if err != nil {
		logrus.Errorf("failed to build chart info : %s", err.Error())
		return nil, err
	}
	// support meta pretty parameters
	configValues := releaseRequest.ConfigValues
	if releaseRequest.MetaInfoParams != nil {
		metaInfoConfigs, err := releaseRequest.MetaInfoParams.BuildConfigValues(chartInfo.MetaInfo)
		if err != nil {
			logrus.Errorf("failed to get meta info parameters : %s", err.Error())
			return nil, err
		}
		util.MergeValues(configValues, metaInfoConfigs, false)
	}

	dependencies := releaseRequest.Dependencies
	releaseLabels := releaseRequest.ReleaseLabels
	releasePlugins := releaseRequest.Plugins
	if update {
		// reuse config values, dependencies, release labels, walm plugins
		configValues, dependencies, releaseLabels, releasePlugins, err = reuseReleaseRequest(oldReleaseInfo, releaseRequest)
		if err != nil {
			logrus.Errorf("failed to reuse release request : %s", err.Error())
			return nil, err
		}
	}

	if chartInfo.MetaInfo != nil {
		releasePlugins, err = mergeReleasePlugins(releasePlugins, chartInfo.MetaInfo.Plugins)
		if err != nil {
			logrus.Errorf("failed to merge chart default plugins : %s", err.Error())
			return nil, err
		}
	}

	// get all the dependency releases' output configs from ReleaseConfig
	dependencyConfigs, err := helmImpl.getDependencyOutputConfigs(namespace, dependencies, chartInfo.MetaInfo)
	if err != nil {
		logrus.Errorf("failed to get all the dependency releases' output configs : %s", err.Error())
		return nil, err
	}

	err = transwarpjsonnet.ProcessJsonnetChart(releaseRequest.RepoName, rawChart, namespace, releaseRequest.Name, configValues,
		dependencyConfigs, dependencies, releaseLabels, releaseRequest.ChartImage)
	if err != nil {
		logrus.Errorf("failed to ProcessJsonnetChart : %s", err.Error())
		return nil, err
	}

	if paused != nil {
		if *paused {
			releasePlugins, err = mergeReleasePlugins([]*release.ReleasePlugin{
				{
					Name:    plugins.PauseReleasePluginName,
					Version: "1.0",
				},
			}, releasePlugins)
		} else {
			releasePlugins, err = mergeReleasePlugins([]*release.ReleasePlugin{
				{
					Name:    plugins.PauseReleasePluginName,
					Version: "1.0",
					Disable: true,
				},
			}, releasePlugins)
		}
	}
	// add default plugin
	releasePlugins = append(releasePlugins, &release.ReleasePlugin{
		Name: plugins.ValidateReleaseConfigPluginName,
	})

	valueOverride := map[string]interface{}{}
	util.MergeValues(valueOverride, configValues, false)
	util.MergeValues(valueOverride, dependencyConfigs, false)
	valueOverride[walm.WalmPluginConfigKey] = releasePlugins
	releaseCache, err := helmImpl.doInstallUpgradeReleaseFromChart(namespace, releaseRequest, rawChart, valueOverride, update, dryRun)
	if err != nil {
		logrus.Errorf("failed to create or update release from chart : %s", err.Error())
		return nil, err
	}
	return releaseCache, nil
}

func (helmImpl *Helm) doInstallUpgradeReleaseFromChart(namespace string,
	releaseRequest *release.ReleaseRequestV2, rawChart *chart.Chart, valueOverride map[string]interface{},
	update bool, dryRun bool) (releaseCache *release.ReleaseCache, err error) {

	currentHelmClient, err := helmImpl.getCurrentHelmClient(namespace)
	if err != nil {
		logrus.Errorf("failed to get helm client : %s", err.Error())
		return nil, err
	}

	var helmRelease *helmRelease.Release
	if update {
		helmRelease, err = currentHelmClient.UpdateReleaseFromChart(
			releaseRequest.Name,
			rawChart,
			helm.UpdateValueOverrides(valueOverride),
			helm.UpgradeDryRun(dryRun),
		)
		if err != nil {
			logrus.Errorf("failed to upgrade release %s/%s from chart : %s", namespace, releaseRequest.Name, err.Error())
			return nil, err
		}
	} else {
		helmRelease, err = currentHelmClient.InstallReleaseFromChart(
			rawChart,
			namespace,
			helm.ValueOverrides(valueOverride),
			helm.ReleaseName(releaseRequest.Name),
			helm.InstallDryRun(dryRun),
		)
		if err != nil {
			logrus.Errorf("failed to install release %s/%s from chart : %s", namespace, releaseRequest.Name, err.Error())
			if !dryRun {
				opts := []helm.UninstallOption{
					helm.UninstallPurge(true),
				}
				_, err1 := currentHelmClient.UninstallRelease(
					releaseRequest.Name, opts...,
				)
				if err1 != nil {
					logrus.Errorf("failed to rollback to delete release %s/%s : %s", namespace, releaseRequest.Name, err1.Error())
				}
			}
			return nil, err
		}
	}
	return convertHelmRelease(helmRelease)
}

func convertHelmRelease(helmRelease *helmRelease.Release) (releaseCache *release.ReleaseCache, err error) {
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
		logrus.Errorf("failed to get computed values : %s", err.Error())
		return nil, err
	}

	releaseCache.MetaInfoValues, _ = buildMetaInfoValues(helmRelease.Chart, releaseCache.ComputedValues)
	releaseCache.ReleaseResourceMetas, err = getReleaseResourceMetas(helmRelease)
	releaseCache.Manifest = helmRelease.Manifest
	return
}

func getReleaseResourceMetas(helmRelease *helmRelease.Release) (resources []release.ReleaseResourceMeta, err error) {
	resources = []release.ReleaseResourceMeta{}
	results, err := client.GetKubeClient(helmRelease.Namespace).BuildUnstructured(helmRelease.Namespace, bytes.NewBufferString(helmRelease.Manifest))
	if err != nil {
		logrus.Errorf("failed to get release resource metas of %s", helmRelease.Name)
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

func buildMetaInfoValues(chart *chart.Chart, computedValues map[string]interface{}) (*release.MetaInfoParams, error) {
	chartMetaInfo, err := getChartMetaInfo(chart)
	if err != nil {
		return nil, err
	}
	if chartMetaInfo != nil {
		metaInfoParams, err := chartMetaInfo.BuildMetaInfoParams(computedValues)
		if err != nil {
			return nil, err
		}
		return metaInfoParams, nil
	}

	return nil, nil
}

func (helmImpl *Helm) getCurrentHelmClient(namespace string) (*helm.Client, error) {
	if c, ok := helmImpl.helmClients.Get(namespace); ok {
		return c.(*helm.Client), nil
	} else {
		kc := client.GetKubeClient(namespace)
		clientset, err := kc.KubernetesClientSet()
		if err != nil {
			return nil, err
		}

		d := driver.NewConfigMaps(clientset.CoreV1().ConfigMaps(namespace))
		client := helm.NewClient(
			helm.KubeClient(kc),
			helm.Driver(d),
			helm.Discovery(clientset.Discovery()),
		)
		client.GetTiller().Log = logrus.Infof
		helmImpl.helmClients.Add(namespace, client)
		return client, nil
	}
}

func reuseReleaseRequest(releaseInfo *release.ReleaseInfoV2, releaseRequest *release.ReleaseRequestV2) (
	configValues map[string]interface{}, dependencies map[string]string, releaseLabels map[string]string, walmPlugins []*release.ReleasePlugin, err error) {

	configValues = map[string]interface{}{}
	util.MergeValues(configValues, releaseInfo.ConfigValues, false)
	util.MergeValues(configValues, releaseRequest.ConfigValues, false)

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
	return
}

func (helmImpl *Helm) getDependencyOutputConfigs(namespace string, dependencies map[string]string, chartMetaInfo *release.ChartMetaInfo) (dependencyConfigs map[string]interface{}, err error) {
	dependencyConfigs = map[string]interface{}{}
	if chartMetaInfo == nil {
		return
	}

	chartDependencies := chartMetaInfo.ChartDependenciesInfo
	dependencyAliasConfigVars := map[string]string{}
	for _, chartDependency := range chartDependencies {
		dependencyAliasConfigVars[chartDependency.Name] = chartDependency.AliasConfigVar
	}

	for dependencyKey, dependency := range dependencies {
		dependencyAliasConfigVar, ok := dependencyAliasConfigVars[dependencyKey]
		if !ok {
			continue
		}

		ss := strings.Split(dependency, "/")
		if len(ss) > 2 {
			err = fmt.Errorf("dependency value %s should not contains more than 1 \"/\"", dependency)
			return
		}
		dependencyNamespace, dependencyName := "", ""
		if len(ss) == 2 {
			dependencyNamespace = ss[0]
			dependencyName = ss[1]
		} else {
			dependencyNamespace = namespace
			dependencyName = ss[0]
		}
		dependencyReleaseConfigResource, err := helmImpl.k8sCache.GetResource(k8sModel.ReleaseConfigKind, dependencyNamespace, dependencyName)
		if err != nil {
			if errorModel.IsNotFoundError(err) {
				logrus.Warnf("release config %s/%s is not found", dependencyNamespace, dependencyName)
				continue
			}
			logrus.Errorf("failed to get release config %s/%s : %s", dependencyNamespace, dependencyName, err.Error())
			return nil, err
		}

		dependencyReleaseConfig := dependencyReleaseConfigResource.(k8sModel.ReleaseConfig)
		if len(dependencyReleaseConfig.OutputConfig) > 0 {
			dependencyConfigs[dependencyAliasConfigVar] = dependencyReleaseConfig.OutputConfig
		}
	}
	return
}

func mergeReleasePlugins(plugins, defaultPlugins []*release.ReleasePlugin) (mergedPlugins []*release.ReleasePlugin, err error) {
	releasePluginsMap := map[string]*release.ReleasePlugin{}
	for _, plugin := range plugins {
		if _, ok := releasePluginsMap[plugin.Name]; ok {
			return nil, fmt.Errorf("more than one plugin %s is not allowed", plugin.Name)
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
