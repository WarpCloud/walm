package framework

import (
	"path/filepath"
	"github.com/sirupsen/logrus"
	"strings"
	"github.com/go-resty/resty"
	"WarpCloud/walm/pkg/setting"

	"errors"
	"os"
	"WarpCloud/walm/pkg/models/common"
)

func GetLocalTomcatChartPath() (string, error) {
	currentFilePath, err := GetCurrentFilePath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(currentFilePath), "../../resources/helm/tomcat-0.2.0.tgz"), nil
}

func GetLocalV1ZookeeperChartPath() (string, error) {
	currentFilePath, err := GetCurrentFilePath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(currentFilePath), "../../resources/helm/v1/zookeeper-5.2.0.tgz"), nil
}

func GetLocalV2ZookeeperChartPath() (string, error) {
	currentFilePath, err := GetCurrentFilePath()
	if err != nil {
		return "", nil
	}
	return filepath.Join(filepath.Dir(currentFilePath), "../../resources/helm/zookeeper-6.1.0.tgz"), nil
}

func PushChartToRepo(repoBaseUrl, chartPath string) error{
	logrus.Infof("start to push %s to repo %s", chartPath, repoBaseUrl)
	if !strings.HasSuffix(repoBaseUrl, "/") {
		repoBaseUrl += "/"
	}

	fullUrl := repoBaseUrl + "api/charts"

	resp, err := resty.R().SetHeader("Content-Type", "multipart/form-data" ).
		SetFile("chart", chartPath).Post(fullUrl)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		logrus.Errorf("status code : %d", resp.StatusCode())
		return errors.New(resp.String())
	}
	return nil
}

func GetTomcatChartImage() string {
	chartImageRegistry := setting.Config.ChartImageRegistry
	if !strings.HasSuffix(chartImageRegistry, "/") {
		chartImageRegistry += "/"
	}
	return chartImageRegistry + tomcatChartImageSuffix
}

func LoadChartArchive(name string) ([]*common.BufferedFile, error) {
	if fi, err := os.Stat(name); err != nil {
		return nil, err
	} else if fi.IsDir() {
		return nil, errors.New("cannot load a directory")
	}

	raw, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer raw.Close()
	return common.LoadArchive(raw)
}