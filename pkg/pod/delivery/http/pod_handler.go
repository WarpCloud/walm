package http

import (
	"WarpCloud/walm/pkg/k8s"
	"github.com/emicklei/go-restful"
	"WarpCloud/walm/pkg/models/http"
	"github.com/emicklei/go-restful-openapi"
	k8sModel "WarpCloud/walm/pkg/models/k8s"
	"strconv"
	"github.com/sirupsen/logrus"
	httpUtils "WarpCloud/walm/pkg/util/http"
	"fmt"
)

type PodHandler struct {
	k8sCache k8s.Cache
	k8sOperator k8s.Operator
}

func RegisterPodHandler(k8sCache k8s.Cache, k8sOperator k8s.Operator) *restful.WebService {
	handler := &PodHandler{
		k8sCache: k8sCache,
		k8sOperator: k8sOperator,
	}

	ws := new(restful.WebService)

	ws.Path(http.ApiV1 + "/pod").
		Consumes(restful.MIME_JSON).
		Produces("*/*")

	tags := []string{"pod"}

	ws.Route(ws.GET("/{namespace}/name/{pod}/events").To(handler.GetPodEvents).
		Doc("获取Pod对应的事件").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("pod", "pod名字").DataType("string")).
		Writes(k8sModel.EventList{}).
		Returns(200, "OK", k8sModel.EventList{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{namespace}/name/{pod}/logs").To(handler.GetPodLogs).
		Doc("获取Pod对应的事件").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("pod", "pod名字").DataType("string")).
		Param(ws.QueryParameter("container", "container名字").DataType("string")).
		Param(ws.QueryParameter("tail", "最后几行").DataType("integer")).
		Writes("").
		Returns(200, "OK", "").
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{namespace}/name/{pod}/restart").To(handler.RestartPod).
		Doc("重启Pod").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("namespace", "租户名字").DataType("string")).
		Param(ws.PathParameter("pod", "pod名字").DataType("string")).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	return ws
}

func (handler *PodHandler)RestartPod(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("pod")
	err := handler.k8sOperator.RestartPod(namespace, name)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to restart pod %s: %s", name, err.Error()))
		return
	}
}

func (handler *PodHandler)GetPodEvents(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("pod")
	events, err := handler.k8sCache.GetPodEventList(namespace, name)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get pod events %s: %s", name, err.Error()))
		return
	}
	response.WriteEntity(*events)
}

func (handler *PodHandler)GetPodLogs(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("pod")
	containerName := request.QueryParameter("container")
	tailLines, err := getTailLinesQueryParam(request)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("query param tail is not valid : %s", err.Error()))
		return
	}

	logs, err := handler.k8sCache.GetPodLogs(namespace, name, containerName, tailLines)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get pod logs %s: %s", name, err.Error()))
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