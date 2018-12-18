package v1

import (
	"github.com/emicklei/go-restful"
	"walm/pkg/release/manager/helm"
	helmv2 "walm/pkg/release/v2/helm"
	"walm/pkg/release"
	"fmt"
	walmerr "walm/pkg/util/error"
	"strconv"
	"github.com/sirupsen/logrus"
)

func DeleteRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("release")
	deletePvcs, err := getDeletePvcsQueryParam(request)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("query param deletePvcs value is not valid : %s", err.Error()))
		return
	}
	err = helm.GetDefaultHelmClient().DeleteRelease(namespace, name, false, deletePvcs)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to delete release: %s", err.Error()))
		return
	}
}

func getDeletePvcsQueryParam(request *restful.Request) (deletePvcs bool, err error) {
	deletePvcsStr := request.QueryParameter("deletePvcs")
	if len(deletePvcsStr) > 0 {
		deletePvcs, err = strconv.ParseBool(deletePvcsStr)
		if err != nil {
			logrus.Errorf("failed to parse query parameter deletePvcs %s : %s", deletePvcsStr, err.Error())
			return
		}
	}
	return
}

func InstallRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	releaseRequest := &release.ReleaseRequest{}
	err := request.ReadEntity(releaseRequest)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}
	err = helmv2.GetDefaultHelmClientV2().InstallRelease(namespace, releaseRequest, false)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to install release: %s", err.Error()))
	}
}

func UpgradeRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	releaseRequest := &release.ReleaseRequest{}
	err := request.ReadEntity(releaseRequest)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}
	err = helm.GetDefaultHelmClient().UpgradeRealese(namespace, releaseRequest)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to upgrade release: %s", err.Error()))
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
