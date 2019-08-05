package tenant

import "WarpCloud/walm/pkg/models/tenant"

type UseCase interface {
	CreateTenant(tenantName string, tenantParams *tenant.TenantParams) error
	GetTenant(tenantName string) (*tenant.TenantInfo, error)
	ListTenants() (*tenant.TenantInfoList, error)
	DeleteTenant(tenantName string) error
	UpdateTenant(tenantName string, tenantParams *tenant.TenantParams) error
}