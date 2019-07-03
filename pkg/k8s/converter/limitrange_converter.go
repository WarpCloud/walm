package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func ConvertLimitRangeToK8s(limitRange *k8s.LimitRange) (*v1.LimitRange, error) {
	resourceList := v1.ResourceList{}
	for key, value := range limitRange.DefaultLimit {
		resourceList[v1.ResourceName(key)] = resource.MustParse(value)
	}

	return &v1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: limitRange.Namespace,
			Name: limitRange.Name,
		},
		Spec: v1.LimitRangeSpec{
			Limits: []v1.LimitRangeItem{
				{
					Type: v1.LimitTypeContainer,
					Default: resourceList,
				},
			},
		},
	}, nil
}
