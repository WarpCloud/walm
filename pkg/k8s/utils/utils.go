package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/api/core/v1"
)

func ConvertLabelSelectorToStr(labelSelector *metav1.LabelSelector) (string, error) {
	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return "", err
	}
	return selector.String(), nil
}

func ConvertLabelSelectorToSelector(labelSelector *metav1.LabelSelector) (labels.Selector, error) {
	if labelSelector == nil {
		return labels.NewSelector(), nil
	}
	return metav1.LabelSelectorAsSelector(labelSelector)
}

func MergeLabels(labels map[string]string, newLabels map[string]string, remove []string) map[string]string{
	if labels == nil {
		labels = make(map[string]string)
	}
	for key, value := range newLabels {
		labels[key] = value
	}
	for _, label := range remove {
		delete(labels, label)
	}
	return labels
}

type SortableEvents []v1.Event

func (list SortableEvents) Len() int {
	return len(list)
}

func (list SortableEvents) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list SortableEvents) Less(i, j int) bool {
	return list[i].LastTimestamp.Time.Before(list[j].LastTimestamp.Time)
}