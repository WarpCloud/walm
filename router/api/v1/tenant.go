package v1

import (
	"github.com/emicklei/go-restful"
	"walm/pkg/k8s/handler"
	"walm/pkg/tenant"
	"net/http"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
}

func DeleteTenant(request *restful.Request, response *restful.Response) {
}

func GetQuotas(request *restful.Request, response *restful.Response) {
}

func UpdateQuotas(request *restful.Request, response *restful.Response) {
}
