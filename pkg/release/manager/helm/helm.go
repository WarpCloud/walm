package helm

import (
	"time"
	"fmt"
	"io/ioutil"

	"walm/pkg/setting"

	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/storage/driver"
	"walm/pkg/release/manager/helm/cache"
	"walm/pkg/redis"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sclient "walm/pkg/k8s/client"
	"k8s.io/helm/pkg/chart"
	"k8s.io/helm/pkg/chart/loader"
)

const (
	helmCacheDefaultResyncInterval time.Duration = 5 * time.Minute
	multiTenantClientsMaxSize int = 128
)

type ChartRepository struct {
	Name     string
	URL      string
	Username string
	Password string
}

type HelmClient struct {
	chartRepoMap            map[string]*ChartRepository
	dryRun                  bool
	helmCache               *cache.HelmCache
	helmCacheResyncInterval time.Duration
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

		helmClient = &HelmClient{
			chartRepoMap:            chartRepoMap,
			dryRun:                  false,
			helmCache:               helmCache,
			helmCacheResyncInterval: helmCacheDefaultResyncInterval,
		}
	}
	return helmClient
}

//func InitHelmByParams(tillerHost string, chartRepoMap map[string]*ChartRepository, dryRun bool) {
//	client := helm.NewClient(helm.Host(tillerHost))
//
//	helmClient = &HelmClient{
//		systemClient: client,
//		chartRepoMap: chartRepoMap,
//		dryRun:       dryRun,
//	}
//}

func (client *HelmClient) GetDryRun() bool {
	return client.dryRun
}

func (client *HelmClient) GetDependencies(repoName, chartName, chartVersion string) (subChartNames []string, err error) {
	logrus.Debugf("Enter GetDependencies %s %s\n", chartName, chartVersion)
	chartRequested, err := client.GetChartRequest(repoName, chartName, chartVersion)
	if err != nil {
		return nil, err
	}
	dependencies, err := parseChartDependencies(chartRequested)
	if err != nil {
		return nil, err
	}
	return dependencies, nil
}

func (client *HelmClient) GetHelmCache() *cache.HelmCache {
	return client.helmCache
}

//TODO to improve
func(client *HelmClient) GetCurrentHelmClient(namespace string) (*helm.Client, error) {
	kc := k8sclient.GetKubeClient(namespace)
	clientset, err := kc.KubernetesClientSet()
	if err != nil {
		return nil, err
	}

	d := driver.NewSecrets(clientset.CoreV1().Secrets(namespace))

	return helm.NewClient(
		helm.KubeClient(kc),
		helm.Driver(d),
		helm.Discovery(clientset.Discovery()),
	), nil
}

func (client *HelmClient) DownloadChart(repoName, charName, version string) (string, error) {
	if repoName == "" {
		repoName = "stable"
	}
	repo, ok := client.chartRepoMap[repoName]
	if !ok {
		return "", fmt.Errorf("can not find repo name: %s", repoName)
	}
	chartURL, httpGetter, err := FindChartInChartMuseumRepoURL(repo.URL, "", "", charName, version)
	if err != nil {
		return "", err
	}

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	filename, err := ChartMuseumDownloadTo(chartURL, tmpDir, httpGetter)
	if err != nil {
		logrus.Printf("DownloadTo err %v", err)
		return "", err
	}

	return filename, nil
}

func (client *HelmClient) GetChartRequest(repoName, chartName, chartVersion string) (*chart.Chart, error) {
	chartPath, err := client.DownloadChart(repoName, chartName, chartVersion)
	if err != nil {
		logrus.Errorf("failed to download chart : %s", err.Error())
		return nil, err
	}
	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		logrus.Errorf("failed to load chart : %s", err.Error())
		return nil, err
	}
	return chartRequested, nil
}

func (client *HelmClient) StartResyncReleaseCaches(stopCh <-chan struct{}) {
	logrus.Infof("start to resync release cache every %v", client.helmCacheResyncInterval)
	// first time should be sync
	client.helmCache.Resync()
	firstTime := true
	go wait.Until(func() {
		if firstTime {
			time.Sleep(client.helmCacheResyncInterval)
			firstTime = false
		}
		client.helmCache.Resync()
	}, client.helmCacheResyncInterval, stopCh)
}


