package operator

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"WarpCloud/walm/test/e2e/framework"
	"k8s.io/api/core/v1"
	"WarpCloud/walm/pkg/k8s/operator"
)

var _ = Describe("K8sOperatorPod", func() {

	var (
		namespace string
		k8sOperator  *operator.Operator
		err       error
		pod *v1.Pod
	)

	BeforeEach(func() {
		By("create namespace")
		namespace, err = framework.CreateRandomNamespace("k8sOperatorPodTest", nil)
		Expect(err).NotTo(HaveOccurred())
		k8sOperator = operator.NewOperator(framework.GetK8sClient(), nil, nil, nil)
		pod, err = framework.CreatePod(namespace, "test-pod")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		By("delete namespace")
		err = framework.DeleteNamespace(namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	It("test delete pod", func() {

		By(" delete pod")
		err = k8sOperator.DeletePod(namespace, pod.Name)
		Expect(err).NotTo(HaveOccurred())

		By( "delete pod which does not exist")
		err = k8sOperator.DeletePod(namespace, "not-existed")
		Expect(err).NotTo(HaveOccurred())
	})

	It("test restart pod", func() {

		By(" restart pod")
		err = k8sOperator.RestartPod(namespace, pod.Name)
		Expect(err).NotTo(HaveOccurred())

		By( "restart pod which does not exist")
		err = k8sOperator.RestartPod(namespace,"not-existed")
		Expect(err).NotTo(BeNil())
	})
})
