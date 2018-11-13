package v1

import (
	"fmt"

	"github.com/emicklei/go-restful"
	"walm/pkg/tenant"
	walmerr "walm/pkg/util/error"
)

func ListTenants(request *restful.Request, response *restful.Response) {
	tenantInfoList, err := tenant.ListTenants()
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to list tenants : %s", err.Error()))
		return
	}
	response.WriteEntity(tenantInfoList)
}

func CreateTenant(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("tenantName")
	tenantParams := new(tenant.TenantParams)
	err := request.ReadEntity(&tenantParams)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to read tenant params : %s", err.Error()))
		return
	}
	err = tenant.CreateTenant(tenantName, tenantParams)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to create tenant : %s", err.Error()))
		return
	}
}

func GetTenant(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("tenantName")
	tenantInfo, err := tenant.GetTenant(tenantName)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			WriteNotFoundResponse(response, -1, fmt.Sprintf("tenant %s is not found", tenantName))
			return
		}
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to get tenant : %s", err.Error()))
		return
	}
	response.WriteEntity(tenantInfo)
}

func DeleteTenant(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("tenantName")
	err := tenant.DeleteTenant(tenantName)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to delete namespace %s: %s" , tenantName, err.Error()))
		return
	}

	response.WriteEntity(nil)
}

func UpdateTenant(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("tenantName")
	tenantParams := new(tenant.TenantParams)
	err := request.ReadEntity(&tenantParams)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to read tenant params : %s", err.Error()))
		return
	}
	err = tenant.UpdateTenant(tenantName, tenantParams)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to update tenant : %s", err.Error()))
		return
	}
}
