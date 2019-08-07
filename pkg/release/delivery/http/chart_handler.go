package http

import (
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"WarpCloud/walm/pkg/models/release"
	errorModel "WarpCloud/walm/pkg/models/error"
	"WarpCloud/walm/pkg/models/http"
	httpUtils "WarpCloud/walm/pkg/util/http"
	"fmt"
	"WarpCloud/walm/pkg/helm"
)

type ChartHandler struct {
	helm helm.Helm
}

func RegisterChartHandler(helm helm.Helm) *restful.WebService {
	handler := &ChartHandler{helm: helm}

	ws := new(restful.WebService)

	ws.Path(http.ApiV1 + "/chart").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON, restful.MIME_XML)
	tags := []string{"chart"}

	ws.Route(ws.GET("/repolist").To(handler.GetRepoList).
		Doc("获取chart-repo列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(release.RepoInfoList{}).
		Returns(200, "OK", release.RepoInfoList{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{repo-name}/list").To(handler.GetChartList).
		Doc("获取chart列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("repo-name", "Repo名字").DataType("string")).
		Writes(release.ChartInfoList{}).
		Returns(200, "OK", release.ChartInfoList{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{repo-name}/chart/{chart-name}").To(handler.GetChartInfo).
		Doc("获取chart详细信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("repo-name", "Repo名字").DataType("string")).
		Param(ws.PathParameter("chart-name", "Chart名字").DataType("string")).
		Param(ws.QueryParameter("chart-version", "chart版本").DataType("string").DefaultValue("")).
		Writes(release.ChartInfo{}).
		Returns(200, "OK", release.ChartInfo{}).
		Returns(404, "Not Found", http.ErrorMessageResponse{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.GET("/image/").To(handler.GetChartInfoByImage).
		Doc("获取chart image的详细信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.QueryParameter("chart-image", "chart image url").DataType("string").Required(true)).
		Writes(release.ChartInfo{}).
		Returns(200, "OK", release.ChartInfo{}).
		Returns(404, "Not Found", http.ErrorMessageResponse{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{repo-name}/chart/{chart-name}/icon").To(handler.GetChartIcon).
		Doc("获取chart图标信息").
		Produces("multipart/form-data").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("repo-name", "Repo名字").DataType("string")).
		Param(ws.PathParameter("chart-name", "Chart名字").DataType("string")).
		Param(ws.QueryParameter("chart-version", "chart版本").DataType("string").DefaultValue("")).
		Writes("").
		Returns(200, "OK", "").
		Returns(404, "Not Found", http.ErrorMessageResponse{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{repo-name}/chart/{chart-name}/advantage").To(handler.GetChartAdvantage).
		Doc("获取chart产品优势信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("repo-name", "Repo名字").DataType("string")).
		Param(ws.PathParameter("chart-name", "Chart名字").DataType("string")).
		Param(ws.QueryParameter("chart-version", "chart版本").DataType("string").DefaultValue("")).
		Writes("").
		Returns(200, "OK", "").
		Returns(404, "Not Found", http.ErrorMessageResponse{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{repo-name}/chart/{chart-name}/architecture").To(handler.GetChartArchitecture).
		Doc("获取chart架构信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("repo-name", "Repo名字").DataType("string")).
		Param(ws.PathParameter("chart-name", "Chart名字").DataType("string")).
		Param(ws.QueryParameter("chart-version", "chart版本").DataType("string").DefaultValue("")).
		Writes("").
		Returns(200, "OK", "").
		Returns(404, "Not Found", http.ErrorMessageResponse{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	return ws
}

func (handler *ChartHandler) GetRepoList(request *restful.Request, response *restful.Response) {
	repoInfoList := handler.helm.GetRepoList()
	response.WriteEntity(&repoInfoList)
}

func (handler *ChartHandler) GetChartList(request *restful.Request, response *restful.Response) {
	repoName := request.PathParameter("repo-name")
	chartList, err := handler.helm.GetChartList(repoName)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get chart list: %s", err.Error()))
		return
	}
	response.WriteEntity(chartList)
}

func (handler *ChartHandler) GetChartInfo(request *restful.Request, response *restful.Response) {
	repoName := request.PathParameter("repo-name")
	chartName := request.PathParameter("chart-name")
	chartVersion := request.QueryParameter("chart-version")

	chartDetailInfo, err := handler.helm.GetChartDetailInfo(repoName, chartName, chartVersion)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			httpUtils.WriteNotFoundResponse(response, -1, fmt.Sprintf("Chart %s-%s is not found in repo %s", chartName, chartVersion, repoName))
			return
		}
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get chart: %s", err.Error()))
		return
	}
	response.WriteEntity(chartDetailInfo.ChartInfo)
}

func (handler *ChartHandler) GetChartInfoByImage(request *restful.Request, response *restful.Response) {
	chartImage := request.QueryParameter("chart-image")
	chartDetailInfo, err := handler.helm.GetDetailChartInfoByImage(chartImage)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			httpUtils.WriteNotFoundResponse(response, -1, fmt.Sprintf("Chart %s is not found", chartImage))
			return
		}
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get chart: %s", err.Error()))
		return
	}
	response.WriteEntity(chartDetailInfo.ChartInfo)
}

func (handler *ChartHandler) GetChartIcon(request *restful.Request, response *restful.Response) {
	repoName := request.PathParameter("repo-name")
	chartName := request.PathParameter("chart-name")
	chartVersion := request.QueryParameter("chart-version")

	chartDetailInfo, err := handler.helm.GetChartDetailInfo(repoName, chartName, chartVersion)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			httpUtils.WriteNotFoundResponse(response, -1, fmt.Sprintf("Chart %s-%s is not found in repo %s", chartName, chartVersion, repoName))
			return
		}
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get chart: %s", err.Error()))
		return
	}
	response.WriteEntity(chartDetailInfo.Icon)
}

func (handler *ChartHandler) GetChartAdvantage(request *restful.Request, response *restful.Response) {
	repoName := request.PathParameter("repo-name")
	chartName := request.PathParameter("chart-name")
	chartVersion := request.QueryParameter("chart-version")

	chartDetailInfo, err := handler.helm.GetChartDetailInfo(repoName, chartName, chartVersion)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			httpUtils.WriteNotFoundResponse(response, -1, fmt.Sprintf("Chart %s-%s is not found in repo %s", chartName, chartVersion, repoName))
			return
		}
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get chart: %s", err.Error()))
		return
	}
	response.WriteEntity(chartDetailInfo.Advantage)
}

func (handler *ChartHandler) GetChartArchitecture(request *restful.Request, response *restful.Response) {
	repoName := request.PathParameter("repo-name")
	chartName := request.PathParameter("chart-name")
	chartVersion := request.QueryParameter("chart-version")

	chartDetailInfo, err := handler.helm.GetChartDetailInfo(repoName, chartName, chartVersion)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			httpUtils.WriteNotFoundResponse(response, -1, fmt.Sprintf("Chart %s-%s is not found in repo %s", chartName, chartVersion, repoName))
			return
		}
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get chart: %s", err.Error()))
		return
	}
	response.WriteEntity(chartDetailInfo.Architecture)
}
