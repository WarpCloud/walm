package helm

import (
	"testing"
	"k8s.io/helm/pkg/helm"
	"walm/pkg/release"
	"encoding/json"
	"io/ioutil"
	"github.com/sirupsen/logrus"
)

func TestInstallRelease(t *testing.T) {
	hc := &HelmClient{
		systemClient: helm.NewClient(helm.Host("172.26.0.5:31225")),
		chartRepoMap: map[string]*ChartRepository{"stable" : &ChartRepository{
			Name:     "stable",
			URL:      "http://172.16.1.41:8882/stable/",
		}},
	}
	commonTemplateFilesPath = "../../../../test/ksonnet-lib"

	releaseRequest := &release.ReleaseRequest{
		Name: "cy-redis-test",
		ChartVersion: "5.2.0",
		ChartName: "redis",
		RepoName: "stable",
		ConfigValues: map[string]interface{}{},
	}

	testRedisConfig, err := ioutil.ReadFile("./test_redis_config.json")
	if err != nil {
		logrus.Fatalf("Read config file failed! %s", err.Error())
	}
	err = json.Unmarshal(testRedisConfig, &(releaseRequest.ConfigValues))
	if err != nil {
		logrus.Fatalf("failed to unmarshal test redis config : %s", err.Error())
	}

	err = hc.InstallRelease("cytest", releaseRequest, true)
	if err!= nil {
		logrus.Fatalf("failed to install release : %s", err.Error())
	}

}