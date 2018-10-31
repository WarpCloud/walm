package helm

import (
	"io/ioutil"
	"fmt"
	"os"
	"net/url"
	"strings"

	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/repo"
	"github.com/ghodss/yaml"
	"path/filepath"
	"walm/pkg/release"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/transwarp"
)

func GetChartIndexFile(repoURL, username, password string) (*repo.IndexFile, error) {
	repoIndex := &repo.IndexFile{}
	tempIndexFile, err := ioutil.TempFile("", "tmp-repo-file")
	if err != nil {
		return nil, fmt.Errorf("cannot write index file for repository requested")
	}
	defer os.Remove(tempIndexFile.Name())

	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return nil, err
	}
	parsedURL.Path = strings.TrimSuffix(parsedURL.Path, "/") + "/index.yaml"

	indexURL := parsedURL.String()
	httpGetter, err := getter.NewHTTPGetter(repoURL, "", "", "")
	httpGetter.SetCredentials(username, password)
	resp, err := httpGetter.Get(indexURL)
	if err != nil {
		return nil, err
	}
	index, err := ioutil.ReadAll(resp)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(index, repoIndex); err != nil {
		return nil, err
	}
	return repoIndex, nil
}

func FindChartInChartMuseumRepoURL(repoURL, username, password, chartName, chartVersion string) (string, getter.Getter, error) {
	tempIndexFile, err := ioutil.TempFile("", "tmp-repo-file")
	if err != nil {
		return "", nil, fmt.Errorf("cannot write index file for repository requested")
	}
	defer os.Remove(tempIndexFile.Name())

	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return "", nil, err
	}
	parsedURL.Path = strings.TrimSuffix(parsedURL.Path, "/") + "/index.yaml"

	indexURL := parsedURL.String()
	httpGetter, err := getter.NewHTTPGetter(repoURL, "", "", "")
	httpGetter.SetCredentials(username, password)
	resp, err := httpGetter.Get(indexURL)
	if err != nil {
		return "", nil, err
	}
	index, err := ioutil.ReadAll(resp)
	if err != nil {
		return "", nil, err
	}
	repoIndex := &repo.IndexFile{}
	if err := yaml.Unmarshal(index, repoIndex); err != nil {
		return "", nil, err
	}
	errMsg := fmt.Sprintf("chart %q", chartName)
	if chartVersion != "" {
		errMsg = fmt.Sprintf("%s version %q", errMsg, chartVersion)
	}
	cv, err := repoIndex.Get(chartName, chartVersion)
	if err != nil {
		return "", nil, fmt.Errorf("%s not found in %s repository", errMsg, repoURL)
	}
	if len(cv.URLs) == 0 {
		return "", nil, fmt.Errorf("%s has no downloadable URLs", errMsg)
	}
	chartURL := cv.URLs[0]

	absoluteChartURL, err := repo.ResolveReferenceURL(repoURL, chartURL)
	if err != nil {
		return "", nil, fmt.Errorf("failed to make chart URL absolute: %v", err)
	}

	return absoluteChartURL, httpGetter, nil
}

func ChartMuseumDownloadTo(ref, dest string, getter getter.Getter) (string, error) {
	data, err := getter.Get(ref)
	if err != nil {
		return "", err
	}

	name := filepath.Base(ref)
	destfile := filepath.Join(dest, name)
	if err := ioutil.WriteFile(destfile, data.Bytes(), 0644); err != nil {
		return destfile, err
	}

	return destfile, nil
}

func GetRepoList() *release.RepoInfoList {
	repoInfoList := new(release.RepoInfoList)
	repoInfoList.Items = make([]*release.RepoInfo, 0)
	for _, v := range GetDefaultHelmClient().chartRepoMap {
		repoInfo := release.RepoInfo{}
		repoInfo.TenantRepoName = v.Name
		repoInfo.TenantRepoURL = v.URL
		repoInfoList.Items = append(repoInfoList.Items, &repoInfo)
	}
	return repoInfoList
}

func GetChartList(TenantRepoName string) (*release.ChartInfoList, error) {
	chartInfoList := new(release.ChartInfoList)
	chartInfoList.Items = make([]*release.ChartInfo, 0)
	chartRepository, ok := GetDefaultHelmClient().chartRepoMap[TenantRepoName]
	if !ok {
		return nil, fmt.Errorf("can't find repo name %s", TenantRepoName)
	}
	indexFile, err := GetChartIndexFile(chartRepository.URL, chartRepository.Username, chartRepository.Password)
	if err != nil {
		return nil, err
	}
	for _, cvs := range indexFile.Entries {
		for _, cv := range cvs {
			chartInfo := new(release.ChartInfo)
			chartInfo.ChartName = cv.Name
			chartInfo.ChartVersion = cv.Version
			chartInfo.ChartAppVersion = cv.AppVersion
			chartInfo.ChartEngine = cv.Engine
			chartInfo.ChartDescription = cv.Description
			chartInfoList.Items = append(chartInfoList.Items, chartInfo)
		}
	}
	return chartInfoList, nil
}

func GetChartInfo(TenantRepoName, ChartName, ChartVersion string) (*release.ChartInfo, error) {
	chartInfo := new(release.ChartInfo)
	appMetaInfo := transwarp.AppDependency{}

	chartRequest, err := GetDefaultHelmClient().getChartRequest(TenantRepoName, ChartName, ChartVersion)
	if err != nil {
		return nil, err
	}
	chartInfo.ChartName = chartRequest.Metadata.Name
	chartInfo.ChartVersion = chartRequest.Metadata.Version
	chartInfo.ChartAppVersion = chartRequest.Metadata.AppVersion
	chartInfo.ChartEngine = chartRequest.Metadata.Engine
	chartInfo.ChartDescription = chartRequest.Metadata.Description
	chartInfo.DefaultValue = chartRequest.Values.Raw
	chartInfo.DependencyCharts = make([]string, 0)
	for _, file := range chartRequest.Files {
		if file.TypeUrl == "transwarp-app-yaml" {
			err := yaml.Unmarshal(file.Value, &appMetaInfo)
			if err != nil {
				return nil, err
			}
			for _, dependency := range appMetaInfo.Dependencies {
				logrus.Infof("%+v", dependency)
				chartInfo.DependencyCharts = append(chartInfo.DependencyCharts, dependency.Name)
			}
		}
	}
	return chartInfo, nil
}