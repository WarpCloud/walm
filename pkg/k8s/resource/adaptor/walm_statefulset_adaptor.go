package adaptor

import (
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	"fmt"
	"k8s.io/api/apps/v1"
	"walm/pkg/k8s/handler"
)

type WalmStatefulSetAdaptor struct {
	statefulSetHandler *handler.StatefulSetHandler
	podHandler         *handler.PodHandler
}

func (adaptor WalmStatefulSetAdaptor) GetResource(namespace string, name string) (WalmResource, error) {
	statefulSet, err := adaptor.statefulSetHandler.GetStatefulSet(namespace, name)
	if err != nil {
		if isNotFoundErr(err) {
			return WalmStatefulSet{
				WalmMeta: buildNotFoundWalmMeta("StatefulSet", namespace, name),
			}, nil
		}
		return WalmStatefulSet{}, err
	}

	return adaptor.buildWalmStatefulSet(statefulSet)
}

func (adaptor WalmStatefulSetAdaptor) buildWalmStatefulSet(statefulSet *appsv1beta1.StatefulSet) (walmStatefulSet WalmStatefulSet, err error) {
	walmStatefulSet = WalmStatefulSet{
		WalmMeta: buildWalmMetaWithoutState("StatefulSet", statefulSet.Namespace, statefulSet.Name),
	}

	walmStatefulSet.Pods, err = WalmPodAdaptor{adaptor.podHandler}.GetWalmPods(statefulSet.Namespace, statefulSet.Spec.Selector)
	walmStatefulSet.State = buildWalmStatefulSetState(statefulSet, walmStatefulSet.Pods)
	return walmStatefulSet, err
}

func buildWalmStatefulSetState(statefulSet *appsv1beta1.StatefulSet, pods []*WalmPod) (walmState WalmState) {
	if isStatefulSetReady(statefulSet) {
		walmState = buildWalmState("Ready", "", "")
	} else {
		if len(pods) == 0 {
			walmState = buildWalmState("Pending", "PodNotCreated", "There is no pod created")
		} else {
			allPodsTerminating, unknownPod, pendingPod, runningPod := parsePods(pods)

			if allPodsTerminating {
				walmState = buildWalmState("Terminating", "", "")
			} else {
				if unknownPod != nil {
					walmState = buildWalmState("Pending", "PodUnknown", fmt.Sprintf("Pod %s/%s is in state Unknown", unknownPod.Namespace, unknownPod.Name))
				} else if pendingPod != nil {
					walmState = buildWalmState("Pending", "PodPending", fmt.Sprintf("Pod %s/%s is in state Pending", pendingPod.Namespace, pendingPod.Name))
				} else if runningPod != nil {
					walmState = buildWalmState("Pending", "PodRunning", fmt.Sprintf("Pod %s/%s is in state Running", runningPod.Namespace, runningPod.Name))
				} else {
					walmState = buildWalmState("Pending", "StatefulSetUpdating", "StatefulSet is updating")
				}
			}
		}
	}
	return walmState
}
func isStatefulSetReady(statefulSet *appsv1beta1.StatefulSet) bool {
	if statefulSet.Status.ReadyReplicas < *statefulSet.Spec.Replicas && *statefulSet.Spec.Replicas > 0 {
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
