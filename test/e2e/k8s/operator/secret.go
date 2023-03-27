package operator

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"WarpCloud/walm/test/e2e/framework"
	"WarpCloud/walm/pkg/k8s/operator"
	"WarpCloud/walm/pkg/models/k8s"
	"encoding/base64"
)

func buildTestSecret() *k8s.CreateSecretRequestBody {
	return &k8s.CreateSecretRequestBody{
		Name: "test-name",
		Type: "Opaque",
		Data: map[string]string{"test": base64.StdEncoding.EncodeToString([]byte("test1"))},
	}
}

var _ = Describe("K8sOperatorSecret", func() {

	var (
		namespace   string
		k8sOperator *operator.Operator
		err         error
	)

	BeforeEach(func() {
		By("create namespace")
		namespace, err = framework.CreateRandomNamespace("k8sOperatorSecretTest", nil)
		Expect(err).NotTo(HaveOccurred())
		k8sOperator = operator.NewOperator(framework.GetK8sClient(), nil, nil, nil)
	})

	AfterEach(func() {
		By("delete namespace")
		err = framework.DeleteNamespace(namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	It("test secret: create & update & delete", func() {

		By("create secret")
		secretRequestBody := buildTestSecret()
		err := k8sOperator.CreateSecret(namespace, secretRequestBody)
		Expect(err).NotTo(HaveOccurred())

		k8sSecret, err := framework.GetSecret(namespace, secretRequestBody.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sSecret.Data).To(Equal(map[string][]byte{"test": []byte("test1")}))

		By("update secret")
		secretRequestBody.Data["test"] = base64.StdEncoding.EncodeToString([]byte("test2"))
		err = k8sOperator.UpdateSecret(namespace, secretRequestBody)
		Expect(err).NotTo(HaveOccurred())

		k8sSecret, err = framework.GetSecret(namespace, secretRequestBody.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sSecret.Data).To(Equal(map[string][]byte{"test": []byte("test2")}))

		By("delete secret")
		err = k8sOperator.DeleteSecret(namespace, secretRequestBody.Name)
		Expect(err).NotTo(HaveOccurred())

		err = k8sOperator.DeleteSecret(namespace, "not-existed")
		Expect(err).NotTo(HaveOccurred())
	})
})
