package handler

import (
	clientsetex "transwarp/release-config/pkg/client/clientset/versioned"
	"transwarp/release-config/pkg/apis/transwarp/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sutils "walm/pkg/k8s/utils"
	listv1beta1 "transwarp/release-config/pkg/client/listers/transwarp/v1beta1"
)

type ReleaseConfigHandler struct {
	client *clientsetex.Clientset
	lister listv1beta1.ReleaseConfigLister
}

func (handler *ReleaseConfigHandler) GetReleaseConfig(namespace string, name string) (*v1beta1.ReleaseConfig, error) {
	return handler.lister.ReleaseConfigs(namespace).Get(name)
}

func (handler *ReleaseConfigHandler) ListReleaseConfigs(namespace string, labelSelector *metav1.LabelSelector) ([]*v1beta1.ReleaseConfig, error) {
	selector, err := k8sutils.ConvertLabelSelectorToSelector(labelSelector)
	if err != nil {
		return nil, err
	}
	return handler.lister.ReleaseConfigs(namespace).List(selector)
}

