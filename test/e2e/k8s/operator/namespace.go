package operator

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"WarpCloud/walm/test/e2e/framework"
	"k8s.io/api/core/v1"
	"WarpCloud/walm/pkg/k8s/operator"
	"WarpCloud/walm/pkg/models/k8s"
)

func buildTestNamespace(namespace string) *k8s.Namespace {
	return &k8s.Namespace{
		Meta: k8s.Meta{
			Namespace: namespace,
			Name:      namespace,
		},
		Labels: map[string]string{
			"master": "true",
		},
		Annotations: map[string]string{
			"test": "true",
		},
	}
}

var _ = Describe("K8sOperatorNamespace", func() {

	var (
		k8sOperator *operator.Operator
	)

	BeforeEach(func() {
		k8sOperator = operator.NewOperator(framework.GetK8sClient(), nil, nil, nil)
	})

	AfterEach(func() {
	})

	It("test namespace lifecycle: create, update, delete", func() {

		By("create namespace")
		namespace := buildTestNamespace("test-namespace")
		err := k8sOperator.CreateNamespace(namespace)
		Expect(err).NotTo(HaveOccurred())

		k8sNamespace, err := framework.GetNamespace(namespace.Name)
		Expect(err).NotTo(HaveOccurred())
		assertK8sNamespace(k8sNamespace, "test-namespace", map[string]string{
			"master": "true",
		}, map[string]string{
			"test": "true",
		}, )

		namespace.Labels["master"] = "false"
		namespace.Annotations["test1"] = "true"
		delete(namespace.Annotations, "test")
		err = k8sOperator.UpdateNamespace(namespace)
		Expect(err).NotTo(HaveOccurred())

		k8sNamespace, err = framework.GetNamespace(namespace.Name)
		Expect(err).NotTo(HaveOccurred())
		assertK8sNamespace(k8sNamespace, "test-namespace", map[string]string{
			"master": "false",
		}, map[string]string{
			"test1": "true",
		}, )

		err = k8sOperator.DeleteNamespace(namespace.Name)
		Expect(err).NotTo(HaveOccurred())

		err = k8sOperator.DeleteNamespace("not-existed")
		Expect(err).NotTo(HaveOccurred())
	})
})

func assertK8sNamespace(k8sNamespace *v1.Namespace, name string, labels map[string]string, annotations map[string]string) {
	Expect(k8sNamespace.Name).To(Equal(name))
	Expect(k8sNamespace.Labels).To(Equal(labels))
	Expect(k8sNamespace.Annotations).To(Equal(annotations))
}
