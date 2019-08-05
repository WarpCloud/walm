package http

import (
	"WarpCloud/walm/pkg/tenant"
	"github.com/emicklei/go-restful"
	httpUtils "WarpCloud/walm/pkg/util/http"
	"WarpCloud/walm/pkg/models/http"
	"github.com/emicklei/go-restful-openapi"
	tenantModel "WarpCloud/walm/pkg/models/tenant"
	"fmt"
	errorModel "WarpCloud/walm/pkg/models/error"
)

type TenantHandler struct {
	usecase tenant.UseCase
}

func RegisterTenantHandler(usecase tenant.UseCase) *restful.WebService {
	handler := TenantHandler{usecase:usecase}

	ws := new(restful.WebService)

	ws.Path(http.ApiV1 + "/tenant").
		Doc("租户相关操作").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"tenant"}

	ws.Route(ws.GET("/").To(handler.ListTenants).
		Doc("获取租户列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(tenantModel.TenantInfoList{}).
		Returns(200, "OK", tenantModel.TenantInfoList{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))


	ws.Route(ws.GET("/{tenantName}").To(handler.GetTenant).
		Doc("获取租户状态").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("tenantName", "租户名字").DataType("string")).
		Writes(tenantModel.TenantInfo{}).
		Returns(200, "OK", tenantModel.TenantInfo{}).
		Returns(404, "Not Found", http.ErrorMessageResponse{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.DELETE("/{tenantName}").To(handler.DeleteTenant).
		Doc("删除租户").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("tenantName", "租户名字").DataType("string")).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{tenantName}").To(handler.CreateTenant).
		Doc("创建租户").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("tenantName", "租户名字").DataType("string")).
		Reads(tenantModel.TenantParams{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.PUT("/{tenantName}").To(handler.UpdateTenant).
		Doc("更新租户信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("tenantName", "租户名字").DataType("string")).
		Reads(tenantModel.TenantParams{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	return ws
}

func (handler *TenantHandler)ListTenants(request *restful.Request, response *restful.Response) {
	tenantInfoList, err := handler.usecase.ListTenants()
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to list tenants : %s", err.Error()))
		return
	}
	response.WriteEntity(tenantInfoList)
}

func (handler *TenantHandler)CreateTenant(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("tenantName")
	tenantParams := new(tenantModel.TenantParams)
	err := request.ReadEntity(&tenantParams)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read tenant params : %s", err.Error()))
		return
	}
	err = handler.usecase.CreateTenant(tenantName, tenantParams)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to create tenant : %s", err.Error()))
		return
	}
}

func (handler *TenantHandler)GetTenant(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("tenantName")
	tenantInfo, err := handler.usecase.GetTenant(tenantName)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			httpUtils.WriteNotFoundResponse(response, -1, fmt.Sprintf("tenant %s is not found", tenantName))
			return
		}
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get tenant : %s", err.Error()))
		return
	}
	response.WriteEntity(tenantInfo)
}

func (handler *TenantHandler)DeleteTenant(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("tenantName")
	err := handler.usecase.DeleteTenant(tenantName)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to delete namespace %s: %s" , tenantName, err.Error()))
		return
	}

	response.WriteEntity(nil)
}

func (handler *TenantHandler)UpdateTenant(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("tenantName")
	tenantParams := new(tenantModel.TenantParams)
	err := request.ReadEntity(&tenantParams)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read tenant params : %s", err.Error()))
		return
	}
	err = handler.usecase.UpdateTenant(tenantName, tenantParams)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to update tenant : %s", err.Error()))
		return
	}
}