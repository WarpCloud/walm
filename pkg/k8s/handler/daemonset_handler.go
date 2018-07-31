package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	listv1beta1 "k8s.io/client-go/listers/extensions/v1beta1"
)

type DaemonSetHandler struct {
	client *kubernetes.Clientset
	lister listv1beta1.DaemonSetLister
}

func (handler DaemonSetHandler) GetDaemonSet(namespace string, name string) (*v1beta1.DaemonSet, error) {
	return handler.lister.DaemonSets(namespace).Get(name)
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

func NewDaemonSetHandler(client *kubernetes.Clientset, lister listv1beta1.DaemonSetLister) (DaemonSetHandler) {
	return DaemonSetHandler{client: client, lister: lister}
}