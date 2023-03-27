package operator

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"WarpCloud/walm/test/e2e/framework"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"WarpCloud/walm/pkg/k8s/operator"
	"WarpCloud/walm/pkg/models/k8s"
)

func buildTestLimitRange(namespace string) *k8s.LimitRange {
	return &k8s.LimitRange{
		Meta: k8s.Meta{
			Namespace: namespace,
			Name: "test-name",
		},
		DefaultLimit: map[k8s.ResourceName]string{
			k8s.ResourceMemory : "100Mi",
		},
	}
}

var _ = Describe("K8sOperatorLimitRange", func() {

	var (
		namespace string
		k8sOperator  *operator.Operator
		err       error
	)

	BeforeEach(func() {
		By("create namespace")
		namespace, err = framework.CreateRandomNamespace("k8sOperatorLimitRangeTest", nil)
		Expect(err).NotTo(HaveOccurred())
		k8sOperator = operator.NewOperator(framework.GetK8sClient(), nil, nil, nil)
	})

	AfterEach(func() {
		By("delete namespace")
		err = framework.DeleteNamespace(namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	It("test create limit range", func() {

		By("create limitRange")
		limitRange := buildTestLimitRange(namespace)
		err := k8sOperator.CreateLimitRange(limitRange)
		Expect(err).NotTo(HaveOccurred())

		k8sLimitRange, err := framework.GetLimitRange(namespace, limitRange.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sLimitRange.Spec.Limits).To(Equal([]v1.LimitRangeItem{
			{
				Type: v1.LimitTypeContainer,
				Default: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("100Mi"),
				},
				DefaultRequest: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("100Mi"),
				},
			},
		}))
	})
})
