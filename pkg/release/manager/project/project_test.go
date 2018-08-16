package project

import (
	"os"
	"fmt"
	"testing"

	"github.com/twmb/algoimpl/go/graph"
	"github.com/ghodss/yaml"

		"walm/pkg/release"
	"walm/pkg/release/manager/helm"
	"walm/pkg/setting"
	"walm/pkg/k8s/informer"
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

func Test_Project_Create(t *testing.T) {
	commonValuesVal := map[string]interface{}{}
	chartList := []string{ "zookeeper", "hdfs", "hyperbase", "inceptor" }

	yaml.Unmarshal([]byte(testSimpleCommonValuesStr), &commonValuesVal)
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
	err := GetDefaultProjectManager().CreateProject("project-test", "test-one", &projectParams)
	fmt.Printf("%v\n", err)
}

func Test_Project_List(t *testing.T) {
	projectInfoList, err := GetDefaultProjectManager().ListProjects("project-test")
	for _, projectInfo := range projectInfoList.Items {
		fmt.Printf("%+v %v\n", *projectInfo, err)
		for _, releaseInfo := range projectInfo.Releases {
			fmt.Printf("ReleaseInfo Name: %s\n", releaseInfo.Name)
			fmt.Printf("ReleaseInfo Chart: %s\n", releaseInfo.ChartName)
			fmt.Printf("ReleaseInfo ChartVersion: %s\n", releaseInfo.ChartVersion)
			fmt.Printf("ReleaseInfo Ready: %v\n", releaseInfo.Ready)
		}
	}
}

func TestMain(m *testing.M) {
	chartRepoMap := make(map[string]*helm.ChartRepository)
	chartRepository := helm.ChartRepository{
		Name: "stable",
		URL: "http://172.16.1.41:8882/stable/",
		Username: "",
		Password: "",
	}
	chartRepoMap["stable"] = &chartRepository
	helm.InitHelmByParams("172.26.0.5:31221", chartRepoMap, false)

	setting.Config.KubeConfig = &setting.KubeConfig{
		Config: "/home/bianyu/user/code/opensource/goproject/src/walm/test/k8sconfig/kubeconfig",
	}
	informer.InitInformer()
	os.Exit(m.Run())
}

const testSimpleCommonValuesStr = `
Transwarp_License_Address: 172.16.1.41:2181
Transwarp_Cni_Network: overlay
Transwarp_Config:
  security:
    auth_type: "none"
`

const testGuardianCommonValuesStr = `
Transwarp_License_Address: 172.16.1.41:2181
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
