package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sutils "walm/pkg/k8s/utils"
	listv1 "k8s.io/client-go/listers/core/v1"
)

type LimitRangeHandler struct {
	client *kubernetes.Clientset
	lister listv1.LimitRangeLister
}

func (handler *LimitRangeHandler) GetLimitRange(namespace string, name string) (*v1.LimitRange, error) {
	return handler.lister.LimitRanges(namespace).Get(name)
}

func (handler *LimitRangeHandler) ListLimitRanges(namespace string, labelSelector *metav1.LabelSelector) ([]*v1.LimitRange, error) {
	selector, err := k8sutils.ConvertLabelSelectorToSelector(labelSelector)
	if err != nil {
		return nil, err
	}
	return handler.lister.LimitRanges(namespace).List(selector)
}

func (handler *LimitRangeHandler) CreateLimitRange(namespace string, limitRange *v1.LimitRange) (*v1.LimitRange, error) {
	return handler.client.CoreV1().LimitRanges(namespace).Create(limitRange)
}

func (handler *LimitRangeHandler) UpdateLimitRange(namespace string, limitRange *v1.LimitRange) (*v1.LimitRange, error) {
	return handler.client.CoreV1().LimitRanges(namespace).Update(limitRange)
}

func (handler *LimitRangeHandler) DeleteLimitRange(namespace string, name string) (error) {
	return handler.client.CoreV1().LimitRanges(namespace).Delete(name, &metav1.DeleteOptions{})
}

func (handler *LimitRangeHandler) GetLimitRangeFromK8s(namespace string, name string) (*v1.LimitRange, error) {
	return handler.client.CoreV1().LimitRanges(namespace).Get(name, metav1.GetOptions{})
}

