package release

import (
	"encoding/json"
	"go/build"
	"io/ioutil"
	"os"
	"walm/pkg/k8s/handler"
	"walm/pkg/release/v2"
	helmv2 "walm/pkg/release/v2/helm"

	"github.com/bitly/go-simplejson"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Release", func() {

	var (
		namespace      string
		gopath         string
		releaseName    string
		releaseRequest v2.ReleaseRequestV2
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
		releaseChartByte, err := ioutil.ReadFile(gopath + "/src/walm/test/resources/simpleTest/TXSQL/release.yaml")
		Expect(err).NotTo(HaveOccurred())

		err = json.Unmarshal(releaseChartByte, &releaseRequest)
		Expect(err).NotTo(HaveOccurred())

		releaseRequest.Name = releaseRequest.Name + "-" + randomId[:8]
		releaseName = releaseRequest.Name
		helmv2.GetDefaultHelmClientV2().InstallUpgradeReleaseV2(namespace, &releaseRequest, false, nil)

		releaseInfo, err := helmv2.GetDefaultHelmClientV2().GetReleaseV2(namespace, releaseName)
		Expect(releaseInfo.Name).To(Equal(releaseName))
	})

	AfterEach(func() {

		By("delete release")
		err := helmv2.GetDefaultHelmClientV2().DeleteRelease(namespace, releaseName, false, true)
		Expect(err).NotTo(HaveOccurred())

		_, err = helmv2.GetDefaultHelmClientV2().GetReleaseV2(namespace, releaseName)
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

			jsonConfigValue.Get("App").Get("txsql").Set("replicas", 2)
			jsonConfigValue.Get("App").Get("txsql").Get("resources").Set("memory_request", 2)

			ConfigValue, err = jsonConfigValue.MarshalJSON()
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal(ConfigValue, &releaseRequest.ConfigValues)
			Expect(err).NotTo(HaveOccurred())

			helmv2.GetDefaultHelmClientV2().InstallUpgradeReleaseV2(namespace, &releaseRequest, false, nil)

			By("validate release value")

			newRelease, err := helmv2.GetDefaultHelmClientV2().GetReleaseV2(namespace, releaseName)
			Expect(err).NotTo(HaveOccurred())
			newConfigValue, err := json.Marshal(newRelease.ConfigValues)
			Expect(err).NotTo(HaveOccurred())
			newJsonConfigValue, err := simplejson.NewJson(newConfigValue)
			Expect(err).NotTo(HaveOccurred())

			replicasValue, err := newJsonConfigValue.Get("App").Get("txsql").Get("replicas").Int()
			Expect(err).NotTo(HaveOccurred())
			memoryValue, err := newJsonConfigValue.Get("App").Get("txsql").Get("resources").Get("memory_request").Int()
			Expect(err).NotTo(HaveOccurred())

			Expect(replicasValue).To(Equal(2))
			Expect(memoryValue).To(Equal(2))

		})

	})
})
