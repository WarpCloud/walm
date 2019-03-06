package helm

import (
	"path"
	"testing"
	"fmt"
	"io/ioutil"
	"os"
	"walm/pkg/k8s/informer"
	"walm/pkg/setting"
)

func Test_downloadChart(t *testing.T) {
	chartURL, httpGetter, _ := FindChartInChartMuseumRepoURL("http://172.16.1.41:8882/stable/",
		"", "", "hdfs", "")
	fmt.Printf("chartURL %s\n", chartURL)

	tmpDir, _ := ioutil.TempDir("", "")

	filename, _ := ChartMuseumDownloadTo(chartURL, tmpDir, httpGetter)
	fmt.Printf("filename %s\n", filename)
}

func Test_GetDependencies(t *testing.T) {
	subCharts, _ := GetDefaultHelmClient().GetAutoDependencies("stable", "inceptor", "")
	fmt.Printf("%v\n", subCharts)
}

func TestMain(m *testing.M) {
	gopath := os.Getenv("GOPATH")

	setting.Config.RepoList = make([]*setting.ChartRepo, 0)
	setting.Config.RepoList = append(setting.Config.RepoList, &setting.ChartRepo{
		Name: "stable",
		URL: "http://172.16.1.41:8882/stable/",
	})
	setting.Config.RedisConfig = &setting.RedisConfig{
		Addr: "172.16.1.45:6379",
		DB: 0,
	}
	setting.Config.KubeConfig = &setting.KubeConfig{
		Config: path.Join(gopath, "src/walm/", "test/k8sconfig/kubeconfig"),
	}
	stopChan := make(chan struct{})
	informer.StartInformer(stopChan)

	os.Exit(m.Run())
}
