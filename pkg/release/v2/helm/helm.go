package helm

import (
	"k8s.io/helm/pkg/proto/hapi/chart"
	"walm/pkg/release/manager/helm/cache"
	"walm/pkg/redis"
	"walm/pkg/k8s/client"
	"time"
	"k8s.io/helm/pkg/helm"
	"walm/pkg/setting"
	"walm/pkg/release"
	"github.com/sirupsen/logrus"
	"fmt"
	"io/ioutil"
	"k8s.io/helm/pkg/chartutil"
	helmv1 "walm/pkg/release/manager/helm"
	"gopkg.in/yaml.v2"
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
		}
	}
	return helmClient
}

func (hc *HelmClient) InstallRelease(namespace string, releaseRequest *release.ReleaseRequest, isSystem bool) error {
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
	dependencyConfigs, err := getDependencyConfigs(releaseRequest.Dependencies)
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

	if isJsonnetChart {
		chart, err = convertJsonnetChart(namespace, jsonnetChart, releaseRequest.ConfigValues, dependencyConfigs)
		if err != nil {
			logrus.Errorf("failed to convert jsonnet chart %s-%s from %s : %s", releaseRequest.ChartName, releaseRequest.ChartVersion, releaseRequest.RepoName, err.Error())
			return err
		}
	}

	valueOverride := map[string]interface{}{}
	mergeValues(valueOverride, dependencyConfigs)
	mergeValues(valueOverride, releaseRequest.ConfigValues)
	valueOverrideBytes, err := yaml.Marshal(valueOverride)

	logrus.Debugf("convert %s takes %v",releaseRequest.Name, time.Now().Sub(now))
	_, err = hc.systemClient.InstallReleaseFromChart(
		chart,
		namespace,
		helm.ValueOverrides(valueOverrideBytes),
		helm.ReleaseName(releaseRequest.Name),
		helm.InstallDryRun(hc.dryRun),
	)
	if err != nil {
		logrus.Errorf("failed to install release %s from chart : %s", releaseRequest.Name, err.Error())
		opts := []helm.DeleteOption{
			helm.DeletePurge(true),
		}
		_, err1 := hc.systemClient.DeleteRelease(
			releaseRequest.Name, opts...,
		)
		if err1 != nil {
			logrus.Errorf("failed to rollback to delete release %s : %s", releaseRequest.Name, err1.Error())
		}
		return err
	}

	return nil
}

func getDependencyConfigs(dependencies map[string]string) (dependencyConfigs map[string]interface{}, err error) {
	//TODO
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









