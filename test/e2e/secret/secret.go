package secret

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"WarpCloud/walm/pkg/k8s/adaptor"
	"encoding/base64"
	"WarpCloud/walm/test/e2e/framework"
	"time"
	"WarpCloud/walm/pkg/k8s/handler"
)

var _ = Describe("Secret", func() {

	var (
		namespace string
		err       error
		secretAdaptor   *adaptor.WalmSecretAdaptor
		getWalmSecret func(namespace, name string) (adaptor.WalmSecret, error)
	)

	BeforeEach(func() {
		By("create namespace")
		namespace, err = framework.CreateRandomNamespace("secretTest")
		Expect(err).NotTo(HaveOccurred())
		secretAdaptor = adaptor.GetDefaultAdaptorSet().GetAdaptor("Secret").(*adaptor.WalmSecretAdaptor)
		Expect(secretAdaptor).NotTo(BeNil())
		getWalmSecret = func(namespace, name string) (adaptor.WalmSecret, error) {
			time.Sleep(time.Second * 1)
			res, err := secretAdaptor.GetResource(namespace, name)
			if err != nil {
				return adaptor.WalmSecret{}, err
			}
			return res.(adaptor.WalmSecret), nil
		}
	})

	AfterEach(func() {
		By("delete namespace")
		err = framework.DeleteNamespace(namespace, true)
		Expect(err).NotTo(HaveOccurred())
	})

	It("test secret lifecycle", func() {

		testData := map[string]string{
			"test": base64.StdEncoding.EncodeToString([]byte("test1")),
		}
		walmSecret := adaptor.WalmSecret{
			Type: "Opaque",
			WalmMeta: adaptor.WalmMeta{
				Namespace: namespace,
				Name:      "test-secret",
			},
			Data: testData,
		}

		By("create secret")
		err = secretAdaptor.CreateSecret(&walmSecret)
		Expect(err).NotTo(HaveOccurred())

		walmSecret, err =  getWalmSecret(namespace, walmSecret.Name)
		expectedWalmSecret := adaptor.WalmSecret{
			Type: "Opaque",
			WalmMeta: adaptor.WalmMeta{
				Namespace: namespace,
				Name:      "test-secret",
				Kind: "Secret",
				State: adaptor.WalmState{
					Status: "Ready",
				},
			},
			Data: map[string]string{
				"test": "test1",
			},
		}
		Expect(err).NotTo(HaveOccurred())
		Expect(walmSecret).To(Equal(expectedWalmSecret))

		By("update secret")
		walmSecret.Data["test"] = base64.StdEncoding.EncodeToString([]byte("test2"))
		err = secretAdaptor.UpdateSecret(&walmSecret)
		Expect(err).NotTo(HaveOccurred())

		walmSecret, err =  getWalmSecret(namespace, walmSecret.Name)
		expectedWalmSecret.Data["test"] = "test2"
		Expect(err).NotTo(HaveOccurred())
		Expect(walmSecret).To(Equal(expectedWalmSecret))

		err = handler.GetDefaultHandlerSet().GetSecretHandler().DeleteSecret(namespace, walmSecret.Name)
		Expect(err).NotTo(HaveOccurred())

	})
})

