package handler

import (
	clientsetex "transwarp/application-instance/pkg/client/clientset/versioned"
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sutils "walm/pkg/k8s/utils"
)

type InstanceHandler struct {
	client *clientsetex.Clientset
}

func (handler InstanceHandler) GetInstance(namespace string, name string) (*v1beta1.ApplicationInstance, error) {
	return handler.client.TranswarpV1beta1().ApplicationInstances(namespace).Get(name, v1.GetOptions{})
}

func (handler InstanceHandler) ListInstances(namespace string, labelSelector *metav1.LabelSelector) (*v1beta1.ApplicationInstanceList, error) {
	selectorStr, err := k8sutils.ConvertLabelSelectorToStr(labelSelector)
	if err != nil {
		return nil, err
	}
	return handler.client.TranswarpV1beta1().ApplicationInstances(namespace).List(metav1.ListOptions{LabelSelector:selectorStr})
}

func NewInstanceHandler(client *clientsetex.Clientset) InstanceHandler {
	return InstanceHandler{client: client}
}
