package framework

import (
	"strings"
	"walm/pkg/k8s/handler"
	"fmt"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
)

const (
	maxNameLength                = 62
	randomLength                 = 5
	maxGeneratedRandomNameLength = maxNameLength - randomLength
)

func GenerateRandomName(base string) string {
	if len(base) > maxGeneratedRandomNameLength {
		base = base[:maxGeneratedRandomNameLength]
	}
	return fmt.Sprintf("%s-%s", strings.ToLower(base), utilrand.String(randomLength))
}

func CreateRandomNamespace(base string) (string, error) {
	namespace := GenerateRandomName(base)
	ns := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := handler.GetDefaultHandlerSet().GetNamespaceHandler().CreateNamespace(&ns)
	return namespace, err
}

func DeleteNamespace(namespace string) (error) {
	return handler.GetDefaultHandlerSet().GetNamespaceHandler().DeleteNamespace(namespace)
}
