package adaptor

import (
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"fmt"
	"walm/pkg/k8s/handler"
)

type WalmDaemonSetAdaptor struct {
	daemonSetHandler *handler.DaemonSetHandler
	podAdaptor       *WalmPodAdaptor
}

func (adaptor *WalmDaemonSetAdaptor) GetResource(namespace string, name string) (WalmResource, error) {
	daemonSet, err := adaptor.daemonSetHandler.GetDaemonSet(namespace, name)
	if err != nil {
		if IsNotFoundErr(err) {
			return WalmDaemonSet{
				WalmMeta: buildNotFoundWalmMeta("DaemonSet", namespace, name),
			}, nil
		}
		return WalmDaemonSet{}, err
	}

	return adaptor.BuildWalmDaemonSet(daemonSet)
}

func (adaptor *WalmDaemonSetAdaptor) BuildWalmDaemonSet(daemonSet *extv1beta1.DaemonSet) (walmDaemonSet WalmDaemonSet, err error) {
	walmDaemonSet = WalmDaemonSet{
		WalmMeta:               buildWalmMetaWithoutState("DaemonSet", daemonSet.Namespace, daemonSet.Name),
		DesiredNumberScheduled: daemonSet.Status.DesiredNumberScheduled,
		UpdatedNumberScheduled: daemonSet.Status.UpdatedNumberScheduled,
		NumberAvailable:        daemonSet.Status.NumberAvailable,
	}

	walmDaemonSet.Pods, err = adaptor.podAdaptor.GetWalmPods(daemonSet.Namespace, daemonSet.Spec.Selector)
	walmDaemonSet.State = BuildWalmDaemonSetState(daemonSet, walmDaemonSet.Pods)
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
					walmState = buildWalmState("Pending", "DeploymentUpdating", "Deployment is updating")
				}
			}
		}
	}
	return walmState
}
