package release

import (
	"walm/pkg/k8s/handler"

	. "walm/pkg/project"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"

	"go/build"
	"os"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"walm/pkg/release"
	"walm/pkg/release/manager/helm"

	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Release", func() {

	var (
		namespace   string
		project     string
		gopath      string
		releaseName string
		isExists    bool
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
			releaseName = strings.Join([]string{project, projectParams.Releases[index].Name}, "--")
			instanceMap[releaseName] = 1
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

	Describe("install project release", func() {
		It("install project release success", func() {

			releaseRaw, err := ioutil.ReadFile(gopath + "/src/walm/test/resources/simpleTest/TXSQL/release.yaml")
			Expect(err).NotTo(HaveOccurred())

			releaseRequest := release.ReleaseRequest{}
			json.Unmarshal(releaseRaw, &releaseRequest)

			err = helm.GetDefaultHelmClient().InstallUpgradeRelease(namespace, &releaseRequest, true)
			Expect(err).NotTo(HaveOccurred())

			By("get release info fail, install success")
			_, err = helm.GetDefaultHelmClient().GetRelease(namespace, releaseRequest.Name)
			if err != nil {
				fmt.Fprintln(GinkgoWriter, "install project release success")
			}
		})
	})
})
