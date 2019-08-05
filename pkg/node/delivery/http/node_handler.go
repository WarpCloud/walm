package http

import (
	"github.com/emicklei/go-restful"
	"WarpCloud/walm/pkg/models/http"
	k8sModel "WarpCloud/walm/pkg/models/k8s"
	"github.com/emicklei/go-restful-openapi"
	"WarpCloud/walm/pkg/k8s"
	"fmt"
	httpUtils "WarpCloud/walm/pkg/util/http"
	errorModel "WarpCloud/walm/pkg/models/error"
)

type NodeHandler struct {
	k8sCache    k8s.Cache
	k8sOperator k8s.Operator
}

func RegisterNodeHandler(k8sCache k8s.Cache, k8sOperator k8s.Operator) *restful.WebService {
	handler := &NodeHandler{
		k8sOperator: k8sOperator,
		k8sCache:    k8sCache,
	}

	ws := new(restful.WebService)

	ws.Path(http.ApiV1 + "/node").
		Doc("Kubernetes节点相关操作").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"node"}

	ws.Route(ws.GET("/").To(handler.GetNodes).
		Doc("获取节点列表").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.QueryParameter("labelselector", "节点标签过滤").DataType("string")).
		Writes(k8sModel.NodeList{}).
		Returns(200, "OK", k8sModel.NodeList{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.GET("/{nodename}").To(handler.GetNode).
		Doc("获取节点详细信息").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("nodename", "节点名字").DataType("string")).
		Writes(k8sModel.Node{}).
		Returns(200, "OK", k8sModel.Node{}).
		Returns(404, "Not Found", http.ErrorMessageResponse{}).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{nodename}/labels").To(handler.LabelNode).
		Doc("修改节点Labels").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("nodename", "节点名字").DataType("string")).
		Reads(k8sModel.LabelNodeRequestBody{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	ws.Route(ws.POST("/{nodename}/annotations").To(handler.AnnotateNode).
		Doc("修改节点Annotations").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("nodename", "节点名字").DataType("string")).
		Reads(k8sModel.AnnotateNodeRequestBody{}).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", http.ErrorMessageResponse{}))

	return ws
}

func (handler *NodeHandler) GetNodes(request *restful.Request, response *restful.Response) {
	labelSelectorStr := request.QueryParameter("labelselector")
	nodes, err := handler.k8sCache.GetNodes(labelSelectorStr)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get nodes: %s", err.Error()))
		return
	}
	response.WriteEntity(k8sModel.NodeList{nodes})
}

func (handler *NodeHandler) GetNode(request *restful.Request, response *restful.Response) {
	nodeName := request.PathParameter("nodename")
	if nodeName == "" {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("node name can not be empty"))
		return
	}
	resource, err := handler.k8sCache.GetResource(k8sModel.NodeKind, "", nodeName)
	if err != nil {
		if errorModel.IsNotFoundError(err) {
			httpUtils.WriteNotFoundResponse(response, -1, fmt.Sprintf("node %s is not found", nodeName))
			return
		}
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to get node: %s", err.Error()))
		return
	}
	response.WriteEntity(resource.(*k8sModel.Node))
}

func (handler *NodeHandler) LabelNode(request *restful.Request, response *restful.Response) {
	nodeName := request.PathParameter("nodename")
	if nodeName == "" {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("node name can not be empty"))
		return
	}
	labelNodeRequest := &k8sModel.LabelNodeRequestBody{}
	err := request.ReadEntity(labelNodeRequest)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}

	err = handler.k8sOperator.LabelNode(nodeName, labelNodeRequest.AddLabels, labelNodeRequest.RemoveLabels)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to label node %s: %s", nodeName, err.Error()))
		return
	}
}

func (handler *NodeHandler) AnnotateNode(request *restful.Request, response *restful.Response) {
	nodeName := request.PathParameter("nodename")
	if nodeName == "" {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("node name can not be empty"))
		return
	}
	annotateNodeRequest := &k8sModel.AnnotateNodeRequestBody{}
	err := request.ReadEntity(annotateNodeRequest)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}

	err = handler.k8sOperator.AnnotateNode(nodeName, annotateNodeRequest.AddAnnotations, annotateNodeRequest.RemoveAnnotations)
	if err != nil {
		httpUtils.WriteErrorResponse(response, -1, fmt.Sprintf("failed to annotate node %s: %s", nodeName, err.Error()))
		return
	}
}
