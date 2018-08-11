package v1

import (
	"github.com/emicklei/go-restful"

	"walm/pkg/release/manager/project"
	releasetypes "walm/pkg/release"
	"net/http"
)

func GetProjectAllNamespaces(request *restful.Request, response *restful.Response) {
	//project.CreateProject()
}

func GetProjectByNamespace(request *restful.Request, response *restful.Response) {
	//project.CreateProject()
}

func DeployProject(request *restful.Request, response *restful.Response) {
	projectParams := new(releasetypes.ProjectParams)
	tenantName := request.PathParameter("namespace")
	projectName := request.PathParameter("project")
	err := request.ReadEntity(&projectParams)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	err = project.CreateProject(tenantName, projectName, projectParams)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
}

func GetProject(request *restful.Request, response *restful.Response) {
}

func DeleteProject(request *restful.Request, response *restful.Response) {
}

func DeployInstanceInProject(request *restful.Request, response *restful.Response) {
}
