package handler

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"walm/pkg/k8s/handler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"walm/test/e2e/framework"
	"transwarp/release-config/pkg/apis/transwarp/v1beta1"
)

func BuildTestReleaseConfig() *v1beta1.ReleaseConfig {
	return &v1beta1.ReleaseConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
		},
	}
}

var _ = Describe("ReleaseConfigHandler", func() {

	var (
		namespace            string
		releaseConfigHandler *handler.ReleaseConfigHandler
		err                  error
	)

	BeforeEach(func() {
		By("create namespace")
		namespace, err = framework.CreateRandomNamespace("releaseConfigHandlerTest")
		Expect(err).NotTo(HaveOccurred())
		releaseConfigHandler = handler.GetDefaultHandlerSet().GetReleaseConfigHandler()
		Expect(releaseConfigHandler).NotTo(BeNil())
	})

	AfterEach(func() {
		By("delete namespace")
		err = framework.DeleteNamespace(namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	It("test release config lifecycle: create, annotate, delete", func() {

		By("create releaseConfig")
		releaseConfig := BuildTestReleaseConfig()
		_, err := releaseConfigHandler.CreateReleaseConfig(namespace, releaseConfig)
		Expect(err).NotTo(HaveOccurred())

		By("annotate releaseConfig")

		err = releaseConfigHandler.AnnotateReleaseConfig(namespace, releaseConfig.Name,
			map[string]string{"test_key1": "test_value1", "test_key2": "test_value2"}, nil)
		Expect(err).NotTo(HaveOccurred())

		releaseConfig, err = releaseConfigHandler.GetReleaseConfigFromK8s(namespace, releaseConfig.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(releaseConfig.Annotations).To(Equal(map[string]string{"test_key1": "test_value1", "test_key2": "test_value2"}))

		err = releaseConfigHandler.AnnotateReleaseConfig(namespace, releaseConfig.Name,
			nil, []string{"test_key1"})
		Expect(err).NotTo(HaveOccurred())

		releaseConfig, err = releaseConfigHandler.GetReleaseConfigFromK8s(namespace, releaseConfig.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(releaseConfig.Annotations).To(Equal(map[string]string{"test_key2": "test_value2"}))

		err = releaseConfigHandler.AnnotateReleaseConfig(namespace, releaseConfig.Name,
			map[string]string{"test_key1": "test_value1", "test_key2": "test_value22"}, nil)
		Expect(err).NotTo(HaveOccurred())

		releaseConfig, err = releaseConfigHandler.GetReleaseConfigFromK8s(namespace, releaseConfig.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(releaseConfig.Annotations).To(Equal(map[string]string{"test_key2": "test_value22", "test_key1": "test_value1"}))


		By("delete releaseConfig")
		err = releaseConfigHandler.DeleteReleaseConfig(namespace, releaseConfig.Name)
		Expect(err).NotTo(HaveOccurred())
	})
})
