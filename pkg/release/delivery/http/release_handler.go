package http

import (
	"WarpCloud/walm/pkg/release"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	releaseModel "WarpCloud/walm/pkg/models/release"
	errorModel "WarpCloud/walm/pkg/models/error"
	"WarpCloud/walm/pkg/models/http"
	httpUtils "WarpCloud/walm/pkg/util/http"
	"fmt"
	"WarpCloud/walm/pkg/models/common"
	"encoding/json"
)

type ReleaseHandler struct {
	usecase release.UseCase
}

func RegisterReleaseHandler(usecase release.UseCase) *restful.WebService {
	releaseHandler := &ReleaseHandler{usecase: usecase}
	ws := new(restful.WebService)

	ws.Path(http.ApiV1 + "/release").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"release"}

	ws.Route(ws.GET("/").To(releaseHandler.ListRelease).
		Doc("获取所有Release列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.QueryParameter("labelselector", "标签过滤").DataType("string")).
		Writes(releaseModel.ReleaseInfoV2List{}).
		Returns(200, "OK", releaseModel.ReleaseInfoV2List{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{namespace}").To(releaseHandler.ListReleaseByNamespace).
		Doc("获取Namepaces下的所有Release列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.QueryParameter("labelselector", "标签过滤").DataType("string")).
		Writes(releaseModel.ReleaseInfoV2List{}).
		Returns(200, "OK", releaseModel.ReleaseInfoV2List{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{namespace}/name/{release}").To(releaseHandler.GetRelease).
		Doc("获取对应Release的详细信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("release", "Release名字").DataType("string")).
		Writes(releaseModel.ReleaseInfoV2{}).
		Returns(200, "OK", releaseModel.ReleaseInfoV2{}).
		Returns(404, "Not Found", http.ErrorMessageResponse{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.PUT("/{namespace}").To(releaseHandler.UpgradeRelease).
		Doc("升级一个Release").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.QueryParameter("async", "异步与否").DataType("boolean").Required(false)).
		Param(ws.QueryParameter("timeoutSec", "超时时间").DataType("integer").Required(false)).
		Reads(releaseModel.ReleaseRequestV2{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.PUT("/{namespace}/withchart").To(releaseHandler.UpgradeReleaseWithChart).
		Consumes("multipart/form-data").
		Doc("用本地chart升级一个Release").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.FormParameter("release", "Release名字").DataType("string").Required(true)).
		Param(ws.FormParameter("chart", "chart").DataType("file").Required(true)).
		Param(ws.FormParameter("body", "request").DataType("string")).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.DELETE("/{namespace}/name/{release}").To(releaseHandler.DeleteRelease).
		Doc("删除一个Release").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("release", "Release名字").DataType("string")).
		Param(ws.QueryParameter("async", "异步与否").DataType("boolean").Required(false)).
		Param(ws.QueryParameter("timeoutSec", "超时时间").DataType("integer").Required(false)).
		Param(ws.QueryParameter("deletePvcs", "是否删除release管理的statefulSet关联的所有pvc").DataType("boolean").Required(false)).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}").To(releaseHandler.InstallRelease).
		Doc("安装一个Release").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.QueryParameter("async", "异步与否").DataType("boolean").Required(false)).
		Param(ws.QueryParameter("timeoutSec", "超时时间").DataType("integer").Required(false)).
		Reads(releaseModel.ReleaseRequestV2{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/withchart").Consumes().To(releaseHandler.InstallReleaseWithChart).
		Consumes("multipart/form-data").
		Doc("用本地chart安装一个Release").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.FormParameter("release", "Release名字").DataType("string").Required(true)).
		Param(ws.FormParameter("chart", "chart").DataType("file").Required(true)).
		Param(ws.FormParameter("body", "request").DataType("string")).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/dryrun").To(releaseHandler.DryRunRelease).
		Doc("模拟安装一个Release").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Reads(releaseModel.ReleaseRequestV2{}).
		Returns(200, "OK", []map[string]interface{}{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/dryrun/withchart").Consumes().To(releaseHandler.DryRunReleaseWithChart).
		Consumes("multipart/form-data").
		Doc("模拟用本地chart安装一个Release").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.FormParameter("release", "Release名字").DataType("string").Required(true)).
		Param(ws.FormParameter("chart", "chart").DataType("file").Required(true)).
		Param(ws.FormParameter("body", "request").DataType("string")).
		Returns(200, "OK", []map[string]interface{}{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/dryrun/resources").To(releaseHandler.ComputeResourcesByDryRunRelease).
		Doc("模拟计算安装一个Release需要多少资源").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Reads(releaseModel.ReleaseRequestV2{}).
		Returns(200, "OK", releaseModel.ReleaseResources{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/dryrun/withchart/resources").Consumes().To(releaseHandler.ComputeResourcesByDryRunReleaseWithChart).
		Consumes("multipart/form-data").
		Doc("模拟计算用本地chart安装一个Release需要多少资源").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.FormParameter("release", "Release名字").DataType("string").Required(true)).
		Param(ws.FormParameter("chart", "chart").DataType("file").Required(true)).
		Param(ws.FormParameter("body", "request").DataType("string")).
		Returns(200, "OK", releaseModel.ReleaseResources{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	//ws.Route(ws.POST("/{namespace}/name/{release}/version/{version}/rollback").To(releaseHandler.RollBackRelease).
	//	Doc("RollBack　Release版本").
	//	Metadata(restfulspec.KeyOpenAPITags, tags).
	//	Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
	//	Param(ws.PathParameter("release", "Release名字").DataType("string")).
	//	Param(ws.PathParameter("version", "版本号").DataType("string")).
	//	Returns(200, "OK", nil).
	//	Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/name/{release}/restart").To(releaseHandler.RestartRelease).
		Doc("Restart　Release关联的所有pod").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("release", "Release名字").DataType("string")).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/name/{release}/pause").To(releaseHandler.PauseRelease).
		Doc("暂停Release服务").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("release", "Release名字").DataType("string")).
		Param(ws.QueryParameter("async", "异步与否").DataType("boolean").Required(false)).
		Param(ws.QueryParameter("timeoutSec", "超时时间").DataType("integer").Required(false)).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/name/{release}/recover").To(releaseHandler.RecoverRelease).
		Doc("恢复Release服务").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("release", "Release名字").DataType("string")).
		Param(ws.QueryParameter("async", "异步与否").DataType("boolean").Required(false)).
		Param(ws.QueryParameter("timeoutSec", "超时时间").DataType("integer").Required(false)).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	return ws
}

func (handler *ReleaseHandler) DeleteRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("release")
	deletePvcs, err := httpUtils.GetDeletePvcsQueryParam(request)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("query param deletePvcs value is not valid : %s", err.Error()))
		return
	}
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

	err = handler.usecase.DeleteRelease(namespace, name, deletePvcs, async, timeoutSec)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to delete release: %s", err.Error()))
		return
	}
}

func (handler *ReleaseHandler) InstallRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
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
	releaseRequest := &releaseModel.ReleaseRequestV2{}
	err = request.ReadEntity(releaseRequest)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}
	err = handler.usecase.InstallUpgradeRelease(namespace, releaseRequest, nil, async, timeoutSec, nil)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to install release: %s", err.Error()))
	}
}

func (handler *ReleaseHandler) InstallReleaseWithChart(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	chartArchive, _, err := request.Request.FormFile("chart")
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read chart archive: %s", err.Error()))
		return
	}
	defer chartArchive.Close()
	chartFiles, err := common.LoadArchive(chartArchive)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to load chart archive: %s", err.Error()))
		return
	}
	releaseName := request.Request.FormValue("release")
	body := request.Request.FormValue("body")
	releaseRequest := &releaseModel.ReleaseRequestV2{}
	if body != "" {
		err = json.Unmarshal([]byte(body), releaseRequest)
		if err != nil {
			httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read release request: %s", err.Error()))
			return
		}
	}
	releaseRequest.Name = releaseName

	err = handler.usecase.InstallUpgradeRelease(namespace, releaseRequest, chartFiles, false, 0, nil)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to install release: %s", err.Error()))
	}
}

func (handler *ReleaseHandler) DryRunRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	releaseRequest := &releaseModel.ReleaseRequestV2{}
	err := request.ReadEntity(releaseRequest)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}
	manifest, err := handler.usecase.DryRunRelease(namespace, releaseRequest, nil)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to dry run release: %s", err.Error()))
		return
	}
	response.WriteEntity(manifest)
}

func (handler *ReleaseHandler) ComputeResourcesByDryRunRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	releaseRequest := &releaseModel.ReleaseRequestV2{}
	err := request.ReadEntity(releaseRequest)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}
	resources, err := handler.usecase.ComputeResourcesByDryRunRelease(namespace, releaseRequest, nil)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to compute resources by dry run release: %s", err.Error()))
		return
	}
	response.WriteEntity(resources)
}

func (handler *ReleaseHandler) DryRunReleaseWithChart(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	chartArchive, _, err := request.Request.FormFile("chart")
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read chart archive: %s", err.Error()))
		return
	}
	defer chartArchive.Close()
	chartFiles, err := common.LoadArchive(chartArchive)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to load chart archive: %s", err.Error()))
		return
	}
	releaseName := request.Request.FormValue("release")
	body := request.Request.FormValue("body")
	releaseRequest := &releaseModel.ReleaseRequestV2{}
	if body != "" {
		err = json.Unmarshal([]byte(body), releaseRequest)
		if err != nil {
			httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read release request: %s", err.Error()))
			return
		}
	}
	releaseRequest.Name = releaseName

	manifest, err := handler.usecase.DryRunRelease(namespace, releaseRequest, chartFiles)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to dry run install release: %s", err.Error()))
		return
	}
	response.WriteEntity(manifest)
}

func (handler *ReleaseHandler) ComputeResourcesByDryRunReleaseWithChart(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	chartArchive, _, err := request.Request.FormFile("chart")
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read chart archive: %s", err.Error()))
		return
	}
	defer chartArchive.Close()
	chartFiles, err := common.LoadArchive(chartArchive)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to load chart archive: %s", err.Error()))
		return
	}
	releaseName := request.Request.FormValue("release")
	body := request.Request.FormValue("body")
	releaseRequest := &releaseModel.ReleaseRequestV2{}
	if body != "" {
		err = json.Unmarshal([]byte(body), releaseRequest)
		if err != nil {
			httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read release request: %s", err.Error()))
			return
		}
	}

	releaseRequest.Name = releaseName

	manifest, err := handler.usecase.ComputeResourcesByDryRunRelease(namespace, releaseRequest, chartFiles)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to compute resources by dry run install release: %s", err.Error()))
		return
	}
	response.WriteEntity(manifest)
}

func (handler *ReleaseHandler) UpgradeRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
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
	releaseRequest := &releaseModel.ReleaseRequestV2{}
	err = request.ReadEntity(releaseRequest)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}
	err = handler.usecase.InstallUpgradeRelease(namespace, releaseRequest, nil, async, timeoutSec, nil)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to upgrade release: %s", err.Error()))
	}
}

func (handler *ReleaseHandler) UpgradeReleaseWithChart(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	releaseName := request.Request.FormValue("release")
	chartArchive, _, err := request.Request.FormFile("chart")
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read chart archive: %s", err.Error()))
		return
	}
	defer chartArchive.Close()
	chartFiles, err := common.LoadArchive(chartArchive)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to load chart archive: %s", err.Error()))
		return
	}

	body := request.Request.FormValue("body")
	releaseRequest := &releaseModel.ReleaseRequestV2{}

	if body != "" {
		err = json.Unmarshal([]byte(body), releaseRequest)
		if err != nil {
			httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read release request: %s", err.Error()))
			return
		}
	}

	releaseRequest.Name = releaseName

	err = handler.usecase.InstallUpgradeRelease(namespace, releaseRequest, chartFiles, false, 0, nil)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to upgrade release: %s", err.Error()))
	}
}

func (handler *ReleaseHandler) ListReleaseByNamespace(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	labelSelectorStr := request.QueryParameter("labelselector")
	var infos []*releaseModel.ReleaseInfoV2
	var err error
	if labelSelectorStr == "" {
		infos, err = handler.usecase.ListReleases(namespace)
		if err != nil {
			httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to list release: %s", err.Error()))
			return
		}
	} else {
		infos, err = handler.usecase.ListReleasesByLabels(namespace, labelSelectorStr)
		if err != nil {
			httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to list release: %s", err.Error()))
			return
		}
	}

	response.WriteEntity(releaseModel.ReleaseInfoV2List{len(infos), infos})
}

func (handler *ReleaseHandler) ListRelease(request *restful.Request, response *restful.Response) {
	labelSelectorStr := request.QueryParameter("labelselector")
	var infos []*releaseModel.ReleaseInfoV2
	var err error
	if labelSelectorStr == "" {
		infos, err = handler.usecase.ListReleases("")
		if err != nil {
			httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to list release: %s", err.Error()))
			return
		}
	} else {
		infos, err = handler.usecase.ListReleasesByLabels("", labelSelectorStr)
		if err != nil {
			httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to list release: %s", err.Error()))
			return
		}
	}

	response.WriteEntity(releaseModel.ReleaseInfoV2List{len(infos), infos})
}

func (handler *ReleaseHandler) GetRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("release")
	info, err := handler.usecase.GetRelease(namespace, name)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			httpUtils.WriteNotFoundResponse(response, -1, fmt.Sprintf("release %s is not found", name))
			return
		}
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get release %s: %s", name, err.Error()))
		return
	}
	response.WriteEntity(info)
}

func (handler *ReleaseHandler) RestartRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("release")
	err := handler.usecase.RestartRelease(namespace, name)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to restart release %s: %s", name, err.Error()))
		return
	}
}

func (handler *ReleaseHandler) PauseRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("release")
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
	err = handler.usecase.PauseRelease(namespace, name, async, timeoutSec)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to pause release %s: %s", name, err.Error()))
		return
	}
}

func (handler *ReleaseHandler) RecoverRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("release")
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
	err = handler.usecase.RecoverRelease(namespace, name, async, timeoutSec)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to recover release %s: %s", name, err.Error()))
		return
	}
}
