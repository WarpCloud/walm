package handler

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"walm/pkg/k8s/handler"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/api/extensions/v1beta1"
	"walm/test/e2e/framework"
)

func BuildTestNginxDeployment() *v1beta1.Deployment {
	replicas := int32(3)
	labels := map[string]string{
		"app": "nginx",
	}
	return &v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-deployment",
			Labels: labels,
		},
		Spec: v1beta1.DeploymentSpec{
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

var _ = Describe("DeploymentHandler", func() {

	var (
		namespace string
		deploymentHandler *handler.DeploymentHandler
		err error
	)

	BeforeEach(func() {
		By("create namespace")
		namespace, err = framework.CreateRandomNamespace("deploymentHandlerTest")
		Expect(err).NotTo(HaveOccurred())
		deploymentHandler = handler.GetDefaultHandlerSet().GetDeploymentHandler()
		Expect(deploymentHandler).NotTo(BeNil())
	})

	AfterEach(func() {
		By("delete namespace")
		err = framework.DeleteNamespace(namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	It("test deployment lifecycle: create, update, scale, delete", func() {

		By("create deployment")
		deployment := BuildTestNginxDeployment()
		newDeployment, err := deploymentHandler.CreateDeployment(namespace, deployment)
		Expect(err).NotTo(HaveOccurred())
		Expect(*newDeployment.Spec.Replicas).To(Equal(int32(3)))

		By("update deployment")
		newReplicas := int32(2)
		newDeployment.Spec.Replicas = &newReplicas
		newDeployment, err = deploymentHandler.UpdateDeployment(namespace, newDeployment)
		Expect(err).NotTo(HaveOccurred())
		Expect(*newDeployment.Spec.Replicas).To(Equal(int32(2)))

		By("scale deployment")
		_, err = deploymentHandler.Scale(namespace, newDeployment.Name, int32(3))
		Expect(err).NotTo(HaveOccurred())

		newDeployment,err = deploymentHandler.GetDeploymentFromK8s(namespace, newDeployment.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(*newDeployment.Spec.Replicas).To(Equal(int32(3)))

		By("delete deployment")
		err = deploymentHandler.DeleteDeployment(namespace, newDeployment.Name)
		Expect(err).NotTo(HaveOccurred())
	})
})
