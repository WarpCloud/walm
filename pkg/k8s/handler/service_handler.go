package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	listv1 "k8s.io/client-go/listers/core/v1"
)

type ServiceHandler struct {
	client *kubernetes.Clientset
	lister listv1.ServiceLister
}

func (handler ServiceHandler) GetService(namespace string, name string) (*v1.Service, error) {
	return handler.lister.Services(namespace).Get(name)
}

func (handler ServiceHandler) CreateService(namespace string, service *v1.Service) (*v1.Service, error) {
	return handler.client.CoreV1().Services(namespace).Create(service)
}

func (handler ServiceHandler) UpdateService(namespace string, service *v1.Service) (*v1.Service, error) {
	return handler.client.CoreV1().Services(namespace).Update(service)
}

func (handler ServiceHandler) DeleteService(namespace string, name string) (error) {
	return handler.client.CoreV1().Services(namespace).Delete(name, &metav1.DeleteOptions{})
}

func NewServiceHandler(client *kubernetes.Clientset, lister listv1.ServiceLister) (ServiceHandler) {
	return ServiceHandler{client: client, lister: lister}
}