package adaptor

import (
	"testing"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
)

func TestBuildWalmStateByPods(t *testing.T) {
	tests := []struct {
		pods           []*WalmPod
		controllerKind string
		result         WalmState
	}{
		{
			pods:   []*WalmPod{},
			result: buildWalmState("Pending", "PodNotCreated", "There is no pod created"),
		},
		{
			pods:   []*WalmPod{buildWalmPod("test1", "Pending")},
			result: buildWalmState("Pending", "PodPending", "Pod default/test1 is in state Pending"),
		},
		{
			pods:   []*WalmPod{buildWalmPod("test1", "Running")},
			result: buildWalmState("Pending", "PodRunning", "Pod default/test1 is in state Running"),
		},
		{
			pods:   []*WalmPod{buildWalmPod("test1", "Unknown")},
			result: buildWalmState("Pending", "PodUnknown", "Pod default/test1 is in state Unknown"),
		},
		{
			pods:   []*WalmPod{buildWalmPod("test1", "Terminating")},
			result: buildWalmState("Terminating", "", ""),
		},
		{
			pods:   []*WalmPod{buildWalmPod("test1", "Ready")},
			controllerKind: "Deployment",
			result: buildWalmState("Pending", "DeploymentUpdating", "Deployment is updating"),
		},
	}

	for _, test := range tests {
		result := buildWalmStateByPods(test.pods, test.controllerKind)
		assert.Equal(t, test.result, result)
	}
}

func buildWalmPod(name string, state string) *WalmPod {
	return &WalmPod{
		WalmMeta: WalmMeta{Namespace: "default", Name: name, State: WalmState{Status: state}},
	}
}

func TestIsNotFoundErr(t *testing.T) {
	tests := []struct {
		err error
		result bool
	} {
		{
			err : fmt.Errorf("unknown error"),
			result: false,
		},
		{
			err: &errors.StatusError{ErrStatus: metav1.Status{
				Reason: metav1.StatusReasonNotFound ,
			}},
			result: true,
		},
	}

	for _, test:= range tests {
		result := IsNotFoundErr(test.err)
		assert.Equal(t, test.result, result)
	}
}

func TestFormatEventSource(t *testing.T) {
	tests := []struct {
		es v1.EventSource
		result string
	} {
		{
			es: v1.EventSource{
				Host: "172.0.0.1",
				Component: "kubelet",
			},
			result: "kubelet, 172.0.0.1",
		},
		{
			es: v1.EventSource{
				Component: "kubelet",
			},
			result: "kubelet",
		},
	}

	for _, test := range tests {
		result := formatEventSource(test.es)
		assert.Equal(t, test.result, result)
	}

}