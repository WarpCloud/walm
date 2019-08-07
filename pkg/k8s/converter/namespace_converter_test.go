package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestConvertNamespaceToK8s(t *testing.T) {
	tests := []struct{
		oriNamespace	*k8s.Namespace
		namespace       *v1.Namespace
		err             error
	}{
		{
			oriNamespace: &k8s.Namespace{
				Meta: k8s.Meta{
					Name: "test1",
					Namespace: "test-namespace1",
				},
			},
			namespace: &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test1",
					Namespace: "test-namespace1",
				},
			},
			err: nil,
		},
		{
			oriNamespace: &k8s.Namespace{},
			namespace: &v1.Namespace{},
			err: nil,
		},
	}

	for _, test := range tests {
		namespace, err := ConvertNamespaceToK8s(test.oriNamespace)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.namespace, namespace)
	}
}