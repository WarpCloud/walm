package utils

import "testing"

func TestMergeLabels(t *testing.T) {
	labels := MergeLabels(nil, map[string]string{"test1": "test1", "test2": "test2"}, nil)
	if len(labels) != 2 {
		t.Fail()
	}
	if labels["test1"] != "test1" || labels["test2"] != "test2" {
		t.Fail()
	}

	labels = MergeLabels(labels, map[string]string{"test3": "test3"}, nil)
	if len(labels) != 3 {
		t.Fail()
	}
	if labels["test1"] != "test1" || labels["test2"] != "test2" || labels["test3"] != "test3" {
		t.Fail()
	}

	labels = MergeLabels(labels, map[string]string{"test3": "test33"}, nil)
	if len(labels) != 3 {
		t.Fail()
	}
	if labels["test1"] != "test1" || labels["test2"] != "test2" || labels["test3"] != "test33" {
		t.Fail()
	}

	labels = MergeLabels(labels, nil, []string{"test3"})
	if len(labels) != 2 {
		t.Fail()
	}
	if labels["test1"] != "test1" || labels["test2"] != "test2" {
		t.Fail()
	}
 }
