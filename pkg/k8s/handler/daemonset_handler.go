package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DaemonSetHandler struct {
	client *kubernetes.Clientset
}

func (handler DaemonSetHandler) GetDaemonSet(namespace string, name string) (*v1beta1.DaemonSet, error) {
	return handler.client.ExtensionsV1beta1().DaemonSets(namespace).Get(name, metav1.GetOptions{})
}

func (handler DaemonSetHandler) CreateDaemonSet(namespace string, daemonSet *v1beta1.DaemonSet) (*v1beta1.DaemonSet, error) {
	return handler.client.ExtensionsV1beta1().DaemonSets(namespace).Create(daemonSet)
}

func (handler DaemonSetHandler) UpdateDaemonSet(namespace string, daemonSet *v1beta1.DaemonSet) (*v1beta1.DaemonSet, error) {
	return handler.client.ExtensionsV1beta1().DaemonSets(namespace).Update(daemonSet)
}

func (handler DaemonSetHandler) DeleteDaemonSet(namespace string, name string) (error) {
	return handler.client.ExtensionsV1beta1().DaemonSets(namespace).Delete(name, &metav1.DeleteOptions{})
}

func NewDaemonSetHandler(client *kubernetes.Clientset) (DaemonSetHandler) {
	return DaemonSetHandler{client: client}
}