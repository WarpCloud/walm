package project

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"walm/pkg/release"
	"strings"
	consulapi "github.com/hashicorp/consul/api"
	. "walm/pkg/release/manager/project"
	"github.com/ghodss/yaml"
	"encoding/json"
	"walm/pkg/release/manager/helm"
	"github.com/bitly/go-simplejson"
)


var _ = Describe("Project", func() {

	var (
		consulBaseKey string
		consulServerUrl string
		consulClient *consulapi.Client
		namespace string
		project string
		releaseName string
		err error
		isExists bool
		updatedConfig int
	)

	const testSimpleCommonValuesStr = `
			Transwarp_License_Address: 172.16.1.41:2181
			Transwarp_Cni_Network: overlay
			Transwarp_Config:
  			security:
    			auth_type: "none"
	`
	BeforeEach(func() {

		consulBaseKey = "/zhiyang_test"
		consulServerUrl = "http://172.16.1.73:8500"
		namespace = "p1131"
		project = "p1131-app"
		releaseName = "p1131-app--yarn"
		config := consulapi.DefaultConfig()
		config.Address = consulServerUrl
		consulClient, _ = consulapi.NewClient(config)

		By("fetch the consul key-value")
		consulKV := consulClient.KV()
		consulKValue, _, err := consulKV.Get(consulBaseKey + "/HDFS/releases", nil)
		Expect(err).NotTo(HaveOccurred())

		By("construct projectParams")
		commonValuesVal := map[string]interface{}{}
		var data []release.ReleaseRequest
		json.Unmarshal(consulKValue.Value, &data)
		yaml.Unmarshal([]byte(testSimpleCommonValuesStr), &commonValuesVal)

		projectParams := release.ProjectParams{
			CommonValues: commonValuesVal,
			Releases:     make([]*release.ReleaseRequest, len(data)),
		}

		instanceMap := make(map[string]int)
		for index := range data {
			projectParams.Releases[index] = &data[index]
			instanceMap[strings.Join([]string{project, projectParams.Releases[index].Name}, "--")] = 1
		}

		By("start create project")
		err = GetDefaultProjectManager().CreateProject(namespace, project, &projectParams, false, 50000)
		projectInfo, err := GetDefaultProjectManager().GetProjectInfo(namespace, project)
		Expect(err).NotTo(HaveOccurred())

	Loop:
		for i := range projectInfo.Releases {
			walmApplicationInstance := projectInfo.Releases[i].Status.Instances
			for j := range walmApplicationInstance {
				instance := walmApplicationInstance[j].WalmMeta.GetName()
				if _, isExists = instanceMap[instance]; !isExists {
					break Loop
				}
			}
		}

		Expect(isExists).To(BeTrue())
	})



	Describe("update project release", func() {

		It("update project release success", func() {
			By("get project release")
			releasesValue, _, err := consulClient.KV().Get(consulBaseKey + "/HDFS/releases", nil)
			Expect(err).NotTo(HaveOccurred())
			var releasesInfo []release.ReleaseRequest
			json.Unmarshal(releasesValue.Value, &releasesInfo)

			By("update project release")
			for index := range releasesInfo {
				if releasesInfo[index].ChartName == "yarn" {

					ConfigValue, _ := json.Marshal(releasesInfo[index].ConfigValues)
					jsonConfigValue, err := simplejson.NewJson(ConfigValue)
					Expect(err).NotTo(HaveOccurred())

					jsonConfigValue.Get("App").Get("yarnnm").Get("resources").Set("memory_request", 10)
					ConfigValue, err = jsonConfigValue.MarshalJSON()
					Expect(err).NotTo(HaveOccurred())

					err = json.Unmarshal(ConfigValue, &releasesInfo[index].ConfigValues)
					Expect(err).NotTo(HaveOccurred())

					releasesInfo[index].Name = releaseName
					err = helm.GetDefaultHelmClient().InstallUpgradeRealese(namespace, &releasesInfo[index], false)
					Expect(err).NotTo(HaveOccurred())
				}
			}
			By("validate project release")
			newRelease, err := helm.GetDefaultHelmClient().GetRelease(namespace, releaseName)
			Expect(err).NotTo(HaveOccurred())

			newConfigValue, err := json.Marshal(newRelease.ConfigValues)
			Expect(err).NotTo(HaveOccurred())

			newJsonConfigValue, err := simplejson.NewJson(newConfigValue)
			Expect(err).NotTo(HaveOccurred())

			updatedConfig, err = newJsonConfigValue.Get("App").Get("yarnnm").Get("resources").Get("memory_request").Int()
			Expect(err).NotTo(HaveOccurred())

			By("update project release success")
			Expect(updatedConfig).To(Equal(10))
		})

	})

	Describe("delete project", func() {

		It("delete project success", func() {
			err = GetDefaultProjectManager().DeleteProject(namespace, project, true, 5000)
			Expect(err).NotTo(HaveOccurred())
			_, err = GetDefaultProjectManager().GetProjectInfo(namespace, project)
			Expect(err).NotTo(HaveOccurred())
		})
	})

})