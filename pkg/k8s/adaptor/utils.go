package adaptor

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"strings"
	"k8s.io/api/core/v1"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func buildWalmState(state string, reason string, message string) WalmState {
	return WalmState{
		Status:  state,
		Reason:  reason,
		Message: message,
	}
}

func buildWalmStateByPods(pods []*WalmPod, controllerKind string) (walmState WalmState) {
	if len(pods) == 0 {
		walmState = buildWalmState("Pending", "PodNotCreated", "There is no pod created")
		return
	}

	allPodsTerminating := true
	for _, pod := range pods {
		if pod.State.Status != "Terminating" {
			allPodsTerminating = false
			if pod.State.Status == "Unknown" {
				walmState = buildWalmState("Pending", "PodUnknown", fmt.Sprintf("Pod %s/%s is in state Unknown", pod.Namespace, pod.Name))
				return
			} else if pod.State.Status == "Pending" {
				walmState = buildWalmState("Pending", "PodPending", fmt.Sprintf("Pod %s/%s is in state Pending", pod.Namespace, pod.Name))
				return
			} else if pod.State.Status == "Running" {
				walmState = buildWalmState("Pending", "PodRunning", fmt.Sprintf("Pod %s/%s is in state Running", pod.Namespace, pod.Name))
				return
			}
		}
	}

	if allPodsTerminating {
		walmState = buildWalmState("Terminating", "", "")
	} else {
		walmState = buildWalmState("Pending", controllerKind + "Updating", controllerKind + " is updating")
	}
	return
}

func IsNotFoundErr(err error) bool {
	if e, ok := err.(*errors.StatusError); ok {
		if e.Status().Reason == metav1.StatusReasonNotFound {
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

func formatEventSource(es v1.EventSource) string {
	EventSourceString := []string{es.Component}
	if len(es.Host) > 0 {
		EventSourceString = append(EventSourceString, es.Host)
	}
	return strings.Join(EventSourceString, ", ")
}
