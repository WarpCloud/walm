package release

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"strings"
	"walm/pkg/release"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"encoding/json"
	"walm/pkg/k8s/handler"
	"walm/pkg/release/manager/helm"
	. "walm/pkg/release/manager/project"

	"go/build"
	"io/ioutil"
	"os"

	"github.com/bitly/go-simplejson"
	"github.com/ghodss/yaml"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

var _ = Describe("Release", func() {

	var (
		namespace     string
		project       string
		releaseName   string
		isExists      bool
		updatedConfig int
		gopath        string
	)

	BeforeEach(func() {

		By("create namespace")
		randomId := uuid.Must(uuid.NewV4()).String()
		namespace = "test-" + randomId[:8]

		ns := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      namespace,
			},
		}
		_, err := handler.GetDefaultHandlerSet().GetNamespaceHandler().CreateNamespace(&ns)
		Expect(err).NotTo(HaveOccurred())

		By("params construct")
		project = namespace + "-app"
		releaseName = project + "--yarn"
		gopath = os.Getenv("GOPATH")
		if gopath == "" {
			gopath = build.Default.GOPATH
		}

		commonValuesVal := map[string]interface{}{}
		commonValuesValStr, err := ioutil.ReadFile(gopath + "/src/walm/test/resources/simpleTest/commonValues.yaml")
		yaml.Unmarshal(commonValuesValStr, &commonValuesVal)

		var data []release.ReleaseRequest
		releaseValue, err := ioutil.ReadFile(gopath + "/src/walm/test/resources/simpleTest/HDFS/releases.yaml")
		Expect(err).NotTo(HaveOccurred())
		json.Unmarshal(releaseValue, &data)

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

	AfterEach(func() {
		err := handler.GetDefaultHandlerSet().GetNamespaceHandler().DeleteNamespace(namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("update project release", func() {

		It("update project release success", func() {
			By("get project release")

			releasesValue, err := ioutil.ReadFile(gopath + "/src/walm/test/resources/simpleTest/HDFS/releases.yaml")
			Expect(err).NotTo(HaveOccurred())
			var releasesInfo []release.ReleaseRequest
			json.Unmarshal(releasesValue, &releasesInfo)

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
					err = helm.GetDefaultHelmClient().InstallUpgradeRealese(namespace, &releasesInfo[index], true)
					Expect(err).NotTo(HaveOccurred())
				}
			}
			By("validate project release")
			newRelease, err := helm.GetDefaultHelmClient().GetRelease(namespace, releaseName)
			logrus.Infof("%s", err)
			Expect(err).NotTo(HaveOccurred())

			newConfigValue, err := json.Marshal(newRelease.ConfigValues)
			Expect(err).NotTo(HaveOccurred())

			newJsonConfigValue, err := simplejson.NewJson(newConfigValue)
			Expect(err).NotTo(HaveOccurred())

			updatedConfig, err = newJsonConfigValue.Get("App").Get("yarnnm").Get("resources").Get("memory_request").Int()
			Expect(err).NotTo(HaveOccurred())

			By("update project release success")
			Expect(updatedConfig).To(Equal(10))

			By("delete project")
			err = GetDefaultProjectManager().DeleteProject(namespace, project, true, 5000)
			Expect(err).NotTo(HaveOccurred())
			_, err = GetDefaultProjectManager().GetProjectInfo(namespace, project)
			Expect(err).NotTo(HaveOccurred())

		})

	})

})
