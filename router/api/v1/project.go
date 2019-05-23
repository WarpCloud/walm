package v1

import (
	"github.com/emicklei/go-restful"

	"WarpCloud/walm/pkg/project"
	"fmt"
	walmerr "WarpCloud/walm/pkg/util/error"
	"github.com/sirupsen/logrus"
	"strconv"
	"WarpCloud/walm/router/api"
	"WarpCloud/walm/pkg/release"
)

func ListProjectAllNamespaces(request *restful.Request, response *restful.Response) {
	projectList, err := project.GetDefaultProjectManager().ListProjects("")
	if err != nil {
		api.WriteErrorResponse(response,-1, fmt.Sprintf("failed to list all projects : %s", err.Error()))
		return
	}
	response.WriteEntity(projectList)
}

func ListProjectByNamespace(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("namespace")
	projectList, err := project.GetDefaultProjectManager().ListProjects(tenantName)
	if err != nil {
		api.WriteErrorResponse(response,-1, fmt.Sprintf("failed to list projects in tenant %s : %s", tenantName, err.Error()))
		return
	}
	response.WriteEntity(projectList)
}

func getAsyncQueryParam(request *restful.Request) (async bool, err error) {
	asyncStr := request.QueryParameter("async")
	if len(asyncStr) > 0 {
		async, err = strconv.ParseBool(asyncStr)
		if err != nil {
			logrus.Errorf("failed to parse query parameter async %s : %s",asyncStr, err.Error())
			return
		}
	}
	return
}

func getTimeoutSecQueryParam(request *restful.Request) (timeoutSec int64, err error) {
	timeoutStr := request.QueryParameter("timeoutSec")
	if len(timeoutStr) > 0 {
		timeoutSec, err = strconv.ParseInt(timeoutStr, 10, 64)
		if err != nil {
			logrus.Errorf("failed to parse query parameter timeoutSec %s : %s", timeoutStr, err.Error())
			return
		}
	}
	return
}

func DeployProject(request *restful.Request, response *restful.Response) {
	projectParams := new(project.ProjectParams)
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	async, err := getAsyncQueryParam(request)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("query param async value is not valid : %s", err.Error()))
		return
	}

	timeoutSec, err := getTimeoutSecQueryParam(request)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("query param timeoutSec value is not valid : %s", err.Error()))
		return
	}

	err = request.ReadEntity(&projectParams)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body : %s", err.Error()))
		return
	}

	if projectParams.CommonValues == nil {
		projectParams.CommonValues = make(map[string]interface{})
	}
	if projectParams.Releases == nil {
		api.WriteErrorResponse(response,-1, "project params releases can not be empty")
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

	err = project.GetDefaultProjectManager().CreateProject(tenantName, projectName, projectParams, async, timeoutSec)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to create project : %s", err.Error()))
		return
	}
}

func GetProjectInfo(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	projectInfo, err := project.GetDefaultProjectManager().GetProjectInfo(tenantName, projectName)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			api.WriteNotFoundResponse(response, -1, fmt.Sprintf("project %s/%s is not found", tenantName, projectName))
			return
		}
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get project info : %s", err.Error()))
		return
	}
	response.WriteEntity(projectInfo)
}

func DeleteProject(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	async, err := getAsyncQueryParam(request)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("query param async value is not valid : %s", err.Error()))
		return
	}
	timeoutSec, err := getTimeoutSecQueryParam(request)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("query param timeoutSec value is not valid : %s", err.Error()))
		return
	}
	deletePvcs, err := getDeletePvcsQueryParam(request)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("query param deletePvcs value is not valid : %s", err.Error()))
		return
	}
	err = project.GetDefaultProjectManager().DeleteProject(tenantName, projectName, async, timeoutSec, deletePvcs)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to delete project : %s", err.Error()))
		return
	}
}

func DeployInstanceInProject(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	async, err := getAsyncQueryParam(request)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("query param async value is not valid : %s", err.Error()))
		return
	}
	timeoutSec, err := getTimeoutSecQueryParam(request)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("query param timeoutSec value is not valid : %s", err.Error()))
		return
	}
	releaseRequest := &release.ReleaseRequestV2{}
	err = request.ReadEntity(releaseRequest)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}
	err = project.GetDefaultProjectManager().AddReleaseInProject(tenantName, projectName, releaseRequest, async, timeoutSec)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to add release in project : %s", err.Error()))
		return
	}
}

func UpgradeInstanceInProject(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	async, err := getAsyncQueryParam(request)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("query param async value is not valid : %s", err.Error()))
		return
	}
	timeoutSec, err := getTimeoutSecQueryParam(request)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("query param timeoutSec value is not valid : %s", err.Error()))
		return
	}
	releaseRequest := &release.ReleaseRequestV2{}
	err = request.ReadEntity(releaseRequest)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}
	err = project.GetDefaultProjectManager().UpgradeReleaseInProject(tenantName, projectName, releaseRequest, async, timeoutSec)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to upgrade release in project : %s", err.Error()))
		return
	}
}

func DeployProjectInProject(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	async, err := getAsyncQueryParam(request)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("query param async value is not valid : %s", err.Error()))
		return
	}
	timeoutSec, err := getTimeoutSecQueryParam(request)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("query param timeoutSec value is not valid : %s", err.Error()))
		return
	}
	projectParams := &project.ProjectParams{}
	err = request.ReadEntity(projectParams)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}
	err = project.GetDefaultProjectManager().AddReleasesInProject(tenantName, projectName, projectParams, async, timeoutSec)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to add releases in project : %s", err.Error()))
		return
	}
}

func DeleteInstanceInProject(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	releaseName := request.PathParameter("release")
	async, err := getAsyncQueryParam(request)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("query param async value is not valid : %s", err.Error()))
		return
	}
	timeoutSec, err := getTimeoutSecQueryParam(request)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("query param timeoutSec value is not valid : %s", err.Error()))
		return
	}
	deletePvcs, err := getDeletePvcsQueryParam(request)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("query param deletePvcs value is not valid : %s", err.Error()))
		return
	}
	err = project.GetDefaultProjectManager().RemoveReleaseInProject(tenantName, projectName, releaseName, async, timeoutSec, deletePvcs)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to delete release in project : %s", err.Error()))
		return
	}
}
