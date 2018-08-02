package adaptor

import (
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"walm/pkg/instance/lister"
	"fmt"
)

type WalmDaemonSetAdaptor struct{
	Lister lister.K8sResourceLister
}

func(adaptor WalmDaemonSetAdaptor) GetWalmModule(module v1beta1.ResourceReference) (WalmModule, error) {
	walmDaemonSet, err := adaptor.GetWalmDaemonSet(module.ResourceRef.Namespace, module.ResourceRef.Name)
	if err != nil {
		if isNotFoundErr(err) {
			return buildNotFoundWalmModule(module), nil
		}
		return WalmModule{}, err
	}

	return WalmModule{Kind: module.ResourceRef.Kind, Resource: walmDaemonSet, ModuleState: walmDaemonSet.DaemonSetState}, nil
}

func (adaptor WalmDaemonSetAdaptor) GetWalmDaemonSet(namespace string, name string) (WalmDaemonSet, error) {
	daemonSet, err := adaptor.Lister.GetDaemonSet(namespace, name)
	if err != nil {
		return WalmDaemonSet{}, err
	}

	return adaptor.BuildWalmDaemonSet(daemonSet)
}

func (adaptor WalmDaemonSetAdaptor) BuildWalmDaemonSet(daemonSet *extv1beta1.DaemonSet) (walmDaemonSet WalmDaemonSet, err error){
	walmDaemonSet = WalmDaemonSet{
		WalmMeta: WalmMeta{Name: daemonSet.Name, Namespace: daemonSet.Namespace},
	}

	walmDaemonSet.Pods, err = WalmPodAdaptor{adaptor.Lister}.GetWalmPods(daemonSet.Namespace, daemonSet.Spec.Selector)
	walmDaemonSet.DaemonSetState = BuildWalmDaemonSetState(daemonSet, walmDaemonSet.Pods)
	return walmDaemonSet, err
}

func isDaemonSetReady(daemon *extv1beta1.DaemonSet) bool {
	if daemon.Status.UpdatedNumberScheduled < daemon.Status.DesiredNumberScheduled {
		return false
	}
	if daemon.Status.NumberAvailable < daemon.Status.DesiredNumberScheduled {
		return false
	}
	return true
}

func BuildWalmDaemonSetState(daemonSet *extv1beta1.DaemonSet, pods []*WalmPod) (walmState WalmState) {
	if isDaemonSetReady(daemonSet) {
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
					walmState = BuildWalmState("Pending", "DeploymentUpdating", "Deployment is updating")
				}
			}
		}
	}
	return walmState
}
