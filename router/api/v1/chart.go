package v1

import (
	"fmt"

	"github.com/emicklei/go-restful"
	"WarpCloud/walm/pkg/release/manager/helm"
	walmerr "WarpCloud/walm/pkg/util/error"
	"WarpCloud/walm/router/api"
)

func GetRepoList(request *restful.Request, response *restful.Response) {
	repoInfoList := helm.GetRepoList()
	response.WriteEntity(&repoInfoList)
}

func GetChartList(request *restful.Request, response *restful.Response) {
	repoName := request.PathParameter("repo-name")
	chartList, err := helm.GetChartList(repoName)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get chart list: %s", err.Error()))
		return
	}
	response.WriteEntity(chartList)
}

func GetChartInfo(request *restful.Request, response *restful.Response) {
	repoName := request.PathParameter("repo-name")
	chartName := request.PathParameter("chart-name")
	chartVersion := request.QueryParameter("chart-version")

	chartDetailInfo, err := helm.GetDetailChartInfo(repoName, chartName, chartVersion)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			api.WriteNotFoundResponse(response, -1, fmt.Sprintf("Chart %s-%s is not found in repo %s", chartName, chartVersion, repoName))
			return
		}
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get chart: %s", err.Error()))
		return
	}
	response.WriteEntity(chartDetailInfo.ChartInfo)
}
func GetChartInfoByImage(request *restful.Request, response *restful.Response) {
	chartImage := request.QueryParameter("chart-image")
	chartDetailInfo, err := helm.GetDetailChartInfoByImage(chartImage)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			api.WriteNotFoundResponse(response, -1, fmt.Sprintf("Chart %s is not found", chartImage))
			return
		}
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get chart: %s", err.Error()))
		return
	}
	response.WriteEntity(chartDetailInfo.ChartInfo)
}


func GetChartIcon(request *restful.Request, response *restful.Response) {
	repoName := request.PathParameter("repo-name")
	chartName := request.PathParameter("chart-name")
	chartVersion := request.QueryParameter("chart-version")

	chartDetailInfo, err := helm.GetDetailChartInfo(repoName, chartName, chartVersion)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			api.WriteNotFoundResponse(response, -1, fmt.Sprintf("Chart %s-%s is not found in repo %s", chartName, chartVersion, repoName))
			return
		}
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get chart: %s", err.Error()))
		return
	}
	response.WriteEntity(chartDetailInfo.Icon)
	//r := bytes.NewReader(chartDetailInfo.Icon)
	//http.ServeContent(response.ResponseWriter, request.Request, "Icon", time.Now(), r)
}

func GetChartAdvantage(request *restful.Request, response *restful.Response) {
	repoName := request.PathParameter("repo-name")
	chartName := request.PathParameter("chart-name")
	chartVersion := request.QueryParameter("chart-version")

	chartDetailInfo, err := helm.GetDetailChartInfo(repoName, chartName, chartVersion)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			api.WriteNotFoundResponse(response, -1, fmt.Sprintf("Chart %s-%s is not found in repo %s", chartName, chartVersion, repoName))
			return
		}
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get chart: %s", err.Error()))
		return
	}
	response.WriteEntity(chartDetailInfo.Advantage)
	//r := bytes.NewReader(chartDetailInfo.Advantage)
	//http.ServeContent(response.ResponseWriter, request.Request, "Advantage", time.Now(), r)
}

func GetChartArchitecture(request *restful.Request, response *restful.Response) {
	repoName := request.PathParameter("repo-name")
	chartName := request.PathParameter("chart-name")
	chartVersion := request.QueryParameter("chart-version")

	chartDetailInfo, err := helm.GetDetailChartInfo(repoName, chartName, chartVersion)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			api.WriteNotFoundResponse(response, -1, fmt.Sprintf("Chart %s-%s is not found in repo %s", chartName, chartVersion, repoName))
			return
		}
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get chart: %s", err.Error()))
		return
	}
	response.WriteEntity(chartDetailInfo.Architecture)
	//r := bytes.NewReader(chartDetailInfo.Architecture)
	//http.ServeContent(response.ResponseWriter, request.Request, "Architecture", time.Now(), r)
}
