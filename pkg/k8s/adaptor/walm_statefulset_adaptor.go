package adaptor

import (
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	"k8s.io/api/apps/v1"
	"WarpCloud/walm/pkg/k8s/handler"
)

type WalmStatefulSetAdaptor struct {
	statefulSetHandler *handler.StatefulSetHandler
	podAdaptor         *WalmPodAdaptor
}

func (adaptor *WalmStatefulSetAdaptor) GetResource(namespace string, name string) (WalmResource, error) {
	statefulSet, err := adaptor.statefulSetHandler.GetStatefulSet(namespace, name)
	if err != nil {
		if IsNotFoundErr(err) {
			return WalmStatefulSet{
				WalmMeta: buildNotFoundWalmMeta("StatefulSet", namespace, name),
			}, nil
		}
		return WalmStatefulSet{}, err
	}

	return adaptor.buildWalmStatefulSet(statefulSet)
}

func (adaptor *WalmStatefulSetAdaptor) buildWalmStatefulSet(statefulSet *appsv1beta1.StatefulSet) (walmStatefulSet WalmStatefulSet, err error) {
	walmStatefulSet = WalmStatefulSet{
		WalmMeta:       buildWalmMetaWithoutState("StatefulSet", statefulSet.Namespace, statefulSet.Name),
		Labels:         statefulSet.Labels,
		Annotations:    statefulSet.Annotations,
		ReadyReplicas:  statefulSet.Status.ReadyReplicas,
		CurrentVersion: statefulSet.Status.CurrentRevision,
		UpdateVersion:  statefulSet.Status.UpdateRevision,
		Selector:       statefulSet.Spec.Selector,
	}

	if statefulSet.Spec.Replicas == nil {
		walmStatefulSet.ExpectedReplicas = 1
	} else {
		walmStatefulSet.ExpectedReplicas = *statefulSet.Spec.Replicas
	}

	walmStatefulSet.Pods, err = adaptor.podAdaptor.GetWalmPods(statefulSet.Namespace, statefulSet.Spec.Selector)
	walmStatefulSet.State = buildWalmStatefulSetState(statefulSet, walmStatefulSet.Pods)
	return walmStatefulSet, err
}

func buildWalmStatefulSetState(statefulSet *appsv1beta1.StatefulSet, pods []*WalmPod) (walmState WalmState) {
	if isStatefulSetReady(statefulSet) {
		walmState = buildWalmState("Ready", "", "")
	} else {
		walmState = buildWalmStateByPods(pods, "StatefulSet")
	}
	return walmState
}
func isStatefulSetReady(statefulSet *appsv1beta1.StatefulSet) bool {
	if statefulSet.Spec.Replicas != nil && statefulSet.Status.ReadyReplicas < *statefulSet.Spec.Replicas {
		return false
	}

	if statefulSet.Spec.UpdateStrategy.Type == v1.RollingUpdateStatefulSetStrategyType && statefulSet.Spec.UpdateStrategy.RollingUpdate != nil {
		if statefulSet.Spec.Replicas != nil && statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
			if statefulSet.Status.UpdatedReplicas < (*statefulSet.Spec.Replicas - *statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition) {
				return false
			}
			return true
		}
	}

	if statefulSet.Status.UpdateRevision != statefulSet.Status.CurrentRevision {
		return false
	}

	return true
}
