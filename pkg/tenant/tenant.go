package tenant

import (
	"walm/pkg/k8s/handler"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"walm/pkg/release"
	"walm/pkg/release/manager/helm"
	"github.com/sirupsen/logrus"
	"fmt"
	"walm/pkg/k8s/adaptor"
	"walm/pkg/setting"
)

func ListTenants() (TenantInfoList, error) {
	var tenantInfoList TenantInfoList

	namespaces, err := handler.GetDefaultHandlerSet().GetNamespaceHandler().ListNamespaces(nil)
	if err != nil {
		return tenantInfoList, err
	}
	for _, namespace := range namespaces {
		tenantInfo, err := GetTenant(namespace.Name)
		if err != nil {
			logrus.Errorf("ListTenants getTenant %s error %v", namespace.Name, err)
		}
		tenantInfoList.Items = append(tenantInfoList.Items, tenantInfo)
	}

	return tenantInfoList, nil
}


func GetTenant(tenantName string) (*TenantInfo, error) {
	namespace, err := handler.GetDefaultHandlerSet().GetNamespaceHandler().GetNamespace(tenantName)
	if err != nil {
		if adaptor.IsNotFoundErr(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	tenantInfo := TenantInfo{
		TenantName: namespace.Name,
		TenantCreationTime: namespace.CreationTimestamp,
		TenantLabels: namespace.Labels,
		TenantAnnotitions: namespace.Annotations,
		TenantStatus: namespace.Status.String(),
	}

	_, ok := namespace.Labels["multi-tenant"]
	if ok {
		tenantInfo.MultiTenant = true
	} else {
		tenantInfo.MultiTenant = false
	}
	tenantInfo.Ready = true

	return &tenantInfo, nil
}

// CreateTenant initialize the namespace for the tenant
// and installs the essential components
func CreateTenant(tenantParams *TenantParams) error {
	tenantLabel := make(map[string]string, 0)
	for k, v := range tenantParams.TenantLabels {
		tenantLabel[k] = v
	}
	tenantLabel["multi-tenant"] = fmt.Sprintf("tenant-tiller-%s", tenantParams.TenantName)
	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tenantParams.TenantName,
			Name: tenantParams.TenantName,
			Labels: tenantLabel,
			Annotations: tenantParams.TenantAnnotitions,
		},
	}

	_, err := handler.GetDefaultHandlerSet().GetNamespaceHandler().GetNamespace(tenantParams.TenantName)
	if err != nil {
		_, err2 := handler.GetDefaultHandlerSet().GetNamespaceHandler().CreateNamespace(&namespace)
		if err2 != nil {
			return err2
		}
	}

	err = deployTillerCharts(tenantParams.TenantName)
	if err != nil {
		return err
	}

	return nil
}

func DeleteTenant(tenantName string) error {
	logrus.Infof("DeleteTenant %s start\n", tenantName)

	err := helm.GetDefaultHelmClient().DeleteRelease(tenantName, fmt.Sprintf("tenant-tiller-%s", tenantName))
	if err != nil {
		logrus.Errorf("DeleteTenant %s error %v", tenantName, err)
	}

	namespace, err := handler.GetDefaultHandlerSet().GetNamespaceHandler().GetNamespace(tenantName)
	if err != nil {
		if adaptor.IsNotFoundErr(err) {
			return nil
		} else {
			return err
		}
	}

	err = handler.GetDefaultHandlerSet().GetNamespaceHandler().DeleteNamespace(tenantName)
	if err != nil {
		logrus.Errorf("DeleteTenant %s error %v", namespace.Name, err)
		return err
	}

	return nil
}

func UpdateTenant(tenantParams *TenantParams) error {
	tenantInfo, err := GetTenant(tenantParams.TenantName)
	if err != nil {
		logrus.Errorf("UpdateTenant getTenant %s error %v", tenantParams.TenantName, err)
		return err
	}
	if tenantInfo == nil {
		return fmt.Errorf("UpdateTenant tenant %s not found", tenantParams.TenantName)
	}

	return nil
}

func deployTillerCharts(namespace string) error {
	tillerRelease := release.ReleaseRequest{}
	tillerRelease.Name = fmt.Sprintf("tenant-tiller-%s", namespace)
	tillerRelease.ChartName = "helm-tiller-tenant"
	tillerRelease.ConfigValues = make(map[string]interface{}, 0)
	tillerRelease.ConfigValues["tiller"] = map[string]string {
		"image": setting.Config.MultiTenantConfig.TillerImage,
	}
	err := helm.GetDefaultHelmClient().InstallUpgradeRealese(namespace, &tillerRelease)
	logrus.Infof("tenant %s deploy tiller %v\n", namespace, err)

	return err
}
