package helm

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	. "walm/pkg/util/log"

	"walm/pkg/setting"

	"github.com/ghodss/yaml"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/downloader"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/proto/hapi/chart"
	rls "k8s.io/helm/pkg/proto/hapi/services"
	"k8s.io/helm/pkg/repo"
	"k8s.io/helm/pkg/storage/driver"
	"k8s.io/helm/pkg/transwarp"
	"k8s.io/helm/pkg/strvals"
	"walm/pkg/k8s/client"
	"bytes"
	"k8s.io/helm/pkg/timeconv"
	"k8s.io/helm/pkg/engine"
	"walm/pkg/release"
	hapi_release5 "k8s.io/helm/pkg/proto/hapi/release"
	"walm/pkg/k8s/adaptor"
)

type Client struct {
	helmClient      *helm.Client
	chartRepository *repo.ChartRepository
}

var Helm *Client

func init() {
	tillerHost := setting.Config.Helm.TillerHost
	client := helm.NewClient(helm.Host(tillerHost))
	helmHome := helmpath.Home("/tmp/helmhome")
	ensureDirectories(helmHome)
	ensureDefaultRepos(helmHome)
	cif := helmHome.CacheIndex("stable")
	c := repo.Entry{
		Name:  setting.Config.Repo.Name,
		Cache: cif,
		URL:   setting.Config.Repo.URL,
	}
	r, _ := repo.NewChartRepository(&c, getter.All(environment.EnvSettings{
		Home: helmpath.Home("/tmp/helmhome")}))
	r.DownloadIndexFile(helmHome.Cache())
	Helm = &Client{
		helmClient:      client,
		chartRepository: r,
	}
}

func ListReleases(namespace string) ([]release.ReleaseInfo, error) {
	Log.Debugf("Enter ListReleases %s\n", namespace)
	res, err := Helm.helmClient.ListReleases(
		helm.ReleaseListNamespace(namespace),
	)
	if err != nil {
		return nil, err
	}

	releases, err := fillReleaseInfo(res)
	if err != nil {
		return nil, err
	}
	return releases, nil
}

func GetReleaseInfo(namespace, releaseName string) (release.ReleaseInfo, error) {
	Log.Debugf("Enter GetReleaseInfo %s %s\n", namespace, releaseName)
	var release release.ReleaseInfo

	res, err := Helm.helmClient.ListReleases(
		helm.ReleaseListFilter(releaseName),
		helm.ReleaseListNamespace(namespace),
	)
	if err != nil {
		return release, err
	}

	releases, err := fillReleaseInfo(res)
	if err != nil {
		return release, err
	}
	for _, rel := range releases {
		if rel.Name == releaseName {
			release = rel
			break
		}
	}
	return release, nil
}

func InstallUpgradeRealese(releaseRequest release.ReleaseRequest) error {
	Log.Debugf("Enter InstallUpgradeRealese %v\n", releaseRequest)
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


func ValidateChart(releaseRequest release.ReleaseRequest) (release.ChartValicationInfo, error) {

	Log.Debugf("Begin ValidateChart %v\n", releaseRequest)

	var chartValicationInfo release.ChartValicationInfo
	chartValicationInfo.ChartName = releaseRequest.ChartName
	chartValicationInfo.Name = releaseRequest.Name
	chartValicationInfo.ConfigValues = releaseRequest.ConfigValues
	chartValicationInfo.ChartVersion = releaseRequest.ChartVersion
	chartValicationInfo.Dependencies = releaseRequest.Dependencies
	chartValicationInfo.Namespace = releaseRequest.Namespace


	chartPath, err := downloadChart(releaseRequest.ChartName, releaseRequest.ChartVersion)
	if err != nil {
		return chartValicationInfo, err
	}
	chartRequested, err := chartutil.Load(chartPath)
	if err != nil {
		return chartValicationInfo, err
	}

	if releaseRequest.Namespace == "" {
		releaseRequest.Namespace = "default"
	}

	rawVals, err := yaml.Marshal(releaseRequest.ConfigValues)
	if err != nil {
		return chartValicationInfo, err
	}
	config := &chart.Config{Raw: string(rawVals), Values: map[string]*chart.Value{}}

	var links []string
	for k, v := range releaseRequest.Dependencies {
		tmpStr := k + "=" + v
		links = append(links, tmpStr)
	}

	out := make(map[string]string)
	if chartRequested.Metadata.Engine == "jsonnet" {

		if len(links) > 0 {

			out, err = renderWithDependencies(chartRequested, releaseRequest.Namespace, rawVals, "1.9", "", links)

		} else {

			out, err = render(chartRequested, releaseRequest.Namespace, rawVals, "1.9")
		}

	} else {

		if req, err := chartutil.LoadRequirements(chartRequested); err == nil {

			if err := checkDependencies(chartRequested, req); err != nil {
				return chartValicationInfo, fmt.Errorf("checkDependencies: %v", err)
			}

		} else if err != chartutil.ErrRequirementsNotFound {
			return  chartValicationInfo, fmt.Errorf("checkDependencies: %v", err)
		}

		options := chartutil.ReleaseOptions{
			Name:      releaseRequest.Name,
			IsInstall: false,
			IsUpgrade: false,
			Time:      timeconv.Now(),
			Namespace: releaseRequest.Namespace,
		}

		err = chartutil.ProcessRequirementsEnabled(chartRequested, config)
		if err != nil {
			return chartValicationInfo, err
		}
		err = chartutil.ProcessRequirementsImportValues(chartRequested)
		if err != nil {
			return chartValicationInfo, err
		}

		// Set up engine.
		renderer := engine.New()

		caps := &chartutil.Capabilities{
			APIVersions:   chartutil.DefaultVersionSet,
			KubeVersion:   chartutil.DefaultKubeVersion,
		}

		vals, err := chartutil.ToRenderValuesCaps(chartRequested, config, options, caps)
		if err != nil {
			return chartValicationInfo, err
		}

		out, err = renderer.Render(chartRequested, vals)
		if err != nil {
			return chartValicationInfo, err
		}

	}

	if err != nil {
		return chartValicationInfo, err
	}

	chartValicationInfo.RenderStatus = "ok"
	chartValicationInfo.RenderResult = out

	resultMap, errFlag := dryRunK8sResource(out, releaseRequest.Namespace)
	if errFlag {
		chartValicationInfo.DryRunStatus = "failed"
		chartValicationInfo.ErrorMessage = "dry run check fail"
	}else {
		chartValicationInfo.DryRunStatus = "ok"
		chartValicationInfo.ErrorMessage = " test pass "
	}

	chartValicationInfo.DryRunResult = resultMap
	return chartValicationInfo, nil

}

func dryRunK8sResource(out map[string]string, namespace string) (map[string]string, bool) {

	resultMap := make(map[string]string)
	errFlag := false
	kubeClient := client.GetKubeClient()
	for name, content := range out {

		if strings.HasSuffix(name, "NOTES.txt") {
			continue
		}

		r := bytes.NewReader([]byte(content))
		_, err := kubeClient.BuildUnstructured(namespace, r)
		if err != nil {
			resultMap[name] = err.Error()
			errFlag = true
		}else {
			resultMap[name] = " dry run suc"
		}

	}

	return  resultMap, errFlag

}

func RollbackRealese(namespace, releaseName, version string) error {
	return nil
}

func PatchUpgradeRealese(releaseRequest release.ReleaseRequest) error {
	return nil
}

func DeleteRealese(namespace, releaseName string) error {
	Log.Debugf("Enter DeleteRealese %s %s\n", namespace, releaseName)
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
		releaseName, opts...,
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
		Getters: getter.All(environment.EnvSettings{
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

func GetDependencies(chartName, chartVersion string) (chartNames, chartVersions []string, err error) {
	return []string{}, []string{}, nil
}

func parseDependencies(chart *chart.Chart) ([]string, error) {

	var dependencies []string
	//dependencies := make([]string, 0)
	for _, chartFile := range chart.Files {
		Log.Printf("Chartfile %s \n", chartFile.TypeUrl)
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

func fillReleaseInfo(helmListReleaseResponse *rls.ListReleasesResponse) ([]release.ReleaseInfo, error) {
	var releaseInfos []release.ReleaseInfo
	depLinks := make(map[string]string)

	for _, helmRelease := range helmListReleaseResponse.GetReleases() {
		release := release.ReleaseInfo{}
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

		release.Status, err = buildReleaseStatus(helmRelease)
		if err != nil {
			Log.Errorf(fmt.Sprintf("Failed to build the status of release: %s", release.Name))
			return releaseInfos, err
		}

		releaseInfos = append(releaseInfos, release)
	}

	return releaseInfos, nil
}

func buildReleaseStatus(helmRelease *hapi_release5.Release) (release.ReleaseStatus, error) {
	status := release.ReleaseStatus{[]release.ReleaseResource{}}
	for _, resourceMeta := range getReleaseResourceMetas(helmRelease) {
		resource, err := adaptor.GetDefaultAdaptorSet().GetAdaptor(resourceMeta.Kind).GetResource(resourceMeta.Namespace, resourceMeta.Name)
		if err != nil {
			return status, err
		}

		status.Resources = append(status.Resources, release.ReleaseResource{Kind: resource.GetKind(), Resource: resource})
	}
	return status, nil
}

// TODO
func getReleaseResourceMetas(helmRelease *hapi_release5.Release) []release.ReleaseResourceMeta {
	return []release.ReleaseResourceMeta{
		release.ReleaseResourceMeta{
			Kind: "ApplicationInstance",
			Namespace: helmRelease.Namespace,
			Name: helmRelease.Name,
		},
	}
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

func renderWithDependencies(chartRequested *chart.Chart, namespace string, userVals []byte, kubeVersion string, kubeContext string, links []string) (map[string]string, error) {

	depLinks := map[string]interface{}{}
	for _, value := range links {
		if err := strvals.ParseInto(value, depLinks); err != nil {
			return nil, fmt.Errorf("failed parsing --set data: %s", err)
		}
	}

	err := transwarp.CheckDepencies(chartRequested, depLinks)
	if err != nil {
		return nil, err
	}

	// init k8s transwarp client
	k8sTranswarpClient := client.GetDefaultClientEx()

	// init k8s client
	k8sClient := client.GetDefaultClient()

	depVals, err := transwarp.GetDepenciesConfig(k8sTranswarpClient, k8sClient, namespace, depLinks)
	if err != nil {
		return nil, err
	}

	newVals, err := transwarp.MergeDepenciesValue(depVals, userVals)
	if err != nil {
		return nil, err
	}

	return transwarp.Render(chartRequested, namespace, newVals, kubeVersion)

}


func render(chartRequested *chart.Chart, namespace string, userVals []byte, kubeVersion string) (map[string]string, error) {

	return transwarp.Render(chartRequested, namespace, userVals, kubeVersion)
}


func checkDependencies(ch *chart.Chart, reqs *chartutil.Requirements) error {
	missing := []string{}

	deps := ch.GetDependencies()
	for _, r := range reqs.Dependencies {
		found := false
		for _, d := range deps {
			if d.Metadata.Name == r.Name {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, r.Name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("found in requirements.yaml, but missing in charts/ directory: %s", strings.Join(missing, ", "))
	}
	return nil
}