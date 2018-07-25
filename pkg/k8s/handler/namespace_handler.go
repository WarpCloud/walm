package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sutils "walm/pkg/k8s/utils"
)

type NamespaceHandler struct {
	client *kubernetes.Clientset
}

func (handler NamespaceHandler) GetNamespace(name string) (*v1.Namespace, error) {
	return handler.client.CoreV1().Namespaces().Get(name, metav1.GetOptions{})
}

func (handler NamespaceHandler) ListNamespaces(labelSelector *metav1.LabelSelector) (*v1.NamespaceList, error) {
	selectorStr, err := k8sutils.ConvertLabelSelectorToStr(labelSelector)
	if err != nil {
		return nil, err
	}
	return handler.client.CoreV1().Namespaces().List(metav1.ListOptions{LabelSelector:selectorStr})
}

func (handler NamespaceHandler) CreateNamespace(Namespace *v1.Namespace) (*v1.Namespace, error) {
	return handler.client.CoreV1().Namespaces().Create(Namespace)
}

func (handler NamespaceHandler) UpdateNamespace(Namespace *v1.Namespace) (*v1.Namespace, error) {
	return handler.client.CoreV1().Namespaces().Update(Namespace)
}

func (handler NamespaceHandler) DeleteNamespace(name string) (error) {
	return handler.client.CoreV1().Namespaces().Delete(name, &metav1.DeleteOptions{})
}

func NewNamespaceHandler(client *kubernetes.Clientset) (NamespaceHandler) {
	return NamespaceHandler{client:client}
}
