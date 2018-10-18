package v1

import (
	"github.com/emicklei/go-restful"

	"walm/pkg/release/manager/project"
	releasetypes "walm/pkg/release"
	"net/http"
	"fmt"
	walmerr "walm/pkg/util/error"
	"github.com/sirupsen/logrus"
)

func ListProjectAllNamespaces(request *restful.Request, response *restful.Response) {
	projectList, err := project.GetDefaultProjectManager().ListProjects("")
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	response.WriteEntity(projectList)
}

func ListProjectByNamespace(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("namespace")
	projectList, err := project.GetDefaultProjectManager().ListProjects(tenantName)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	response.WriteEntity(projectList)
}

func DeployProject(request *restful.Request, response *restful.Response) {
	projectParams := new(releasetypes.ProjectParams)
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	err := request.ReadEntity(&projectParams)
	if projectParams.CommonValues == nil {
		projectParams.CommonValues = make(map[string]interface{})
	}
	if projectParams.Releases == nil {
		response.WriteError(http.StatusInternalServerError,
			fmt.Errorf("invalid project params releases is nil %+v", projectParams))
	}
	for _, releaseInfo := range projectParams.Releases {
		if releaseInfo.Dependencies == nil {
			releaseInfo.Dependencies = make(map[string]string)
		}
		if releaseInfo.ConfigValues == nil {
			releaseInfo.ConfigValues = make(map[string]interface{})
		}
	}
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	err = project.GetDefaultProjectManager().CreateProject(tenantName, projectName, projectParams)
	if err != nil {
		logrus.Infof("DeployProject %+v", err)
		response.WriteError(http.StatusInternalServerError, err)
	}
}

func GetProjectInfo(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	projectInfo, err := project.GetDefaultProjectManager().GetProjectInfo(tenantName, projectName)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			WriteNotFoundResponse(response, -1, fmt.Sprintf("project %s/%s is not found", tenantName, projectName))
			return
		}
		response.WriteError(http.StatusInternalServerError, err)
	}
	response.WriteEntity(projectInfo)
}

func DeleteProject(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	err := project.GetDefaultProjectManager().DeleteProject(tenantName, projectName)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
}

func DeployInstanceInProject(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	releaseRequest := &releasetypes.ReleaseRequest{}
	err := request.ReadEntity(releaseRequest)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}
	err = project.GetDefaultProjectManager().AddReleaseInProject(tenantName, projectName, releaseRequest)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
}

func DeployProjectInProject(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	projectParams := &releasetypes.ProjectParams{}
	err := request.ReadEntity(projectParams)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}
	err = project.GetDefaultProjectManager().AddReleasesInProject(tenantName, projectName, projectParams)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
}

func DeleteInstanceInProject(request *restful.Request, response *restful.Response) {
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	releaseName := request.PathParameter("release")
	err := project.GetDefaultProjectManager().RemoveReleaseInProject(tenantName, projectName, releaseName)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
}
