package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"k8s.io/api/core/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"WarpCloud/walm/pkg/k8s/utils"
)

func ConvertStatefulSetFromK8s(oriStatefulSet *appsv1beta1.StatefulSet, pods []*v1.Pod) (walmStatefulSet *k8s.StatefulSet, err error) {
	if oriStatefulSet == nil {
		return
	}
	statefulSet := oriStatefulSet.DeepCopy()

	walmStatefulSet = &k8s.StatefulSet{
		Meta:       k8s.NewEmptyStateMeta(k8s.StatefulSetKind, statefulSet.Namespace, statefulSet.Name),
		Labels:         statefulSet.Labels,
		Annotations:    statefulSet.Annotations,
		ReadyReplicas:  statefulSet.Status.ReadyReplicas,
		CurrentVersion: statefulSet.Status.CurrentRevision,
		UpdateVersion:  statefulSet.Status.UpdateRevision,
	}

	walmStatefulSet.Selector, err = utils.ConvertLabelSelectorToStr(statefulSet.Spec.Selector)
	if err != nil {
		return
	}

	if statefulSet.Spec.Replicas == nil {
		walmStatefulSet.ExpectedReplicas = 1
	} else {
		walmStatefulSet.ExpectedReplicas = *statefulSet.Spec.Replicas
	}

	for _, pod := range pods {
		walmPod, err := ConvertPodFromK8s(pod)
		if err != nil {
			return nil, err
		}
		walmStatefulSet.Pods = append(walmStatefulSet.Pods, walmPod)
	}
	walmStatefulSet.State = buildWalmStatefulSetState(statefulSet, walmStatefulSet.Pods)
	return walmStatefulSet, nil
}

func buildWalmStatefulSetState(statefulSet *appsv1beta1.StatefulSet, pods []*k8s.Pod) (walmState k8s.State) {
	if isStatefulSetReady(statefulSet) {
		walmState = k8s.NewState("Ready", "", "")
	} else {
		walmState = buildWalmStateByPods(pods, "StatefulSet")
	}
	return walmState
}
func isStatefulSetReady(statefulSet *appsv1beta1.StatefulSet) bool {
	if statefulSet.Spec.Replicas != nil && statefulSet.Status.ReadyReplicas < *statefulSet.Spec.Replicas {
		return false
	}

	if statefulSet.Spec.UpdateStrategy.Type == appsv1.RollingUpdateStatefulSetStrategyType && statefulSet.Spec.UpdateStrategy.RollingUpdate != nil {
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

