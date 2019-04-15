package release

import (
	"encoding/json"
	"go/build"
	"io/ioutil"
	"os"
	"walm/pkg/k8s/handler"

	"github.com/bitly/go-simplejson"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"walm/pkg/release/manager/helm"
	"walm/pkg/release"
)

var _ = Describe("Release", func() {

	var (
		namespace      string
		gopath         string
		releaseName    string
		releaseRequest release.ReleaseRequestV2
		err            error
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

		By("create a release")

		gopath = os.Getenv("GOPATH")
		if gopath == "" {
			gopath = build.Default.GOPATH
		}
		releaseChartByte, err := ioutil.ReadFile(gopath + "/src/walm/test/resources/releases/smartbi.yaml")
		Expect(err).NotTo(HaveOccurred())

		err = json.Unmarshal(releaseChartByte, &releaseRequest)
		Expect(err).NotTo(HaveOccurred())

		releaseRequest.Name = releaseRequest.Name + "-" + randomId[:8]
		releaseName = releaseRequest.Name
		helm.GetDefaultHelmClient().InstallUpgradeRelease(namespace, &releaseRequest, false, nil, false, 0)

		releaseInfo, err := helm.GetDefaultHelmClient().GetRelease(namespace, releaseName)
		Expect(releaseInfo.Name).To(Equal(releaseName))
	})

	AfterEach(func() {

		By("delete release")
		err := helm.GetDefaultHelmClient().DeleteRelease(namespace, releaseName, false, true, false, 0)
		Expect(err).NotTo(HaveOccurred())

		_, err = helm.GetDefaultHelmClient().GetRelease(namespace, releaseName)
		Expect(err).To(HaveOccurred())

		By("delete namespace")
		err = handler.GetDefaultHandlerSet().GetNamespaceHandler().DeleteNamespace(namespace)
		Expect(err).NotTo(HaveOccurred())

	})

	Describe("update release", func() {

		It("update release success", func() {

			By("update release value")

			ConfigValue, err := json.Marshal(releaseRequest.ConfigValues)
			Expect(err).NotTo(HaveOccurred())
			jsonConfigValue, err := simplejson.NewJson(ConfigValue)
			Expect(err).NotTo(HaveOccurred())

			jsonConfigValue.Get("appConfig").Get("mysql").Set("replicas", 2)
			jsonConfigValue.Get("appConfig").Get("mysql").Get("resources").Set("memory_request", 2)

			ConfigValue, err = jsonConfigValue.MarshalJSON()
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal(ConfigValue, &releaseRequest.ConfigValues)
			Expect(err).NotTo(HaveOccurred())

			helm.GetDefaultHelmClient().InstallUpgradeRelease(namespace, &releaseRequest, false, nil, false, 0)

			By("validate release value")

			newRelease, err := helm.GetDefaultHelmClient().GetRelease(namespace, releaseName)
			Expect(err).NotTo(HaveOccurred())
			newConfigValue, err := json.Marshal(newRelease.ConfigValues)
			Expect(err).NotTo(HaveOccurred())
			newJsonConfigValue, err := simplejson.NewJson(newConfigValue)
			Expect(err).NotTo(HaveOccurred())

			replicasValue, err := newJsonConfigValue.Get("appConfig").Get("mysql").Get("replicas").Int()
			Expect(err).NotTo(HaveOccurred())
			memoryValue, err := newJsonConfigValue.Get("appConfig").Get("mysql").Get("resources").Get("memory_request").Int()
			Expect(err).NotTo(HaveOccurred())

			Expect(replicasValue).To(Equal(2))
			Expect(memoryValue).To(Equal(2))

		})

	})
})
