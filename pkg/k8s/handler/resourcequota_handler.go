package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	listv1 "k8s.io/client-go/listers/core/v1"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
)

type ResourceQuotaHandler struct {
	client *kubernetes.Clientset
	lister listv1.ResourceQuotaLister
}

func (handler *ResourceQuotaHandler) GetResourceQuota(namespace string, name string) (*v1.ResourceQuota, error) {
	return handler.lister.ResourceQuotas(namespace).Get(name)
}

func (handler *ResourceQuotaHandler) CreateResourceQuota(namespace string, resourceQuota *v1.ResourceQuota) (*v1.ResourceQuota, error) {
	return handler.client.CoreV1().ResourceQuotas(namespace).Create(resourceQuota)
}

func (handler *ResourceQuotaHandler) UpdateResourceQuota(namespace string, resourceQuota *v1.ResourceQuota) (*v1.ResourceQuota, error) {
	return handler.client.CoreV1().ResourceQuotas(namespace).Update(resourceQuota)
}

func (handler *ResourceQuotaHandler) DeleteResourceQuota(namespace string, name string) (error) {
	return handler.client.CoreV1().ResourceQuotas(namespace).Delete(name, &metav1.DeleteOptions{})
}

type ResourceQuotaBuilder struct {
	name        string
	namespace   string
	labels      map[string]string
	annotations map[string]string
	hard        v1.ResourceList
	scopes      []v1.ResourceQuotaScope
}

func (builder *ResourceQuotaBuilder) Namespace(namespace string) *ResourceQuotaBuilder {
	builder.namespace = namespace
	return builder
}

func (builder *ResourceQuotaBuilder) Name(name string) *ResourceQuotaBuilder {
	builder.name = name
	return builder
}

func (builder *ResourceQuotaBuilder) AddLabel(key string, value string) *ResourceQuotaBuilder {
	if builder.labels == nil {
		builder.labels = map[string]string{}
	}
	builder.labels[key] = value
	return builder
}

func (builder *ResourceQuotaBuilder) AddAnnotations(key string, value string) *ResourceQuotaBuilder {
	if builder.annotations == nil {
		builder.annotations = map[string]string{}
	}
	builder.annotations[key] = value
	return builder
}

func (builder *ResourceQuotaBuilder) AddHardResourceLimit(resourceName v1.ResourceName, quantity k8sresource.Quantity) *ResourceQuotaBuilder {
	if builder.hard == nil {
		builder.hard = map[v1.ResourceName]k8sresource.Quantity{}
	}
	builder.hard[resourceName] = quantity
	return builder
}

func (builder *ResourceQuotaBuilder) AddResourceQuotaScope(scope v1.ResourceQuotaScope) *ResourceQuotaBuilder {
	if builder.scopes == nil {
		builder.scopes = []v1.ResourceQuotaScope{}
	}
	builder.scopes = append(builder.scopes, scope)
	return builder
}

func (builder *ResourceQuotaBuilder) Build() *v1.ResourceQuota {
	meta := metav1.ObjectMeta{
		Namespace:   builder.namespace,
		Name:        builder.name,
		Labels:      builder.labels,
		Annotations: builder.annotations,
	}

	spec := v1.ResourceQuotaSpec{
		Hard:   builder.hard,
		Scopes: builder.scopes,
	}

	return &v1.ResourceQuota{
		ObjectMeta: meta,
		Spec:       spec,
	}
}

func NewFakeResourceQuotaHandler(client *kubernetes.Clientset,
	lister listv1.ResourceQuotaLister) *ResourceQuotaHandler {
	return &ResourceQuotaHandler{client, lister}
}
