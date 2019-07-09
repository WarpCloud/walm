package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"k8s.io/api/storage/v1"
)

func ConvertStorageClassFromK8s(oriStorageClass *v1.StorageClass) (*k8s.StorageClass, error) {
	if oriStorageClass == nil {
		return nil, nil
	}
	storageClass := oriStorageClass.DeepCopy()

	walmStorageClass := &k8s.StorageClass{
		Meta:    k8s.NewMeta(k8s.StorageClassKind, storageClass.Namespace, storageClass.Name, k8s.NewState("Ready", "", "")),
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
	return walmStorageClass, nil
}
