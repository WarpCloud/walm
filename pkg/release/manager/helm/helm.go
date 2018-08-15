package helm

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"bytes"

	"walm/pkg/k8s/client"
	"walm/pkg/setting"
	"walm/pkg/release"
	"walm/pkg/k8s/adaptor"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/repo"
	"k8s.io/helm/pkg/storage/driver"
	"k8s.io/helm/pkg/transwarp"
	hapiRelease "k8s.io/helm/pkg/proto/hapi/release"
	"sync"
	"errors"
)

type ChartRepository struct {
	Name string
	URL string
	Username string
	Password string
}

type Client struct {
	helmClient      *helm.Client
	chartRepoMap map[string]*ChartRepository
	DryRun bool
}

var Helm *Client

func InitHelm() {
	tillerHost := setting.Config.SysHelm.TillerHost
	helmClient := helm.NewClient(helm.Host(tillerHost))
	chartRepoMap := make(map[string]*ChartRepository)

	for _, chartRepo := range setting.Config.RepoList {
		chartRepository := ChartRepository{
			Name: chartRepo.Name,
			URL: chartRepo.URL,
			Username: "",
			Password: "",
		}
		chartRepoMap[chartRepo.Name] = &chartRepository
	}

	Helm = &Client{
		helmClient: helmClient,
		chartRepoMap: chartRepoMap,
		DryRun: false,
	}
}

func InitHelmByParams(tillerHost string, chartRepoMap map[string]*ChartRepository) {
	helmClient := helm.NewClient(helm.Host(tillerHost))

	Helm = &Client{
		helmClient: helmClient,
		chartRepoMap: chartRepoMap,
		DryRun: true,
	}
}

func ListReleases(option *release.ReleaseListOption) ([]*release.ReleaseInfo, error) {
	logrus.Debugf("Enter ListReleases %v\n", option)
	options := BuildReleaseListOptions(option)
	resp, err := Helm.helmClient.ListReleases(options...)
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
			info, err1 := BuildReleaseInfo(helmRelease)
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

func BuildReleaseInfo(helmRelease *hapiRelease.Release) (releaseInfo *release.ReleaseInfo, err error) {
	emptyChart := chart.Chart{}
	depLinks := make(map[string]string)
	releaseInfo = &release.ReleaseInfo{}
	releaseInfo.Name = helmRelease.Name
	releaseInfo.Namespace = helmRelease.Namespace
	releaseInfo.Version = helmRelease.Version
	releaseInfo.ChartVersion = helmRelease.Chart.Metadata.Version
	releaseInfo.ChartName = helmRelease.Chart.Metadata.Name
	releaseInfo.ChartAppVersion = helmRelease.Chart.Metadata.AppVersion
	cvals, err := chartutil.CoalesceValues(&emptyChart, helmRelease.Config)
	if err != nil {
		logrus.Errorf("parse raw values error %s\n", helmRelease.Config.Raw)
		return
	}
	releaseInfo.ConfigValues = cvals
	depValue, ok := helmRelease.Config.Values["dependencies"]
	if ok {
		yaml.Unmarshal([]byte(depValue.Value), &depLinks)
		releaseInfo.Dependencies = depLinks
	}

	if helmRelease.Info.Status.Code == hapiRelease.Status_DEPLOYED {
		releaseInfo.Status, err = buildReleaseStatus(helmRelease)
		if err != nil {
			logrus.Errorf(fmt.Sprintf("Failed to build the status of releaseInfo: %s", releaseInfo.Name))
			return
		}
		releaseInfo.Ready = isReleaseReady(releaseInfo.Status)
	}

	return
}
func isReleaseReady(status *release.ReleaseStatus) bool {
	ready := true
	for _, resource := range status.Resources {
		if resource.Resource.GetState().Status != "Ready" {
			ready = false
			break
		}
	}
	return ready
}

func GetRelease(namespace, releaseName string) (*release.ReleaseInfo, error) {
	logrus.Debugf("Enter GetRelease %s %s\n", namespace, releaseName)
	var release *release.ReleaseInfo

	res, err := Helm.helmClient.ReleaseContent(releaseName)
	if err != nil {
		logrus.Errorf(fmt.Sprintf("Failed to get release %s", releaseName))
		return release, err
	}

	release, err = BuildReleaseInfo(res.Release)
	if err != nil {
		logrus.Errorf("Failed to build release info: %s\n", err.Error())
		return release, err
	}

	return release, nil
}

func InstallUpgradeRealese(releaseRequest *release.ReleaseRequest) error {
	logrus.Infof("Enter InstallUpgradeRealese %v\n", releaseRequest)
	chartRequested, err := getChartRequest(releaseRequest.ChartName, releaseRequest.ChartVersion)
	if err != nil {
		return err
	}
	dependencies, err := parseChartDependencies(chartRequested)
	if err != nil {
		return err
	}
	logrus.Printf("InstallUpgradeRealese Dependency %v\n", dependencies)
	depLinks := make(map[string]interface{})
	for k, v := range releaseRequest.Dependencies {
		depLinks[k] = v
	}
	err = installChart(releaseRequest.Name, releaseRequest.Namespace, releaseRequest.ConfigValues, depLinks, chartRequested)
	if err != nil {
		return err
	}
	return nil
}

func RollbackRealese(namespace, releaseName, version string) error {
	return nil
}

func PatchUpgradeRealese(releaseRequest release.ReleaseRequest) error {
	return nil
}

func DeleteRealese(namespace, releaseName string) error {
	logrus.Debugf("Enter DeleteRealese %s %s\n", namespace, releaseName)

	releaseInfo, err := GetRelease(namespace, releaseName)
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
	res, err := Helm.helmClient.DeleteRelease(
		releaseName, opts...,
	)
	if res != nil && res.Info != "" {
		logrus.Println(res.Info)
	}

	return err
}


func GetDependencies(chartName, chartVersion string) (subChartNames []string, err error) {
	logrus.Debugf("Enter GetDependencies %s %s\n", chartName, chartVersion)
	chartRequested, err := getChartRequest(chartName, chartVersion)
	if err != nil {
		return nil, err
	}
	dependencies, err := parseChartDependencies(chartRequested)
	if err != nil {
		return nil, err
	}
	return dependencies, nil
}

func downloadChart(name, version string) (string, error) {
	chartURL, httpGetter, err := FindChartInChartMuseumRepoURL(Helm.chartRepoMap["stable"].URL, "", "", name, version)
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

func getChartRequest(chartName, chartVersion string) (*chart.Chart, error) {
	chartPath, err := downloadChart(chartName, chartVersion)
	if err != nil {
		return nil, err
	}
	chartRequested, err := chartutil.Load(chartPath)
	if err != nil {
		return nil, err
	}

	return chartRequested, nil
}

func parseChartDependencies(chart *chart.Chart) ([]string, error) {
	var dependencies []string

	for _, chartFile := range chart.Files {
		logrus.Printf("Chartfile %s \n", chartFile.TypeUrl)
		if chartFile.TypeUrl == "transwarp-app-yaml" {
			app := &release.AppDependency{}
			err := yaml.Unmarshal(chartFile.Value, &app)
			if err != nil {
				return dependencies, err
			}
			for _, dependency := range app.Dependencies {
				dependencies = append(dependencies, dependency.Name)
			}
		}
	}

	return dependencies, nil
}

func installChart(releaseName, namespace string, configValues map[string]interface{}, depLinks map[string]interface{}, chart *chart.Chart) error {
	rawVals, err := yaml.Marshal(configValues)
	if err != nil {
		logrus.Printf("installChart Marshal Error %v\n", err)
		return err
	}
	err = transwarp.ProcessAppCharts(Helm.helmClient, chart, releaseName, namespace, string(rawVals[:]), depLinks)
	if err != nil {
		return err
	}
	releaseHistory, err := Helm.helmClient.ReleaseHistory(releaseName, helm.WithMaxHistory(1))
	if err == nil {
		previousReleaseNamespace := releaseHistory.Releases[0].Namespace
		if previousReleaseNamespace != namespace {
			logrus.Printf("WARNING: Namespace %q doesn't match with previous. Release will be deployed to %s\n",
				namespace, previousReleaseNamespace,
			)
		}
	}
	if err != nil && strings.Contains(err.Error(), driver.ErrReleaseNotFound(releaseName).Error()) {
		resp, err := Helm.helmClient.InstallReleaseFromChart(
			chart,
			namespace,
			helm.ValueOverrides(rawVals),
			helm.ReleaseName(releaseName),
			helm.InstallDryRun(Helm.DryRun),
		)
		if err != nil {
			return fmt.Errorf("installChart INSTALL FAILED: %v", err)
		}
		logrus.Infof("installChart Response %+v\n", resp)
	} else {
		resp, err := Helm.helmClient.UpdateReleaseFromChart(
			releaseName,
			chart,
			helm.UpdateValueOverrides(rawVals),
			helm.ReuseValues(true),
			helm.UpgradeDryRun(Helm.DryRun),
		)
		if err != nil {
			return fmt.Errorf("installChart UPGRADE FAILED: %v", err)
		}
		logrus.Infof("installChart Response %+v\n", resp)
	}

	return nil
}

func buildReleaseStatus(helmRelease *hapiRelease.Release) (*release.ReleaseStatus, error) {
	status := &release.ReleaseStatus{[]release.ReleaseResource{}}
	resourceMetas, err := getReleaseResourceMetas(helmRelease)
	if err != nil {
		return status, err
	}
	for _, resourceMeta := range resourceMetas {
		resource, err := adaptor.GetDefaultAdaptorSet().GetAdaptor(resourceMeta.Kind).GetResource(resourceMeta.Namespace, resourceMeta.Name)
		if err != nil {
			return status, err
		}
		if resource.GetState().Status == "Unknown" && resource.GetState().Reason == "NotSupportedKind" {
			continue
		}
		status.Resources = append(status.Resources, release.ReleaseResource{Kind: resource.GetKind(), Resource: resource})
	}
	return status, nil
}

func getReleaseResourceMetas(helmRelease *hapiRelease.Release) (resources []release.ReleaseResourceMeta, err error) {
	resources = []release.ReleaseResourceMeta{}
	results, err := client.GetKubeClient().BuildUnstructured(helmRelease.Namespace, bytes.NewBufferString(helmRelease.Manifest))
	if err != nil {
		return resources, err
	}
	for _, result := range results {
		resource := release.ReleaseResourceMeta{
			Kind:      result.Object.GetObjectKind().GroupVersionKind().Kind,
			Namespace: result.Namespace,
			Name:      result.Name,
		}
		resources = append(resources, resource)
	}
	return
}

func ensureDirectories(home helmpath.Home) error {
	configDirectories := []string{
		home.String(),
		home.Repository(),
		home.Cache(),
		home.LocalRepository(),
		home.Plugins(),
		home.Starters(),
		home.Archive(),
	}
	for _, p := range configDirectories {
		if fi, err := os.Stat(p); err != nil {
			fmt.Printf("Creating %s \n", p)
			if err := os.MkdirAll(p, 0755); err != nil {
				return fmt.Errorf("could not create %s: %s", p, err)
			}
		} else if !fi.IsDir() {
			return fmt.Errorf("%s must be a directory", p)
		}
	}

	return nil
}

func ensureDefaultRepos(home helmpath.Home) error {
	repoFile := home.RepositoryFile()
	if fi, err := os.Stat(repoFile); err != nil {
		fmt.Printf("Creating %s \n", repoFile)
		f := repo.NewRepoFile()
		if err := f.WriteFile(repoFile, 0644); err != nil {
			return err
		}
	} else if !fi.IsDir() {
		return fmt.Errorf("%s must be a directory", repoFile)
	}

	return nil
}

