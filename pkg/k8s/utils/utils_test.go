package utils

import (
	"testing"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"sort"
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
