package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sutils "walm/pkg/k8s/utils"
	listv1 "k8s.io/client-go/listers/storage/v1"
)

type StorageClassHandler struct {
	client *kubernetes.Clientset
	lister listv1.StorageClassLister
}

func (handler *StorageClassHandler) GetStorageClass(name string) (*v1.StorageClass, error){
	return handler.lister.Get(name)
}

func (handler *StorageClassHandler) ListStorageClasss(labelSelector *metav1.LabelSelector) ([]*v1.StorageClass, error){
	selector, err := k8sutils.ConvertLabelSelectorToSelector(labelSelector)
	if err != nil {
		return nil, err
	}
	return handler.lister.List(selector)
}




