package tenant

import (
	"fmt"
	"walm/pkg/k8s"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetTenantInfo(name string) (TenantInfo, error) {
	if namespace, err := k8s.GetDefaultClient().CoreV1().Namespaces().Get(name, metaV1.GetOptions{}); err != nil {
		return TenantInfo{}, err
	} else {

		return TenantInfo{
			TenantName:         namespace.GetName(),
			TenantUid:          string(namespace.GetUID()),
			TenantCreationTime: namespace.GetCreationTimestamp(),
			TenantLabels:       MapToKvString(namespace.GetLabels()),
			TenantStatus:       string(namespace.Status.Phase),
		}, nil
	}

}

func DeleteTenant(name string) error {
	return k8s.GetDefaultClient().CoreV1().Namespaces().Delete(name, &metaV1.DeleteOptions{})
}

/*
func CreateTenant(info *TenantInfo) error {

	ns := &api.Namespace{
		ObjectMeta: metaV1.ObjectMeta{
			Name: spec.Name,
		},
	}
	if _, err := k8s.GetDefaultClient().CoreV1().Namespaces().Create(ns); err != nil {
		return err
	} else {
		return nil
	}
}
*/

func MapToKvString(m map[string]string) string {
	var str string
	for k, v := range m {
		str = fmt.Sprintf("%s %s=%v ", str, k, v)
	}
	return str

}
