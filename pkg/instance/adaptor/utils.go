package adaptor

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
)

func BuildWalmState(state string, reason string, message string) WalmState {
	return WalmState{
		State: state,
		Reason: reason,
		Message: message,
	}
}

func parsePods(pods []*WalmPod) (bool, *WalmPod, *WalmPod, *WalmPod) {
	allPodsTerminating := true
	var unknownPod, pendingPod, runningPod *WalmPod
	for _, pod := range pods {
		if pod.PodState.State != "Terminating" {
			allPodsTerminating = false
			if pod.PodState.State == "Unknown" {
				unknownPod = pod
				break
			} else if pod.PodState.State == "Pending" {
				pendingPod = pod
			} else if pod.PodState.State == "Running" {
				runningPod = pod
			}
		}
	}
	return allPodsTerminating, unknownPod, pendingPod, runningPod
}

func isNotFoundErr(err error) bool {
	if e, ok := err.(*errors.StatusError); ok {
		if e.Status().Reason == "NotFound" {
			return true
		}
	}
	return false
}

func buildNotFoundWalmModule(module v1beta1.ResourceReference) WalmModule {
	return WalmModule{
		Kind:        module.ResourceRef.Kind,
		ModuleState: BuildWalmState("NotFound", "NotFound", fmt.Sprintf("%s %s/%s is not found", module.ResourceRef.Kind, module.ResourceRef.Namespace, module.ResourceRef.Name)),
	}
}