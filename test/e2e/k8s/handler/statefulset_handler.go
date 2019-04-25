package handler

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"walm/pkg/k8s/handler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"walm/test/e2e/framework"
	"k8s.io/api/apps/v1beta1"
)

func BuildTestStatefulSet() *v1beta1.StatefulSet {
	replicas := int32(3)
	labels := map[string]string{
		"app": "nginx",
	}
	return &v1beta1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-ss",
		},
		Spec: v1beta1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "nginx",
							Image: "nginx:1.7.9",
						},
					},
				},
			},
		},
	}
}

var _ = Describe("StatefulSetHandler", func() {

	var (
		namespace          string
		statefulSetHandler *handler.StatefulSetHandler
		err                error
	)

	BeforeEach(func() {
		By("create namespace")
		namespace, err = framework.CreateRandomNamespace("StatefulSetHandlerTest")
		Expect(err).NotTo(HaveOccurred())
		statefulSetHandler = handler.GetDefaultHandlerSet().GetStatefulSetHandler()
		Expect(statefulSetHandler).NotTo(BeNil())
	})

	AfterEach(func() {
		By("delete namespace")
		err = framework.DeleteNamespace(namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	It("test statefulSet lifecycle: create, scale, delete", func() {

		By("create statefulSet")
		statefulSet := BuildTestStatefulSet()
		_, err := statefulSetHandler.CreateStatefulSet(namespace, statefulSet)
		Expect(err).NotTo(HaveOccurred())
		statefulSet, err = statefulSetHandler.GetStatefulSetFromK8s(namespace, statefulSet.Name)
		Expect(*statefulSet.Spec.Replicas).To(Equal(int32(3)))

		By("scale statefulSet")
		err = statefulSetHandler.Scale(namespace, statefulSet.Name, 2)
		Expect(err).NotTo(HaveOccurred())
		statefulSet, err = statefulSetHandler.GetStatefulSetFromK8s(namespace, statefulSet.Name)
		Expect(*statefulSet.Spec.Replicas).To(Equal(int32(2)))

		By("delete statefulSet")
		err = statefulSetHandler.DeleteStatefulSet(namespace, statefulSet.Name)
		Expect(err).NotTo(HaveOccurred())
	})
})
