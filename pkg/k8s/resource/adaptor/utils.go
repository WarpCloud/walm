package adaptor

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"walm/pkg/k8s/handler"
)

func buildWalmState(state string, reason string, message string) WalmState {
	return WalmState{
		Status:  state,
		Reason:  reason,
		Message: message,
	}
}

func parsePods(pods []*WalmPod) (bool, *WalmPod, *WalmPod, *WalmPod) {
	allPodsTerminating := true
	var unknownPod, pendingPod, runningPod *WalmPod
	for _, pod := range pods {
		if pod.State.Status != "Terminating" {
			allPodsTerminating = false
			if pod.State.Status == "Unknown" {
				unknownPod = pod
				break
			} else if pod.State.Status == "Pending" {
				pendingPod = pod
			} else if pod.State.Status == "Running" {
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

func buildWalmMeta(kind string, namespace string, name string, state WalmState) WalmMeta {
	return WalmMeta{
		Kind:      kind,
		Namespace: namespace,
		Name:      name,
		State:     state,
	}
}

func buildWalmMetaWithoutState(kind string, namespace string, name string) WalmMeta {
	return buildWalmMeta(kind, namespace, name, WalmState{})
}

func buildNotFoundWalmMeta(kind string, namespace string, name string) WalmMeta {
	return buildWalmMeta(kind, namespace, name, buildWalmState("NotFound", "", ""))
}

func GetAdaptor(kind string, handlerSet *handler.HandlerSet) (resourceAdaptor ResourceAdaptor) {
	switch kind {
	case "ApplicationInstance":
		resourceAdaptor = WalmInstanceAdaptor{handlerSet}
	case "Deployment":
		resourceAdaptor = WalmDeploymentAdaptor{handlerSet.GetDeploymentHandler(), handlerSet.GetPodHandler()}
	case "Service":
		resourceAdaptor = WalmServiceAdaptor{handlerSet.GetServiceHandler()}
	case "StatefulSet":
		resourceAdaptor = WalmStatefulSetAdaptor{handlerSet.GetStatefulSetHandler(),handlerSet.GetPodHandler()}
	case "DaemonSet":
		resourceAdaptor = WalmDaemonSetAdaptor{handlerSet.GetDaemonSetHandler(), handlerSet.GetPodHandler()}
	case "Job":
		resourceAdaptor = WalmJobAdaptor{handlerSet.GetJobHandler(),handlerSet.GetPodHandler()}
	case "ConfigMap":
		resourceAdaptor = WalmConfigMapAdaptor{handlerSet.GetConfigMapHandler()}
	case "Ingress":
		resourceAdaptor = WalmIngressAdaptor{handlerSet.GetIngressHandler()}
	case "Secret":
		resourceAdaptor = WalmSecretAdaptor{handlerSet.GetSecretHandler()}
	default:
		resourceAdaptor = WalmDefaultAdaptor{kind}
	}
	return
}
