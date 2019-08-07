package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestConvertConfigMapFromK8s(t *testing.T) {
	tests := []struct{
		oriConfigMap   *corev1.ConfigMap
		configMap      *k8s.ConfigMap
		err            error
	}{
		{
			oriConfigMap:  &corev1.ConfigMap{
				TypeMeta:   metav1.TypeMeta{
					Kind:  "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-configMap",
					Namespace: "test-namespace",

				},
				Data: map[string]string{"test1": "test1"},
			},

			configMap:  &k8s.ConfigMap{
				Meta: k8s.Meta{
					Name: "test-configMap",
					Namespace: "test-namespace",
					Kind: "ConfigMap",
					State: k8s.State{
						Status:  "Ready",
						Reason:  "",
						Message: "",
					},
				},
				Data: map[string]string{"test1": "test1"},
			},
			err:  nil,
		},
		{
			oriConfigMap: nil,
			configMap:    nil,
			err: 		  nil,
		},
	}

	for _, test := range tests {
		configMap, err := ConvertConfigMapFromK8s(test.oriConfigMap)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.configMap, configMap)
	}
}