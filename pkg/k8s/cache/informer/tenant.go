package informer

import (
	"github.com/sirupsen/logrus"
	"WarpCloud/walm/pkg/k8s/utils"
	"WarpCloud/walm/pkg/k8s/converter"
	"WarpCloud/walm/pkg/models/k8s"
	"WarpCloud/walm/pkg/models/tenant"
	errorModel "WarpCloud/walm/pkg/models/error"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (informer *Informer) ListTenants(labelSelectorStr string) (*tenant.TenantInfoList, error) {
	tenantInfoList := &tenant.TenantInfoList{}
	selector, err := labels.Parse(labelSelectorStr)
	if err != nil {
		logrus.Errorf("failed to parse label string %s : %s", labelSelectorStr, err.Error())
		return nil, err
	}
	namespaces, err := informer.namespaceLister.List(selector)
	if err != nil {
		logrus.Errorf("failed to list namespaces : %s", err.Error())
		return nil, err
	}

	for _, namespace := range namespaces {
		tenantInfo, err := informer.buildTenantInfo(namespace)
		if err != nil {
			logrus.Errorf("failed to build tenant info %s : %s", namespace.Name, err)
			return nil, err
		}
		tenantInfoList.Items = append(tenantInfoList.Items, tenantInfo)
	}

	return tenantInfoList, nil
}

func (informer *Informer) GetTenant(tenantName string) (*tenant.TenantInfo, error) {
	namespace, err := informer.namespaceLister.Get(tenantName)
	if err != nil {
		if utils.IsK8sResourceNotFoundErr(err) {
			logrus.Warnf("namespace %s is not found", tenantName)
			return nil, errorModel.NotFoundError{}
		} else {
			logrus.Errorf("failed to get namespace %s : %s", tenantName, err.Error())
			return nil, err
		}
	}

	return informer.buildTenantInfo(namespace)
}

func (informer *Informer)buildTenantInfo(namespace *corev1.Namespace)(*tenant.TenantInfo, error) {
	tenantInfo := buildBasicTenantInfo(namespace)

	resourceQuotas, err := informer.resourceQuotaLister.ResourceQuotas(namespace.Name).List(labels.NewSelector())
	if err != nil {
		logrus.Errorf("failed to get resource quotas : %s", err.Error())
		return nil, err
	}

	tenantInfo.TenantQuotas, tenantInfo.UnifyUnitTenantQuotas, err = buildTenantQuotas(resourceQuotas)
	if err != nil {
		return nil, err
	}

	return tenantInfo, nil
}

func buildBasicTenantInfo(namespace *corev1.Namespace) *tenant.TenantInfo {
	tenantInfo := &tenant.TenantInfo{
		TenantName:            namespace.Name,
		TenantCreationTime:    namespace.CreationTimestamp.String(),
		TenantLabels:          namespace.Labels,
		TenantAnnotitions:     namespace.Annotations,
		TenantStatus:          namespace.Status.String(),
		TenantQuotas:          []*tenant.TenantQuota{},
		UnifyUnitTenantQuotas: []*tenant.UnifyUnitTenantQuota{},
	}
	if tenantInfo.TenantLabels == nil {
		tenantInfo.TenantLabels = map[string]string{}
	}
	if tenantInfo.TenantAnnotitions == nil {
		tenantInfo.TenantAnnotitions = map[string]string{}
	}
	if _, ok := namespace.Labels[tenant.MultiTenantLabelKey]; ok {
		tenantInfo.MultiTenant = true
	}
	if namespace.Status.Phase == corev1.NamespaceActive {
		tenantInfo.Ready = true
	}
	return tenantInfo
}

func buildTenantQuotas(resourceQuotas []*corev1.ResourceQuota) ([]*tenant.TenantQuota, []*tenant.UnifyUnitTenantQuota, error) {
	tenantQuotas := []*tenant.TenantQuota{}
	unifyUnitTenantQuotas := []*tenant.UnifyUnitTenantQuota{}
	for _, resourceQuota := range resourceQuotas {
		walmResourceQuota, err := converter.ConvertResourceQuotaFromK8s(resourceQuota)
		if err != nil {
			logrus.Errorf("failed to convert resource quota %s: %s", resourceQuota.Name, err.Error())
			return nil, nil, err
		}
		hard := &tenant.TenantQuotaInfo{
			Pods:            walmResourceQuota.ResourceLimits[k8s.ResourcePods],
			LimitCpu:        walmResourceQuota.ResourceLimits[k8s.ResourceLimitsCPU],
			LimitMemory:     walmResourceQuota.ResourceLimits[k8s.ResourceLimitsMemory],
			RequestsStorage: walmResourceQuota.ResourceLimits[k8s.ResourceRequestsStorage],
			RequestsMemory:  walmResourceQuota.ResourceLimits[k8s.ResourceRequestsMemory],
			RequestsCPU:     walmResourceQuota.ResourceLimits[k8s.ResourceRequestsCPU],
		}
		used := &tenant.TenantQuotaInfo{
			Pods:            walmResourceQuota.ResourceUsed[k8s.ResourcePods],
			LimitCpu:        walmResourceQuota.ResourceUsed[k8s.ResourceLimitsCPU],
			LimitMemory:     walmResourceQuota.ResourceUsed[k8s.ResourceLimitsMemory],
			RequestsStorage: walmResourceQuota.ResourceUsed[k8s.ResourceRequestsStorage],
			RequestsMemory:  walmResourceQuota.ResourceUsed[k8s.ResourceRequestsMemory],
			RequestsCPU:     walmResourceQuota.ResourceUsed[k8s.ResourceRequestsCPU],
		}
		tenantQuotas = append(tenantQuotas, &tenant.TenantQuota{
			QuotaName: walmResourceQuota.Name,
			Hard:      hard,
			Used:      used})
		unifyUnitTenantQuotas = append(unifyUnitTenantQuotas, buildUnifyUnitTenantQuota(walmResourceQuota.Name, hard, used))
	}
	return tenantQuotas, unifyUnitTenantQuotas, nil
}

func buildUnifyUnitTenantQuota(name string, hard *tenant.TenantQuotaInfo, used *tenant.TenantQuotaInfo) *tenant.UnifyUnitTenantQuota {
	return &tenant.UnifyUnitTenantQuota{
		QuotaName: name,
		Hard:      buildUnifyUnitTenantInfo(hard),
		Used:      buildUnifyUnitTenantInfo(used),
	}
}

func buildUnifyUnitTenantInfo(info *tenant.TenantQuotaInfo) *tenant.UnifyUnitTenantQuotaInfo {
	return &tenant.UnifyUnitTenantQuotaInfo{
		RequestsCPU:     utils.ParseK8sResourceCpu(info.RequestsCPU),
		RequestsMemory:  utils.ParseK8sResourceMemory(info.RequestsMemory),
		RequestsStorage: utils.ParseK8sResourceStorage(info.RequestsStorage),
		LimitMemory:     utils.ParseK8sResourceMemory(info.LimitMemory),
		LimitCpu:        utils.ParseK8sResourceCpu(info.LimitCpu),
		Pods:            utils.ParseK8sResourcePod(info.Pods),
	}
}
