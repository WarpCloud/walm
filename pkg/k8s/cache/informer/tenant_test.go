package informer

import (
	"testing"
	"WarpCloud/walm/pkg/models/tenant"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

func Test_BuildBasicTenantInfo(t *testing.T) {
	tests := []struct {
		namespace  *corev1.Namespace
		tenantInfo *tenant.TenantInfo
	}{
		{
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-namespace",
					CreationTimestamp: metav1.NewTime(time.Date(2019, 1, 1, 1, 0, 0, 0, time.UTC)),
				},
				Status: corev1.NamespaceStatus{
					Phase: corev1.NamespaceActive,
				},
			},
			tenantInfo: &tenant.TenantInfo{
				TenantName:            "test-namespace",
				Ready:                 true,
				MultiTenant:           false,
				TenantStatus:          "&NamespaceStatus{Phase:Active,}",
				TenantCreationTime:    "2019-01-01 01:00:00 +0000 UTC",
				UnifyUnitTenantQuotas: []*tenant.UnifyUnitTenantQuota{},
				TenantQuotas:          []*tenant.TenantQuota{},
				TenantLabels:          map[string]string{},
				TenantAnnotitions:     map[string]string{},
			},
		},
		{
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-namespace",
					CreationTimestamp: metav1.NewTime(time.Date(2019, 1, 1, 1, 0, 0, 0, time.UTC)),
					Labels:            map[string]string{tenant.MultiTenantLabelKey: "true"},
				},
				Status: corev1.NamespaceStatus{
					Phase: corev1.NamespaceTerminating,
				},
			},
			tenantInfo: &tenant.TenantInfo{
				TenantName:            "test-namespace",
				Ready:                 false,
				MultiTenant:           true,
				TenantStatus:          "&NamespaceStatus{Phase:Terminating,}",
				TenantCreationTime:    "2019-01-01 01:00:00 +0000 UTC",
				UnifyUnitTenantQuotas: []*tenant.UnifyUnitTenantQuota{},
				TenantQuotas:          []*tenant.TenantQuota{},
				TenantLabels:          map[string]string{tenant.MultiTenantLabelKey: "true"},
				TenantAnnotitions:     map[string]string{},
			},
		},
	}

	for _, test := range tests {
		tenantInfo := buildBasicTenantInfo(test.namespace)
		assert.Equal(t, test.tenantInfo, tenantInfo)
	}
}

func Test_BuildTenantQuotas(t *testing.T) {
	tests := []struct {
		resourceQuotas        []*corev1.ResourceQuota
		tenantQuotas          []*tenant.TenantQuota
		unifyUnitTenantQuotas []*tenant.UnifyUnitTenantQuota
		err                   error
	}{
		{
			resourceQuotas: []*corev1.ResourceQuota{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-rq",
					},
					Spec: corev1.ResourceQuotaSpec{
						Hard: corev1.ResourceList{
							corev1.ResourceRequestsMemory:  resource.MustParse("100Gi"),
							corev1.ResourceLimitsMemory:    resource.MustParse("200Gi"),
							corev1.ResourceRequestsCPU:     resource.MustParse("100"),
							corev1.ResourceLimitsCPU:       resource.MustParse("200"),
							corev1.ResourceRequestsStorage: resource.MustParse("1000Gi"),
							corev1.ResourcePods:            resource.MustParse("100"),
						},
					},
					Status: corev1.ResourceQuotaStatus{
						Used: corev1.ResourceList{
							corev1.ResourceRequestsMemory:  resource.MustParse("10Gi"),
							corev1.ResourceLimitsMemory:    resource.MustParse("20Gi"),
							corev1.ResourceRequestsCPU:     resource.MustParse("10"),
							corev1.ResourceLimitsCPU:       resource.MustParse("20"),
							corev1.ResourceRequestsStorage: resource.MustParse("100Gi"),
							corev1.ResourcePods:            resource.MustParse("10"),
						},
					},
				},
			},
			tenantQuotas: []*tenant.TenantQuota{
				{
					QuotaName: "test-rq",
					Hard: &tenant.TenantQuotaInfo{
						Pods:            "100",
						LimitCpu:        "200",
						LimitMemory:     "200Gi",
						RequestsStorage: "1000Gi",
						RequestsMemory:  "100Gi",
						RequestsCPU:     "100",
					},
					Used: &tenant.TenantQuotaInfo{
						Pods:            "10",
						LimitCpu:        "20",
						LimitMemory:     "20Gi",
						RequestsStorage: "100Gi",
						RequestsMemory:  "10Gi",
						RequestsCPU:     "10",
					},
				},
			},
			unifyUnitTenantQuotas: []*tenant.UnifyUnitTenantQuota{
				{
					QuotaName: "test-rq",
					Hard: &tenant.UnifyUnitTenantQuotaInfo{
						Pods:            100,
						LimitCpu:        200,
						LimitMemory:     204800,
						RequestsStorage: 1000,
						RequestsMemory:  102400,
						RequestsCPU:     100,
					},
					Used: &tenant.UnifyUnitTenantQuotaInfo{
						Pods:            10,
						LimitCpu:        20,
						LimitMemory:     20480,
						RequestsStorage: 100,
						RequestsMemory:  10240,
						RequestsCPU:     10,
					},
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		tenantQuotas, unifyTenantQuotas, err := buildTenantQuotas(test.resourceQuotas)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.tenantQuotas, tenantQuotas)
		assert.Equal(t, test.unifyUnitTenantQuotas, unifyTenantQuotas)
	}
}
