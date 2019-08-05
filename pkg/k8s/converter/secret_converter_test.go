package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)


func TestConvertSecretFromK8s(t *testing.T) {
	tests := []struct{
		oriSecret *corev1.Secret
		secret    *k8s.Secret
		err       error
	}{
		{
			oriSecret: &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-secret",
					Namespace: "test-namespace",
				},
				Type: "Opaque",
				Data: map[string][]byte{"username": []byte("admin"), "password": []byte("1f2d1e2e67df")},
			},
			secret:   &k8s.Secret{
				Meta: k8s.Meta{
					Name: "test-secret",
					Namespace: "test-namespace",
					Kind:      "Secret",
					State:     k8s.State{
						Status:  "Ready",
						Reason:  "",
						Message: "",
					},
				},
				Data: map[string]string{"username": "admin", "password": "1f2d1e2e67df"},
				Type: "Opaque",
			},
			err: nil,
		},
		{
			oriSecret: nil,
			secret:    nil,
			err:       nil,
		},
	}

	for _, test := range tests {
		secret, err := ConvertSecretFromK8s(test.oriSecret)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.secret, secret)
	}
}