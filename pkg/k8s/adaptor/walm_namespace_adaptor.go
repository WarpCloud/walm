package adaptor

import (
	corev1 "k8s.io/api/core/v1"
	"WarpCloud/walm/pkg/k8s/handler"
)

type WalmNamespaceAdaptor struct {
	handler *handler.NamespaceHandler
}

func (adaptor *WalmNamespaceAdaptor) GetResource(namespace string, name string) (WalmResource, error) {
	return nil, nil
}

func BuildWalmNamespace(resourceQuota *corev1.ResourceQuota) *WalmResourceQuota {
	return nil
}

func BuildNamespace(walmResourceQuota *WalmResourceQuota) (*corev1.ResourceQuota, error) {
	return nil, nil
}
