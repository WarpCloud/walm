package usecase

import (
	"WarpCloud/walm/pkg/models/tenant"
	"github.com/sirupsen/logrus"
	"WarpCloud/walm/pkg/k8s"
	k8sModel "WarpCloud/walm/pkg/models/k8s"
	errorModel "WarpCloud/walm/pkg/models/error"
	"fmt"
	"sync"
	"WarpCloud/walm/pkg/release"
)

type Tenant struct {
	k8sCache k8s.Cache
	k8sOperator k8s.Operator
	releaseUsecase release.UseCase
}

func (tenantImpl *Tenant) CreateTenant(tenantName string, tenantParams *tenant.TenantParams) error {
	_, err := tenantImpl.GetTenant(tenantName)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			namespace := buildNamespace(tenantParams, tenantName)
			err = tenantImpl.k8sOperator.CreateNamespace(namespace)
			if err != nil {
				logrus.Errorf("failed to create namespace %s : %s", tenantName, err.Error())
				return err
			}

			err = tenantImpl.doCreateTenant(tenantName, tenantParams)
			if err != nil {
				// rollback
				err1 := tenantImpl.k8sOperator.DeleteNamespace(tenantName)
				if err1 != nil {
					logrus.Warnf("failed to rollback deleting namespace %s", err1.Error())
				}
				return err
			}
			logrus.Infof("succeed to create tenant %s", tenantName)
			return nil
		}
		logrus.Errorf("failed to get tenant : %s", err.Error())
		return err

	}
	logrus.Warnf("namespace %s exists", tenantName)
	return nil
}

func buildNamespace(tenantParams *tenant.TenantParams, tenantName string) *k8sModel.Namespace {
	namespace := &k8sModel.Namespace{
		Meta: k8sModel.Meta{
			Namespace: tenantName,
			Name:      tenantName,
		},
		Labels:      tenantParams.TenantLabels,
		Annotations: tenantParams.TenantAnnotations,
	}
	if namespace.Labels == nil {
		namespace.Labels = map[string]string{}
	}
	namespace.Labels[tenant.MultiTenantLabelKey] = fmt.Sprintf("tenant-tiller-%s", tenantName)
	return namespace
}

func (tenantImpl *Tenant)doCreateTenant(tenantName string, tenantParams *tenant.TenantParams) error {
	for _, tenantQuota := range tenantParams.TenantQuotas {
		err := tenantImpl.createResourceQuota(tenantName, tenantQuota)
		if err != nil {
			logrus.Errorf("failed to create resource quota : %s", err.Error())
			return err
		}
	}

	err := tenantImpl.k8sOperator.CreateLimitRange(getDefaultLimitRange(tenantName))
	if err != nil {
		logrus.Errorf("failed to create limitrange : %s", err.Error())
		return err
	}

	return nil
}

const(
	limitRangeDefaultMem = "128Mi"
	limitRangeDefaultCpu = "0.1"
	LimitRangeDefaultName = "walm-default-limitrange"
)

func getDefaultLimitRange(namespace string) *k8sModel.LimitRange {
	return &k8sModel.LimitRange{
		Meta: k8sModel.Meta{
			Namespace: namespace,
			Name: LimitRangeDefaultName,
		},
		DefaultLimit: map[k8sModel.ResourceName]string{
			k8sModel.ResourceCPU: limitRangeDefaultCpu,
			k8sModel.ResourceMemory: limitRangeDefaultMem,
		},
	}
}

func (tenantImpl *Tenant)createResourceQuota(tenantName string, tenantQuota *tenant.TenantQuotaParams) error {
	resourceQuota := buildResourceQuota(tenantName, tenantQuota)
	err := tenantImpl.k8sOperator.CreateResourceQuota(resourceQuota)
	if err != nil {
		logrus.Errorf("failed to create resource quota : %s", err.Error())
		return err
	}
	return nil
}

func buildResourceQuota(tenantName string, tenantQuota *tenant.TenantQuotaParams) *k8sModel.ResourceQuota {
	resourceQuota := &k8sModel.ResourceQuota{
		Meta: k8sModel.Meta{
			Namespace: tenantName,
			Name:      tenantQuota.QuotaName,
			Kind:      k8sModel.ResourceQuotaKind,
		},
		ResourceLimits: map[k8sModel.ResourceName]string{
			k8sModel.ResourcePods:            tenantQuota.Hard.Pods,
			k8sModel.ResourceLimitsCPU:       tenantQuota.Hard.LimitCpu,
			k8sModel.ResourceLimitsMemory:    tenantQuota.Hard.LimitMemory,
			k8sModel.ResourceRequestsCPU:     tenantQuota.Hard.RequestsCPU,
			k8sModel.ResourceRequestsMemory:  tenantQuota.Hard.RequestsMemory,
			k8sModel.ResourceRequestsStorage: tenantQuota.Hard.RequestsStorage,
		},
	}
	return resourceQuota
}

func (tenantImpl *Tenant)updateResourceQuota(tenantName string, tenantQuota *tenant.TenantQuotaParams) error {
	resourceQuota := buildResourceQuota(tenantName, tenantQuota)
	err := tenantImpl.k8sOperator.CreateOrUpdateResourceQuota(resourceQuota)
	if err != nil {
		logrus.Errorf("failed to update resource quota : %s", err.Error())
		return err
	}
	return nil
}

func (tenantImpl *Tenant) GetTenant(tenantName string) (*tenant.TenantInfo, error) {
	return tenantImpl.k8sCache.GetTenant(tenantName)
}

func (tenantImpl *Tenant) ListTenants() (*tenant.TenantInfoList, error) {
	return tenantImpl.k8sCache.ListTenants("")
}

func (tenantImpl *Tenant) DeleteTenant(tenantName string) error {
	_, err := tenantImpl.k8sCache.GetTenant(tenantName)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			return nil
		} else {
			return err
		}
	}

	releases, err := tenantImpl.releaseUsecase.ListReleases(tenantName)
	if err != nil {
		logrus.Errorf("failed to get releases in tenant %s : %s", tenantName, err.Error())
		return err
	}

	var wg sync.WaitGroup
	for _, release := range releases {
		wg.Add(1)
		go func(releaseName string) {
			defer wg.Done()
			err1 := tenantImpl.releaseUsecase.DeleteReleaseWithRetry(tenantName, releaseName,  false, false, 0)
			if err1 != nil {
				err = fmt.Errorf("failed to delete release %s under tenant %s : %s", releaseName, tenantName, err1.Error())
				logrus.Error(err.Error())
			}
		}(release.Name)
	}
	wg.Wait()

	if err != nil {
		return err
	}

	err = tenantImpl.k8sOperator.DeleteNamespace(tenantName)
	if err != nil {
		logrus.Errorf("failed to delete namespace %s : %s", tenantName, err.Error())
		return err
	}

	logrus.Infof("succeed to delete tenant %s", tenantName)
	return nil
}

func (tenantImpl *Tenant) UpdateTenant(tenantName string, tenantParams *tenant.TenantParams) error {
	_, err := tenantImpl.k8sCache.GetTenant(tenantName)
	if err != nil {
		logrus.Errorf("failed to get tenantInfo : %s", err.Error())
		return err
	}

	namespace := buildNamespace(tenantParams, tenantName)
	err = tenantImpl.k8sOperator.UpdateNamespace(namespace)
	if err != nil {
		logrus.Errorf("failed to update namespace : %s", err.Error())
		return err
	}

	for _, tenantQuota := range tenantParams.TenantQuotas {
		err := tenantImpl.updateResourceQuota(tenantName, tenantQuota)
		if err != nil {
			logrus.Errorf("failed to update resource quota : %s", err.Error())
			return err
		}
	}
	logrus.Infof("succeed to update tenant %s", tenantName)
	return nil
}
