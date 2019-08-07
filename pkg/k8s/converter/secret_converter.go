package converter

import (
	corev1 "k8s.io/api/core/v1"
	"WarpCloud/walm/pkg/models/k8s"
)

func ConvertSecretFromK8s(oriSecret *corev1.Secret) (*k8s.Secret, error) {
	if oriSecret == nil {
		return nil, nil
	}
	secret := oriSecret.DeepCopy()

	return &k8s.Secret{
		Meta: k8s.NewMeta(k8s.SecretKind, secret.Namespace, secret.Name, k8s.NewState("Ready", "", "")),
		Data:     convertDataToStringData(secret.Data),
		Type:     string(secret.Type),
	}, nil
}

func convertDataToStringData(data map[string][]byte) (stringData map[string]string) {
	stringData = map[string]string{}
	for key, value := range data {
		stringData[key] = string(value)
	}
	return
}