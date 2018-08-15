package v1

import (
	"github.com/emicklei/go-restful"
	"walm/pkg/release/manager/helm"
	"walm/pkg/release"
	"fmt"
	"strconv"
)

func DeleteRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("release")
	err := helm.DeleteRelease(namespace, name)
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
	err = helm.InstallUpgradeRealese(namespace, releaseRequest)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to install release: %s", err.Error()))
	}
}

func ListReleaseByNamespace(request *restful.Request, response *restful.Response) {
	option := &release.ReleaseListOption{}
	namespace := request.PathParameter("namespace")
	option.Namespace = &namespace
	if filter := request.QueryParameter("filter"); len(filter) > 0 {
		option.Filter = &filter
	}
	if limit := request.QueryParameter("limit"); len(limit) > 0 {
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			WriteErrorResponse(response, -1, fmt.Sprintf("failed to parse limit query parameter %s: %s", limit, err.Error()))
			return
		}
		option.Limit = &limitInt
	}
	infos, err := helm.ListReleases(option)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to list release: %s", err.Error()))
		return
	}
	response.WriteEntity(release.ReleaseInfoList{len(infos), infos})
}

func ListRelease(request *restful.Request, response *restful.Response) {
	option := &release.ReleaseListOption{}
	if filter := request.QueryParameter("filter"); len(filter) > 0 {
		option.Filter = &filter
	}
	if limit := request.QueryParameter("limit"); len(limit) > 0 {
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			WriteErrorResponse(response, -1, fmt.Sprintf("failed to parse limit query parameter %s: %s", limit, err.Error()))
			return
		}
		option.Limit = &limitInt
	}
	infos, err := helm.ListReleases(option)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to list release: %s", err.Error()))
		return
	}
	response.WriteEntity(release.ReleaseInfoList{len(infos), infos})
}

func GetRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("release")
	info, err := helm.GetRelease(namespace, name)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to get release: %s", err.Error()))
		return
	}
	response.WriteEntity(info)
}

func RollBackRelease(request *restful.Request, response *restful.Response) {
}

func UpgradeRelease(request *restful.Request, response *restful.Response) {
	InstallRelease(request, response)
}
