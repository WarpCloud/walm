package helm

import (
	"os"
	"io/ioutil"
	"fmt"
	"strings"

	. "walm/pkg/util/log"

	"github.com/ghodss/yaml"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/downloader"
	"k8s.io/helm/pkg/repo"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/storage/driver"
	"k8s.io/helm/pkg/transwarp"
	"k8s.io/helm/pkg/proto/hapi/chart"
	rls "k8s.io/helm/pkg/proto/hapi/services"
)

type Client struct {
	helmClient *helm.Client
	chartRepository *repo.ChartRepository
}

var Helm *Client

func init() {
	tillerHost := "172.26.0.5:31221"
	client := helm.NewClient(helm.Host(tillerHost))
	helmHome := helmpath.Home("/tmp/helmhome")
	ensureDirectories(helmHome)
	ensureDefaultRepos(helmHome)
	cif := helmHome.CacheIndex("stable")
	c := repo.Entry{
		Name:     "stable",
		Cache:    cif,
		URL:      "http://172.16.1.41:8880",
	}
	r, _ := repo.NewChartRepository(&c, getter.All(environment.EnvSettings{
		Home: helmpath.Home("/tmp/helmhome"),}))
	r.DownloadIndexFile(helmHome.Cache())
	Helm = &Client{
		helmClient: client,
		chartRepository: r,
	}
}

func ListReleases(namespace string) ([]ReleaseInfo, error) {
	res, err := Helm.helmClient.ListReleases(
		helm.ReleaseListNamespace(namespace),
	)
	if err != nil {
		return nil, err
	}

	releases := fillReleaseInfo(res)
	return releases, nil
}

func GetReleaseInfo(namespace, releaseName string) (ReleaseInfo, error) {
	var release ReleaseInfo

	res, err := Helm.helmClient.ListReleases(
		helm.ReleaseListFilter(releaseName),
		helm.ReleaseListNamespace(namespace),
	)
	if err != nil {
		return release, err
	}

	releases := fillReleaseInfo(res)
	for _, rel := range releases {
		if rel.Name == releaseName {
			release = rel
			break
		}
	}
	return ReleaseInfo{}, nil
}

func InstallUpgradeRealese(releaseRequest ReleaseRequest) error {
	chartPath, err := downloadChart(releaseRequest.ChartName, releaseRequest.ChartVersion)
	if err != nil {
		return err
	}
	chartRequested, err := chartutil.Load(chartPath)
	if err != nil {
		return err
	}
	dependencies, err := parseDependencies(chartRequested)
	if err != nil {
		return err
	}
	Log.Printf("Dep %v\n", dependencies)
	depLinks := make(map[string]interface{})
	for k, v := range releaseRequest.Dependencies {
		depLinks[k] = v
	}
	installChart(releaseRequest.Name, releaseRequest.Namespace, releaseRequest.ConfigValues, depLinks, chartRequested)
	return nil
}

func RollbackRealese(namespace, releaseName, version string) error {
	return nil
}

func PatchUpgradeRealese(releaseRequest ReleaseRequest) error {
	return nil
}

func DeleteRealese(namespace, releaseName string) error {
	release, err := GetReleaseInfo(namespace, releaseName)
	if err != nil {
		return err
	}

	if release.Name == "" {
		Log.Printf("Can't found %s in ns %s\n", releaseName, namespace)
		return nil
	}

	opts := []helm.DeleteOption{
		helm.DeletePurge(true),
	}
	res, err := Helm.helmClient.DeleteRelease(
		releaseName, opts...
	)
	if res != nil && res.Info != "" {
		Log.Println(res.Info)
	}

	return err
}

func downloadChart(name, version string) (string, error) {
	dl := downloader.ChartDownloader{
		Out:      os.Stdout,
		HelmHome: helmpath.Home("/tmp/helmhome"),
		Getters:  getter.All(environment.EnvSettings{
			Home: helmpath.Home("/tmp/helmhome"),
		}),
	}

	chartURL, err := repo.FindChartInRepoURL(Helm.chartRepository.Config.URL, name, version,
		"", "", "", getter.All(environment.EnvSettings{
			Home: helmpath.Home("/tmp/helmhome"),
		}))
	if err != nil {
		return "", err
	}

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	filename, _, err := dl.DownloadTo(chartURL, version, tmpDir)
	if err != nil {
		Log.Printf("DownloadTo err %v", err)
		return "", err
	}

	return filename, nil
}

func parseDependencies(chart *chart.Chart) ([]string, error) {
	dependencies := make([]string, 1)
	for _, chartFile := range chart.Files {
		Log.Printf("Chartfile %s \n", chartFile.TypeUrl)
		if chartFile.TypeUrl == "transwarp-app-yaml" {
			app := &AppDependency{}
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
		Log.Printf("Marshal Error %v\n", err)
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
			Log.Printf("WARNING: Namespace %q doesn't match with previous. Release will be deployed to %s\n",
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
		)
		if err != nil {
			return fmt.Errorf("INSTALL FAILED: %v", err)
		}
		Log.Printf("%+v\n", resp)
	}
	resp, err := Helm.helmClient.UpdateReleaseFromChart(
		releaseName,
		chart,
		helm.UpdateValueOverrides(rawVals),
		helm.ReuseValues(true),
	)
	if err != nil {
		return fmt.Errorf("UPGRADE FAILED: %v", err)
	}
	Log.Printf("%+v\n", resp)

	return nil
}

func fillReleaseInfo(helmListReleaseResponse *rls.ListReleasesResponse) []ReleaseInfo {
	var releaseInfos []ReleaseInfo
	depLinks := make(map[string]string)

	for _, helmRelease := range helmListReleaseResponse.GetReleases() {
		release := ReleaseInfo{}
		emptyChart := chart.Chart{}

		release.Name = helmRelease.Name
		release.Namespace = helmRelease.Namespace
		release.Version = helmRelease.Version
		release.ChartVersion = helmRelease.Chart.Metadata.Version
		release.ChartName = helmRelease.Chart.Metadata.Name
		release.ChartAppVersion = helmRelease.Chart.Metadata.AppVersion
		cvals, err := chartutil.CoalesceValues(&emptyChart, helmRelease.Config)
		if err != nil {
			Log.Errorf("parse raw values error %s\n", helmRelease.Config.Raw)
			continue
		}
		release.ConfigValues = cvals
		release.Statuscode = int32(helmRelease.Info.Status.Code)
		depValue, ok := helmRelease.Config.Values["dependencies"]
		if ok {
			yaml.Unmarshal([]byte(depValue.Value), &depLinks)
			release.Dependencies = depLinks
		}

		releaseInfos = append(releaseInfos, release)
	}

	return releaseInfos
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
				return fmt.Errorf("Could not create %s: %s", p, err)
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
