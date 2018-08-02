package adaptor

import (
	corev1 "k8s.io/api/core/v1"
	"walm/pkg/instance/lister"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WalmPodAdaptor struct{
	Lister lister.K8sResourceLister
}

func (adaptor WalmPodAdaptor) GetWalmPods(namespace string, labelSelector *metav1.LabelSelector) ([]*WalmPod, error) {
	podList, err := adaptor.Lister.GetPods(namespace, labelSelector)
	if err != nil {
		return nil, err
	}

	walmPods := []*WalmPod{}
	if podList != nil {
		for _, pod := range podList {
			walmPod := BuildWalmPod(*pod)
			walmPods = append(walmPods, walmPod)
		}
	}

	return walmPods, nil
}

func BuildWalmPod(pod corev1.Pod) *WalmPod {
	walmPod := WalmPod{
		WalmMeta: WalmMeta{pod.Name, pod.Namespace},
		PodIp:    pod.Status.PodIP,
		PodState: BuildWalmPodState(pod),
	}
	return &walmPod
}

// Pending, Running, Ready, Succeeded, Failed, Terminating, Unknown
func BuildWalmPodState(pod corev1.Pod) WalmState {
	podState := WalmState{}
	podState.State = string(pod.Status.Phase)
	if pod.DeletionTimestamp != nil {
		podState.State = "Terminating"
	}

	if podState.State == "Pending" {
		podState.Reason, podState.Message = getPendingReason(pod)
	}

	if podState.State == "Running" {
		if ready, reason, message := isPodReady(pod); ready {
			podState.State = "Ready"
		} else {
			podState.Reason = reason
			podState.Message = message
		}
	}

	if podState.State == "Failed" {
		podState.Reason, podState.Message = getFailedReason(pod)
	}

	return podState
}

func getFailedReason(pod corev1.Pod) (reason string, message string) {
	for _, containerState := range getContainerStates(pod) {
		if containerState.Terminated != nil && containerState.Terminated.ExitCode != 0{
			return containerState.Terminated.Reason, containerState.Terminated.Message
		}
	}

	return
}
func isPodReady(pod corev1.Pod) (ready bool, reason string, message string){
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

func getPendingReason(pod corev1.Pod) (reason string, message string) {
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

func getContainerStates(pod corev1.Pod) []corev1.ContainerState {
	containerStates := []corev1.ContainerState{}
	for _, status := range pod.Status.ContainerStatuses {
		containerStates = append(containerStates, status.State)
	}
	return containerStates
}


