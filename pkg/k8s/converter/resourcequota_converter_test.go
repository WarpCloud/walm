package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestConvertResourceQuotaToK8s(t *testing.T) {
	tests := []struct {
		oriQuota *k8s.ResourceQuota
		quota    *corev1.ResourceQuota
		err      error
	}{
		{
			oriQuota: &k8s.ResourceQuota{
				Meta: k8s.Meta{
					Name:      "test-resourceQuota",
					Namespace: "test-namespace",
					Kind:      "ResourceQuota",
				},
				ResourceLimits: map[k8s.ResourceName]string{
					k8s.ResourcePods:           "20",
					k8s.ResourceRequestsCPU:    "20",
					k8s.ResourceRequestsMemory: "100Gi",
					k8s.ResourceLimitsCPU:      "40",
					k8s.ResourceLimitsMemory:   "200Gi",
				},
			},
			quota: &corev1.ResourceQuota{
				TypeMeta: metav1.TypeMeta{
					Kind: "",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-resourceQuota",
					Namespace: "test-namespace",
				},
				Spec: corev1.ResourceQuotaSpec{
					Hard: corev1.ResourceList{
						corev1.ResourceName(k8s.ResourcePods):           resource.MustParse("20"),
						corev1.ResourceName(k8s.ResourceRequestsCPU):    resource.MustParse("20"),
						corev1.ResourceName(k8s.ResourceRequestsMemory): resource.MustParse("100Gi"),
						corev1.ResourceName(k8s.ResourceLimitsCPU):      resource.MustParse("40"),
						corev1.ResourceName(k8s.ResourceLimitsMemory):   resource.MustParse("200Gi"),
					},
				},
			},
			err: nil,
		},
		{
			oriQuota: &k8s.ResourceQuota{},
			quota:    &corev1.ResourceQuota{
				Spec: corev1.ResourceQuotaSpec{
					Hard: corev1.ResourceList{},
				},
			},
			err:      nil,
		},
	}

	for _, test := range tests {
		quota, err := ConvertResourceQuotaToK8s(test.oriQuota)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.quota, quota)
	}
}

func TestConvertResourceQuotaFromK8s(t *testing.T) {
	tests := []struct {
		oriQuota *corev1.ResourceQuota
		quota    *k8s.ResourceQuota
		err      error
	}{
		{
			oriQuota: &corev1.ResourceQuota{
				TypeMeta: metav1.TypeMeta{
					Kind: "ResourceQuota",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-resourceQuota",
					Namespace: "test-namespace",
				},
				Spec: corev1.ResourceQuotaSpec{
					Hard: corev1.ResourceList{
						corev1.ResourceName(k8s.ResourcePods):           resource.MustParse("20"),
						corev1.ResourceName(k8s.ResourceRequestsCPU):    resource.MustParse("20"),
						corev1.ResourceName(k8s.ResourceRequestsMemory): resource.MustParse("100Gi"),
						corev1.ResourceName(k8s.ResourceLimitsCPU):      resource.MustParse("40"),
						corev1.ResourceName(k8s.ResourceLimitsMemory):   resource.MustParse("200Gi"),
					},
				},
				Status: corev1.ResourceQuotaStatus{
					Used: corev1.ResourceList{
						corev1.ResourceName(k8s.ResourcePods):           resource.MustParse("10"),
						corev1.ResourceName(k8s.ResourceRequestsCPU):    resource.MustParse("10"),
						corev1.ResourceName(k8s.ResourceRequestsMemory): resource.MustParse("50Gi"),
						corev1.ResourceName(k8s.ResourceLimitsCPU):      resource.MustParse("20"),
						corev1.ResourceName(k8s.ResourceLimitsMemory):   resource.MustParse("100Gi"),
					},
				},
			},

			quota: &k8s.ResourceQuota{
				Meta: k8s.Meta{
					Name: "test-resourceQuota",
					Namespace: "test-namespace",
					Kind: "ResourceQuota",
					State: k8s.State{
						Status:  "Ready",
						Reason:  "",
						Message: "",
					},
				},
				ResourceLimits: map[k8s.ResourceName]string{
					k8s.ResourcePods:           "20",
					k8s.ResourceRequestsCPU:    "20",
					k8s.ResourceRequestsMemory: "100Gi",
					k8s.ResourceLimitsCPU:      "40",
					k8s.ResourceLimitsMemory:   "200Gi",
				},
				ResourceUsed: map[k8s.ResourceName]string{
					k8s.ResourcePods:           "10",
					k8s.ResourceRequestsCPU:    "10",
					k8s.ResourceRequestsMemory: "50Gi",
					k8s.ResourceLimitsCPU:      "20",
					k8s.ResourceLimitsMemory:   "100Gi",
				},
			},
			err: nil,
		},
		{
			oriQuota: nil,
			quota:    nil,
			err:      nil,
		},
	}

	for _, test := range tests {
		quota, err := ConvertResourceQuotaFromK8s(test.oriQuota)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.quota, quota)

	}
}
