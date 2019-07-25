package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"github.com/stretchr/testify/assert"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func TestConvertIngressFromK8s(t *testing.T) {
	testIngressRuleValue := extv1beta1.IngressRuleValue{
		HTTP: &extv1beta1.HTTPIngressRuleValue{
			Paths: []extv1beta1.HTTPIngressPath{
				{
					Path: "/",
					Backend: extv1beta1.IngressBackend{
						ServiceName: "test-service",
						ServicePort: intstr.IntOrString{
							Type: intstr.Int,
							IntVal: 80,
						},
					},
				},
			},
		},
	}
	tests := []struct{
		oriIngress *extv1beta1.Ingress
		ingress    *k8s.Ingress
		err        error
	}{
		{
			oriIngress: &extv1beta1.Ingress{
				TypeMeta:   metav1.TypeMeta{
					Kind: "Ingress",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-ingress",
					Namespace: "test-namespace",
				},
				Spec:       extv1beta1.IngressSpec{
					Rules: []extv1beta1.IngressRule{
						{
							Host: "test-ingress.example.com",
							IngressRuleValue: testIngressRuleValue,
						},
					},
				},
				Status:     extv1beta1.IngressStatus{},
			},
			ingress: &k8s.Ingress{
				Meta:        k8s.Meta{
					Name: "test-ingress",
					Namespace: "test-namespace",
					Kind: "Ingress",
					State: k8s.State{
						Status:  "Ready",
						Reason:  "",
						Message: "",
					},
				},
				Host:        "test-ingress.example.com",
				Path:        "/",
				ServiceName: "test-service",
				ServicePort: "80",
			},
			err: nil,
		},
		{
			oriIngress: nil,
			ingress: nil,
			err: nil,
		},
	}

	for _, test := range tests {
		ingress, err := ConvertIngressFromK8s(test.oriIngress)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.ingress, ingress)
	}
}