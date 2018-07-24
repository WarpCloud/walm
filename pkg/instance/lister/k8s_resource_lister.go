package lister

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/extensions/v1beta1"
	corev1 "k8s.io/api/core/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	"walm/pkg/k8s/handler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sutils "walm/pkg/k8s/utils"
)

type K8sResourceLister struct {
	Client *kubernetes.Clientset
}

func (lister K8sResourceLister) GetDeployment(namespace string, name string) (*v1beta1.Deployment, error) {
	return handler.NewDeploymentHandler(lister.Client).GetDeployment(namespace, name)
}

func (lister K8sResourceLister)GetPods(namespace string, labelSelector *metav1.LabelSelector) (*corev1.PodList, error) {
	selectorStr, err := k8sutils.ConvertLabelSelectorToStr(labelSelector)
	if err != nil {
		return nil, err
	}
	return handler.NewPodHandler(lister.Client).ListPods(namespace, selectorStr)
}

func (lister K8sResourceLister)GetService(namespace string, name string) (*corev1.Service, error) {
	return handler.NewServiceHandler(lister.Client).GetService(namespace, name)
}

func (lister K8sResourceLister) GetStatefulSet(namespace string, name string) (*appsv1beta1.StatefulSet, error) {
	return handler.NewStatefulSetHandler(lister.Client).GetStatefulSet(namespace, name)
}

func (lister K8sResourceLister) GetDaemonSet(namespace string, name string) (*v1beta1.DaemonSet, error) {
	return handler.NewDaemonSetHandler(lister.Client).GetDaemonSet(namespace, name)
}

func (lister K8sResourceLister)GetJob(namespace string, name string) (*batchv1.Job, error) {
	return handler.NewJobHandler(lister.Client).GetJob(namespace, name)
}

func (lister K8sResourceLister)GetConfigMap(namespace string, name string) (*corev1.ConfigMap, error) {
	return handler.NewConfigMapHandler(lister.Client).GetConfigMap(namespace, name)
}

func (lister K8sResourceLister) GetIngress(namespace string, name string) (*v1beta1.Ingress, error) {
	return handler.NewIngressHandler(lister.Client).GetIngress(namespace, name)
}

func (lister K8sResourceLister)GetSecret(namespace string, name string) (*corev1.Secret, error) {
	return handler.NewSecretHandler(lister.Client).GetSecret(namespace, name)
}