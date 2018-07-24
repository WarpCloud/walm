package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sutils "walm/pkg/k8s/utils"
	"k8s.io/apimachinery/pkg/fields"
)

type NodeHandler struct {
	client *kubernetes.Clientset
}

func (handler NodeHandler) GetNode(name string) (*v1.Node, error){
	return handler.client.CoreV1().Nodes().Get(name, metav1.GetOptions{})
}

func (handler NodeHandler) ListNodes(labelSelector *metav1.LabelSelector) (*v1.NodeList, error){
	selectorStr, err := k8sutils.ConvertLabelSelectorToStr(labelSelector)
	if err != nil {
		return nil, err
	}
	return handler.client.CoreV1().Nodes().List(metav1.ListOptions{LabelSelector: selectorStr})
}

func (handler NodeHandler) LabelNode(name string, labels map[string]string) (*v1.Node, error){
	oldNode, err := handler.client.CoreV1().Nodes().Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	newNode := oldNode.DeepCopy()
	newNode.Labels = labels

	return handler.client.CoreV1().Nodes().Update(newNode)
}

func (handler NodeHandler) GetPodsOnNode(nodeName string, labelSelector *metav1.LabelSelector) (*v1.PodList, error) {
	fieldSelector, err := fields.ParseSelector("spec.nodeName=" + nodeName + ",status.phase!=" + string(v1.PodSucceeded) + ",status.phase!=" + string(v1.PodFailed))
	if err != nil {
		return nil, err
	}
	labelSelectorStr, err := k8sutils.ConvertLabelSelectorToStr(labelSelector)
	if err != nil {
		return nil, err
	}
	listOptions := metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
		LabelSelector: labelSelectorStr,
	}

	return handler.client.CoreV1().Pods(metav1.NamespaceAll).List(listOptions)
}

func NewNodeHandler(client *kubernetes.Clientset) NodeHandler {
	return NodeHandler{client: client}
}



