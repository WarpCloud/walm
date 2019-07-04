package helm

import (
	"fmt"
	"io/ioutil"
	"time"
	"github.com/hashicorp/golang-lru"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/helm/pkg/chart"
	"k8s.io/helm/pkg/chart/loader"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/storage/driver"
	"WarpCloud/walm/pkg/k8s/adaptor"
	"WarpCloud/walm/pkg/k8s/client"
	"WarpCloud/walm/pkg/k8s/handler"
	"WarpCloud/walm/pkg/redis"
	"WarpCloud/walm/pkg/release/manager/helm/cache"
	"WarpCloud/walm/pkg/setting"
	walmerr "WarpCloud/walm/pkg/util/error"

	"WarpCloud/walm/pkg/release/manager/metainfo"
)

const (
	defaultTimeoutSec              int64         = 60 * 5
	defaultSleepTimeSecond         time.Duration = 1 * time.Second
	helmCacheDefaultResyncInterval time.Duration = 5 * time.Minute
)

type ChartRepository struct {
	Name     string
	URL      string
	Username string
	Password string
}

type HelmClient struct {
	chartRepoMap            map[string]*ChartRepository
	helmCache               *cache.HelmCache
	helmCacheResyncInterval time.Duration
	releaseConfigHandler    *handler.ReleaseConfigHandler
	helmClients             *lru.Cache
}

var helmClient *HelmClient

func GetDefaultHelmClient() *HelmClient {
	if helmClient == nil {
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

		helmCache := cache.NewHelmCache(redis.GetDefaultRedisClient())

		helmClients, _ := lru.New(100)

		helmClient = &HelmClient{
			chartRepoMap:            chartRepoMap,
			helmCache:               helmCache,
			helmCacheResyncInterval: helmCacheDefaultResyncInterval,
			releaseConfigHandler:    handler.GetDefaultHandlerSet().GetReleaseConfigHandler(),
			helmClients:             helmClients,
		}
	}
	return helmClient
}

func (hc *HelmClient) GetHelmCache() *cache.HelmCache {
	return hc.helmCache
}

func (hc *HelmClient) GetAutoDependencies(repoName, chartName, chartVersion string) (subChartNames []string, err error) {
	logrus.Debugf("Enter GetAutoDependencies %s %s\n", chartName, chartVersion)

	subChartNames = []string{}
	detailChartInfo, err := GetDetailChartInfo(repoName, chartName, chartVersion)
	if err != nil {
		return nil, err
	}
	if detailChartInfo.MetaInfo != nil && detailChartInfo.MetaInfo.ChartDependenciesInfo != nil {
		for _, dependency := range detailChartInfo.MetaInfo.ChartDependenciesInfo {
			if dependency.AutoDependency() {
				subChartNames = append(subChartNames, dependency.Name)
			}
		}
	}

	return subChartNames, nil
}

func (hc *HelmClient) getCurrentHelmClient(namespace string) (*helm.Client, error) {
	if c, ok := hc.helmClients.Get(namespace); ok {
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
		hc.helmClients.Add(namespace, client)
		return client, nil
	}
}

func (hc *HelmClient) downloadChart(repoName, chartName, version string) (string, error) {
	if repoName == "" {
		repoName = "stable"
	}
	repo, ok := hc.chartRepoMap[repoName]
	if !ok {
		return "", fmt.Errorf("can not find repo name: %s", repoName)
	}
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	filename, err := LoadChartFromRepo(repo.URL, repo.Username, repo.Password, chartName, version, tmpDir)
	if err != nil {
		logrus.Printf("DownloadTo err %v", err)
		return "", err
	}

	return filename, nil
}

func (hc *HelmClient) LoadChart(repoName, chartName, chartVersion string) (rawChart *chart.Chart, err error) {
	chartPath, err := hc.downloadChart(repoName, chartName, chartVersion)
	if err != nil {
		logrus.Errorf("failed to download chart : %s", err.Error())
		return nil, err
	}

	chartLoader, err := loader.Loader(chartPath)
	if err != nil {
		logrus.Errorf("failed to init chartLoader : %s", err.Error())
		return nil, err
	}

	return chartLoader.Load()
}

func (hc *HelmClient) StartResyncReleaseCaches(stopCh <-chan struct{}) {
	logrus.Infof("start to resync release cache every %v", hc.helmCacheResyncInterval)
	// first time should be sync
	hc.helmCache.Resync()
	firstTime := true
	go wait.Until(func() {
		if firstTime {
			time.Sleep(hc.helmCacheResyncInterval)
			firstTime = false
		}
		hc.helmCache.Resync()
	}, hc.helmCacheResyncInterval, stopCh)
}

// reload dependencies config values, if changes, upgrade release
func (hc *HelmClient) ReloadRelease(namespace, name string, isSystem bool) error {
	releaseInfo, err := hc.GetRelease(namespace, name)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			logrus.Warnf("release %s/%s is not found， ignore to reload release", namespace, name)
			return nil
		}
		logrus.Errorf("failed to get release %s/%s : %s", namespace, name, err.Error())
		return err
	}

	chartInfo, err := GetDetailChartInfo(releaseInfo.RepoName, releaseInfo.ChartName, releaseInfo.ChartVersion)
	if err != nil {
		logrus.Errorf("failed to get chart info : %s", err.Error())
		return err
	}

	oldDependenciesConfigValues := releaseInfo.DependenciesConfigValues
	newDependenciesConfigValues, err := hc.getDependencyOutputConfigs(namespace, releaseInfo.Dependencies, chartInfo.MetaInfo)
	if err != nil {
		logrus.Errorf("failed to get dependencies output configs of %s/%s : %s", namespace, name, err.Error())
		return err
	}

	if ConfigValuesDiff(oldDependenciesConfigValues, newDependenciesConfigValues) {
		releaseRequest := releaseInfo.BuildReleaseRequestV2()
		err = hc.InstallUpgradeRelease(namespace, releaseRequest, isSystem, nil, false, 0)
		if err != nil {
			logrus.Errorf("failed to upgrade release v2 %s/%s : %s", namespace, name, err.Error())
			return err
		}
		logrus.Infof("succeed to reload release %s/%s", namespace, name)
	} else {
		logrus.Infof("ignore reloading release %s/%s : dependencies config value does not change", namespace, name)
	}

	return nil
}

func (hc *HelmClient) validateReleaseTask(namespace, name string, allowReleaseTaskNotExist bool) (releaseTask *cache.ReleaseTask, err error) {
	releaseTask, err = hc.helmCache.GetReleaseTask(namespace, name)
	if err != nil {
		if !walmerr.IsNotFoundError(err) {
			logrus.Errorf("failed to get release task : %s", err.Error())
			return
		} else if !allowReleaseTaskNotExist {
			return
		} else {
			err = nil
		}
	} else {
		if releaseTask.LatestReleaseTaskSig != nil && !releaseTask.LatestReleaseTaskSig.IsTaskFinishedOrTimeout() {
			err = fmt.Errorf("please wait for the release latest task %s-%s finished or timeout", releaseTask.LatestReleaseTaskSig.Name, releaseTask.LatestReleaseTaskSig.UUID)
			logrus.Warn(err.Error())
			return
		}
	}
	return
}

func (hc *HelmClient) getDependencyOutputConfigs(namespace string, dependencies map[string]string, chartMetaInfo *metainfo.ChartMetaInfo) (dependencyConfigs map[string]interface{}, err error) {
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

		dependencyNamespace, dependencyName, err := ParseDependedRelease(namespace, dependency)
		if err != nil {
			return nil, err
		}

		dependencyReleaseConfig, err := hc.releaseConfigHandler.GetReleaseConfig(dependencyNamespace, dependencyName)
		if err != nil {
			if adaptor.IsNotFoundErr(err) {
				logrus.Warnf("release config %s/%s is not found", dependencyNamespace, dependencyName)
				continue
			}
			logrus.Errorf("failed to get release config %s/%s : %s", dependencyNamespace, dependencyName, err.Error())
			return nil, err
		}

		if len(dependencyReleaseConfig.Spec.OutputConfig) > 0 {
			dependencyConfigs[dependencyAliasConfigVar] = dependencyReleaseConfig.Spec.OutputConfig
		}
	}
	return
}
