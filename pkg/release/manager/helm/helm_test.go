package helm

import (
	"testing"
	"fmt"
	"io/ioutil"
	"os"
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
	subCharts, _ := GetDefaultHelmClient().GetDependencies("stable", "inceptor", "")
	fmt.Printf("%v\n", subCharts)
}

func TestMain(m *testing.M) {
	chartRepoMap := make(map[string]*ChartRepository)
	chartRepository := ChartRepository{
		Name: "stable",
		URL: "http://172.16.1.41:8882/stable/",
		Username: "",
		Password: "",
	}
	chartRepoMap["stable"] = &chartRepository
	InitHelmByParams("172.26.0.5:31221", chartRepoMap, true)

	os.Exit(m.Run())
}
