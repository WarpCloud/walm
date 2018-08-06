package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	listv1beta1 "k8s.io/client-go/listers/apps/v1beta1"
)

type StatefulSetHandler struct {
	client *kubernetes.Clientset
	lister listv1beta1.StatefulSetLister
}

func (handler *StatefulSetHandler) GetStatefulSet(namespace string, name string) (*v1beta1.StatefulSet, error) {
	return handler.lister.StatefulSets(namespace).Get(name)
}

func (handler *StatefulSetHandler) CreateStatefulSet(namespace string, statefulSet *v1beta1.StatefulSet) (*v1beta1.StatefulSet, error) {
	return handler.client.AppsV1beta1().StatefulSets(namespace).Create(statefulSet)
}

func (handler *StatefulSetHandler) UpdateStatefulSet(namespace string, statefulSet *v1beta1.StatefulSet) (*v1beta1.StatefulSet, error) {
	return handler.client.AppsV1beta1().StatefulSets(namespace).Update(statefulSet)
}

func (handler *StatefulSetHandler) DeleteStatefulSet(namespace string, name string) (error) {
	return handler.client.AppsV1beta1().StatefulSets(namespace).Delete(name, &metav1.DeleteOptions{})
}
