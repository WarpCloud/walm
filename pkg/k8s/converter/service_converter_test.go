package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func TestConvertServiceFromK8s(t *testing.T) {
	tests := []struct {
		oriService  *corev1.Service
		endpoints   *corev1.Endpoints
		walmService *k8s.Service
		err         error
	}{
		{
			oriService: &corev1.Service{
				TypeMeta: metav1.TypeMeta{
					Kind: "Service",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "test-namespace",
				},
				Spec: corev1.ServiceSpec{
					ClusterIP: "10.0.171.239",
					Type:      "ClusterIP",
					Ports: []corev1.ServicePort{
						{
							Name:     "http",
							Protocol: "TCP",
							Port:     80,
							TargetPort: intstr.IntOrString{
								Type:   intstr.Int,
								IntVal: 9376,
							},
						},
					},
				},
				Status: corev1.ServiceStatus{},
			},
			endpoints: &corev1.Endpoints{
				TypeMeta: metav1.TypeMeta{

				},
				ObjectMeta: metav1.ObjectMeta{},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								IP: "1.2.3.4",
							},
						},
						Ports: []corev1.EndpointPort{
							{
								Port: 9376,
								Name: "http",
							},
						},
					},
				},
			},

			walmService: &k8s.Service{
				Meta: k8s.Meta{
					Name:      "test-service",
					Namespace: "test-namespace",
					Kind:      "Service",
					State: k8s.State{
						Status:  "Ready",
						Reason:  "",
						Message: "",
					},
				},
				Ports: []k8s.ServicePort{
					{
						Name:       "http",
						Protocol:   "TCP",
						Port:       80,
						TargetPort: "9376",
						Endpoints:  []string{"1.2.3.4:9376"},
					},
				},
				ClusterIp:   "10.0.171.239",
				ServiceType: "ClusterIP",
			},
			err: nil,
		},
		{
			oriService:  nil,
			endpoints:   nil,
			walmService: nil,
			err:         nil,
		},
	}

	for _, test := range tests {
		walmService, err := ConvertServiceFromK8s(test.oriService, test.endpoints)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.walmService, walmService)
	}
}
