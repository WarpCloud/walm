package handler

import (
	"testing"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/stretchr/testify/assert"
)

func TestResourceQuotaBuilder(t *testing.T) {
	tests := []struct{
		namespace string
		name string
		annotations map[string]string
		labels map[string]string
		hardResourceLimits v1.ResourceList
		resourceQuotaScopes []v1.ResourceQuotaScope
		result *v1.ResourceQuota
	} {
		{
			namespace: "test_namespace",
			name: "test_name",
			annotations: map[string]string{"test_key": "test_anno"},
			labels: map[string]string{"test_key": "test_label"},
			hardResourceLimits: v1.ResourceList{v1.ResourceCPU : resource.MustParse("1")},
			resourceQuotaScopes: []v1.ResourceQuotaScope{v1.ResourceQuotaScopeBestEffort},
			result: &v1.ResourceQuota{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test_namespace",
					Name: "test_name",
					Labels:      map[string]string{"test_key": "test_label"},
					Annotations: map[string]string{"test_key": "test_anno"},
				},
				Spec: v1.ResourceQuotaSpec{
					Hard:   v1.ResourceList{v1.ResourceCPU : resource.MustParse("1")},
					Scopes:[]v1.ResourceQuotaScope{v1.ResourceQuotaScopeBestEffort},
				},
			},
		},
	}

	for _, test := range tests {
		builder := &ResourceQuotaBuilder{}
		builder.Namespace(test.namespace).Name(test.name)
		for key, value := range test.annotations {
			builder.AddAnnotations(key, value)
		}
		for key, value := range test.labels {
			builder.AddLabel(key, value)
		}
		for key, value := range test.hardResourceLimits {
			builder.AddHardResourceLimit(key, value)
		}
		for _, scope := range test.resourceQuotaScopes {
			builder.AddResourceQuotaScope(scope)
		}
		rq := builder.Build()
		assert.Equal(t, test.result, rq)
	}
}
