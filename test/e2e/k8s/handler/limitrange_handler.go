package handler

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"walm/pkg/k8s/handler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"walm/test/e2e/framework"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func BuildTestLimitRange() *v1.LimitRange {
	return &v1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-name",
		},
		Spec: v1.LimitRangeSpec{
			Limits: []v1.LimitRangeItem{
				{
					Type: v1.LimitTypeContainer,
					Default: v1.ResourceList{
						v1.ResourceMemory: resource.MustParse("100Mi"),
					},
				},
			},
		},
	}
}

var _ = Describe("LimitRangeHandler", func() {

	var (
		namespace         string
		limitRangeHandler *handler.LimitRangeHandler
		err               error
	)

	BeforeEach(func() {
		By("create namespace")
		namespace, err = framework.CreateRandomNamespace("limitRangeHandlerTest")
		Expect(err).NotTo(HaveOccurred())
		limitRangeHandler = handler.GetDefaultHandlerSet().GetLimitRangeHandler()
		Expect(limitRangeHandler).NotTo(BeNil())
	})

	AfterEach(func() {
		By("delete namespace")
		err = framework.DeleteNamespace(namespace, false)
		Expect(err).NotTo(HaveOccurred())
	})

	It("test limit range lifecycle: create, update, delete", func() {

		By("create limitRange")
		limitRange := BuildTestLimitRange()
		_, err := limitRangeHandler.CreateLimitRange(namespace, limitRange)
		Expect(err).NotTo(HaveOccurred())

		limitRange, err = limitRangeHandler.GetLimitRangeFromK8s(namespace, limitRange.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(limitRange.Spec.Limits).To(Equal([]v1.LimitRangeItem{
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

		By("update limitRange")
		limitRange.Spec.Limits[0].Default[v1.ResourceMemory] = resource.MustParse("200Mi")
		_, err = limitRangeHandler.UpdateLimitRange(namespace, limitRange)
		Expect(err).NotTo(HaveOccurred())

		limitRange, err = limitRangeHandler.GetLimitRangeFromK8s(namespace, limitRange.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(limitRange.Spec.Limits).To(Equal([]v1.LimitRangeItem{
			{
				Type: v1.LimitTypeContainer,
				Default: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("200Mi"),
				},
				DefaultRequest: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("100Mi"),
				},
			},
		}))

		By("delete limitRange")
		err = limitRangeHandler.DeleteLimitRange(namespace, limitRange.Name)
		Expect(err).NotTo(HaveOccurred())
	})
})
