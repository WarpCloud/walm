package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"strings"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	K8sResourceMemoryScale  int64   = 1024 * 1024
	K8sResourceStorageScale int64   = 1024 * 1024 * 1024
	K8sResourceCpuScale     float64 = 1000

	// k8s resource memory unit
	K8sResourceMemoryUnit = "Mi"

	// k8s resource storage unit
	K8sResourceStorageUnit = "Gi"
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

func MergeLabels(labels map[string]string, newLabels map[string]string, remove []string) map[string]string {
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

func IsK8sResourceNotFoundErr(err error) bool {
	if e, ok := err.(*errors.StatusError); ok {
		if e.Status().Reason == metav1.StatusReasonNotFound {
			return true
		}
	}
	return false
}

func FormatEventSource(es v1.EventSource) string {
	EventSourceString := []string{es.Component}
	if len(es.Host) > 0 {
		EventSourceString = append(EventSourceString, es.Host)
	}
	return strings.Join(EventSourceString, ", ")
}

func GetPodRequestsAndLimits(podSpec v1.PodSpec) (reqs map[v1.ResourceName]resource.Quantity, limits map[v1.ResourceName]resource.Quantity) {
	reqs, limits = map[v1.ResourceName]resource.Quantity{}, map[v1.ResourceName]resource.Quantity{}
	for _, container := range podSpec.Containers {
		for name, quantity := range container.Resources.Requests {
			if value, ok := reqs[name]; !ok {
				reqs[name] = *quantity.Copy()
			} else {
				value.Add(quantity)
				reqs[name] = value
			}
		}
		for name, quantity := range container.Resources.Limits {
			if value, ok := limits[name]; !ok {
				limits[name] = *quantity.Copy()
			} else {
				value.Add(quantity)
				limits[name] = value
			}
		}
	}
	// init containers define the minimum of any resource
	for _, container := range podSpec.InitContainers {
		for name, quantity := range container.Resources.Requests {
			value, ok := reqs[name]
			if !ok {
				reqs[name] = *quantity.Copy()
				continue
			}
			if quantity.Cmp(value) > 0 {
				reqs[name] = *quantity.Copy()
			}
		}
		for name, quantity := range container.Resources.Limits {
			value, ok := limits[name]
			if !ok {
				limits[name] = *quantity.Copy()
				continue
			}
			if quantity.Cmp(value) > 0 {
				limits[name] = *quantity.Copy()
			}
		}
	}
	return
}

func ParseK8sResourceMemory(strValue string) (int64) {
	quantity := resource.MustParse(strValue)
	return quantity.Value() / K8sResourceMemoryScale
}

func ParseK8sResourceCpu(strValue string) (float64) {
	quantity := resource.MustParse(strValue)
	return float64(quantity.MilliValue()) / K8sResourceCpuScale
}

func ParseK8sResourceStorage(strValue string) (int64) {
	quantity := resource.MustParse(strValue)
	return quantity.Value() / K8sResourceStorageScale
}

func ParseK8sResourcePod(strValue string) (int64) {
	quantity := resource.MustParse(strValue)
	return quantity.Value()
}

