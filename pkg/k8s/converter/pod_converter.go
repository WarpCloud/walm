package converter

import (
	corev1 "k8s.io/api/core/v1"
	"time"
	"k8s.io/apimachinery/pkg/util/duration"
	"WarpCloud/walm/pkg/models/k8s"
	"fmt"
)

func ConvertPodFromK8s(oriPod *corev1.Pod) (*k8s.Pod, error) {
	if oriPod == nil {
		return nil, nil
	}
	pod := oriPod.DeepCopy()

	walmPod := &k8s.Pod{
		Meta:        k8s.NewMeta(k8s.PodKind, pod.Namespace, pod.Name, buildWalmPodState(pod)),
		Labels:      map[string]string{},
		Annotations: map[string]string{},
		PodIp:       pod.Status.PodIP,
		HostIp:      pod.Status.HostIP,
		Containers:  buildWalmPodContainers(pod),
		Age:         duration.ShortHumanDuration(time.Since(pod.CreationTimestamp.Time)),
	}
	if len(pod.Labels) > 0 {
		walmPod.Labels = pod.Labels
	}
	if len(pod.Annotations) > 0 {
		walmPod.Annotations = pod.Annotations
	}
	return walmPod, nil
}

func buildWalmPodContainers(pod *corev1.Pod) (walmContainers []k8s.Container) {
	walmContainers = []k8s.Container{}
	for _, container := range pod.Status.ContainerStatuses {
		walmContainer := k8s.Container{
			Name:         container.Name,
			Ready:        container.Ready,
			Image:        container.Image,
			RestartCount: container.RestartCount,
			State:        buildPodContainerState(container.State),
		}
		walmContainers = append(walmContainers, walmContainer)
	}
	return
}

func buildPodContainerState(state corev1.ContainerState) (walmState k8s.State) {
	if state.Terminated != nil {
		walmState = k8s.NewState("Terminated", state.Terminated.Message, state.Terminated.Reason)
	} else if state.Waiting != nil {
		walmState = k8s.NewState("Waiting", state.Waiting.Message, state.Waiting.Reason)
	} else if state.Running != nil {
		walmState = k8s.NewState("Running", "", "")
	}
	return
}

// Pending, Running, Ready, Succeeded, Failed, Terminating, Unknown
func buildWalmPodState(pod *corev1.Pod) k8s.State {
	podState := k8s.State{}
	podState.Status = string(pod.Status.Phase)
	if pod.DeletionTimestamp != nil {
		podState.Status = "Terminating"
	}

	if podState.Status == "Pending" {
		podState.Reason, podState.Message = getPendingReason(pod)
	}

	if podState.Status == "Running" {
		if ready, reason, message := isPodReady(pod); ready {
			podState.Status = "Ready"
		} else {
			podState.Reason = reason
			podState.Message = message
		}
	}

	if podState.Status == "Failed" {
		podState.Reason, podState.Message = getFailedReason(pod)
	}

	return podState
}

func getFailedReason(pod *corev1.Pod) (reason string, message string) {
	for _, containerState := range getContainerStates(pod) {
		if containerState.Terminated != nil && containerState.Terminated.ExitCode != 0 {
			return containerState.Terminated.Reason, containerState.Terminated.Message
		}
	}

	return
}
func isPodReady(pod *corev1.Pod) (ready bool, reason string, message string) {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == "Ready" {
			if condition.Status == "True" {
				ready = true
			} else {
				reason = condition.Reason
				message = condition.Message
			}
			break
		}
	}

	return
}

func getPendingReason(pod *corev1.Pod) (reason string, message string) {
	for _, condition := range pod.Status.Conditions {
		if (condition.Type == "PodScheduled" || condition.Type == "Initialized") && condition.Status != "True" {
			return condition.Reason, condition.Message
		}
	}

	for _, containerState := range getContainerStates(pod) {
		if containerState.Waiting != nil {
			return containerState.Waiting.Reason, containerState.Waiting.Message
		}
	}

	return
}

func getContainerStates(pod *corev1.Pod) []corev1.ContainerState {
	containerStates := []corev1.ContainerState{}
	for _, status := range pod.Status.ContainerStatuses {
		containerStates = append(containerStates, status.State)
	}
	return containerStates
}

func buildWalmStateByPods(pods []*k8s.Pod, controllerKind string) (walmState k8s.State) {
	if len(pods) == 0 {
		walmState = k8s.NewState("Pending", "PodNotCreated", "There is no pod created")
		return
	}

	allPodsTerminating := true
	for _, pod := range pods {
		if pod.State.Status != "Terminating" {
			allPodsTerminating = false
			if pod.State.Status == "Unknown" {
				walmState = k8s.NewState("Pending", "PodUnknown", fmt.Sprintf("Pod %s/%s is in state Unknown", pod.Namespace, pod.Name))
				return
			} else if pod.State.Status == "Pending" {
				walmState = k8s.NewState("Pending", "PodPending", fmt.Sprintf("Pod %s/%s is in state Pending", pod.Namespace, pod.Name))
				return
			} else if pod.State.Status == "Running" {
				walmState = k8s.NewState("Pending", "PodRunning", fmt.Sprintf("Pod %s/%s is in state Running", pod.Namespace, pod.Name))
				return
			}
		}
	}

	if allPodsTerminating {
		walmState = k8s.NewState("Terminating", "", "")
	} else {
		walmState = k8s.NewState("Pending", controllerKind + "Updating", controllerKind + " is updating")
	}
	return
}