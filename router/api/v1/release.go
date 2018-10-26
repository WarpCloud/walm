package v1

import (
	"github.com/emicklei/go-restful"
	"walm/pkg/release/manager/helm"
	"walm/pkg/release"
	"fmt"
	walmerr "walm/pkg/util/error"
)

func DeleteRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("release")
	err := helm.GetDefaultHelmClient().DeleteRelease(namespace, name)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to delete release: %s", err.Error()))
		return
	}
}

func InstallRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	releaseRequest := &release.ReleaseRequest{}
	err := request.ReadEntity(releaseRequest)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}
	err = helm.GetDefaultHelmClient().InstallUpgradeRealese(namespace, releaseRequest)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to install or upgrade release: %s", err.Error()))
	}
}

func ListReleaseByNamespace(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	infos, err := helm.GetDefaultHelmClient().ListReleases(namespace, "")
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to list release: %s", err.Error()))
		return
	}
	response.WriteEntity(release.ReleaseInfoList{len(infos), infos})
}

func ListRelease(request *restful.Request, response *restful.Response) {
	infos, err := helm.GetDefaultHelmClient().ListReleases("", "")
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to list release: %s", err.Error()))
		return
	}
	response.WriteEntity(release.ReleaseInfoList{len(infos), infos})
}

func GetRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("release")
	info, err := helm.GetDefaultHelmClient().GetRelease(namespace, name)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			WriteNotFoundResponse(response, -1, fmt.Sprintf("release %s is not found", name))
			return
		}
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to get release %s: %s", name, err.Error()))
		return
	}
	response.WriteEntity(info)
}

func RestartRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("release")
	err := helm.GetDefaultHelmClient().RestartRelease(namespace, name)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to restart release %s: %s", name, err.Error()))
		return
	}
}

func RollBackRelease(request *restful.Request, response *restful.Response) {
}

func UpgradeRelease(request *restful.Request, response *restful.Response) {
	InstallRelease(request, response)
}
