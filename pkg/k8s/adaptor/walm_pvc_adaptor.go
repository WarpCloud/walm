package adaptor

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"walm/pkg/k8s/handler"
	"github.com/sirupsen/logrus"
	"fmt"
)

type WalmPersistentVolumeClaimAdaptor struct {
	handler *handler.PersistentVolumeClaimHandler
	statefulsetHandler *handler.StatefulSetHandler
}

func (adaptor *WalmPersistentVolumeClaimAdaptor) GetResource(namespace string, name string) (WalmResource, error) {
	persistentVolumeClaim, err := adaptor.handler.GetPersistentVolumeClaim(namespace, name)
	if err != nil {
		if IsNotFoundErr(err) {
			return WalmPersistentVolumeClaim{
				WalmMeta: buildNotFoundWalmMeta("PersistentVolumeClaim", namespace, name),
			}, nil
		}
		return WalmPersistentVolumeClaim{}, err
	}

	return adaptor.BuildWalmPersistentVolumeClaim(persistentVolumeClaim), nil
}

func (adaptor *WalmPersistentVolumeClaimAdaptor) GetWalmPersistentVolumeClaimAdaptors(namespace string, labelSelector *metav1.LabelSelector) ([]*WalmPersistentVolumeClaim, error) {
	pvcList, err := adaptor.handler.ListPersistentVolumeClaims(namespace, labelSelector)
	if err != nil {
		return nil, err
	}

	walmPvcs := []*WalmPersistentVolumeClaim{}
	if pvcList != nil {
		for _, pvc := range pvcList {
			walmPvc := adaptor.BuildWalmPersistentVolumeClaim(pvc)
			walmPvcs = append(walmPvcs, &walmPvc)
		}
	}

	return walmPvcs, nil
}

func (adaptor *WalmPersistentVolumeClaimAdaptor) BuildWalmPersistentVolumeClaim(persistentVolumeClaim *corev1.PersistentVolumeClaim) (walmPersistentVolumeClaim WalmPersistentVolumeClaim) {
	walmPersistentVolumeClaim = WalmPersistentVolumeClaim{
		WalmMeta:    buildWalmMeta("PersistentVolumeClaim", persistentVolumeClaim.Namespace, persistentVolumeClaim.Name, buildWalmState(string(persistentVolumeClaim.Status.Phase), "", "")),
		AccessModes: persistentVolumeClaim.Status.AccessModes,
		VolumeName: persistentVolumeClaim.Spec.VolumeName,
	}
	if persistentVolumeClaim.Status.Capacity != nil {
		for key, value :=range persistentVolumeClaim.Status.Capacity {
			if key == corev1.ResourceStorage {
				walmPersistentVolumeClaim.Capacity = value.String()
				break
			}
		}
	}
	if persistentVolumeClaim.Spec.StorageClassName != nil {
		walmPersistentVolumeClaim.StorageClass = *(persistentVolumeClaim.Spec.StorageClassName)
	}
	if persistentVolumeClaim.Spec.VolumeMode != nil {
		walmPersistentVolumeClaim.VolumeMode = string(*persistentVolumeClaim.Spec.VolumeMode)
	}

	return
}

// if the pvc is related any existed stateful set, it can not be deleted
func  (adaptor *WalmPersistentVolumeClaimAdaptor)DeletePvc(namespace, name string) error{
	pvc, err := adaptor.handler.GetPersistentVolumeClaim(namespace, name)
	if err != nil {
		logrus.Errorf("failed to get pvc %s/%s : %s", namespace, name, err.Error())
		return err
	}

	if len(pvc.Labels) > 0 {
		statefulSets, err := adaptor.statefulsetHandler.ListStatefulSet(namespace, &metav1.LabelSelector{
			MatchLabels: pvc.Labels,
		})
		if err != nil {
			logrus.Errorf("failed to list stateful set : %s", err.Error())
			return err
		}
		if len(statefulSets) > 0 {
			statefulSetNames := make([]string, len(statefulSets))
			for _, statefulSet := range statefulSets {
				statefulSetNames = append(statefulSetNames, statefulSet.Namespace + "/" + statefulSet.Name)
			}
			err = fmt.Errorf("pvc %s/%s can not be deleted, it is still used by statefulsets %v", namespace, name, statefulSetNames)
			return err
		}
	}
	err = adaptor.handler.DeletePersistentVolumeClaim(namespace, name)
	if err != nil {
		logrus.Errorf("failed to delete pvc %s/%s : %s", namespace, name, err.Error())
		return err
	}
	logrus.Infof("succeed to delete pvc %s/%s", namespace, name)
	return nil
}

