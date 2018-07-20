package walmlister

import (
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	corev1 "k8s.io/api/core/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
)


type K8sResourceLister interface {
	GetDeployment(namespace string, name string) (*extv1beta1.Deployment, error)
	GetPods(namespace string, labelSelectorStr string) (*corev1.PodList, error)
	GetService(namespace string, name string) (*corev1.Service, error)
	GetStatefulSet(namespace string, name string) (*appsv1beta1.StatefulSet, error)
	GetDaemonSet(namespace string, name string) (*extv1beta1.DaemonSet, error)
	GetJob(namespace string, name string) (*batchv1.Job, error)
	GetConfigMap(namespace string, name string) (*corev1.ConfigMap, error)
	GetIngress(namespace string, name string) (*extv1beta1.Ingress, error)
	GetSecret(namespace string, name string) (*corev1.Secret, error)
}