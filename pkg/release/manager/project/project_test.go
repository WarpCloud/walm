package project

import (
	"fmt"
	"github.com/twmb/algoimpl/go/graph"
	"testing"
	"walm/pkg/release"
	"github.com/ghodss/yaml"
	"walm/pkg/release/manager/helm"
	"os"
)

func Test_ExampleGraph_TopologicalSort(t *testing.T) {
	g := graph.New(graph.Directed)

	clothes := make(map[string]graph.Node, 0)
	// Make a mapping from strings to a node
	clothes["hat"] = g.MakeNode()
	clothes["shirt"] = g.MakeNode()
	clothes["tie"] = g.MakeNode()
	clothes["jacket"] = g.MakeNode()
	clothes["belt"] = g.MakeNode()
	clothes["watch"] = g.MakeNode()
	clothes["undershorts"] = g.MakeNode()
	clothes["pants"] = g.MakeNode()
	clothes["shoes"] = g.MakeNode()
	clothes["socks"] = g.MakeNode()
	// Make references back to the string values
	for key, node := range clothes {
		*node.Value = key
	}
	// Connect the elements
	g.MakeEdge(clothes["shirt"], clothes["tie"])
	g.MakeEdge(clothes["tie"], clothes["jacket"])
	g.MakeEdge(clothes["shirt"], clothes["belt"])
	g.MakeEdge(clothes["belt"], clothes["jacket"])
	g.MakeEdge(clothes["undershorts"], clothes["pants"])
	g.MakeEdge(clothes["undershorts"], clothes["shoes"])
	g.MakeEdge(clothes["pants"], clothes["belt"])
	g.MakeEdge(clothes["pants"], clothes["shoes"])
	g.MakeEdge(clothes["socks"], clothes["shoes"])
	sorted := g.TopologicalSort()
	for i := range sorted {
		fmt.Println(*sorted[i].Value)
	}
	// Output:
	// socks
	// undershorts
	// pants
	// shoes
	// watch
	// shirt
	// belt
	// tie
	// jacket
	// hat
}

func Test_ProjectGraph_Dependency(t *testing.T) {
	chartRepoMap := make(map[string]*helm.ChartRepository)
	chartRepository := helm.ChartRepository{
		Name: "stable",
		URL: "http://172.16.1.41:8882/stable/",
		Username: "",
		Password: "",
	}
	chartRepoMap["stable"] = &chartRepository
	helm.InitHelmByParams("172.26.0.5:31221", chartRepoMap)

	commonValuesVal := map[string]interface{}{}
	chartList := []string{ "zookeeper", "hdfs", "hyperbase", "inceptor" }

	yaml.Unmarshal([]byte(testCommonValuesStr), &commonValuesVal)
	projectParams := release.ProjectParams{
		CommonValues: commonValuesVal,
	}
	for _, chartName := range chartList {
		releaseInfo := release.ReleaseRequest{
			Name: fmt.Sprintf("%s-%s", chartName, "test"),
			ChartName: chartName,
		}
		releaseInfo.ConfigValues = make(map[string]interface{})
		releaseInfo.Dependencies = make(map[string]string)
		projectParams.Releases = append(projectParams.Releases, &releaseInfo)
	}
	err := CreateProject("project-test", "test-one", &projectParams)
	fmt.Printf("%v\n", err)
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

const testCommonValuesStr = `
Transwarp_License_Address: 172.16.3.231:2191
Transwarp_Cni_Network: overlay
Transwarp_Config:
  Transwarp_Auto_Injected_Volumes:
  - name: "keytab"
    volumeName: "keytab"
    secretname: all-keytab
  security:
    auth_type: "kerberos"
    guardian_plugin_enable: "true"
`
