package helm

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/repo"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"walm/pkg/release"
	walmerr "walm/pkg/util/error"
	"walm/pkg/util/transwarpjsonnet"
	"encoding/json"
	"k8s.io/helm/pkg/chart"
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
			chartInfo.ChartEngine = "transwarp"
			chartInfo.ChartDescription = cv.Description
			chartInfoList.Items = append(chartInfoList.Items, chartInfo)
		}
	}
	return chartInfoList, nil
}

func GetDetailChartInfo(TenantRepoName, ChartName, ChartVersion string) (*release.ChartDetailInfo, error) {
	rawChart, err := GetDefaultHelmClient().LoadChart(TenantRepoName, ChartName, ChartVersion)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			err = walmerr.NotFoundError{}
		}
		return nil, err
	}

	return BuildChartInfo(rawChart)
}

func BuildChartInfo(rawChart *chart.Chart) (*release.ChartDetailInfo, error) {
	chartDetailInfo := new(release.ChartDetailInfo)

	chartDetailInfo.ChartName = rawChart.Metadata.Name
	chartDetailInfo.ChartVersion = rawChart.Metadata.Version
	chartDetailInfo.ChartAppVersion = rawChart.Metadata.AppVersion
	chartDetailInfo.ChartEngine = "transwarp"
	chartDetailInfo.ChartDescription = rawChart.Metadata.Description

	if len(rawChart.Values) != 0 {
		defaultValueBytes, err := json.Marshal(rawChart.Values)
		if err != nil {
			logrus.Errorf("failed to marshal raw chart values: %s", err.Error())
			return nil, err
		}
		chartDetailInfo.DefaultValue = string(defaultValueBytes)
	}

	for _, f := range rawChart.Files {
		if strings.HasPrefix(f.Name, transwarpjsonnet.TranswarpMetadataDir) {
			cname := strings.TrimPrefix(f.Name, transwarpjsonnet.TranswarpMetadataDir)
			if strings.IndexAny(cname, "._") == 0 {
				// Ignore charts/ that start with . or _.
				continue
			}

			if strings.HasPrefix(cname, "icon") {
				chartDetailInfo.Icon = f.Data
			}
			if cname == "advantage.html" {
				chartDetailInfo.Advantage = f.Data
			}
			if cname == "architecture.html" {
				chartDetailInfo.Architecture = f.Data
			}
			if cname == "metainfo.yaml" {
				chartMetaInfo := release.ChartMetaInfo{}
				err := yaml.Unmarshal(f.Data, &chartMetaInfo)
				if err != nil {
					logrus.Error(errors.Wrapf(err, "chartMetaInfo Unmarshal metainfo.yaml error"))
				}
				chartDetailInfo.MetaInfo = &chartMetaInfo
			}
		}
	}
	return chartDetailInfo, nil
}

func GetChartInfo(TenantRepoName, ChartName, ChartVersion string) (*release.ChartInfo, error) {
	chartInfo := new(release.ChartInfo)

	chartDetailInfo, err := GetDetailChartInfo(TenantRepoName, ChartName, ChartVersion)
	if chartDetailInfo != nil {
		chartInfo = &chartDetailInfo.ChartInfo
	}

	return chartInfo, err
}
