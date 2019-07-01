package converter

import (
	corev1 "k8s.io/api/core/v1"
	"WarpCloud/walm/pkg/models/k8s"
)

func ConvertPvcFromK8s(oriPersistentVolumeClaim *corev1.PersistentVolumeClaim) (pvc *k8s.PersistentVolumeClaim, err error) {
	if oriPersistentVolumeClaim == nil {
		return
	}
	persistentVolumeClaim := oriPersistentVolumeClaim.DeepCopy()

	pvc = &k8s.PersistentVolumeClaim{
		Meta:    k8s.NewMeta(k8s.PersistentVolumeClaimKind, persistentVolumeClaim.Namespace, persistentVolumeClaim.Name, k8s.NewState(string(persistentVolumeClaim.Status.Phase), "", "")),
		VolumeName: persistentVolumeClaim.Spec.VolumeName,
	}
	if persistentVolumeClaim.Spec.AccessModes != nil {
		pvc.AccessModes = []string{}
		for _, accessMode := range persistentVolumeClaim.Spec.AccessModes {
			pvc.AccessModes = append(pvc.AccessModes, string(accessMode))
		}
	}
	if persistentVolumeClaim.Status.Capacity != nil {
		for key, value :=range persistentVolumeClaim.Status.Capacity {
			if key == corev1.ResourceStorage {
				pvc.Capacity = value.String()
				break
			}
		}
	}
	if persistentVolumeClaim.Spec.StorageClassName != nil {
		pvc.StorageClass = *(persistentVolumeClaim.Spec.StorageClassName)
	}
	if persistentVolumeClaim.Spec.VolumeMode != nil {
		pvc.VolumeMode = string(*persistentVolumeClaim.Spec.VolumeMode)
	}

	return
}


