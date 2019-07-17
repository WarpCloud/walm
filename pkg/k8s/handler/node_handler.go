package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sutils "WarpCloud/walm/pkg/k8s/utils"
	"k8s.io/apimachinery/pkg/fields"
	listv1 "k8s.io/client-go/listers/core/v1"
	"encoding/json"
	"reflect"
	"github.com/sirupsen/logrus"
)

type NodeHandler struct {
	client *kubernetes.Clientset
	lister listv1.NodeLister
}

func (handler *NodeHandler) GetNode(name string) (*v1.Node, error) {
	return handler.lister.Get(name)
}

func (handler *NodeHandler) ListNodes(labelSelector *metav1.LabelSelector) ([]*v1.Node, error) {
	selector, err := k8sutils.ConvertLabelSelectorToSelector(labelSelector)
	if err != nil {
		return nil, err
	}
	return handler.lister.List(selector)
}

func (handler *NodeHandler) LabelNode(name string, labels map[string]string, remove []string) (node *v1.Node, err error) {
	if len(labels) == 0 && len(remove) == 0 {
		return
	}

	node, err = handler.client.CoreV1().Nodes().Get(name, metav1.GetOptions{})
	if err != nil {
		return
	}

	oldLabels, err := json.Marshal(node.Labels)
	if err != nil {
		return
	}

	node.Labels = k8sutils.MergeLabels(node.Labels, labels, remove)
	newLabels, err := json.Marshal(node.Labels)
	if err != nil {
		return
	}

	if !reflect.DeepEqual(oldLabels, newLabels) {
		return handler.client.CoreV1().Nodes().Update(node)
	}

	return
}

func (handler *NodeHandler) AnnotateNode(name string, annos map[string]string, remove []string) (node *v1.Node, err error) {
	if len(annos) == 0 && len(remove) == 0 {
		return
	}

	node, err = handler.client.CoreV1().Nodes().Get(name, metav1.GetOptions{})
	if err != nil {
		return
	}

	oldAnnos, err := json.Marshal(node.Annotations)
	if err != nil {
		return
	}

	node.Annotations = k8sutils.MergeLabels(node.Annotations, annos, remove)
	newAnnos, err := json.Marshal(node.Annotations)
	if err != nil {
		return
	}

	if !reflect.DeepEqual(oldAnnos, newAnnos) {
		return handler.client.CoreV1().Nodes().Update(node)
	}

	return
}

func (handler *NodeHandler) GetPodsOnNode(nodeName string, labelSelector *metav1.LabelSelector) (*v1.PodList, error) {
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

func (handler *NodeHandler) GetPodsOnNodeUsingCache(nodeName string, labelSelector *metav1.LabelSelector) (*v1.PodList, error) {
	pods, err := GetDefaultHandlerSet().GetPodHandler().ListPods("", labelSelector)
	if err != nil {
		logrus.Errorf("failed to list pods : %s", err.Error())
		return nil, err
	}

	podList := &v1.PodList{
		Items: []v1.Pod{},
	}

	for _, pod := range pods {
		if pod.Spec.NodeName == nodeName && pod.Status.Phase != v1.PodSucceeded && pod.Status.Phase != v1.PodFailed {
			podList.Items = append(podList.Items, *pod)
		}
	}

	return podList, nil
}
