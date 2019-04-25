package handler

import (
	clientsetex "transwarp/release-config/pkg/client/clientset/versioned"
	"transwarp/release-config/pkg/apis/transwarp/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sutils "walm/pkg/k8s/utils"
	listv1beta1 "transwarp/release-config/pkg/client/listers/transwarp/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/evanphx/json-patch"
	"encoding/json"
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

func (handler *ReleaseConfigHandler) CreateReleaseConfig(namespace string, releaseConfig *v1beta1.ReleaseConfig) (*v1beta1.ReleaseConfig, error) {
	return handler.client.TranswarpV1beta1().ReleaseConfigs(namespace).Create(releaseConfig)
}

func (handler *ReleaseConfigHandler) PatchReleaseConfig(namespace, name string, pt types.PatchType, data []byte, subResources... string) (*v1beta1.ReleaseConfig, error) {
	return handler.client.TranswarpV1beta1().ReleaseConfigs(namespace).Patch(name, pt, data, subResources...)
}

func (handler *ReleaseConfigHandler) DeleteReleaseConfig(namespace string, name string) (error) {
	return handler.client.TranswarpV1beta1().ReleaseConfigs(namespace).Delete(name, &metav1.DeleteOptions{})
}

func (handler *ReleaseConfigHandler) GetReleaseConfigFromK8s(namespace string, name string) (*v1beta1.ReleaseConfig, error) {
	return handler.client.TranswarpV1beta1().ReleaseConfigs(namespace).Get(name, metav1.GetOptions{})
}

func (handler *ReleaseConfigHandler) AnnotateReleaseConfig(namespace, name string, annosToAdd map[string]string, annosToRemove []string) error {
	if len(annosToAdd) == 0 && len(annosToRemove) == 0 {
		return nil
	}

	releaseConfig, err := handler.GetReleaseConfigFromK8s(namespace, name)
	if err != nil {
		logrus.Errorf("failed to get release config from k8s : %s", err.Error())
		return err
	}

	oldData, err := json.Marshal(releaseConfig)
	if err != nil {
		return err
	}

	releaseConfig.Annotations = k8sutils.MergeLabels(releaseConfig.Annotations, annosToAdd, annosToRemove)
	newData, err := json.Marshal(releaseConfig)
	if err != nil {
		return err
	}

	patchBytes, err := jsonpatch.CreateMergePatch(oldData, newData)
	if err != nil {
		return err
	}

	_, err = handler.PatchReleaseConfig(namespace, name, types.MergePatchType, patchBytes)
	if err != nil {
		return err
	}

	return nil
}