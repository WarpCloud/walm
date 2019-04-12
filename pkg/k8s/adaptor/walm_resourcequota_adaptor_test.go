package adaptor

import (
	"testing"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"github.com/stretchr/testify/assert"
)

func TestBuildWalmResourceQuota(t *testing.T) {
	tests := []struct {
		resourceQuota *v1.ResourceQuota
		result        *WalmResourceQuota
	}{
		{
			resourceQuota: &v1.ResourceQuota{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test_ns",
					Name:      "test_name",
				},
				Spec: v1.ResourceQuotaSpec{
					Hard: v1.ResourceList{v1.ResourceMemory: resource.MustParse("10Gi")},
				},
				Status: v1.ResourceQuotaStatus{
					Used: v1.ResourceList{v1.ResourceMemory: resource.MustParse("5Gi")},
				},
			},
			result: &WalmResourceQuota{
				WalmMeta: WalmMeta{
					Namespace: "test_ns",
					Name:      "test_name",
					Kind:      "ResourceQuota",
					State:     buildWalmState("Ready", "", ""),
				},
				ResourceUsed:   map[v1.ResourceName]string{v1.ResourceMemory: "5Gi"},
				ResourceLimits: map[v1.ResourceName]string{v1.ResourceMemory: "10Gi"},
			},
		},
	}

	for _, test := range tests {
		result := BuildWalmResourceQuota(test.resourceQuota)
		assert.Equal(t, test.result, result)
	}
}

func TestBuildResourceQuota(t *testing.T) {
	tests := []struct {
		resourceQuota *WalmResourceQuota
		err           error
		result        *v1.ResourceQuota
	}{
		{
			resourceQuota: &WalmResourceQuota{
				WalmMeta: WalmMeta{
					Namespace: "test_ns",
					Name:      "test_name",
				},
				ResourceLimits: map[v1.ResourceName]string{v1.ResourceMemory: "10Gi"},
			},
			result: &v1.ResourceQuota{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test_ns",
					Name:      "test_name",
				},
				Spec: v1.ResourceQuotaSpec{
					Hard: v1.ResourceList{v1.ResourceMemory: resource.MustParse("10Gi")},
				},
			},
		},
	}

	for _, test := range tests {
		result, err := BuildResourceQuota(test.resourceQuota)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.result, result)
	}
}
