package v1

import (
	"github.com/emicklei/go-restful"
	"fmt"
	"walm/pkg/k8s/adaptor"
)

func ExecShell(request *restful.Request, response *restful.Response) {
}

func GetPodEvents(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("pod")
	events, err := adaptor.GetDefaultAdaptorSet().GetAdaptor("Pod").(*adaptor.WalmPodAdaptor).GetWalmPodEventList(namespace, name)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to get pod events %s: %s", name, err.Error()))
		return
	}
	response.WriteEntity(*events)
}
