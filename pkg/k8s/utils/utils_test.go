package utils

import (
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"testing"
)

func TestMergeLabels(t *testing.T) {
	tests := []struct {
		oriLabels    map[string]string
		addLabels    map[string]string
		removeLabels []string
		mergedLabels map[string]string
	}{
		{
			oriLabels:    nil,
			addLabels:    map[string]string{"test1": "test1", "test2": "test2"},
			removeLabels: nil,
			mergedLabels: map[string]string{"test1": "test1", "test2": "test2"},},
		{
			oriLabels:    map[string]string{"test1": "test1", "test2": "test2"},
			addLabels:    map[string]string{"test3": "test3"},
			removeLabels: nil,
			mergedLabels: map[string]string{"test1": "test1", "test2": "test2", "test3": "test3"},},
		{
			oriLabels:    map[string]string{"test1": "test1", "test2": "test2", "test3": "test3"},
			addLabels:    nil,
			removeLabels: []string{"test3"},
			mergedLabels: map[string]string{"test1": "test1", "test2": "test2"},},
	}

	for _, test := range tests {
		mergedLabels := MergeLabels(test.oriLabels, test.addLabels, test.removeLabels)
		assert.Equal(t, test.mergedLabels, mergedLabels)
	}
}

func TestConvertLabelSelectorToStr(t *testing.T) {
	tests := []struct {
		labelSelector *metav1.LabelSelector
		result        string
		err           error
	}{
		{
			labelSelector: &v1.LabelSelector{MatchLabels: map[string]string{"test1": "test1", "test2": "test2"}},
			result:        "test1=test1,test2=test2",
			err:           nil,
		},
	}

	for _, test := range tests {
		result, err := ConvertLabelSelectorToStr(test.labelSelector)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.result, result)
	}
}

func TestConvertLabelSelectorToSelector(t *testing.T) {
	tests := []struct {
		labelSelector *metav1.LabelSelector
		result        string
		err           error
	}{
		{
			labelSelector: &v1.LabelSelector{MatchLabels: map[string]string{"test1": "test1", "test2": "test2"}},
			result:        "test1=test1,test2=test2",
			err:           nil,
		},
		{
			labelSelector: nil,
			result:        "",
			err:           nil,
		},
	}

	for _, test := range tests {
		result, err := ConvertLabelSelectorToSelector(test.labelSelector)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.result, result.String())
	}
}

func TestSortableEvents(t *testing.T) {
	tests := []struct {
		events []corev1.Event
		result []corev1.Event
	}{
		{
			events: []corev1.Event{
				{LastTimestamp: metav1.Unix(4000000, 0)},
				{LastTimestamp: metav1.Unix(2000000, 0)},
			},
			result: []corev1.Event{
				{LastTimestamp: metav1.Unix(2000000, 0)},
				{LastTimestamp: metav1.Unix(4000000, 0)},
			},
		},
	}
	for _, test := range tests {
		sort.Sort(SortableEvents(test.events))
		assert.Equal(t, test.result, test.events)
	}
}

func TestSortableEvents_Len(t *testing.T) {
	tests := []struct{
		events SortableEvents
		result int
	}{
		{
			events: []corev1.Event{
				{LastTimestamp: metav1.Unix(4000000, 0)},
				{LastTimestamp: metav1.Unix(2000000, 0)},
			},
			result: 2,
		},
		{
			events: []corev1.Event{},
			result: 0,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.events.Len(), test.result)
	}
}

func TestSortableEvents_Swap(t *testing.T) {
	tests := []struct{
		events SortableEvents
		result SortableEvents
	}{
		{
			events: []corev1.Event{
				{LastTimestamp: metav1.Unix(4000000, 0)},
				{LastTimestamp: metav1.Unix(2000000, 0)},
			},
			result: []corev1.Event{
				{LastTimestamp: metav1.Unix(2000000, 0)},
				{LastTimestamp: metav1.Unix(4000000, 0)},
			},
		},
		{
			events: []corev1.Event{
				{LastTimestamp: metav1.Unix(4000000, 0)},
				{LastTimestamp: metav1.Unix(3000000, 0)},
			},
			result: []corev1.Event{
				{LastTimestamp: metav1.Unix(3000000, 0)},
				{LastTimestamp: metav1.Unix(4000000, 0)},
			},
		},
	}
	for _, test := range tests {
		test.events.Swap(0, 1)
		assert.Equal(t, test.events, test.result)
	}
}

func TestSortableEvents_Less(t *testing.T) {
	tests := []struct{
		events SortableEvents
		result bool
	}{
		{
			events: []corev1.Event{
				{LastTimestamp: metav1.Unix(4000000, 0)},
				{LastTimestamp: metav1.Unix(3000000, 0)},
			},
			result: false,
		},
		{
			events: []corev1.Event{
				{LastTimestamp: metav1.Unix(3000000, 0)},
				{LastTimestamp: metav1.Unix(4000000, 0)},
			},
			result: true,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.events.Less(0, 1), test.result)
	}
}

func TestIsK8sResourceNotFoundErr(t *testing.T) {
	tests := []struct{
		err error
		result bool
	}{
		{
			err: &errors.StatusError{
				ErrStatus: metav1.Status{
					TypeMeta: metav1.TypeMeta{
						Kind: "Status",
					},
					Status:   "Failed",
					Message:  "",
					Reason:   metav1.StatusReasonNotFound,
				},
			},
			result: true,
		},
		{
			err: nil,
			result: false,
		},
	}
	for _, test := range tests {
		result := IsK8sResourceNotFoundErr(test.err)
		assert.Equal(t, test.result, result)
	}
}

func TestFormatEventSource(t *testing.T) {
	tests := []struct{
		es corev1.EventSource
		result string
	}{
		{
			es: corev1.EventSource{
				Component: "kubelet",
				Host:      "gke-knative-auto-cluster-default-pool-23c23c4f-xdj0",
			},
			result: "kubelet, gke-knative-auto-cluster-default-pool-23c23c4f-xdj0",
		},
		{
			es: corev1.EventSource{
				Component: "kubelet",
				Host:      "",
			},
			result: "kubelet",
		},
	}

	for _, test := range tests {
		result := FormatEventSource(test.es)
		assert.Equal(t, test.result, result)
	}
}

func TestGetPodRequestsAndLimits(t *testing.T) {
	tests := []struct{
		podSpec corev1.PodSpec
		reqs map[corev1.ResourceName]resource.Quantity
		limits map[corev1.ResourceName]resource.Quantity
	}{
		{
			podSpec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "test-container1",
						Resources: corev1.ResourceRequirements{
							Limits:   corev1.ResourceList{
								corev1.ResourceCPU: resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("2Gi"),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU: resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
						},
					},
					{
						Name: "test-container2",
						Resources: corev1.ResourceRequirements{
							Limits:   corev1.ResourceList{
								corev1.ResourceCPU: resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("5Gi"),
								corev1.ResourceStorage: resource.MustParse("10Gi"),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU: resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("3Gi"),
								corev1.ResourceStorage: resource.MustParse("5Gi"),
							},
						},
					},
				},

				InitContainers: []corev1.Container{
					{
						Name: "test-initContainer1",
						Resources: corev1.ResourceRequirements{
							Limits:   corev1.ResourceList{
								corev1.ResourceCPU: resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("3Gi"),
								corev1.ResourceStorage: resource.MustParse("5Gi"),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU: resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("3Gi"),
								corev1.ResourceStorage: resource.MustParse("5Gi"),
							},
						},
					},
					{
						Name: "test-initContainer2",
						Resources: corev1.ResourceRequirements{
							Limits:	  corev1.ResourceList{
								corev1.ResourceCPU: resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("3Gi"),
								corev1.ResourceStorage: resource.MustParse("5Gi"),
								corev1.ResourceEphemeralStorage: resource.MustParse("10Gi"),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU: resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("3Gi"),
								corev1.ResourceStorage: resource.MustParse("5Gi"),
								corev1.ResourceEphemeralStorage: resource.MustParse("10Gi"),
							},
						},
					},
				},
			},
			reqs: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU: *resource.NewQuantity(2, resource.DecimalSI),
				corev1.ResourceMemory: *resource.NewQuantity(4 * 1024 * K8sResourceMemoryScale, resource.BinarySI),
				corev1.ResourceStorage: resource.MustParse("5Gi"),
				corev1.ResourceEphemeralStorage: resource.MustParse("10Gi"),
			},
			limits: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU: *resource.NewQuantity(2, resource.DecimalSI),
				corev1.ResourceMemory: *resource.NewQuantity(7 * 1024 * K8sResourceMemoryScale, resource.BinarySI),
				corev1.ResourceStorage: resource.MustParse("10Gi"),
				corev1.ResourceEphemeralStorage: resource.MustParse("10Gi"),
			},
		},
	}

	for _, test := range tests {
		reqs, limits := GetPodRequestsAndLimits(test.podSpec)
		assert.Equal(t, test.reqs, reqs)
		assert.Equal(t, test.limits, limits)
	}
}

func TestParseK8sResourceMemory(t *testing.T) {
	tests := []struct{
		strValue string
		result int64
	}{
		{
			strValue: "",
			result: 0,
		},
		{
			strValue: "100k23",
			result: 0,
		},
		{
			strValue: "1Gi",
			result: 1024,
		},
	}

	for _, test := range tests {
		result := ParseK8sResourceMemory(test.strValue)
		assert.Equal(t, test.result, result)
	}
}

func TestParseK8sResourceCpu(t *testing.T) {
	tests := []struct{
		strValue string
		result float64
	}{
		{
			strValue: "",
			result: 0,
		},
		{
			strValue: "500m",
			result: 0.5,
		},
		{
			strValue: "1",
			result: 1,
		},
		{
			strValue: "100k2",
			result: 0,
		},
	}

	for _, test := range tests {
		result := ParseK8sResourceCpu(test.strValue)
		assert.Equal(t, test.result, result)
	}}

func TestParseK8sResourceStorage(t *testing.T) {
	tests := []struct{
		strValue string
		result int64
	}{
		{
			strValue: "10Gi",
			result: 10,
		},
		{
			strValue: "",
			result: 0,
		},
		{
			strValue: "100k23",
			result: 0,
		},
	}

	for _, test := range tests {
		result := ParseK8sResourceStorage(test.strValue)
		assert.Equal(t, test.result, result)
	}
}

func TestParseK8sResourcePod(t *testing.T) {
	tests := []struct{
		strValue string
		result int64
	}{
		{
			strValue: "",
			result: 0,
		},
		{
			strValue: "2",
			result: 2,
		},
		{
			strValue: "100k23",
			result: 0,
		},
	}

	for _, test := range tests {
		result := ParseK8sResourcePod(test.strValue)
		assert.Equal(t, test.result, result)
	}
}
