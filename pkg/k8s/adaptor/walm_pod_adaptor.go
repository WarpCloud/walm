package adaptor

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"WarpCloud/walm/pkg/k8s/handler"
	"github.com/sirupsen/logrus"
	"sort"
	"WarpCloud/walm/pkg/k8s/utils"
)

type WalmPodAdaptor struct {
	handler      *handler.PodHandler
	eventHandler *handler.EventHandler
}

func (adaptor *WalmPodAdaptor) GetResource(namespace string, name string) (WalmResource, error) {
	pod, err := adaptor.handler.GetPod(namespace, name)
	if err != nil {
		if IsNotFoundErr(err) {
			return WalmPod{
				WalmMeta: buildNotFoundWalmMeta("Pod", namespace, name),
			}, nil
		}
		return WalmPod{}, err
	}

	return BuildWalmPod(*pod), nil
}

func (adaptor *WalmPodAdaptor) GetWalmPods(namespace string, labelSelector *metav1.LabelSelector) ([]*WalmPod, error) {
	podList, err := adaptor.handler.ListPods(namespace, labelSelector)
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

func (adaptor *WalmPodAdaptor) GetWalmPodEventList(namespace, name string) (*WalmEventList, error) {
	pod, err := adaptor.handler.GetPod(namespace, name)
	if err != nil {
		logrus.Errorf("failed to get pod : %s", err.Error())
		return nil, err
	}
	eventList := &WalmEventList{}
	eventList.Events, err = adaptor.GetWalmPodEvents(pod)
	if err != nil {
		logrus.Errorf("failed to get pod Events : %s", err.Error())
		return nil, err
	}
	return eventList, nil
}

func (adaptor *WalmPodAdaptor) GetWalmPodEvents(pod *corev1.Pod) ([]WalmEvent, error) {
	ref := &corev1.ObjectReference{
		Namespace:       pod.Namespace,
		Name:            pod.Name,
		Kind:            pod.Kind,
		ResourceVersion: pod.ResourceVersion,
		UID:             pod.UID,
		APIVersion:      pod.APIVersion,
	}

	podEvents, err := adaptor.eventHandler.SearchEvents(pod.Namespace, ref)
	if err != nil {
		logrus.Errorf("failed to get Events : %s", err.Error())
		return nil, err
	}
	sort.Sort(utils.SortableEvents(podEvents.Items))

	walmEvents := []WalmEvent{}
	for _, event := range podEvents.Items {
		walmEvent := WalmEvent{
			Type:           event.Type,
			Reason:         event.Reason,
			Message:        event.Message,
			Count:          event.Count,
			FirstTimestamp: event.FirstTimestamp,
			LastTimestamp:  event.LastTimestamp,
			From:           formatEventSource(event.Source),
		}
		walmEvents = append(walmEvents, walmEvent)
	}
	return walmEvents, nil
}

func (adaptor *WalmPodAdaptor) GetWalmPodLogs(namespace, podName, containerName string, tailLines int64) (string, error) {
	podLogOptions := &corev1.PodLogOptions{}
	if containerName != "" {
		podLogOptions.Container = containerName
	}
	if tailLines != 0 {
		podLogOptions.TailLines = &tailLines
	}
	logs, err := adaptor.handler.GetPodLogs(namespace, podName, podLogOptions)
	if err != nil {
		logrus.Errorf("failed to get pod logs : %s", err.Error())
		return "", err
	}
	return string(logs), nil
}

func BuildWalmPod(pod corev1.Pod) *WalmPod {
	walmPod := WalmPod{
		WalmMeta:    buildWalmMeta("Pod", pod.Namespace, pod.Name, BuildWalmPodState(pod)),
		Labels:      map[string]string{},
		Annotations: map[string]string{},
		PodIp:       pod.Status.PodIP,
		HostIp:      pod.Status.HostIP,
		Containers:  buildWalmPodContainers(pod),
	}
	if len(pod.Labels) > 0 {
		walmPod.Labels = pod.Labels
	}
	if len(pod.Annotations) > 0 {
		walmPod.Annotations = pod.Annotations
	}
	return &walmPod
}

func buildWalmPodContainers(pod corev1.Pod) (walmContainers []WalmContainer) {
	walmContainers = []WalmContainer{}
	for _, container := range pod.Status.ContainerStatuses {
		walmContainer := WalmContainer{
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
func buildPodContainerState(state corev1.ContainerState) (walmState WalmState) {
	if state.Terminated != nil {
		walmState = buildWalmState("Terminated", state.Terminated.Message, state.Terminated.Reason)
	} else if state.Waiting != nil {
		walmState = buildWalmState("Waiting", state.Waiting.Message, state.Waiting.Reason)
	} else if state.Running != nil {
		walmState = buildWalmState("Running", "", "")
	}
	return
}

// Pending, Running, Ready, Succeeded, Failed, Terminating, Unknown
func BuildWalmPodState(pod corev1.Pod) WalmState {
	podState := WalmState{}
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

func getFailedReason(pod corev1.Pod) (reason string, message string) {
	for _, containerState := range getContainerStates(pod) {
		if containerState.Terminated != nil && containerState.Terminated.ExitCode != 0 {
			return containerState.Terminated.Reason, containerState.Terminated.Message
		}
	}

	return
}
func isPodReady(pod corev1.Pod) (ready bool, reason string, message string) {
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
