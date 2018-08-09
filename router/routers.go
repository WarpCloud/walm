package router

import (
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
)

var APIPATH = "/walm/api/v1"

func InitRootRouter() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path("/").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"root"}

	ws.Route(ws.GET("/readiness").To(readinessProbe).
		Doc("服务Ready状态检查").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	ws.Route(ws.GET("/liveniess").To(livenessProbe).
		Doc("服务Live状态检查").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	return ws
}

func InitTenantRouter() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path(APIPATH + "/tenant").
		Doc("租户相关操作").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"tenant"}

	ws.Route(ws.GET("/").To(readinessProbe).
		Doc("获取租户列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	ws.Route(ws.GET("/{tenantName}").To(readinessProbe).
		Doc("获取租户状态").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("tenantName", "租户名字").DataType("string")).
		Writes("").
		Returns(200, "OK", nil).
		Returns(400, "Invalid Name", nil).
		Returns(500, "Server Error", ""))

	ws.Route(ws.DELETE("/{tenantName}").To(readinessProbe).
		Doc("删除租户").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("tenantName", "租户名字").DataType("string")).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	ws.Route(ws.GET("/{tenantName}/quotas").To(readinessProbe).
		Doc("获取租户配额").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("tenantName", "租户名字").DataType("string")).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	ws.Route(ws.PUT("/{tenantName}/quotas").To(readinessProbe).
		Doc("更新租户配额").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("tenantName", "租户名字").DataType("string")).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	return ws
}

func InitNodeRouter() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path(APIPATH + "/node").
		Doc("Kubernetes节点相关操作").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"node"}

	ws.Route(ws.GET("/").To(readinessProbe).
		Doc("获取节点列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	ws.Route(ws.PUT("/{node}/labels").To(readinessProbe).
		Doc("修改节点Labels").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	return ws
}

func InitInstanceRouter() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path(APIPATH + "/instance").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"instance"}

	ws.Route(ws.GET("/").To(readinessProbe).
		Doc("获取所有Release列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	ws.Route(ws.GET("/{namespace}").To(readinessProbe).
		Doc("获取Namepaces下的所有Release列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	ws.Route(ws.GET("/{namespace}/name/{appname}").To(readinessProbe).
		Doc("获取对应Release的详细信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	ws.Route(ws.PUT("/{namespace}/name/{appname}").To(readinessProbe).
		Doc("更改一个Release").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	ws.Route(ws.DELETE("/{namespace}/name/{appname}").To(readinessProbe).
		Doc("删除一个Release").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	ws.Route(ws.POST("/{namespace}/name/{appname}").To(readinessProbe).
		Doc("创建一个Release").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	ws.Route(ws.POST("/{namespace}/name/{appname}/version/{version}/rollback").To(readinessProbe).
		Doc("RollBack　Release版本").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	return ws
}

func InitClusterRouter() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path(APIPATH + "/cluster").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"cluster"}

	ws.Route(ws.GET("/").To(readinessProbe).
		Doc("获取所有Cluster列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	ws.Route(ws.GET("/{namespace}/name/{cluster}").To(readinessProbe).
		Doc("获取对应Cluster的详细信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	ws.Route(ws.POST("/{namespace}/name/{cluster}").To(readinessProbe).
		Doc("新创建一个Cluster的详细信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	ws.Route(ws.DELETE("/{namespace}/name/{cluster}").To(readinessProbe).
		Doc("删除一个Release").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	ws.Route(ws.POST("/{namespace}/name/{cluster}/instance").To(readinessProbe).
		Doc("新添加一个Release组件").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	ws.Route(ws.POST("/{namespace}/name/{cluster}/list").To(readinessProbe).
		Doc("新添加一组Release组件").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	return ws
}

func InitPodRouter() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path(APIPATH + "/pod").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"pod"}

	ws.Route(ws.GET("/{namespace}/{pod}/shell/{container}").To(readinessProbe).
		Doc("登录Pod对应的Shell").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes("").
		Returns(200, "OK", nil).
		Returns(500, "Server Error", ""))

	return ws
}

func readinessProbe(request *restful.Request, response *restful.Response) {
	response.WriteEntity("OK")
}

func livenessProbe(request *restful.Request, response *restful.Response) {
	response.WriteEntity("OK")
}

