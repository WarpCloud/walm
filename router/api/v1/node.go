package v1

import (
	"github.com/emicklei/go-restful"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"fmt"
	"walm/pkg/k8s/adaptor"
	"walm/router/api"
	"walm/pkg/k8s/handler"
)

func GetNodes(request *restful.Request, response *restful.Response) {
	labelSelectorStr := request.QueryParameter("labelselector")
	labelSelector, err := metav1.ParseToLabelSelector(labelSelectorStr)
	if err != nil {
		WriteErrorResponse(response, -1,  fmt.Sprintf("parse label selector failed: %s", err.Error()))
		return
	}
	nodes, err := adaptor.GetDefaultAdaptorSet().GetAdaptor("Node").(*adaptor.WalmNodeAdaptor).GetWalmNodes("", labelSelector)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to get nodes: %s", err.Error()))
		return
	}
	response.WriteEntity(adaptor.WalmNodeList{nodes})
}

func GetNode(request *restful.Request, response *restful.Response) {
	nodeName := request.PathParameter("nodename")
	if nodeName == "" {
		WriteErrorResponse(response, -1, fmt.Sprintf("node name can not be empty"))
		return
	}
	node, err := adaptor.GetDefaultAdaptorSet().GetAdaptor("Node").(*adaptor.WalmNodeAdaptor).GetResource("", nodeName)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to get node: %s", err.Error()))
		return
	}
	if node.GetState().Status == "NotFound" {
		WriteNotFoundResponse(response, -1, fmt.Sprintf("node %s is not found", nodeName))
		return
	}
	response.WriteEntity(node)
}

func LabelNode(request *restful.Request, response *restful.Response) {
	nodeName := request.PathParameter("nodename")
	if nodeName == "" {
		WriteErrorResponse(response, -1, fmt.Sprintf("node name can not be empty"))
		return
	}
	labelNodeRequest := &api.LabelNodeRequestBody{}
	err := request.ReadEntity(labelNodeRequest)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}

	_, err = handler.GetDefaultHandlerSet().GetNodeHandler().LabelNode(nodeName, labelNodeRequest.AddLabels, labelNodeRequest.RemoveLabels)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to label node %s: %s", nodeName, err.Error()))
		return
	}
}

func AnnotateNode(request *restful.Request, response *restful.Response) {
	nodeName := request.PathParameter("nodename")
	if nodeName == "" {
		WriteErrorResponse(response, -1, fmt.Sprintf("node name can not be empty"))
		return
	}
	annotateNodeRequest := &api.AnnotateNodeRequestBody{}
	err := request.ReadEntity(annotateNodeRequest)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to read request body: %s", err.Error()))
		return
	}

	_, err = handler.GetDefaultHandlerSet().GetNodeHandler().AnnotateNode(nodeName, annotateNodeRequest.AddAnnotations, annotateNodeRequest.RemoveAnnotations)
	if err != nil {
		WriteErrorResponse(response, -1, fmt.Sprintf("failed to annotate node %s: %s", nodeName, err.Error()))
		return
	}
}
