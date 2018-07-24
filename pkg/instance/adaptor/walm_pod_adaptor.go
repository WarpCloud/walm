package adaptor

import (
	corev1 "k8s.io/api/core/v1"
	"walm/pkg/instance/lister"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WalmPodAdaptor struct{
	Lister lister.K8sResourceLister
}

func (adaptor WalmPodAdaptor) GetWalmPods(namespace string, labelSelector *metav1.LabelSelector) ([]WalmPod, error) {
	podList, err := adaptor.Lister.GetPods(namespace, labelSelector)
	if err != nil {
		return nil, err
	}

	walmPods := []WalmPod{}
	if podList != nil {
		for _, pod := range podList.Items {
			walmPod := BuildWalmPod(pod)
			walmPods = append(walmPods, walmPod)
		}
	}

	return walmPods, nil
}

func BuildWalmPod(pod corev1.Pod) WalmPod {
	walmPod := WalmPod{
		WalmMeta: WalmMeta{pod.Name, pod.Namespace},
		PodIp:    pod.Status.PodIP,
	}
	return walmPod
}
