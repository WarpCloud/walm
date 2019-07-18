package v1

import (
	"github.com/emicklei/go-restful"
	"fmt"
	"WarpCloud/walm/pkg/k8s/adaptor"
	"strconv"
	"github.com/sirupsen/logrus"
	"WarpCloud/walm/pkg/k8s/handler"
	"WarpCloud/walm/router/api"
)

func ExecShell(request *restful.Request, response *restful.Response) {
}

func RestartPod(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("pod")
	err := handler.GetDefaultHandlerSet().GetPodHandler().DeletePod(namespace, name)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to restart pod %s: %s", name, err.Error()))
		return
	}
}

func GetPodEvents(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("pod")
	events, err := adaptor.GetDefaultAdaptorSet().GetAdaptor("Pod").(*adaptor.WalmPodAdaptor).GetWalmPodEventList(namespace, name)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get pod events %s: %s", name, err.Error()))
		return
	}
	response.WriteEntity(*events)
}

func GetPodLogs(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("pod")
	containerName := request.QueryParameter("container")
	tailLines, err := getTailLinesQueryParam(request)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("query param tail is not valid : %s", err.Error()))
		return
	}

	logs, err := adaptor.GetDefaultAdaptorSet().GetAdaptor("Pod").(*adaptor.WalmPodAdaptor).GetWalmPodLogs(namespace, name, containerName, tailLines)
	if err != nil {
		api.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get pod logs %s: %s", name, err.Error()))
		return
	}
	response.WriteEntity(logs)
}

func getTailLinesQueryParam(request *restful.Request) (tailLines int64, err error) {
	tailLinesStr := request.QueryParameter("tail")
	if len(tailLinesStr) > 0 {
		tailLines, err = strconv.ParseInt(tailLinesStr, 10, 64)
		if err != nil {
			logrus.Errorf("failed to parse query parameter tail %s : %s", tailLinesStr, err.Error())
			return
		}
	}
	return
}
