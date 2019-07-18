package helm

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/helm/pkg/repo"
	"net/url"
	"path/filepath"
	"strings"

	"WarpCloud/walm/pkg/release"
	walmerr "WarpCloud/walm/pkg/util/error"
	"WarpCloud/walm/pkg/util/transwarpjsonnet"
	"encoding/json"
	"k8s.io/helm/pkg/chart"
	"github.com/go-resty/resty"
)

func GetChartIndexFile(repoURL, username, password string) (*repo.IndexFile, error) {
	repoIndex := &repo.IndexFile{}
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return nil, err
	}
	parsedURL.Path = strings.TrimSuffix(parsedURL.Path, "/") + "/index.yaml"

	indexURL := parsedURL.String()

	resp, err := resty.R().Get(indexURL)
	if err != nil {
		logrus.Errorf("failed to get index : %s", err.Error())
		return nil, err
	}

	if err := yaml.Unmarshal(resp.Body(), repoIndex); err != nil {
		return nil, err
	}
	return repoIndex, nil
}

func LoadChartFromRepo(repoUrl, username, password, chartName, chartVersion, dest string) (string, error) {
	indexFile, err := GetChartIndexFile(repoUrl, username, password)
	if err != nil {
		logrus.Errorf("failed to get chart index file : %s", err.Error())
		return "", err
	}

	cv, err := indexFile.Get(chartName, chartVersion)
	if err != nil {
		return "", fmt.Errorf("chart %s-%s is not found: %s", chartName, chartVersion, err.Error())
	}
	if len(cv.URLs) == 0 {
		return "", fmt.Errorf("chart %s has no downloadable URLs", chartName)
	}
	chartUrl := cv.URLs[0]
	absoluteChartURL, err := repo.ResolveReferenceURL(repoUrl, chartUrl)
	if err != nil {
		return "", fmt.Errorf("failed to make absolute chart url: %v", err)
	}
	resp, err := resty.R().Get(absoluteChartURL)
	if err != nil {
		logrus.Errorf("failed to get chart : %s", err.Error())
		return "", err
	}

	name := filepath.Base(absoluteChartURL)
	destfile := filepath.Join(dest, name)
	if err := ioutil.WriteFile(destfile, resp.Body(), 0644); err != nil {
		logrus.Errorf("failed to write file : %s", err.Error())
		return "", err
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

func GetDetailChartInfoByImage(chartImage string) (*release.ChartDetailInfo, error) {
	rawChart, err := GetDefaultChartImageClient().GetChart(chartImage)
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
		if f.Name == transwarpjsonnet.TranswarpMetadataDir+transwarpjsonnet.TranswarpArchitectureFileName {
			chartDetailInfo.Architecture = string(f.Data[:])
		}
		if f.Name == transwarpjsonnet.TranswarpMetadataDir+transwarpjsonnet.TranswarpAdvantageFileName {
			chartDetailInfo.Advantage = string(f.Data[:])
		}
		if f.Name == transwarpjsonnet.TranswarpMetadataDir+transwarpjsonnet.TranswarpIconFileName {
			chartDetailInfo.Icon = string(f.Data[:])
		}
	}

	chartMetaInfo, err := transwarpjsonnet.GetChartMetaInfo(rawChart)
	if err != nil {
		logrus.Errorf("failed to get chart meta info : %s", err.Error())
		return nil, err
	}
	chartDetailInfo.MetaInfo = chartMetaInfo

	if chartDetailInfo.MetaInfo != nil {
		chartDetailInfo.MetaInfo.BuildDefaultValue(chartDetailInfo.DefaultValue)
	}

	return chartDetailInfo, nil
}

