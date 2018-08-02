package adaptor

import (
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	"walm/pkg/instance/lister"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	"fmt"
	"k8s.io/api/apps/v1"
)

type WalmStatefulSetAdaptor struct{
	Lister lister.K8sResourceLister
}

func(adaptor WalmStatefulSetAdaptor) GetWalmModule(module v1beta1.ResourceReference) (WalmModule, error) {
	statefulSet, err := adaptor.GetWalmStatefulSet(module.ResourceRef.Namespace, module.ResourceRef.Name)
	if err != nil {
		if isNotFoundErr(err) {
			return buildNotFoundWalmModule(module), nil
		}
		return WalmModule{}, err
	}

	return WalmModule{Kind: module.ResourceRef.Kind, Resource: statefulSet, ModuleState: statefulSet.StatefulSetState}, nil
}

func (adaptor WalmStatefulSetAdaptor) GetWalmStatefulSet(namespace string, name string) (WalmStatefulSet, error) {
	statefulSet, err := adaptor.Lister.GetStatefulSet(namespace, name)
	if err != nil {
		return WalmStatefulSet{}, err
	}

	return adaptor.BuildWalmStatefulSet(statefulSet)
}

func (adaptor WalmStatefulSetAdaptor) BuildWalmStatefulSet(statefulSet *appsv1beta1.StatefulSet) (walmStatefulSet WalmStatefulSet, err error){
	walmStatefulSet = WalmStatefulSet{
		WalmMeta: WalmMeta{Name: statefulSet.Name, Namespace: statefulSet.Namespace},
	}

	walmStatefulSet.Pods, err = WalmPodAdaptor{adaptor.Lister}.GetWalmPods(statefulSet.Namespace, statefulSet.Spec.Selector)
	walmStatefulSet.StatefulSetState = BuildWalmStatefulSetState(statefulSet, walmStatefulSet.Pods)
	return walmStatefulSet, err
}

func BuildWalmStatefulSetState(statefulSet *appsv1beta1.StatefulSet, pods []*WalmPod) (walmState WalmState) {
	if isStatefulSetReady(statefulSet) {
		walmState = BuildWalmState("Ready", "", "")
	} else {
		if len(pods) == 0 {
			walmState = BuildWalmState("Pending","PodNotCreated", "There is no pod created")
		} else {
			allPodsTerminating, unknownPod, pendingPod, runningPod := parsePods(pods)

			if allPodsTerminating {
				walmState = BuildWalmState("Terminating", "", "")
			} else {
				if unknownPod != nil {
					walmState = BuildWalmState("Pending", "PodUnknown", fmt.Sprintf("Pod %s/%s is in state Unknown", unknownPod.Namespace, unknownPod.Name))
				} else if pendingPod != nil {
					walmState = BuildWalmState("Pending", "PodPending", fmt.Sprintf("Pod %s/%s is in state Pending", pendingPod.Namespace, pendingPod.Name))
				} else if runningPod != nil {
					walmState = BuildWalmState("Pending", "PodRunning", fmt.Sprintf("Pod %s/%s is in state Running", runningPod.Namespace, runningPod.Name))
				} else {
					walmState = BuildWalmState("Pending", "StatefulSetUpdating", "StatefulSet is updating")
				}
			}
		}
	}
	return walmState
}
func isStatefulSetReady(statefulSet *appsv1beta1.StatefulSet) bool{
	if statefulSet.Status.ReadyReplicas < *statefulSet.Spec.Replicas && *statefulSet.Spec.Replicas > 0{
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

	if statefulSet.Status.UpdateRevision != statefulSet.Status.CurrentRevision{
		return false
	}

	return true
}
