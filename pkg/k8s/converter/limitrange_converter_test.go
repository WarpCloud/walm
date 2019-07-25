package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestConvertLimitRangeToK8s(t *testing.T) {
	tests := []struct {
		orilimitRange *k8s.LimitRange
		limitRange    *corev1.LimitRange
		err           error
	}{
		{
			orilimitRange: &k8s.LimitRange{
				Meta: k8s.Meta{
					Name:      "test-limit-range",
					Namespace: "test-namespace",
				},
				DefaultLimit: map[k8s.ResourceName]string{
					k8s.ResourceMemory:         "50Gi",
					k8s.ResourceCPU:            "5",
					k8s.ResourceRequestsMemory: "1Gi",
					k8s.ResourceRequestsCPU:    "1",
				},
			},
			limitRange: &corev1.LimitRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-limit-range",
					Namespace: "test-namespace",
				},
				Spec: corev1.LimitRangeSpec{
					Limits: []corev1.LimitRangeItem{
						{
							Type: corev1.LimitTypeContainer,
							Default: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceMemory:         resource.MustParse("50Gi"),
								corev1.ResourceCPU:            resource.MustParse("5"),
								corev1.ResourceRequestsMemory: resource.MustParse("1Gi"),
								corev1.ResourceRequestsCPU:    resource.MustParse("1"),
							},
						},
					},
				},
			},
			err: nil,
		},
		{
			orilimitRange: &k8s.LimitRange{
				Meta:         k8s.Meta{},
				DefaultLimit: nil,
			},
			limitRange: &corev1.LimitRange{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       corev1.LimitRangeSpec{
					Limits: []corev1.LimitRangeItem{
						{
							Type: corev1.LimitTypeContainer,
							Default: corev1.ResourceList{},
						},
					},
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		limitRange, err := ConvertLimitRangeToK8s(test.orilimitRange)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.limitRange, limitRange)
	}
}
