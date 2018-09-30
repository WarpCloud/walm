package project

import (
	"os"
	"fmt"
	"testing"

	"walm/pkg/util/dag"
	"github.com/ghodss/yaml"

	"walm/pkg/release"
	"walm/pkg/release/manager/helm"
	"walm/pkg/setting"
	"walm/pkg/k8s/informer"
	"sync"
)

func Test_ExampleGraph_TopologicalSort(t *testing.T) {
	var g dag.AcyclicGraph

	g.Add("hat")
	g.Add("shirt")
	g.Add("tie")
	g.Add("jacket")
	g.Add("belt")
	g.Add("watch")
	g.Add("undershorts")
	g.Add("pants")
	g.Add("shoes")
	g.Add("socks")

	g.Connect(dag.BasicEdge("hat", "shirt"))
	g.Connect(dag.BasicEdge("shirt", "tie"))
	g.Connect(dag.BasicEdge("tie", "jacket"))
	g.Connect(dag.BasicEdge("undershorts", "pants"))
	g.Connect(dag.BasicEdge("undershorts", "shoes"))
	g.Connect(dag.BasicEdge("pants", "belt"))
	g.Connect(dag.BasicEdge("pants", "shoes"))
	g.Connect(dag.BasicEdge("socks", "shoes"))

	var visits []dag.Vertex
	var lock sync.Mutex
	err := g.Walk(func(v dag.Vertex) error {
		lock.Lock()
		defer lock.Unlock()
		visits = append(visits, v)
		return nil
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	roots, _ := g.Root()
	fmt.Printf("%+v\n", roots)
	marshalBytes, _ := g.MarshalJSON()
	fmt.Printf("%+v\n", string(marshalBytes[:]))
	fmt.Printf("%+v\n", visits)
	fmt.Printf("%+v\n", g.UpEdges("shoes").List())
	fmt.Printf("%+v\n", g.DownEdges("hat").List())
}

func Test_Project_Create(t *testing.T) {
	commonValuesVal := map[string]interface{}{}
	chartList := []string{ "zookeeper", "txsql" }

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
	InitProject()
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
