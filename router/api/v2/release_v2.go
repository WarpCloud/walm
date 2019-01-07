package v2

import (
	"github.com/emicklei/go-restful"
	helmv2 "walm/pkg/release/v2/helm"
	"fmt"
	walmerr "walm/pkg/util/error"
	"walm/router/api"
	"walm/pkg/release/v2"
)

func DeleteRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("release")
	err := helmv2.GetDefaultHelmClientV2().DeleteRelease(namespace, name, false)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to delete release: %s", err.Error()))
		return
	}
}

func InstallRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	releaseRequest := &v2.ReleaseRequestV2{}
	err := request.ReadEntity(releaseRequest)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}
	err = helmv2.GetDefaultHelmClientV2().InstallUpgradeRelease(namespace, releaseRequest, false)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to install release: %s", err.Error()))
	}
}

func UpgradeRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	releaseRequest := &v2.ReleaseRequestV2{}
	err := request.ReadEntity(releaseRequest)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}
	err = helmv2.GetDefaultHelmClientV2().InstallUpgradeRelease(namespace, releaseRequest, false)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to upgrade release: %s", err.Error()))
	}
}

//func ListReleaseByNamespace(request *restful.Request, response *restful.Response) {
//	namespace := request.PathParameter("namespace")
//	infos, err := helm.GetDefaultHelmClient().ListReleases(namespace, "")
//	if err != nil {
//		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to list release: %s", err.Error()))
//		return
//	}
//	response.WriteEntity(release.ReleaseInfoList{len(infos), infos})
//}

//func ListRelease(request *restful.Request, response *restful.Response) {
//	infos, err := helm.GetDefaultHelmClient().ListReleases("", "")
//	if err != nil {
//		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to list release: %s", err.Error()))
//		return
//	}
//	response.WriteEntity(release.ReleaseInfoList{len(infos), infos})
//}

func GetRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("release")
	info, err := helmv2.GetDefaultHelmClientV2().GetRelease(namespace, name)
	if err != nil {
		if walmerr.IsNotFoundError(err) {
			api.WriteNotFoundResponse(response, -1, fmt.Sprintf("release %s is not found", name))
			return
		}
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get release %s: %s", name, err.Error()))
		return
	}
	response.WriteEntity(info)
}

