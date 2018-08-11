package router

import (
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"

	"walm/router/api/v1"
	"walm/router/middleware"
	releasetypes "walm/pkg/release"
	walmtypes "walm/router/api"
	tenanttypes "walm/pkg/tenant"
	k8stypes "walm/pkg/k8s/adaptor"
)

var APIPATH = "/api/v1"

func InitRootRouter() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path("/").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"root"}

	ws.Route(ws.GET("/readiness").To(readinessProbe).
		Doc("服务Ready状态检查").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.GET("/liveniess").To(livenessProbe).
		Doc("服务Live状态检查").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.GET("/stats").To(middleware.ServerStatsData).
		Doc("获取服务Stats").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	return ws
}

func InitTenantRouter() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path(APIPATH + "/tenant").
		Doc("租户相关操作").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"tenant"}

	ws.Route(ws.GET("/").To(v1.ListTenants).
		Doc("获取租户列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(tenanttypes.TenantInfoList{}).
		Returns(200, "OK", tenanttypes.TenantInfoList{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))


	ws.Route(ws.GET("/{tenantName}").To(v1.GetTenant).
		Doc("获取租户状态").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("tenantName", "租户名字").DataType("string")).
		Writes(tenanttypes.TenantInfo{}).
		Returns(200, "OK", tenanttypes.TenantInfo{}).
		Returns(400, "Invalid Name", walmtypes.ErrorMessageResponse{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.DELETE("/{tenantName}").To(v1.DeleteTenant).
		Doc("删除租户").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("tenantName", "租户名字").DataType("string")).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{tenantName}").To(v1.CreateTenant).
		Doc("创建租户").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("tenantName", "租户名字").DataType("string")).
		Reads(tenanttypes.TenantParams{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.PUT("/{tenantName}/quotas").To(v1.UpdateQuotas).
		Doc("更新租户配额").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("tenantName", "租户名字").DataType("string")).
		Reads(tenanttypes.TenantQuotaInfo{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	return ws
}

func InitNodeRouter() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path(APIPATH + "/node").
		Doc("Kubernetes节点相关操作").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"node"}

	ws.Route(ws.GET("/").To(v1.GetNodes).
		Doc("获取节点列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(k8stypes.WalmNodeList{}).
		Returns(200, "OK", k8stypes.WalmNodeList{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.PUT("/{node}").To(v1.GetNode).
		Doc("获取节点详细信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("node", "节点名字").DataType("string")).
		Writes(k8stypes.WalmNode{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.PUT("/{node}/labels").To(v1.PatchNodeLabels).
		Doc("修改节点Labels").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("node", "节点名字").DataType("string")).
		Writes(k8stypes.WalmNode{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	return ws
}

func InitInstanceRouter() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path(APIPATH + "/instance").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"instance"}

	ws.Route(ws.GET("/").To(v1.ListInstanceAllNamespaces).
		Doc("获取所有Release列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(releasetypes.ReleaseInfoList{}).
		Returns(200, "OK", releasetypes.ReleaseInfoList{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{namespace}").To(v1.ListInstanceByNamespace).
		Doc("获取Namepaces下的所有Release列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Writes(releasetypes.ReleaseInfoList{}).
		Returns(200, "OK", releasetypes.ReleaseInfoList{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{namespace}/name/{release}").To(v1.GetInstanceInfo).
		Doc("获取对应Release的详细信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("release", "Release名字").DataType("string")).
		Writes(releasetypes.ReleaseInfo{}).
		Returns(200, "OK", releasetypes.ReleaseInfo{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.PUT("/{namespace}/name/{release}").To(v1.UpdateInstance).
		Doc("更改一个Release").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("release", "Release名字").DataType("string")).
		Reads(releasetypes.ReleaseRequest{}).
		Writes(releasetypes.ReleaseInfo{}).
		Returns(200, "OK", releasetypes.ReleaseInfo{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.DELETE("/{namespace}/name/{release}").To(v1.DeleteInstance).
		Doc("删除一个Release").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("release", "Release名字").DataType("string")).
		Reads(releasetypes.ReleaseRequest{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/name/{release}").To(v1.DeployInstance).
		Doc("创建一个Release").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("release", "Release名字").DataType("string")).
		Reads(releasetypes.ReleaseRequest{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/name/{release}/version/{version}/rollback").To(v1.RollBackInstance).
		Doc("RollBack　Release版本").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("release", "Release名字").DataType("string")).
		Param(ws.PathParameter("version", "版本号").DataType("string")).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	return ws
}

func InitProjectRouter() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path(APIPATH + "/project").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"Project"}

	ws.Route(ws.GET("/").To(v1.GetProjectAllNamespaces).
		Doc("获取所有Project列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(releasetypes.ProjectInfoList{}).
		Returns(200, "OK", releasetypes.ProjectInfoList{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{namespace}").To(v1.GetProjectByNamespace).
		Doc("获取对应租户的Project列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Writes(releasetypes.ProjectInfoList{}).
		Returns(200, "OK", releasetypes.ProjectInfoList{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/name/{project}").To(v1.DeployProject).
		Doc("新创建一个Project的详细信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(releasetypes.ProjectParams{}).
		Returns(200, "OK", releasetypes.ProjectParams{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.DELETE("/{namespace}/name/{project}").To(v1.DeleteProject).
		Doc("删除一个Project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("project", "Project名字").DataType("string")).
		Returns(200, "OK", nil).
		Returns(500, "Server Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/name/{project}/instance").To(v1.DeployInstanceInProject).
		Doc("新添加一个Project组件").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("project", "Project名字").DataType("string")).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	return ws
}

func InitPodRouter() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path(APIPATH + "/pod").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"pod"}

	ws.Route(ws.GET("/{namespace}/{pod}/shell/{container}").To(v1.ExecShell).
		Doc("登录Pod对应的Shell").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", ""))

	return ws
}

func readinessProbe(request *restful.Request, response *restful.Response) {
	response.WriteEntity("OK")
}

func livenessProbe(request *restful.Request, response *restful.Response) {
	response.WriteEntity("OK")
}

