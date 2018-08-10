package adaptor

import (
	corev1 "k8s.io/api/core/v1"
	"walm/pkg/k8s/handler"
	"k8s.io/apimachinery/pkg/api/resource"
)

type WalmResourceQuotaAdaptor struct {
	handler *handler.ResourceQuotaHandler
}

func (adaptor *WalmResourceQuotaAdaptor) GetResource(namespace string, name string) (WalmResource, error) {
	resourceQuota, err := adaptor.handler.GetResourceQuota(namespace, name)
	if err != nil {
		if isNotFoundErr(err) {
			return WalmResourceQuota{
				WalmMeta: buildNotFoundWalmMeta("ResourceQuota", namespace, name),
			}, nil
		}
		return WalmResourceQuota{}, err
	}

	return BuildWalmResourceQuota(resourceQuota), nil
}

func BuildWalmResourceQuota(resourceQuota *corev1.ResourceQuota) *WalmResourceQuota {
	walmResourceQuota := WalmResourceQuota{
		WalmMeta:       buildWalmMeta("ResourceQuota", resourceQuota.Namespace, resourceQuota.Name, buildWalmState("Ready", "", "")),
		ResourceLimits: buildResourceLimits(resourceQuota),
	}
	return &walmResourceQuota
}

func buildResourceLimits(quota *corev1.ResourceQuota) map[corev1.ResourceName]string {
	limits := map[corev1.ResourceName]string{}
	for key, value := range quota.Spec.Hard {
		limits[key] = value.String()
	}
	return limits
}

func BuildResourceQuota(walmResourceQuota *WalmResourceQuota) (*corev1.ResourceQuota, error) {
	builder := (&handler.ResourceQuotaBuilder{}).Namespace(walmResourceQuota.Namespace).Name(walmResourceQuota.Name)
	for key, value := range walmResourceQuota.ResourceLimits {
		limit, err := resource.ParseQuantity(value)
		if err != nil {
			return nil, err
		}
		builder.AddHardResourceLimit(key, limit)
	}
	return builder.Build(), nil
}
