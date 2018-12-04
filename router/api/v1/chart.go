package v1

import (
	"github.com/emicklei/go-restful"
	"walm/pkg/release/manager/helm"
	"fmt"
	walmerr "walm/pkg/util/error"
)

func GetRepoList(request *restful.Request, response *restful.Response) {
	repoInfoList := helm.GetRepoList()
	response.WriteEntity(&repoInfoList)
}

func GetChartList(request *restful.Request, response *restful.Response) {
	repoName := request.PathParameter("repo-name")
	chartList, err := helm.GetChartList(repoName)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to get chart list: %s", err.Error()))
		return
	}
	response.WriteEntity(chartList)
}

func GetChartInfo(request *restful.Request, response *restful.Response) {
	repoName := request.PathParameter("repo-name")
	chartName := request.PathParameter("chart-name")
	chartVersion := request.QueryParameter("chart-version")
	chartInfo, err := helm.GetChartInfo(repoName, chartName, chartVersion)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			WriteNotFoundResponse(response, -1, fmt.Sprintf("Chart %s-%s is not found in repo %s", chartName, chartVersion, repoName))
			return
		}
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to get chart: %s", err.Error()))
		return
	}
	response.WriteEntity(chartInfo)
}