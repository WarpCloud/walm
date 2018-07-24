package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StatefulSetHandler struct {
	client *kubernetes.Clientset
}

func (handler StatefulSetHandler) GetStatefulSet(namespace string, name string) (*v1beta1.StatefulSet, error) {
	return handler.client.AppsV1beta1().StatefulSets(namespace).Get(name, metav1.GetOptions{})
}

func (handler StatefulSetHandler) CreateStatefulSet(namespace string, statefulSet *v1beta1.StatefulSet) (*v1beta1.StatefulSet, error) {
	return handler.client.AppsV1beta1().StatefulSets(namespace).Create(statefulSet)
}

func (handler StatefulSetHandler) UpdateStatefulSet(namespace string, statefulSet *v1beta1.StatefulSet) (*v1beta1.StatefulSet, error) {
	return handler.client.AppsV1beta1().StatefulSets(namespace).Update(statefulSet)
}

func (handler StatefulSetHandler) DeleteStatefulSet(namespace string, name string) (error) {
	return handler.client.AppsV1beta1().StatefulSets(namespace).Delete(name, &metav1.DeleteOptions{})
}

func NewStatefulSetHandler(client *kubernetes.Clientset) (StatefulSetHandler) {
	return StatefulSetHandler{client: client}
}