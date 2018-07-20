package walmlister

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/extensions/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
)

type K8sClientLister struct {
	Client *kubernetes.Clientset
}

func (lister K8sClientLister) GetDeployment(namespace string, name string) (*v1beta1.Deployment, error) {
	return lister.Client.ExtensionsV1beta1().Deployments(namespace).Get(name, metav1.GetOptions{})
}

func (lister K8sClientLister)GetPods(namespace string, labelSelectorStr string) (*corev1.PodList, error) {
	return lister.Client.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: labelSelectorStr})
}

func (lister K8sClientLister)GetService(namespace string, name string) (*corev1.Service, error) {
	return lister.Client.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
}

func (lister K8sClientLister) GetStatefulSet(namespace string, name string) (*appsv1beta1.StatefulSet, error) {
	return lister.Client.AppsV1beta1().StatefulSets(namespace).Get(name, metav1.GetOptions{})
}

func (lister K8sClientLister) GetDaemonSet(namespace string, name string) (*v1beta1.DaemonSet, error) {
	return lister.Client.ExtensionsV1beta1().DaemonSets(namespace).Get(name, metav1.GetOptions{})
}

func (lister K8sClientLister)GetJob(namespace string, name string) (*batchv1.Job, error) {
	return lister.Client.BatchV1().Jobs(namespace).Get(name, metav1.GetOptions{})
}

func (lister K8sClientLister)GetConfigMap(namespace string, name string) (*corev1.ConfigMap, error) {
	return lister.Client.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
}

func (lister K8sClientLister) GetIngress(namespace string, name string) (*v1beta1.Ingress, error) {
	return lister.Client.ExtensionsV1beta1().Ingresses(namespace).Get(name, metav1.GetOptions{})
}

func (lister K8sClientLister)GetSecret(namespace string, name string) (*corev1.Secret, error) {
	return lister.Client.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
}