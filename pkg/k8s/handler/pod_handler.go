package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sutils "walm/pkg/k8s/utils"
)

type PodHandler struct {
	client *kubernetes.Clientset
}

func (handler PodHandler) GetPod(namespace string, name string) (*v1.Pod, error) {
	return handler.client.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
}

func (handler PodHandler) ListPods(namespace string, labelSelector *metav1.LabelSelector) (*v1.PodList, error) {
	selectorStr, err := k8sutils.ConvertLabelSelectorToStr(labelSelector)
	if err != nil {
		return nil, err
	}
	return handler.client.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector:selectorStr})
}

func (handler PodHandler) CreatePod(namespace string, pod *v1.Pod) (*v1.Pod, error) {
	return handler.client.CoreV1().Pods(namespace).Create(pod)
}

func (handler PodHandler) UpdatePod(namespace string, pod *v1.Pod) (*v1.Pod, error) {
	return handler.client.CoreV1().Pods(namespace).Update(pod)
}

func (handler PodHandler) DeletePod(namespace string, name string) (error) {
	return handler.client.CoreV1().Pods(namespace).Delete(name, &metav1.DeleteOptions{})
}

func NewPodHandler(client *kubernetes.Clientset) (PodHandler) {
	return PodHandler{client:client}
}
