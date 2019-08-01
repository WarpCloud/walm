package converter

import (
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"WarpCloud/walm/pkg/models/k8s"
	"k8s.io/api/core/v1"
)

func ConvertDaemonSetFromK8s(oriDaemonSet *extv1beta1.DaemonSet, pods []*v1.Pod) (walmDaemonSet *k8s.DaemonSet, err error) {
	if oriDaemonSet == nil {
		return
	}
	daemonSet := oriDaemonSet.DeepCopy()

	walmDaemonSet = &k8s.DaemonSet{
		Meta:               k8s.NewEmptyStateMeta(k8s.DaemonSetKind, daemonSet.Namespace, daemonSet.Name),
		Labels:                 daemonSet.Labels,
		Annotations:            daemonSet.Annotations,
		DesiredNumberScheduled: daemonSet.Status.DesiredNumberScheduled,
		UpdatedNumberScheduled: daemonSet.Status.UpdatedNumberScheduled,
		NumberAvailable:        daemonSet.Status.NumberAvailable,
	}

	for _, pod := range pods {
		walmPod, err := ConvertPodFromK8s(pod)
		if err != nil {
			return nil, err
		}
		walmDaemonSet.Pods = append(walmDaemonSet.Pods, walmPod)
	}

	walmDaemonSet.State = buildWalmDaemonSetState(daemonSet, walmDaemonSet.Pods)
	return walmDaemonSet, nil
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

func buildWalmDaemonSetState(daemonSet *extv1beta1.DaemonSet, pods []*k8s.Pod) (walmState k8s.State) {
	if isDaemonSetReady(daemonSet) {
		walmState = k8s.NewState("Ready", "", "")
	} else {
		walmState = buildWalmStateByPods(pods, string(k8s.DaemonSetKind))
	}
	return walmState
}
