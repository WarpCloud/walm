package lister

import (
	"k8s.io/api/extensions/v1beta1"
	corev1 "k8s.io/api/core/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"walm/pkg/k8s/informer"
	tranv1beta1 "transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	"walm/pkg/k8s/handler"
	"k8s.io/client-go/kubernetes"
)

type K8sResourceLister struct {
	Factory informer.InformerFactory
	Client *kubernetes.Clientset
}

func (lister K8sResourceLister) GetDeployment(namespace string, name string) (*v1beta1.Deployment, error) {
	return handler.NewDeploymentHandler(lister.Client, lister.Factory.DeploymentLister).GetDeployment(namespace, name)
}

func (lister K8sResourceLister)GetPods(namespace string, labelSelector *metav1.LabelSelector) ([]*corev1.Pod, error) {
	return handler.NewPodHandler(lister.Client, lister.Factory.PodLister).ListPods(namespace, labelSelector)
}

func (lister K8sResourceLister)GetService(namespace string, name string) (*corev1.Service, error) {
	return handler.NewServiceHandler(lister.Client, lister.Factory.ServiceLister).GetService(namespace, name)
}

func (lister K8sResourceLister) GetStatefulSet(namespace string, name string) (*appsv1beta1.StatefulSet, error) {
	return handler.NewStatefulSetHandler(lister.Client, lister.Factory.StatefulSetLister).GetStatefulSet(namespace, name)
}

func (lister K8sResourceLister) GetDaemonSet(namespace string, name string) (*v1beta1.DaemonSet, error) {
	return handler.NewDaemonSetHandler(lister.Client, lister.Factory.DaemonSetLister).GetDaemonSet(namespace, name)
}

func (lister K8sResourceLister)GetJob(namespace string, name string) (*batchv1.Job, error) {
	return handler.NewJobHandler(lister.Client, lister.Factory.JobLister).GetJob(namespace, name)
}

func (lister K8sResourceLister)GetConfigMap(namespace string, name string) (*corev1.ConfigMap, error) {
	return handler.NewConfigMapHandler(lister.Client, lister.Factory.ConfigMapLister).GetConfigMap(namespace, name)
}

func (lister K8sResourceLister) GetIngress(namespace string, name string) (*v1beta1.Ingress, error) {
	return handler.NewIngressHandler(lister.Client, lister.Factory.IngressLister).GetIngress(namespace, name)
}

func (lister K8sResourceLister)GetSecret(namespace string, name string) (*corev1.Secret, error) {
	return handler.NewSecretHandler(lister.Client, lister.Factory.SecretLister).GetSecret(namespace, name)
}

// EventInformer does not support search api
func (lister K8sResourceLister)GetInstanceEvents(inst tranv1beta1.ApplicationInstance) ([]corev1.Event, error) {
	handler := handler.NewEventHandler(lister.Client)
	ref := corev1.ObjectReference{
		Namespace: inst.Namespace,
		Name: inst.Name,
		Kind: inst.Kind,
		ResourceVersion: inst.ResourceVersion,
		UID: inst.UID,
		APIVersion: inst.APIVersion,
	}
	events, err := handler.SearchEvents("txsql3", &ref)
	if err != nil {
		return nil , err
	}
	return events.Items, nil
}