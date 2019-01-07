package helm

import (
	"k8s.io/helm/pkg/proto/hapi/chart"
	hapirelease "k8s.io/helm/pkg/proto/hapi/release"
	"walm/pkg/release/manager/helm/cache"
	"walm/pkg/redis"
	"walm/pkg/k8s/client"
	"time"
	"k8s.io/helm/pkg/helm"
	"walm/pkg/setting"
	"github.com/sirupsen/logrus"
	"fmt"
	"io/ioutil"
	"k8s.io/helm/pkg/chartutil"
	helmv1 "walm/pkg/release/manager/helm"
	"gopkg.in/yaml.v2"
	"strings"
	"walm/pkg/k8s/handler"
	"walm/pkg/k8s/adaptor"
	releasev2 "walm/pkg/release/v2"
	walmerr "walm/pkg/util/error"
)

const (
	helmCacheDefaultResyncInterval time.Duration = 5 * time.Minute
	multiTenantClientsMaxSize      int           = 128
)

type ChartRepository struct {
	Name     string
	URL      string
	Username string
	Password string
}

type HelmClient struct {
	systemClient            *helm.Client
	multiTenantClients      *cache.MultiTenantClientsCache
	chartRepoMap            map[string]*ChartRepository
	dryRun                  bool
	helmCache               *cache.HelmCache
	helmCacheResyncInterval time.Duration
	releaseConfigHandler    *handler.ReleaseConfigHandler
}

var helmClient *HelmClient

func GetDefaultHelmClientV2() *HelmClient {
	if helmClient == nil {
		tillerHost := setting.Config.SysHelm.TillerHost
		client1 := helm.NewClient(helm.Host(tillerHost))
		chartRepoMap := make(map[string]*ChartRepository)

		for _, chartRepo := range setting.Config.RepoList {
			chartRepository := ChartRepository{
				Name:     chartRepo.Name,
				URL:      chartRepo.URL,
				Username: "",
				Password: "",
			}
			chartRepoMap[chartRepo.Name] = &chartRepository
		}

		multiTenantClients := cache.NewMultiTenantClientsCache(multiTenantClientsMaxSize)
		helmCache := cache.NewHelmCache(redis.GetDefaultRedisClient(), client1, multiTenantClients, client.GetKubeClient())

		helmClient = &HelmClient{
			systemClient:            client1,
			multiTenantClients:      multiTenantClients,
			chartRepoMap:            chartRepoMap,
			dryRun:                  false,
			helmCache:               helmCache,
			helmCacheResyncInterval: helmCacheDefaultResyncInterval,
			releaseConfigHandler:    handler.GetDefaultHandlerSet().GetReleaseConfigHandler(),
		}
	}
	return helmClient
}

// reload dependencies config values, if changes, upgrade release
func (hc *HelmClient) GetRelease(namespace, name string) (releaseV2 *releasev2.ReleaseInfoV2, err error) {
	releaseCache, err := hc.helmCache.GetReleaseCache(namespace, name)
	if err != nil {
		logrus.Errorf("failed to get release cache of %s/%s : %s", namespace, name, err.Error())
		return nil, err
	}
	releaseV1, err := helmv1.BuildReleaseInfo(releaseCache)
	if err != nil {
		logrus.Errorf("failed to build release info: %s", err.Error())
		return nil, err
	}
	releaseV2 = &releasev2.ReleaseInfoV2{ReleaseInfo: *releaseV1}
	releaseConfig, err := hc.releaseConfigHandler.GetReleaseConfig(namespace, name)
	if err != nil {
		if adaptor.IsNotFoundErr(err) {
			//TODO
		}
		logrus.Errorf("failed to get release config : %s", err.Error())
		return nil, err
	}
	releaseV2.ConfigValues = releaseConfig.Spec.ConfigValues
	releaseV2.Dependencies = releaseConfig.Spec.Dependencies
	releaseV2.DependenciesConfigValues = releaseConfig.Spec.DependenciesConfigValues
	releaseV2.OutputConfigValues = releaseConfig.Spec.OutputConfig
    releaseV2.ComputedValues = releaseCache.ComputedValues

	return
}

// reload dependencies config values, if changes, upgrade release
func (hc *HelmClient) ReloadRelease(namespace, name string, isSystem bool) error {
	_, err := hc.helmCache.GetReleaseCache(namespace, name)
	if err != nil {
		logrus.Errorf("failed to get release cache of %s/%s : %s", namespace, name, err.Error())
		return err
	}

	releaseConfig, err := handler.GetDefaultHandlerSet().GetReleaseConfigHandler().GetReleaseConfig(namespace, name)
	if err != nil {
		logrus.Errorf("failed to get release config of %s/%s : %s", namespace, name, err.Error())
		return err
	}

	oldDependenciesConfigValues := releaseConfig.Spec.DependenciesConfigValues
	newDependenciesConfigValues, err := hc.getDependencyOutputConfigs(namespace, releaseConfig.Spec.Dependencies)
	if err != nil {
		logrus.Errorf("failed to get dependencies output configs of %s/%s : %s", namespace, name, err.Error())
		return err
	}

	if ConfigValuesDiff(oldDependenciesConfigValues, newDependenciesConfigValues) {
		//TODO add spec RepoName
		chart, err := hc.loadChartFromRepo("", releaseConfig.Spec.ChartName, releaseConfig.Spec.ChartVersion)
		if err != nil {
			logrus.Errorf("failed to load chart %s-%s from %s : %s", releaseConfig.Spec.ChartName, releaseConfig.Spec.ChartVersion, "", err.Error())
			return err
		}

		isJsonnetChart, jsonnetChart, _, err := isJsonnetChart(chart)
		if err != nil {
			logrus.Errorf("failed to check whether is jsonnet chart : %s", err.Error())
			return err
		}

		reuseValues := true
		if isJsonnetChart {
			reuseValues = false
			chart, err = convertJsonnetChart(namespace, name, releaseConfig.Spec.Dependencies, jsonnetChart, releaseConfig.Spec.ConfigValues, newDependenciesConfigValues)
			if err != nil {
				logrus.Errorf("failed to convert jsonnet chart %s-%s from %s : %s", releaseConfig.Spec.ChartName, releaseConfig.Spec.ChartVersion, "", err.Error())
				return err
			}
		}

		valueOverride := map[string]interface{}{}
		mergeValues(valueOverride, releaseConfig.Spec.ConfigValues)
		valueOverrideBytes, err := yaml.Marshal(valueOverride)

		currentHelmClient, err := hc.getCurrentHelmClient(namespace, isSystem)
		if err != nil {
			logrus.Errorf("failed to get current helm client : %s", err.Error())
			return err
		}

		resp, err := currentHelmClient.UpdateReleaseFromChart(
			name,
			chart,
			helm.UpdateValueOverrides(valueOverrideBytes),
			helm.ReuseValues(reuseValues),
			helm.UpgradeDryRun(hc.dryRun),
		)
		if err != nil {
			logrus.Errorf("failed to upgrade release %s/%s from chart : %s", namespace, name, err.Error())
			return err
		}
		err = hc.helmCache.CreateOrUpdateReleaseCache(resp.GetRelease())
		if err != nil {
			logrus.Errorf("failed to update release cache of %s/%s : %s", namespace, name, err.Error())
			return err
		}

		logrus.Infof("succeed to reload release %s/%s", namespace, name)
	} else {
		logrus.Infof("ignore reloading release %s/%s : dependencies config value does not change", namespace, name)
	}

	return nil
}

func (hc *HelmClient) DeleteRelease(namespace, name string, isSystem bool) error {
	_, err := hc.helmCache.GetReleaseCache(namespace, name)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			logrus.Warnf("release %s/%s is not found", namespace, name)
			return nil
		}
		logrus.Errorf("failed to get release cache of %s/%s : %s", namespace, name, err.Error())
		return err
	}

	currentHelmClient, err := hc.getCurrentHelmClient(namespace, isSystem)
	if err != nil {
		logrus.Errorf("failed to get current helm client : %s", err.Error())
		return err
	}

	opts := []helm.DeleteOption{
		helm.DeletePurge(true),
	}
	res, err := currentHelmClient.DeleteRelease(
		name, opts...,
	)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logrus.Warnf("release %s is not found in tiller", name)
		} else {
			logrus.Errorf("failed to delete release : %s", err.Error())
			return err
		}
	}
	if res != nil && res.Info != "" {
		logrus.Println(res.Info)
	}

	err = hc.helmCache.DeleteReleaseCache(namespace, name)
	if err != nil {
		logrus.Errorf("failed to delete release cache of %s : %s", name, err.Error())
		return err
	}

	logrus.Infof("succeed to delete release %s/%s", namespace, name)
	return nil
}

func (hc *HelmClient) InstallUpgradeRelease(namespace string, releaseRequest *releasev2.ReleaseRequestV2, isSystem bool) error {
	update := true
	_, err := hc.helmCache.GetReleaseCache(namespace, releaseRequest.Name)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			update = false
		} else {
			logrus.Errorf("failed to get release cache of %s/%s : %s", namespace, releaseRequest.Name, err.Error())
			return err
		}
	}

	now := time.Now()
	if releaseRequest.ConfigValues == nil {
		releaseRequest.ConfigValues = map[string]interface{}{}
	}
	if releaseRequest.Dependencies == nil {
		releaseRequest.Dependencies = map[string]string{}
	}

	// if jsonnet chart, add template-jsonnet/, app.yaml to chart.Files
	// app.yaml : used to define chart dependency relations
	chart, err := hc.loadChartFromRepo(releaseRequest.RepoName, releaseRequest.ChartName, releaseRequest.ChartVersion)
	if err != nil {
		logrus.Errorf("failed to load chart %s-%s from %s : %s", releaseRequest.ChartName, releaseRequest.ChartVersion, releaseRequest.RepoName, err.Error())
		return err
	}

	// get all the dependency releases' output configs
	dependencyConfigs, err := hc.getDependencyOutputConfigs(namespace, releaseRequest.Dependencies)
	if err != nil {
		logrus.Errorf("failed to get all the dependency releases' output configs : %s", err.Error())
		return err
	}

	// check whether is jsonnet chart
	isJsonnetChart, jsonnetChart, _, err := isJsonnetChart(chart)
	if err != nil {
		logrus.Errorf("failed to check whether is jsonnet chart : %s", err.Error())
		return err
	}

	reuseValue := true
	if isJsonnetChart {
		reuseValue = false
		configValues := map[string]interface{}{}
		if update {
			releaseConfig, err := hc.releaseConfigHandler.GetReleaseConfig(namespace, releaseRequest.Name)
			if err != nil {
				if adaptor.IsNotFoundErr(err) {
					logrus.Warnf("release config %s/%s is not found", namespace, releaseRequest.Name)
				} else {
					logrus.Errorf("failed to get release config : %s", err.Error())
					return err
				}
			}
			mergeValues(configValues, releaseConfig.Spec.ConfigValues)
		}
		mergeValues(configValues, releaseRequest.ConfigValues)

		chart, err = convertJsonnetChart(namespace, releaseRequest.Name, releaseRequest.Dependencies, jsonnetChart, configValues, dependencyConfigs)
		if err != nil {
			logrus.Errorf("failed to convert jsonnet chart %s-%s from %s : %s", releaseRequest.ChartName, releaseRequest.ChartVersion, releaseRequest.RepoName, err.Error())
			return err
		}
	} else {
		//TODO native helm chart如何处理？
	}


	valueOverride := map[string]interface{}{}
	mergeValues(valueOverride, releaseRequest.ConfigValues)
	valueOverrideBytes, err := yaml.Marshal(valueOverride)
	logrus.Infof("convert %s takes %v", releaseRequest.Name, time.Now().Sub(now))

	currentHelmClient, err := hc.getCurrentHelmClient(namespace, isSystem)
	if err != nil {
		logrus.Errorf("failed to get current helm client : %s", err.Error())
		return err
	}

	var release *hapirelease.Release
	if update {
		//TODO update chart value.yaml does not take effect
		resp, err := currentHelmClient.UpdateReleaseFromChart(
			releaseRequest.Name,
			chart,
			helm.UpdateValueOverrides(valueOverrideBytes),
			helm.ReuseValues(reuseValue),
			helm.UpgradeDryRun(hc.dryRun),
		)
		if err != nil {
			logrus.Errorf("failed to upgrade release %s/%s from chart : %s", namespace, releaseRequest.Name, err.Error())
			return err
		}
		release = resp.GetRelease()
	} else {
		resp, err := currentHelmClient.InstallReleaseFromChart(
			chart,
			namespace,
			helm.ValueOverrides(valueOverrideBytes),
			helm.ReleaseName(releaseRequest.Name),
			helm.InstallDryRun(hc.dryRun),
		)
		if err != nil {
			logrus.Errorf("failed to install release %s/%s from chart : %s", namespace, releaseRequest.Name, err.Error())
			opts := []helm.DeleteOption{
				helm.DeletePurge(true),
			}
			_, err1 := hc.systemClient.DeleteRelease(
				releaseRequest.Name, opts...,
			)
			if err1 != nil {
				logrus.Errorf("failed to rollback to delete release %s/%s : %s", namespace, releaseRequest.Name, err1.Error())
			}
			return err
		}
		release = resp.GetRelease()
	}

	err = hc.helmCache.CreateOrUpdateReleaseCache(release)
	if err != nil {
		logrus.Errorf("failed to create of update release cache of %s/%s : %s", namespace, releaseRequest.Name, err.Error())
		return err
	}

	logrus.Infof("succeed to create or update release %s/%s", namespace, releaseRequest.Name)

	return nil
}

func(hc *HelmClient) getCurrentHelmClient(namespace string, isSystem bool) (*helm.Client, error) {
	currentHelmClient := hc.systemClient
	if !isSystem {
		multiTenant, err := cache.IsMultiTenant(namespace)
		if err != nil {
			logrus.Errorf("failed to check whether is multi tenant", err.Error())
			return nil, err
		}
		if multiTenant {
			tillerHosts := fmt.Sprintf("tiller-tenant.%s.svc:44134", namespace)
			currentHelmClient = hc.multiTenantClients.Get(tillerHosts)
		}
	}

	//TODO improve
	retry := 20
	for i := 0; i < retry; i++ {
		err := currentHelmClient.PingTiller()
		if err == nil {
			break
		}
		if i == retry-1 {
			return nil, fmt.Errorf("tiller is not ready, PingTiller timeout: %s", err.Error())
		}
		time.Sleep(500 * time.Millisecond)
	}
	return currentHelmClient, nil
}

func (hc *HelmClient) getDependencyOutputConfigs(namespace string, dependencies map[string]string) (dependencyConfigs map[string]interface{}, err error) {
	dependencyConfigs = map[string]interface{}{}
	for _, dependency := range dependencies {
		ss := strings.Split(dependency, ".")
		if len(ss) > 2 {
			err = fmt.Errorf("dependency value %s should not contains more than 1 \".\"", dependency)
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
		dependencyReleaseConfig, err := hc.releaseConfigHandler.GetReleaseConfig(dependencyNamespace, dependencyName)
		if err != nil {
			if adaptor.IsNotFoundErr(err) {
				continue
			}
			logrus.Errorf("failed to get release config %s/%s : %s", dependencyNamespace, dependencyName, err.Error())
			return nil, err
		}

		// TODO how to deal with key conflict?
		if len(dependencyReleaseConfig.Spec.OutputConfig) > 0 {
			// compatible
			provideConfigValues, ok := dependencyReleaseConfig.Spec.OutputConfig["provides"].(map[string]interface{})
			if ok {
				valueToMerge := make(map[string]interface{}, len(provideConfigValues))
				for key, value := range provideConfigValues {
					if immediateValue, ok := value.(map[string]interface{}); ok {
						if immediateValue["immediate_value"] != nil {
							valueToMerge[key] = immediateValue["immediate_value"]
						}
					}
				}
				mergeValues(dependencyConfigs, valueToMerge)
			} else {
				mergeValues(dependencyConfigs, dependencyReleaseConfig.Spec.OutputConfig)
			}
		}
	}
	return
}

func (hc *HelmClient) loadChartFromRepo(repoName, chartName, chartVersion string) (*chart.Chart, error) {
	chartPath, err := hc.downloadChart(repoName, chartName, chartVersion)
	if err != nil {
		logrus.Errorf("failed to download chart : %s", err.Error())
		return nil, err
	}
	chartRequested, err := chartutil.Load(chartPath)
	if err != nil {
		logrus.Errorf("failed to load chart : %s", err.Error())
		return nil, err
	}

	return chartRequested, nil
}

func (hc *HelmClient) downloadChart(repoName, charName, version string) (string, error) {
	if repoName == "" {
		repoName = "stable"
	}
	repo, ok := hc.chartRepoMap[repoName]
	if !ok {
		return "", fmt.Errorf("can not find repo name: %s", repoName)
	}
	chartURL, httpGetter, err := helmv1.FindChartInChartMuseumRepoURL(repo.URL, "", "", charName, version)
	if err != nil {
		return "", err
	}

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	filename, err := helmv1.ChartMuseumDownloadTo(chartURL, tmpDir, httpGetter)
	if err != nil {
		logrus.Printf("DownloadTo err %v", err)
		return "", err
	}

	return filename, nil
}
