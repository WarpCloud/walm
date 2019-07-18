package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	listv1beta1 "k8s.io/client-go/listers/extensions/v1beta1"
)

type IngressHandler struct {
	client *kubernetes.Clientset
	lister listv1beta1.IngressLister
}

func (handler *IngressHandler) GetIngress(namespace string, name string) (*v1beta1.Ingress, error) {
	return handler.lister.Ingresses(namespace).Get(name)
}

func (handler *IngressHandler) CreateIngress(namespace string, ingress *v1beta1.Ingress) (*v1beta1.Ingress, error) {
	return handler.client.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)
}

func (handler *IngressHandler) UpdateIngress(namespace string, ingress *v1beta1.Ingress) (*v1beta1.Ingress, error) {
	return handler.client.ExtensionsV1beta1().Ingresses(namespace).Update(ingress)
}

func (handler *IngressHandler) DeleteIngress(namespace string, name string) (error) {
	return handler.client.ExtensionsV1beta1().Ingresses(namespace).Delete(name, &metav1.DeleteOptions{})
}
