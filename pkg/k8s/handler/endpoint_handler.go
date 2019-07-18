package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	listv1 "k8s.io/client-go/listers/core/v1"
	k8sutils "WarpCloud/walm/pkg/k8s/utils"
)

type EndpointsHandler struct {
	client *kubernetes.Clientset
	lister listv1.EndpointsLister
}

func (handler *EndpointsHandler) GetEndpoints(namespace string, name string) (*v1.Endpoints, error) {
	return handler.lister.Endpoints(namespace).Get(name)
}

func (handler *EndpointsHandler) ListEndpointss(namespace string, labelSelector *metav1.LabelSelector) ([]*v1.Endpoints, error) {
	selector, err := k8sutils.ConvertLabelSelectorToSelector(labelSelector)
	if err != nil {
		return nil, err
	}
	return handler.lister.Endpoints(namespace).List(selector)
}

func (handler *EndpointsHandler) CreateEndpoints(namespace string, endpoints *v1.Endpoints) (*v1.Endpoints, error) {
	return handler.client.CoreV1().Endpoints(namespace).Create(endpoints)
}

func (handler *EndpointsHandler) UpdateEndpoints(namespace string, endpoints *v1.Endpoints) (*v1.Endpoints, error) {
	return handler.client.CoreV1().Endpoints(namespace).Update(endpoints)
}

func (handler *EndpointsHandler) DeleteEndpoints(namespace string, name string) (error) {
	return handler.client.CoreV1().Endpoints(namespace).Delete(name, &metav1.DeleteOptions{})
}
