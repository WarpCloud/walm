package helm

import (
	"fmt"
	"io/ioutil"
	"strings"

	"walm/pkg/setting"
	"walm/pkg/release"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/storage/driver"
	"k8s.io/helm/pkg/transwarp"
	hapiRelease "k8s.io/helm/pkg/proto/hapi/release"
	"sync"
	"errors"
)

type ChartRepository struct {
	Name     string
	URL      string
	Username string
	Password string
}

type HelmClient struct {
	client       *helm.Client
	chartRepoMap map[string]*ChartRepository
	dryRun       bool
}

var helmClient *HelmClient

func GetDefaultHelmClient() *HelmClient {
	return helmClient
}

func InitHelm() {
	tillerHost := setting.Config.SysHelm.TillerHost
	client := helm.NewClient(helm.Host(tillerHost))
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

	helmClient = &HelmClient{
		client:       client,
		chartRepoMap: chartRepoMap,
		dryRun:       false,
	}
}

func InitHelmByParams(tillerHost string, chartRepoMap map[string]*ChartRepository, dryRun bool) {
	client := helm.NewClient(helm.Host(tillerHost))

	helmClient = &HelmClient{
		client:       client,
		chartRepoMap: chartRepoMap,
		dryRun:       dryRun,
	}
}

func (client *HelmClient)ListReleases(option *release.ReleaseListOption) ([]*release.ReleaseInfo, error) {
	logrus.Debugf("Enter ListReleases %v\n", option)
	options := BuildReleaseListOptions(option)
	resp, err := client.client.ListReleases(options...)
	if err != nil {
		logrus.Errorf("failed to list release: %s\n", err.Error())
		return nil, err
	}

	releaseInfos := []*release.ReleaseInfo{}
	mux := &sync.Mutex{}
	var wg sync.WaitGroup
	for _, helmRelease := range resp.GetReleases() {
		wg.Add(1)
		go func(helmRelease *hapiRelease.Release) {
			defer wg.Done()
			info, err1 := buildReleaseInfo(helmRelease)
			if err1 != nil {
				err = errors.New(fmt.Sprintf("failed to build release info: %s\n", err1.Error()))
				logrus.Error(err.Error())
				return
			}
			mux.Lock()
			releaseInfos = append(releaseInfos, info)
			mux.Unlock()
		}(helmRelease)
	}
	wg.Wait()
	if err != nil {
		return nil, err
	}
	return releaseInfos, nil
}



func (client *HelmClient)GetRelease(namespace, releaseName string) (*release.ReleaseInfo, error) {
	logrus.Debugf("Enter GetRelease %s %s\n", namespace, releaseName)
	var release *release.ReleaseInfo

	res, err := client.client.ReleaseContent(releaseName)
	if err != nil {
		logrus.Errorf(fmt.Sprintf("Failed to get release %s", releaseName))
		return release, err
	}

	release, err = buildReleaseInfo(res.Release)
	if err != nil {
		logrus.Errorf("Failed to build release info: %s\n", err.Error())
		return release, err
	}

	return release, nil
}

func (client *HelmClient)InstallUpgradeRealese(namespace string, releaseRequest *release.ReleaseRequest) error {
	logrus.Infof("Enter InstallUpgradeRealese %v\n", releaseRequest)
	if releaseRequest.ConfigValues == nil {
		releaseRequest.ConfigValues = map[string]interface{}{}
	}
	if releaseRequest.Dependencies == nil {
		releaseRequest.Dependencies = map[string]string{}
	}
	chartRequested, err := client.getChartRequest(releaseRequest.RepoName, releaseRequest.ChartName, releaseRequest.ChartVersion)
	if err != nil {
		return err
	}
	depLinks := make(map[string]interface{})
	for k, v := range releaseRequest.Dependencies {
		depLinks[k] = v
	}
	err = client.installChart(releaseRequest.Name, namespace, releaseRequest.ConfigValues, depLinks, chartRequested)
	if err != nil {
		return err
	}
	return nil
}

func (client *HelmClient)RollbackRealese(namespace, releaseName, version string) error {
	return nil
}

func (client *HelmClient)DeleteRelease(namespace, releaseName string) error {
	logrus.Debugf("Enter DeleteRelease %s %s\n", namespace, releaseName)

	releaseInfo, err := client.GetRelease(namespace, releaseName)
	if err != nil {
		return err
	}

	if releaseInfo.Name == "" {
		logrus.Printf("Can't found %s in ns %s\n", releaseName, namespace)
		return nil
	}

	opts := []helm.DeleteOption{
		helm.DeletePurge(true),
	}
	res, err := client.client.DeleteRelease(
		releaseName, opts...,
	)
	if res != nil && res.Info != "" {
		logrus.Println(res.Info)
	}

	return err
}

func (client *HelmClient)GetDependencies(repoName, chartName, chartVersion string) (subChartNames []string, err error) {
	logrus.Debugf("Enter GetDependencies %s %s\n", chartName, chartVersion)
	chartRequested, err := client.getChartRequest(repoName, chartName, chartVersion)
	if err != nil {
		return nil, err
	}
	dependencies, err := parseChartDependencies(chartRequested)
	if err != nil {
		return nil, err
	}
	return dependencies, nil
}

func (client *HelmClient)downloadChart(repoName, charName, version string) (string, error) {
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

func (client *HelmClient)getChartRequest(repoName, chartName, chartVersion string) (*chart.Chart, error) {
	chartPath, err := client.downloadChart(repoName, chartName, chartVersion)
	if err != nil {
		return nil, err
	}
	chartRequested, err := chartutil.Load(chartPath)
	if err != nil {
		return nil, err
	}

	return chartRequested, nil
}

func (client *HelmClient)installChart(releaseName, namespace string, configValues map[string]interface{}, depLinks map[string]interface{}, chart *chart.Chart) error {
	rawVals, err := yaml.Marshal(configValues)
	if err != nil {
		logrus.Infof("installChart Marshal Error %v\n", err)
		return err
	}
	logrus.Infof("installChart dependency %+v\n", depLinks)
	err = transwarp.ProcessAppCharts(client.client, chart, releaseName, namespace, string(rawVals[:]), depLinks)
	if err != nil {
		logrus.Infof("installChart ProcessAppCharts error %+v\n", err)
		return err
	}
	releaseHistory, err := client.client.ReleaseHistory(releaseName, helm.WithMaxHistory(1))
	if err == nil {
		previousReleaseNamespace := releaseHistory.Releases[0].Namespace
		if previousReleaseNamespace != namespace {
			logrus.Infof("WARNING: Namespace %q doesn't match with previous. Release will be deployed to %s\n",
				namespace, previousReleaseNamespace,
			)
		}
		resp, err := client.client.UpdateReleaseFromChart(
			releaseName,
			chart,
			helm.UpdateValueOverrides(rawVals),
			helm.ReuseValues(true),
			helm.UpgradeDryRun(client.dryRun),
		)
		if err != nil {
			return fmt.Errorf("installChart UPGRADE FAILED: %v", err)
		}
		logrus.Infof("installChart Response %+v\n", resp)
	} else if strings.Contains(err.Error(), driver.ErrReleaseNotFound(releaseName).Error()) {
		resp, err := client.client.InstallReleaseFromChart(
			chart,
			namespace,
			helm.ValueOverrides(rawVals),
			helm.ReleaseName(releaseName),
			helm.InstallDryRun(client.dryRun),
		)
		if err != nil {
			return fmt.Errorf("installChart INSTALL FAILED: %v", err)
		}
		logrus.Infof("installChart Response %+v\n", resp)
	} else {
		return err
	}

	return nil
}


