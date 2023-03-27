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

func buildTestResourceQuota(namespace string) *k8s.ResourceQuota {
	return &k8s.ResourceQuota{
		Meta: k8s.Meta{
			Namespace: namespace,
			Name:      "test-name",
		},
		ResourceLimits: map[k8s.ResourceName]string{
			k8s.ResourceRequestsMemory: "100Mi",
		},
	}
}

var _ = Describe("K8sOperatorResourceQuota", func() {

	var (
		namespace   string
		k8sOperator *operator.Operator
		err         error
	)

	BeforeEach(func() {
		By("create namespace")
		namespace, err = framework.CreateRandomNamespace("k8sOperatorResourceQuotaTest", nil)
		Expect(err).NotTo(HaveOccurred())
		k8sOperator = operator.NewOperator(framework.GetK8sClient(), nil, nil, nil)
	})

	AfterEach(func() {
		By("delete namespace")
		err = framework.DeleteNamespace(namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	It("test resource quota: create & create or update", func() {

		By("create resource quota")
		resourceQuota := buildTestResourceQuota(namespace)
		err := k8sOperator.CreateResourceQuota(resourceQuota)
		Expect(err).NotTo(HaveOccurred())

		k8sResourceQuota, err := framework.GetResourceQuota(namespace, resourceQuota.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sResourceQuota.Spec.Hard).To(Equal(v1.ResourceList{
			v1.ResourceRequestsMemory: resource.MustParse("100Mi"),
		}))

		By("create or update resource quota")
		resourceQuota.Name = "test-name1"
		err = k8sOperator.CreateOrUpdateResourceQuota(resourceQuota)
		Expect(err).NotTo(HaveOccurred())

		k8sResourceQuota, err = framework.GetResourceQuota(namespace, resourceQuota.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sResourceQuota.Spec.Hard).To(Equal(v1.ResourceList{
			v1.ResourceRequestsMemory: resource.MustParse("100Mi"),
		}))

		By("create or update resource quota")
		resourceQuota.ResourceLimits[k8s.ResourceRequestsMemory] = "200Mi"
		err = k8sOperator.CreateOrUpdateResourceQuota(resourceQuota)
		Expect(err).NotTo(HaveOccurred())

		k8sResourceQuota, err = framework.GetResourceQuota(namespace, resourceQuota.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sResourceQuota.Spec.Hard).To(Equal(v1.ResourceList{
			v1.ResourceRequestsMemory: resource.MustParse("200Mi"),
		}))
	})
})
