package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sutils "walm/pkg/k8s/utils"
	listv1 "k8s.io/client-go/listers/core/v1"
)

type NamespaceHandler struct {
	client *kubernetes.Clientset
	lister listv1.NamespaceLister
}

func (handler *NamespaceHandler) GetNamespace(name string) (*v1.Namespace, error) {
	return handler.lister.Get(name)
}

func (handler *NamespaceHandler) ListNamespaces(labelSelector *metav1.LabelSelector) ([]*v1.Namespace, error) {
	selector, err := k8sutils.ConvertLabelSelectorToSelector(labelSelector)
	if err != nil {
		return nil, err
	}
	return handler.lister.List(selector)
}

func (handler *NamespaceHandler) CreateNamespace(Namespace *v1.Namespace) (*v1.Namespace, error) {
	return handler.client.CoreV1().Namespaces().Create(Namespace)
}

func (handler *NamespaceHandler) UpdateNamespace(Namespace *v1.Namespace) (*v1.Namespace, error) {
	return handler.client.CoreV1().Namespaces().Update(Namespace)
}

func (handler *NamespaceHandler) DeleteNamespace(name string) (error) {
	return handler.client.CoreV1().Namespaces().Delete(name, &metav1.DeleteOptions{})
}

