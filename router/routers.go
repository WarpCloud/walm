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
		Returns(404, "Not Found", walmtypes.ErrorMessageResponse{}).
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

	ws.Route(ws.PUT("/{tenantName}").To(v1.UpdateTenant).
		Doc("更新租户信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("tenantName", "租户名字").DataType("string")).
		Reads(tenanttypes.TenantParams{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	return ws
}

func InitSecretRouter() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path(APIPATH + "/secret").
		Doc("Kubernetes Secret相关操作").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"secret"}

	ws.Route(ws.GET("/{namespace}").To(v1.GetSecrets).
		Doc("获取Namepace下的所有Secret列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Writes(k8stypes.WalmSecretList{}).
		Returns(200, "OK", k8stypes.WalmSecretList{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{namespace}/name/{secretname}").To(v1.GetSecret).
		Doc("获取对应Secret的详细信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("secretname", "secret名字").DataType("string")).
		Writes(k8stypes.WalmSecret{}).
		Returns(200, "OK", k8stypes.WalmSecret{}).
		Returns(404, "Not Found", walmtypes.ErrorMessageResponse{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.DELETE("/{namespace}/name/{secretname}").To(v1.DeleteSecret).
		Doc("删除一个Secret").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("secretname", "Secret名字").DataType("string")).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}").To(v1.CreateSecret).
		Doc("创建一个Secret").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Reads(walmtypes.CreateSecretRequestBody{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.PUT("/{namespace}").To(v1.UpdateSecret).
		Doc("更新一个Secret").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Reads(walmtypes.CreateSecretRequestBody{}).
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
		Param(ws.QueryParameter("labelselector", "节点标签过滤").DataType("string")).
		Writes(k8stypes.WalmNodeList{}).
		Returns(200, "OK", k8stypes.WalmNodeList{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{nodename}").To(v1.GetNode).
		Doc("获取节点详细信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("nodename", "节点名字").DataType("string")).
		Writes(k8stypes.WalmNode{}).
		Returns(200, "OK", k8stypes.WalmNode{}).
		Returns(404, "Not Found", walmtypes.ErrorMessageResponse{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{nodename}/labels").To(v1.LabelNode).
		Doc("修改节点Labels").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("nodename", "节点名字").DataType("string")).
		Reads(walmtypes.LabelNodeRequestBody{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{nodename}/annotations").To(v1.AnnotateNode).
		Doc("修改节点Annotations").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("nodename", "节点名字").DataType("string")).
		Reads(walmtypes.AnnotateNodeRequestBody{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	return ws
}

func InitReleaseRouter() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path(APIPATH + "/release").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"release"}

	ws.Route(ws.GET("/").To(v1.ListRelease).
		Doc("获取所有Release列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(releasetypes.ReleaseInfoList{}).
		Returns(200, "OK", releasetypes.ReleaseInfoList{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{namespace}").To(v1.ListReleaseByNamespace).
		Doc("获取Namepaces下的所有Release列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Writes(releasetypes.ReleaseInfoList{}).
		Returns(200, "OK", releasetypes.ReleaseInfoList{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{namespace}/name/{release}").To(v1.GetRelease).
		Doc("获取对应Release的详细信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("release", "Release名字").DataType("string")).
		Writes(releasetypes.ReleaseInfo{}).
		Returns(200, "OK", releasetypes.ReleaseInfo{}).
		Returns(404, "Not Found", walmtypes.ErrorMessageResponse{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.PUT("/{namespace}").To(v1.UpgradeRelease).
		Doc("升级一个Release").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Reads(releasetypes.ReleaseRequest{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.DELETE("/{namespace}/name/{release}").To(v1.DeleteRelease).
		Doc("删除一个Release").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("release", "Release名字").DataType("string")).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}").To(v1.InstallRelease).
		Doc("安装一个Release").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Reads(releasetypes.ReleaseRequest{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/name/{release}/version/{version}/rollback").To(v1.RollBackRelease).
		Doc("RollBack　Release版本").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("release", "Release名字").DataType("string")).
		Param(ws.PathParameter("version", "版本号").DataType("string")).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))
	ws.Route(ws.POST("/{namespace}/name/{release}/restart").To(v1.RestartRelease).
		Doc("Restart　Release关联的所有pod").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("release", "Release名字").DataType("string")).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	return ws
}

func InitProjectRouter() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path(APIPATH + "/project").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"project"}

	ws.Route(ws.GET("/").To(v1.ListProjectAllNamespaces).
		Doc("获取所有Project列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(releasetypes.ProjectInfoList{}).
		Returns(200, "OK", releasetypes.ProjectInfoList{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{namespace}").To(v1.ListProjectByNamespace).
		Doc("获取对应租户的Project列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Writes(releasetypes.ProjectInfoList{}).
		Returns(200, "OK", releasetypes.ProjectInfoList{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{namespace}/name/{project}").To(v1.GetProjectInfo).
		Doc("获取Project的详细信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("project", "Project名字").DataType("string")).
		Returns(200, "OK", releasetypes.ProjectInfo{}).
		Returns(404, "Not Found", walmtypes.ErrorMessageResponse{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/name/{project}").To(v1.DeployProject).
		Doc("创建一个Project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("project", "Project名字").DataType("string")).
		Param(ws.QueryParameter("async", "异步与否").DataType("boolean").Required(false)).
		Param(ws.QueryParameter("timeoutSec", "超时时间").DataType("integer").Required(false)).
		Reads(releasetypes.ProjectParams{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.DELETE("/{namespace}/name/{project}").To(v1.DeleteProject).
		Doc("删除一个Project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("project", "Project名字").DataType("string")).
		Param(ws.QueryParameter("async", "异步与否").DataType("boolean").Required(false)).
		Param(ws.QueryParameter("timeoutSec", "超时时间").DataType("integer").Required(false)).
		Returns(200, "OK", nil).
		Returns(500, "Server Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/name/{project}/instance").To(v1.DeployInstanceInProject).
		Doc("添加一个Project组件").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("project", "Project名字").DataType("string")).
		Param(ws.QueryParameter("async", "异步与否").DataType("boolean").Required(false)).
		Param(ws.QueryParameter("timeoutSec", "超时时间").DataType("integer").Required(false)).
		Reads(releasetypes.ReleaseRequest{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/name/{project}/project").To(v1.DeployProjectInProject).
		Doc("添加多个Project组件").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("project", "Project名字").DataType("string")).
		Param(ws.QueryParameter("async", "异步与否").DataType("boolean").Required(false)).
		Param(ws.QueryParameter("timeoutSec", "超时时间").DataType("integer").Required(false)).
		Reads(releasetypes.ProjectParams{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.DELETE("/{namespace}/name/{project}/instance/{release}").To(v1.DeleteInstanceInProject).
		Doc("删除一个Project组件").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("project", "Project名字").DataType("string")).
		Param(ws.PathParameter("release", "Release名字").DataType("string")).
		Param(ws.QueryParameter("async", "异步与否").DataType("boolean").Required(false)).
		Param(ws.QueryParameter("timeoutSec", "超时时间").DataType("integer").Required(false)).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	return ws
}

func InitChartRouter() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path(APIPATH + "/chart").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"chart"}

	ws.Route(ws.GET("/repolist").To(v1.GetRepoList).
		Doc("获取chart-repo列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(releasetypes.RepoInfoList{}).
		Returns(200, "OK", releasetypes.RepoInfoList{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{repo-name}/list").To(v1.GetChartList).
		Doc("获取chart列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("repo-name", "Repo名字").DataType("string")).
		Writes(releasetypes.ChartInfoList{}).
		Returns(200, "OK", releasetypes.ChartInfoList{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{repo-name}/chart/{chart-name}").To(v1.GetChartInfo).
		Doc("获取chart详细信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("repo-name", "Repo名字").DataType("string")).
		Param(ws.PathParameter("chart-name", "Chart名字").DataType("string")).
		Param(ws.QueryParameter("chart-version", "chart版本").DataType("string").DefaultValue("")).
		Writes(releasetypes.ChartInfo{}).
		Returns(200, "OK", releasetypes.ChartInfo{}).
		Returns(404, "Not Found", walmtypes.ErrorMessageResponse{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	return ws
}

func InitPodRouter() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path(APIPATH + "/pod").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"pod"}

	//ws.Route(ws.GET("/{namespace}/{pod}/shell/{container}").To(v1.ExecShell).
	//	Doc("登录Pod对应的Shell").
	//	Metadata(restfulspec.KeyOpenAPITags, tags).
	//	Writes("").
	//	Returns(200, "OK", nil).
	//	Returns(500, "Internal Error", ""))

	ws.Route(ws.GET("/{namespace}/name/{pod}/events").To(v1.GetPodEvents).
		Doc("获取Pod对应的事件").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("pod", "pod名字").DataType("string")).
		Writes(k8stypes.WalmEventList{}).
		Returns(200, "OK", k8stypes.WalmEventList{}).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{namespace}/name/{pod}/logs").To(v1.GetPodLogs).
		Doc("获取Pod对应的事件").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("pod", "pod名字").DataType("string")).
		Param(ws.QueryParameter("container", "container名字").DataType("string")).
		Param(ws.QueryParameter("tail", "最后几行").DataType("integer")).
		Writes("").
		Returns(200, "OK", "").
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/name/{pod}/restart").To(v1.RestartPod).
		Doc("重启Pod").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("pod", "pod名字").DataType("string")).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", walmtypes.ErrorMessageResponse{}))

	return ws
}

func readinessProbe(request *restful.Request, response *restful.Response) {
	response.WriteEntity("OK")
}

func livenessProbe(request *restful.Request, response *restful.Response) {
	response.WriteEntity("OK")
}

