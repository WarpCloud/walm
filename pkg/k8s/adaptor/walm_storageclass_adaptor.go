package adaptor

import (
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"WarpCloud/walm/pkg/k8s/handler"
	"github.com/sirupsen/logrus"
)

type WalmStorageClassAdaptor struct {
	handler *handler.StorageClassHandler
}

func (adaptor *WalmStorageClassAdaptor) GetResource(namespace string, name string) (WalmResource, error) {
	storageClass, err := adaptor.handler.GetStorageClass(name)
	if err != nil {
		if IsNotFoundErr(err) {
			return WalmStorageClass{
				WalmMeta: buildNotFoundWalmMeta("StorageClass", namespace, name),
			}, nil
		}
		return WalmStorageClass{}, err
	}

	return adaptor.BuildWalmStorageClass(*storageClass)
}

func (adaptor *WalmStorageClassAdaptor) GetWalmStorageClasses(namespace string, labelSelector *metav1.LabelSelector) ([]*WalmStorageClass, error) {
	storageClassList, err := adaptor.handler.ListStorageClasss(labelSelector)
	if err != nil {
		return nil, err
	}

	walmStorageClasss := []*WalmStorageClass{}
	if storageClassList != nil {
		for _, storageClass := range storageClassList {
			walmStorageClass, err := adaptor.BuildWalmStorageClass(*storageClass)
			if err != nil {
				logrus.Errorf("failed to build walm storageClass : %s", err.Error())
				return nil, err
			}
			walmStorageClasss = append(walmStorageClasss, walmStorageClass)
		}
	}

	return walmStorageClasss, nil
}

func (adaptor *WalmStorageClassAdaptor) BuildWalmStorageClass(storageClass storagev1.StorageClass) (walmStorageClass *WalmStorageClass, err error) {
	walmStorageClass = &WalmStorageClass{
		WalmMeta:    buildWalmMeta("StorageClass", storageClass.Namespace, storageClass.Name, buildWalmState("Ready", "", "")),
		Provisioner:      storageClass.Provisioner,
	}

	if storageClass.AllowVolumeExpansion != nil {
		walmStorageClass.AllowVolumeExpansion = *storageClass.AllowVolumeExpansion
	}
	if storageClass.ReclaimPolicy != nil {
		walmStorageClass.ReclaimPolicy = string(*storageClass.ReclaimPolicy)
	}
	if storageClass.VolumeBindingMode != nil {
		walmStorageClass.VolumeBindingMode = string(*storageClass.VolumeBindingMode)
	}
	return
}

