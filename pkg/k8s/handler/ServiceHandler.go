package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceHandler struct {
	client *kubernetes.Clientset
}

func (handler ServiceHandler) GetService(namespace string, name string) (*v1.Service, error) {
	return handler.client.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
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

func NewServiceHandler(client *kubernetes.Clientset) (ServiceHandler) {
	return ServiceHandler{client: client}
}