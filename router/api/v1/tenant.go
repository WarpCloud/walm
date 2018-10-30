package v1

import (
	"net/http"
	"fmt"

	"github.com/emicklei/go-restful"
	"walm/pkg/tenant"
)

func ListTenants(request *restful.Request, response *restful.Response) {
	tenantInfoList, err := tenant.ListTenants()
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	response.WriteEntity(tenantInfoList)
}

func CreateTenant(request *restful.Request, response *restful.Response) {
	tenantParams := new(tenant.TenantParams)
	err := request.ReadEntity(&tenantParams)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	err = tenant.CreateTenant(tenantParams)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
}

func GetTenant(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("tenantName")
	tenantInfo, err := tenant.GetTenant(tenantName)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	if tenantInfo == nil {
		response.WriteError(http.StatusNotFound, fmt.Errorf("namespace %s not found", tenantName))
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
	tenantParams := new(tenant.TenantParams)
	err := request.ReadEntity(&tenantParams)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	err = tenant.UpdateTenant(tenantParams)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
}
