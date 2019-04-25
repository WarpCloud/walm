package v1

import (
	"encoding/json"
	"fmt"
	"github.com/emicklei/go-restful"
	walmerr "walm/pkg/util/error"
	"strconv"
	"github.com/sirupsen/logrus"
	"walm/pkg/util/transwarpjsonnet"
	"walm/router/api"
	"walm/pkg/release/manager/helm"
	"walm/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DeleteRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("release")
	deletePvcs, err := getDeletePvcsQueryParam(request)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("query param deletePvcs value is not valid : %s", err.Error()))
		return
	}
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

	err = helm.GetDefaultHelmClient().DeleteRelease(namespace, name, false, deletePvcs, async, timeoutSec)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to delete release: %s", err.Error()))
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
	err = helm.GetDefaultHelmClient().InstallUpgradeRelease(namespace, releaseRequest, false, nil, async, timeoutSec)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to install release: %s", err.Error()))
	}
}

func InstallReleaseWithChart(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	chartArchive, _, err := request.Request.FormFile("chart")
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read chart archive: %s", err.Error()))
		return
	}
	defer chartArchive.Close()
	chartFiles, err := transwarpjsonnet.LoadArchive(chartArchive)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to load chart archive: %s", err.Error()))
		return
	}
	releaseName := request.Request.FormValue("release")
	body := request.Request.FormValue("body")
	releaseRequest := &release.ReleaseRequestV2{}
	if body != "" {
		err = json.Unmarshal([]byte(body), releaseRequest)
		if err != nil {
			api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read release request: %s", err.Error()))
			return
		}
	}

	releaseRequest.Name = releaseName

	err = helm.GetDefaultHelmClient().InstallUpgradeRelease(namespace, releaseRequest, false, chartFiles, false, 0)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to install release: %s", err.Error()))
	}
}

func UpgradeRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
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
	err = helm.GetDefaultHelmClient().InstallUpgradeRelease(namespace, releaseRequest, false,nil, async, timeoutSec)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to upgrade release: %s", err.Error()))
	}
}

func UpgradeReleaseWithChart(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	releaseName := request.Request.FormValue("release")
	chartArchive, _, err := request.Request.FormFile("chart")
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read chart archive: %s", err.Error()))
		return
	}
	defer chartArchive.Close()
	chartFiles, err := transwarpjsonnet.LoadArchive(chartArchive)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to load chart archive: %s", err.Error()))
		return
	}

	body := request.Request.FormValue("body")
	releaseRequest := &release.ReleaseRequestV2{}

	if body != "" {
		err = json.Unmarshal([]byte(body), releaseRequest)
		if err != nil {
			api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read release request: %s", err.Error()))
			return
		}
	}

	releaseRequest.Name = releaseName

	err = helm.GetDefaultHelmClient().InstallUpgradeRelease(namespace, releaseRequest, false, chartFiles, false, 0)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to upgrade release: %s", err.Error()))
	}
}

func ListReleaseByNamespace(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	labelSelectorStr := request.QueryParameter("labelselector")
	var infos []*release.ReleaseInfoV2
	var err error
	if labelSelectorStr == "" {
		infos, err = helm.GetDefaultHelmClient().ListReleases(namespace, "")
		if err != nil {
			api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to list release: %s", err.Error()))
			return
		}
	} else {
		labelSelector, err := metav1.ParseToLabelSelector(labelSelectorStr)
		if err != nil {
			api.WriteErrorResponse(response, -1,  fmt.Sprintf("parse label selector failed: %s", err.Error()))
			return
		}
		infos, err = helm.GetDefaultHelmClient().ListReleasesByLabels(namespace, labelSelector)
		if err != nil {
			api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to list release: %s", err.Error()))
			return
		}
	}

	response.WriteEntity(release.ReleaseInfoV2List{len(infos), infos})
}

func ListRelease(request *restful.Request, response *restful.Response) {
	labelSelectorStr := request.QueryParameter("labelselector")
	var infos []*release.ReleaseInfoV2
	var err error
	if labelSelectorStr == "" {
		infos, err = helm.GetDefaultHelmClient().ListReleases("", "")
		if err != nil {
			api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to list release: %s", err.Error()))
			return
		}
	} else {
		labelSelector, err := metav1.ParseToLabelSelector(labelSelectorStr)
		if err != nil {
			api.WriteErrorResponse(response, -1,  fmt.Sprintf("parse label selector failed: %s", err.Error()))
			return
		}
		infos, err = helm.GetDefaultHelmClient().ListReleasesByLabels("", labelSelector)
		if err != nil {
			api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to list release: %s", err.Error()))
			return
		}
	}

	response.WriteEntity(release.ReleaseInfoV2List{len(infos), infos})
}

func GetRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("release")
	info, err := helm.GetDefaultHelmClient().GetRelease(namespace, name)
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

func RestartRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("release")
	err := helm.GetDefaultHelmClient().RestartRelease(namespace, name)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to restart release %s: %s", name, err.Error()))
		return
	}
}

func PauseRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("release")
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
	err = helm.GetDefaultHelmClient().PauseRelease(namespace, name, false, async, timeoutSec)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to pause release %s: %s", name, err.Error()))
		return
	}
}

func RecoverRelease(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("release")
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
	err = helm.GetDefaultHelmClient().RecoverRelease(namespace, name, false, async, timeoutSec)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to recover release %s: %s", name, err.Error()))
		return
	}
}

func RollBackRelease(request *restful.Request, response *restful.Response) {
}
