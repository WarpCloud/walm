package converter

import (
	corev1 "k8s.io/api/core/v1"
	"WarpCloud/walm/pkg/models/k8s"
)

func ConvertConfigMapFromK8s(oriConfigMap *corev1.ConfigMap) (*k8s.ConfigMap, error) {
	if oriConfigMap == nil {
		return nil, nil
	}
	configMap := oriConfigMap.DeepCopy()

	return &k8s.ConfigMap{
		Meta: k8s.NewMeta(k8s.ConfigMapKind, configMap.Namespace, configMap.Name, k8s.NewState("Ready", "", "") ),
		Data:     configMap.Data,
	}, nil
}