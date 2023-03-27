package operator

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"WarpCloud/walm/test/e2e/framework"
	"WarpCloud/walm/pkg/k8s/operator"
)

var _ = Describe("K8sOperatorNode", func() {

	var (
		k8sOperator  *operator.Operator
	)

	BeforeEach(func() {
		k8sOperator = operator.NewOperator(framework.GetK8sClient(), nil, nil, nil)
	})

	AfterEach(func() {
	})

	It("test label & annotate node", func() {
		node, err := framework.GetTestNode()
		Expect(err).NotTo(HaveOccurred())

		By("label node")
		err = k8sOperator.LabelNode(node.Name, map[string]string{"walm-test": "true"}, nil)
		Expect(err).NotTo(HaveOccurred())
		node, err = framework.GetNode(node.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(node.Labels).To(HaveKeyWithValue("walm-test", "true"))

		err = k8sOperator.LabelNode(node.Name, map[string]string{"walm-test1": "true"}, []string{"walm-test"})
		Expect(err).NotTo(HaveOccurred())
		node, err = framework.GetNode(node.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(node.Labels).To(HaveKeyWithValue("walm-test1", "true"))
		Expect(node.Labels).NotTo(HaveKey("walm-test"))

		err = k8sOperator.LabelNode(node.Name, nil, []string{"walm-test1"})
		Expect(err).NotTo(HaveOccurred())

		By("annotate node")
		err = k8sOperator.AnnotateNode(node.Name, map[string]string{"walm-test": "true"}, nil)
		Expect(err).NotTo(HaveOccurred())
		node, err = framework.GetNode(node.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(node.Annotations).To(HaveKeyWithValue("walm-test", "true"))

		err = k8sOperator.AnnotateNode(node.Name, map[string]string{"walm-test1": "true"}, []string{"walm-test"})
		Expect(err).NotTo(HaveOccurred())
		node, err = framework.GetNode(node.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(node.Annotations).To(HaveKeyWithValue("walm-test1", "true"))
		Expect(node.Annotations).NotTo(HaveKey("walm-test"))

		err = k8sOperator.AnnotateNode(node.Name, nil, []string{"walm-test1"})
		Expect(err).NotTo(HaveOccurred())
	})
})
