package helm

import (
	"os"
	"io/ioutil"
	"fmt"

	. "walm/pkg/util/log"

	"github.com/ghodss/yaml"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/downloader"
	"k8s.io/helm/pkg/repo"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/proto/hapi/chart"
	rls "k8s.io/helm/pkg/proto/hapi/services"
)

type Interface struct {
	helmClient *helm.Client
	repoURL string
}

var Helm *Interface

func init() {
	tillerHost := "172.26.0.5:31221"
	client := helm.NewClient(helm.Host(tillerHost))
	Helm = &Interface{
		helmClient: client,
		repoURL: "http://172.16.1.41:8880",
	}
	HelmHome := helmpath.Home("/tmp/helmhome")
	ensureDirectories(HelmHome)
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
	parseDependencies(chartRequested)
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

func downloadChart(name, version string) (string, error) {
	dl := downloader.ChartDownloader{
		Out:      os.Stdout,
		HelmHome: helmpath.Home("/tmp/helmhome"),
		Getters:  getter.All(environment.EnvSettings{}),
	}

	repoURL := Helm.repoURL
	if repoURL != "" {
		chartURL, err := repo.FindChartInRepoURL(repoURL, name, version,
			"", "", "", getter.All(environment.EnvSettings{
				Home: helmpath.Home("/tmp/helmhome"),
			}))
		if err != nil {
			return "", err
		}
		name = chartURL
	}

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	filename, _, err := dl.DownloadTo(name, version, tmpDir)

	return filename, nil
}

func parseDependencies(chart *chart.Chart) error {
	Log.Printf("%v\n", chart)
	return nil
}

func installChart(chart *chart.Chart) error {
	resp, err := Helm.helmClient.UpdateReleaseFromChart(
		"",
		chart,
		helm.UpdateValueOverrides([]byte("")),
		helm.ReuseValues(true),
	)
	if err != nil {
		return fmt.Errorf("UPGRADE FAILED: %v", err)
	}
	Log.Printf("%+v\n", resp)

	return nil
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

