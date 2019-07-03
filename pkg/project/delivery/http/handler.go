package http

import (
	"WarpCloud/walm/pkg/project"
	"github.com/emicklei/go-restful"
	"WarpCloud/walm/pkg/models/http"
	"github.com/emicklei/go-restful-openapi"
	projectModel "WarpCloud/walm/pkg/models/project"
	"WarpCloud/walm/pkg/models/release"
	httpUtils "WarpCloud/walm/pkg/util/http"
	"fmt"
	errorModel "WarpCloud/walm/pkg/models/error"
)

type ProjectHandler struct {
	usecase project.UseCase
}

func RegisterProjectHandler(usecase project.UseCase) *restful.WebService {
	handler := &ProjectHandler{usecase: usecase}
	ws := new(restful.WebService)

	ws.Path(http.ApiV1 + "/project").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"project"}

	ws.Route(ws.GET("/").To(handler.ListProject).
		Doc("获取所有Project列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(projectModel.ProjectInfoList{}).
		Returns(200, "OK", projectModel.ProjectInfoList{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{namespace}").To(handler.ListProjectByNamespace).
		Doc("获取对应租户的Project列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Writes(projectModel.ProjectInfoList{}).
		Returns(200, "OK", projectModel.ProjectInfoList{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{namespace}/name/{project}").To(handler.GetProjectInfo).
		Doc("获取Project的详细信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("project", "Project名字").DataType("string")).
		Returns(200, "OK", projectModel.ProjectInfo{}).
		Returns(404, "Not Found", http.ErrorMessageResponse{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/name/{project}").To(handler.CreateProject).
		Doc("创建一个Project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("project", "Project名字").DataType("string")).
		Param(ws.QueryParameter("async", "异步与否").DataType("boolean").Required(false)).
		Param(ws.QueryParameter("timeoutSec", "超时时间").DataType("integer").Required(false)).
		Reads(projectModel.ProjectParams{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.DELETE("/{namespace}/name/{project}").To(handler.DeleteProject).
		Doc("删除一个Project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("project", "Project名字").DataType("string")).
		Param(ws.QueryParameter("async", "异步与否").DataType("boolean").Required(false)).
		Param(ws.QueryParameter("timeoutSec", "超时时间").DataType("integer").Required(false)).
		Param(ws.QueryParameter("deletePvcs", "是否删除Project Releases管理的statefulSet关联的所有pvc").DataType("boolean").Required(false)).
		Returns(200, "OK", nil).
		Returns(500, "Server Error", http.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/name/{project}/instance").To(handler.AddReleaseInProject).
		Doc("添加一个Project组件").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("project", "Project名字").DataType("string")).
		Param(ws.QueryParameter("async", "异步与否").DataType("boolean").Required(false)).
		Param(ws.QueryParameter("timeoutSec", "超时时间").DataType("integer").Required(false)).
		Reads(release.ReleaseRequestV2{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.PUT("/{namespace}/name/{project}/instance").To(handler.UpgradeReleaseInProject).
		Doc("升级一个Project组件").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("project", "Project名字").DataType("string")).
		Param(ws.QueryParameter("async", "异步与否").DataType("boolean").Required(false)).
		Param(ws.QueryParameter("timeoutSec", "超时时间").DataType("integer").Required(false)).
		Reads(release.ReleaseRequestV2{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/name/{project}/project").To(handler.AddReleasesInProject).
		Doc("添加多个Project组件").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("project", "Project名字").DataType("string")).
		Param(ws.QueryParameter("async", "异步与否").DataType("boolean").Required(false)).
		Param(ws.QueryParameter("timeoutSec", "超时时间").DataType("integer").Required(false)).
		Reads(projectModel.ProjectParams{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.DELETE("/{namespace}/name/{project}/instance/{release}").To(handler.DeleteReleaseInProject).
		Doc("删除一个Project组件").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("project", "Project名字").DataType("string")).
		Param(ws.PathParameter("release", "Release名字").DataType("string")).
		Param(ws.QueryParameter("async", "异步与否").DataType("boolean").Required(false)).
		Param(ws.QueryParameter("timeoutSec", "超时时间").DataType("integer").Required(false)).
		Param(ws.QueryParameter("deletePvcs", "是否删除release管理的statefulSet关联的所有pvc").DataType("boolean").Required(false)).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	return ws

}

func (handler *ProjectHandler) ListProject(request *restful.Request, response *restful.Response) {
	projectList, err := handler.usecase.ListProjects("")
	if err != nil {
		httpUtils.WriteErrorResponse(response,-1, fmt.Sprintf("failed to list all projects : %s", err.Error()))
		return
	}
	response.WriteEntity(projectList)
}

func (handler *ProjectHandler)ListProjectByNamespace(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("namespace")
	projectList, err := handler.usecase.ListProjects(tenantName)
	if err != nil {
		httpUtils.WriteErrorResponse(response,-1, fmt.Sprintf("failed to list projects in tenant %s : %s", tenantName, err.Error()))
		return
	}
	response.WriteEntity(projectList)
}

func (handler *ProjectHandler) CreateProject(request *restful.Request, response *restful.Response) {
	projectParams := new(projectModel.ProjectParams)
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	async, err := httpUtils.GetAsyncQueryParam(request)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("query param async value is not valid : %s", err.Error()))
		return
	}

	timeoutSec, err := httpUtils.GetTimeoutSecQueryParam(request)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("query param timeoutSec value is not valid : %s", err.Error()))
		return
	}

	err = request.ReadEntity(&projectParams)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body : %s", err.Error()))
		return
	}

	if projectParams.CommonValues == nil {
		projectParams.CommonValues = make(map[string]interface{})
	}
	if projectParams.Releases == nil {
		httpUtils.WriteErrorResponse(response,-1, "project params releases can not be empty")
		return
	}
	for _, releaseInfo := range projectParams.Releases {
		if releaseInfo.Dependencies == nil {
			releaseInfo.Dependencies = make(map[string]string)
		}
		if releaseInfo.ConfigValues == nil {
			releaseInfo.ConfigValues = make(map[string]interface{})
		}
	}

	err = handler.usecase.CreateProject(tenantName, projectName, projectParams, async, timeoutSec)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to create project : %s", err.Error()))
		return
	}
}

func (handler *ProjectHandler)GetProjectInfo(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	projectInfo, err := handler.usecase.GetProjectInfo(tenantName, projectName)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			httpUtils.WriteNotFoundResponse(response, -1, fmt.Sprintf("project %s/%s is not found", tenantName, projectName))
			return
		}
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get project info : %s", err.Error()))
		return
	}
	response.WriteEntity(projectInfo)
}

func (handler *ProjectHandler)DeleteProject(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	async, err := httpUtils.GetAsyncQueryParam(request)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("query param async value is not valid : %s", err.Error()))
		return
	}
	timeoutSec, err := httpUtils.GetTimeoutSecQueryParam(request)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("query param timeoutSec value is not valid : %s", err.Error()))
		return
	}
	deletePvcs, err := httpUtils.GetDeletePvcsQueryParam(request)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("query param deletePvcs value is not valid : %s", err.Error()))
		return
	}
	err = handler.usecase.DeleteProject(tenantName, projectName, async, timeoutSec, deletePvcs)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to delete project : %s", err.Error()))
		return
	}
}

func (handler *ProjectHandler) AddReleaseInProject(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	async, err := httpUtils.GetAsyncQueryParam(request)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("query param async value is not valid : %s", err.Error()))
		return
	}
	timeoutSec, err := httpUtils.GetTimeoutSecQueryParam(request)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("query param timeoutSec value is not valid : %s", err.Error()))
		return
	}
	releaseRequest := &release.ReleaseRequestV2{}
	err = request.ReadEntity(releaseRequest)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}
	err = handler.usecase.AddReleasesInProject(tenantName, projectName, &projectModel.ProjectParams{Releases: []*release.ReleaseRequestV2{releaseRequest}}, async, timeoutSec)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to add release in project : %s", err.Error()))
		return
	}
}

func (handler *ProjectHandler) UpgradeReleaseInProject(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	async, err := httpUtils.GetAsyncQueryParam(request)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("query param async value is not valid : %s", err.Error()))
		return
	}
	timeoutSec, err := httpUtils.GetTimeoutSecQueryParam(request)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("query param timeoutSec value is not valid : %s", err.Error()))
		return
	}
	releaseRequest := &release.ReleaseRequestV2{}
	err = request.ReadEntity(releaseRequest)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}
	err = handler.usecase.UpgradeReleaseInProject(tenantName, projectName, releaseRequest, async, timeoutSec)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to upgrade release in project : %s", err.Error()))
		return
	}
}

func (handler *ProjectHandler) AddReleasesInProject(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	async, err := httpUtils.GetAsyncQueryParam(request)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("query param async value is not valid : %s", err.Error()))
		return
	}
	timeoutSec, err := httpUtils.GetTimeoutSecQueryParam(request)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("query param timeoutSec value is not valid : %s", err.Error()))
		return
	}
	projectParams := &projectModel.ProjectParams{}
	err = request.ReadEntity(projectParams)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}
	err = handler.usecase.AddReleasesInProject(tenantName, projectName, projectParams, async, timeoutSec)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to add releases in project : %s", err.Error()))
		return
	}
}

func (handler *ProjectHandler) DeleteReleaseInProject(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	releaseName := request.PathParameter("release")
	async, err := httpUtils.GetAsyncQueryParam(request)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("query param async value is not valid : %s", err.Error()))
		return
	}
	timeoutSec, err := httpUtils.GetTimeoutSecQueryParam(request)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("query param timeoutSec value is not valid : %s", err.Error()))
		return
	}
	deletePvcs, err := httpUtils.GetDeletePvcsQueryParam(request)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("query param deletePvcs value is not valid : %s", err.Error()))
		return
	}
	err = handler.usecase.RemoveReleaseInProject(tenantName, projectName, releaseName, async, timeoutSec, deletePvcs)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to delete release in project : %s", err.Error()))
		return
	}
}
