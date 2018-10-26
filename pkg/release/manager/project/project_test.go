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
	"github.com/sirupsen/logrus"
	"walm/pkg/kafka"
	"walm/pkg/redis"
	"walm/pkg/job"
	"encoding/json"
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
	chartList := []string{ "zookeeper", "txsql", "hdfs" }

	yaml.Unmarshal([]byte(testSimpleCommonValuesStr), &commonValuesVal)
	projectParams := release.ProjectParams{
		CommonValues: commonValuesVal,
	}
	for _, chartName := range chartList {
		releaseInfo := release.ReleaseRequest{
			Name: fmt.Sprintf("%s", chartName),
			ChartName: chartName,
		}
		releaseInfo.ConfigValues = make(map[string]interface{})
		releaseInfo.Dependencies = make(map[string]string)
		projectParams.Releases = append(projectParams.Releases, &releaseInfo)
	}
	err := GetDefaultProjectManager().CreateProject("project-test", "test-one", &projectParams, true)
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

func Test_Project_ChartDepParse(t *testing.T) {
	commonValuesVal := map[string]interface{}{}
	chartList := []string{ "zookeeper", "txsql", "hdfs", "guardian" }

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
	releases, _ := GetDefaultProjectManager().brainFuckChartDepParse(&projectParams)
	for i := range releases {
		release := releases[len(releases)-1-i]
		fmt.Printf("%s %s %+v\n", release.Name, release.ChartName, release.Dependencies)
	}
}

func Test_Project_ChartRuntimeDepParse(t *testing.T) {
	projectInfo := release.ProjectInfo{}
	projectInfo.Name = "test0"
	projectInfo.Releases = make([]*release.ReleaseInfo, 0)
	projectInfo.Releases = append(projectInfo.Releases, &release.ReleaseInfo{
		ReleaseSpec: release.ReleaseSpec{
			Name:"zookeeper-test",
			ChartName: "zookeeper",
		},
	})
	projectInfo.Releases = append(projectInfo.Releases, &release.ReleaseInfo{
		ReleaseSpec: release.ReleaseSpec{
			Name:"hdfs-test",
			ChartName: "hdfs",
			Dependencies:map[string]string{
				"zookeeper": "zookeeper-test",
			},
		},
	})
	projectInfo.Releases = append(projectInfo.Releases, &release.ReleaseInfo{
		ReleaseSpec: release.ReleaseSpec{
			Name:"metastore-test",
			ChartName: "metastore",
			Dependencies:map[string]string{
				"hdfs": "hdfs-test",
			},
		},
	})
	projectInfo.Releases = append(projectInfo.Releases, &release.ReleaseInfo{
		ReleaseSpec: release.ReleaseSpec{
			Name:"guardian-test",
			ChartName: "guardian",
		},
	})
	releaseParams := release.ReleaseRequest{
		Name: "yarn-test",
		ChartName: "yarn",
		Dependencies:map[string]string{
			"hdfs": "hdfs-aaa",
		},
	}
	releases, _ := GetDefaultProjectManager().brainFuckRuntimeDepParse(&projectInfo, &releaseParams, false)
	fmt.Printf("%+v\n", releaseParams)
	for i := range releases {
		release := releases[len(releases)-1-i]
		fmt.Printf("%s %s %+v\n", release.Name, release.ChartName, release.Dependencies)
	}
}

func Test_Project_ChartRuntimeDepParse2(t *testing.T) {
	projectInfo := release.ProjectInfo{}
	projectInfo.Name = "test0"
	projectInfo.Releases = make([]*release.ReleaseInfo, 0)
	projectInfo.Releases = append(projectInfo.Releases, &release.ReleaseInfo{
		ReleaseSpec: release.ReleaseSpec{
			Name:"zookeeper-test",
			ChartName: "zookeeper",
		},
	})
	projectInfo.Releases = append(projectInfo.Releases, &release.ReleaseInfo{
		ReleaseSpec: release.ReleaseSpec{
			Name:"hdfs-test",
			ChartName: "hdfs",
			Dependencies:map[string]string{
				"zookeeper": "zookeeper-test",
			},
		},
	})
	projectInfo.Releases = append(projectInfo.Releases, &release.ReleaseInfo{
		ReleaseSpec: release.ReleaseSpec{
			Name:"yarn-test",
			ChartName: "yarn",
			Dependencies:map[string]string{
				"zookeeper": "zookeeper-test",
				"hdfs": "hdfs-test",
			},
		},
	})
	projectInfo.Releases = append(projectInfo.Releases, &release.ReleaseInfo{
		ReleaseSpec: release.ReleaseSpec{
			Name:"metastore-test",
			ChartName: "metastore",
			Dependencies:map[string]string{
				"hdfs": "hdfs-test",
				"yarn": "yarn-test",
			},
		},
	})
	releaseParams := release.ReleaseRequest{
		Name: "yarn-test",
		ChartName: "yarn",
	}
	releases, _ := GetDefaultProjectManager().brainFuckRuntimeDepParse(&projectInfo, &releaseParams, true)
	fmt.Printf("%+v\n", releaseParams)
	for i := range releases {
		release := releases[len(releases)-1-i]
		fmt.Printf("%s %s %+v\n", release.Name, release.ChartName, release.Dependencies)
	}
}

func TestMain(m *testing.M) {
	gopath := os.Getenv("GOPATH")

	setting.Config.KubeConfig = &setting.KubeConfig{
		Config: gopath + "/src/walm/test/k8sconfig/kubeconfig",
	}

	logrus.Infof("loading configuration from [%s]", gopath + "/src/walm/walm.yaml")
	setting.InitConfig(gopath + "/src/walm/walm.yaml")
	settingConfig, err := json.Marshal(setting.Config)
	if err != nil {
		logrus.Fatal("failed to marshal setting config")
	}
	setting.Config.KubeConfig.Config = gopath + "/src/walm/test/k8sconfig/kubeconfig"
	logrus.Infof("finished loading configuration: %s", string(settingConfig))

	kafka.InitKafkaClient(setting.Config.KafkaConfig)
	redis.InitRedisClient()
	job.InitWalmJobManager()
	informer.InitInformer()
	helm.InitHelm()
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
