package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"github.com/sirupsen/logrus"
)

func ConvertResourceQuotaToK8s(quota *k8s.ResourceQuota) (*v1.ResourceQuota, error) {
	k8sQuota := &v1.ResourceQuota{
		ObjectMeta : metav1.ObjectMeta{
			Namespace: quota.Namespace,
			Name: quota.Name,
		},
	}
	k8sQuota.Spec.Hard = v1.ResourceList{}
	for key, value := range quota.ResourceLimits {
		limit, err := resource.ParseQuantity(value)
		if err != nil {
			logrus.Errorf("failed to parse quantity : %s", err.Error())
			return nil, err
		}
		k8sQuota.Spec.Hard[v1.ResourceName(key)] = limit
	}
	return k8sQuota, nil
}

func ConvertResourceQuotaFromK8s(quota *v1.ResourceQuota) (*k8s.ResourceQuota, error) {
	return &k8s.ResourceQuota{
		Meta:       k8s.NewMeta(k8s.ResourceQuotaKind, quota.Namespace, quota.Name, k8s.NewState("Ready", "", "")),
		ResourceLimits: buildResourceLimits(quota),
		ResourceUsed:   buildResourceUsed(quota),
	}, nil
}

func buildResourceLimits(quota *v1.ResourceQuota) map[k8s.ResourceName]string {
	limits := map[k8s.ResourceName]string{}
	for key, value := range quota.Spec.Hard {
		limits[k8s.ResourceName(key)] = value.String()
	}
	return limits
}

func buildResourceUsed(quota *v1.ResourceQuota) map[k8s.ResourceName]string {
	limits := map[k8s.ResourceName]string{}
	for key, value := range quota.Status.Used {
		limits[k8s.ResourceName(key)] = value.String()
	}
	return limits
}