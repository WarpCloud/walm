package adaptor

import (
	extv1beta1 "k8s.io/api/extensions/v1beta1"
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
		Labels:                 daemonSet.Labels,
		Annotations:            daemonSet.Annotations,
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
		walmState = buildWalmStateByPods(pods, "DaemonSet")
	}
	return walmState
}
