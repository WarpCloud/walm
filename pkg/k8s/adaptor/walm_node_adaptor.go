package adaptor

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"walm/pkg/k8s/handler"
)

type WalmNodeAdaptor struct {
	handler *handler.NodeHandler
}

func (adaptor *WalmNodeAdaptor) GetResource(namespace string, name string) (WalmResource, error) {
	node, err := adaptor.handler.GetNode(name)
	if err != nil {
		if IsNotFoundErr(err) {
			return WalmNode{
				WalmMeta: buildNotFoundWalmMeta("Node", namespace, name),
			}, nil
		}
		return WalmNode{}, err
	}

	return BuildWalmNode(*node), nil
}

func (adaptor *WalmNodeAdaptor) GetWalmNodes(namespace string, labelSelector *metav1.LabelSelector) ([]*WalmNode, error) {
	nodeList, err := adaptor.handler.ListNodes(labelSelector)
	if err != nil {
		return nil, err
	}

	walmNodes := []*WalmNode{}
	if nodeList != nil {
		for _, node := range nodeList {
			walmNode := BuildWalmNode(*node)
			walmNodes = append(walmNodes, walmNode)
		}
	}

	return walmNodes, nil
}

func BuildWalmNode(node corev1.Node) *WalmNode {
	walmNode := WalmNode{
		WalmMeta:    buildWalmMeta("Node", node.Namespace, node.Name, BuildWalmNodeState(node)),
		NodeIp:      BuildNodeIp(node),
		Labels:      node.Labels,
		Annotations: node.Annotations,
	}
	return &walmNode
}

func BuildNodeIp(node corev1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == corev1.NodeInternalIP {
			return address.Address
		}
	}
	return ""
}

func BuildWalmNodeState(node corev1.Node) WalmState {
	podState := buildWalmState("NotReady", "Unknown", "")
	for _, condition := range node.Status.Conditions {
		if condition.Type == "Ready" {
			if condition.Status == corev1.ConditionTrue {
				podState = buildWalmState("Ready", "", "")
			} else {
				podState = buildWalmState("NotReady", condition.Reason, condition.Message)
			}
			break
		}
	}
	return podState
}
