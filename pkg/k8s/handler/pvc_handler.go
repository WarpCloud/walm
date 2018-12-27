package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	listv1 "k8s.io/client-go/listers/core/v1"
	k8sutils "walm/pkg/k8s/utils"
)

type PersistentVolumeClaimHandler struct {
	client *kubernetes.Clientset
	lister listv1.PersistentVolumeClaimLister
}

func (handler *PersistentVolumeClaimHandler) GetPersistentVolumeClaim(namespace string, name string) (*v1.PersistentVolumeClaim, error) {
	return handler.lister.PersistentVolumeClaims(namespace).Get(name)
}

func (handler *PersistentVolumeClaimHandler) ListPersistentVolumeClaims(namespace string, labelSelector *metav1.LabelSelector) ([]*v1.PersistentVolumeClaim, error) {
	selector, err := k8sutils.ConvertLabelSelectorToSelector(labelSelector)
	if err != nil {
		return nil, err
	}
	return handler.lister.PersistentVolumeClaims(namespace).List(selector)
}

func (handler *PersistentVolumeClaimHandler) CreatePersistentVolumeClaim(namespace string, persistentVolumeClaim *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	return handler.client.CoreV1().PersistentVolumeClaims(namespace).Create(persistentVolumeClaim)
}

func (handler *PersistentVolumeClaimHandler) UpdatePersistentVolumeClaim(namespace string, persistentVolumeClaim *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	return handler.client.CoreV1().PersistentVolumeClaims(namespace).Update(persistentVolumeClaim)
}

func (handler *PersistentVolumeClaimHandler) DeletePersistentVolumeClaim(namespace string, name string) (error) {
	return handler.client.CoreV1().PersistentVolumeClaims(namespace).Delete(name, &metav1.DeleteOptions{})
}
