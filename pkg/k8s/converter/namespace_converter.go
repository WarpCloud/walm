package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ConvertNamespaceToK8s(namespace *k8s.Namespace) (*v1.Namespace, error) {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace.Namespace,
			Name: namespace.Name,
			Annotations: namespace.Annotations,
			Labels: namespace.Labels,
		},
	}, nil
}
