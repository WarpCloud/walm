package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	listv1beta1 "k8s.io/client-go/listers/apps/v1beta1"
	k8sutils "WarpCloud/walm/pkg/k8s/utils"
	"k8s.io/apimachinery/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/evanphx/json-patch"
	"encoding/json"
)

type StatefulSetHandler struct {
	client *kubernetes.Clientset
	lister listv1beta1.StatefulSetLister
}

func (handler *StatefulSetHandler) GetStatefulSet(namespace string, name string) (*v1beta1.StatefulSet, error) {
	return handler.lister.StatefulSets(namespace).Get(name)
}

func (handler *StatefulSetHandler) ListStatefulSet(namespace string, labelSelector *metav1.LabelSelector) ([]*v1beta1.StatefulSet, error){
	selector, err := k8sutils.ConvertLabelSelectorToSelector(labelSelector)
	if err != nil {
		return nil, err
	}
	return handler.lister.StatefulSets(namespace).List(selector)
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

func (handler *StatefulSetHandler) PatchStatefulSet(namespace, name string, pt types.PatchType, data []byte, subResources... string) (*v1beta1.StatefulSet, error) {
	return handler.client.AppsV1beta1().StatefulSets(namespace).Patch(name, pt, data, subResources...)
}

func (handler *StatefulSetHandler) GetStatefulSetFromK8s(namespace string, name string) (*v1beta1.StatefulSet, error) {
	return handler.client.AppsV1beta1().StatefulSets(namespace).Get(name, metav1.GetOptions{})
}

func (handler *StatefulSetHandler) Scale(namespace string, name string, replicas int32) (error) {
	ss, err := handler.GetStatefulSetFromK8s(namespace, name)
	if err != nil {
		logrus.Errorf("failed to get stateful set from k8s : %s", err.Error())
		return err
	}

	oldData, err := json.Marshal(ss)
	if err != nil {
		return err
	}

	ss.Spec.Replicas = &replicas
	newData, err := json.Marshal(ss)
	if err != nil {
		return err
	}

	patchBytes, err := jsonpatch.CreateMergePatch(oldData, newData)
	if err != nil {
		return err
	}

	_, err = handler.PatchStatefulSet(namespace, name, types.MergePatchType, patchBytes)
	if err != nil {
		return err
	}

	return nil
}