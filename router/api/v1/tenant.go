package v1

import (
	"net/http"
	"fmt"

	"github.com/emicklei/go-restful"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"walm/pkg/k8s/handler"
	"walm/pkg/tenant"
	"walm/pkg/k8s/adaptor"
)

func ListTenants(request *restful.Request, response *restful.Response) {
	var tenantInfoList tenant.TenantInfoList
	namespaces, err := handler.GetDefaultHandlerSet().GetNamespaceHandler().ListNamespaces(nil)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	for _, namespace := range namespaces {
		var tenantInfo tenant.TenantInfo
		tenantInfo.TenantName = namespace.GetName()
		tenantInfo.TenantCreationTime = namespace.GetCreationTimestamp()
		tenantInfo.TenantStatus = namespace.Status.String()
		tenantInfoList.Items = append(tenantInfoList.Items, &tenantInfo)
	}
	response.WriteEntity(tenantInfoList)
}

func CreateTenant(request *restful.Request, response *restful.Response) {
	tenantParams := new(tenant.TenantParams)
	err := request.ReadEntity(&tenantParams)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tenantParams.TenantName,
			Name: tenantParams.TenantName,
		},
	}
	_, err = handler.GetDefaultHandlerSet().GetNamespaceHandler().GetNamespace(tenantParams.TenantName)
	if err != nil {
		_, err2 := handler.GetDefaultHandlerSet().GetNamespaceHandler().CreateNamespace(&namespace)
		if err2 != nil {
			response.WriteError(http.StatusInternalServerError, err2)
		}
	}
}

func GetTenant(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("tenantName")
	namespace, err := handler.GetDefaultHandlerSet().GetNamespaceHandler().GetNamespace(tenantName)
	if err != nil {
		if adaptor.IsNotFoundErr(err) {
			WriteNotFoundResponse(response, -1, fmt.Sprintf("namespace %s is not found", namespace))
			return
		}
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to get namespace %s: %s" , tenantName, err.Error()))
		return
	}

	tenantInfo := tenant.TenantInfo{
		TenantName: namespace.Name,
		TenantCreationTime: namespace.CreationTimestamp,
		//TenantLabels: namespace.Labels,
		TenantStatus: namespace.Status.String(),
	}

	response.WriteEntity(tenantInfo)
}

func DeleteTenant(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("tenantName")
	_, err := handler.GetDefaultHandlerSet().GetNamespaceHandler().GetNamespace(tenantName)
	if err != nil {
		if adaptor.IsNotFoundErr(err) {
			response.WriteEntity(nil)
			return
		}
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to get namespace %s: %s" , tenantName, err.Error()))
		return
	}

	err = handler.GetDefaultHandlerSet().GetNamespaceHandler().DeleteNamespace(tenantName)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to delete namespace %s: %s" , tenantName, err.Error()))
		return
	}

	response.WriteEntity(nil)
}

func GetQuotas(request *restful.Request, response *restful.Response) {
}

func UpdateQuotas(request *restful.Request, response *restful.Response) {
}
