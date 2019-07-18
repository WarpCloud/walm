package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sutils "WarpCloud/walm/pkg/k8s/utils"
	"k8s.io/apimachinery/pkg/runtime"
)

type EventHandler struct {
	client *kubernetes.Clientset
}

func (handler *EventHandler) ListEvents(namespace string, labelSelector *metav1.LabelSelector) (*v1.EventList, error) {
	selectorStr, err := k8sutils.ConvertLabelSelectorToStr(labelSelector)
	if err != nil {
		return nil, err
	}
	return handler.client.CoreV1().Events(namespace).List(metav1.ListOptions{LabelSelector:selectorStr})
}

func (handler *EventHandler) SearchEvents(namespace string,objOrRef runtime.Object)(*v1.EventList, error) {
	return handler.client.CoreV1().Events(namespace).Search(runtime.NewScheme(), objOrRef)
}

