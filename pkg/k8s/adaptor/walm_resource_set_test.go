package adaptor

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestAddToResourceSet(t *testing.T) {
	tests := []struct {
		resource WalmResource
		result   *WalmResourceSet
	}{
		{resource: WalmConfigMap{}, result: &WalmResourceSet{ConfigMaps: []WalmConfigMap{{}}}},
		{resource: WalmStatefulSet{}, result: &WalmResourceSet{StatefulSets: []WalmStatefulSet{{}}}},
		{resource: WalmSecret{}, result: &WalmResourceSet{Secrets: []WalmSecret{{}}}},
		{resource: WalmService{}, result: &WalmResourceSet{Services: []WalmService{{}}}},
		{resource: WalmDeployment{}, result: &WalmResourceSet{Deployments: []WalmDeployment{{}}}},
		{resource: WalmDaemonSet{}, result: &WalmResourceSet{DaemonSets: []WalmDaemonSet{{}}}},
		{resource: WalmIngress{}, result: &WalmResourceSet{Ingresses: []WalmIngress{{}}}},
		{resource: WalmJob{}, result: &WalmResourceSet{Jobs: []WalmJob{{}}}},
	}
	for _, test := range tests {
		resourceSet := &WalmResourceSet{}
		test.resource.AddToWalmResourceSet(resourceSet)
		assert.Equal(t, test.result, resourceSet)
	}
}

func TestResourceSet_GetPodsNeedRestart(t *testing.T) {
	tests := []struct {
		resourceSet *WalmResourceSet
		result      []*WalmPod
	}{
		{
			resourceSet: &WalmResourceSet{
				StatefulSets: []WalmStatefulSet{{Pods: []*WalmPod{{WalmMeta: buildWalmMetaWithoutState("", "", "ss_pod")}}}},
				DaemonSets:   []WalmDaemonSet{{Pods: []*WalmPod{{WalmMeta: buildWalmMetaWithoutState("", "", "ds_pod")}}}},
				Deployments:  []WalmDeployment{{Pods: []*WalmPod{{WalmMeta: buildWalmMetaWithoutState("", "", "dp_pod")}}}},
			},
			result: []*WalmPod{
				{WalmMeta: buildWalmMetaWithoutState("", "", "ss_pod")},
				{WalmMeta: buildWalmMetaWithoutState("", "", "ds_pod")},
				{WalmMeta: buildWalmMetaWithoutState("", "", "dp_pod")},
			},
		},
	}
	for _, test := range tests {
		pods := test.resourceSet.GetPodsNeedRestart()
		assert.ElementsMatch(t, test.result, pods)
	}
}

func TestResourceSet_IsReady(t *testing.T) {
	pendingWalmMeta := buildWalmMeta("", "", "", buildWalmState("Pending", "", ""))
	readyWalmMeta := buildWalmMeta("", "", "", buildWalmState("Ready", "", ""))
	tests := []struct {
		resourceSet      *WalmResourceSet
		result           bool
		notReadyResource WalmResource
	}{
		{
			resourceSet:      &WalmResourceSet{StatefulSets: []WalmStatefulSet{{WalmMeta: pendingWalmMeta}},},
			result:           false,
			notReadyResource: WalmStatefulSet{WalmMeta: pendingWalmMeta},
		},
		{
			resourceSet:      &WalmResourceSet{Deployments: []WalmDeployment{{WalmMeta: pendingWalmMeta}},},
			result:           false,
			notReadyResource: WalmDeployment{WalmMeta: pendingWalmMeta},
		},
		{
			resourceSet:      &WalmResourceSet{DaemonSets: []WalmDaemonSet{{WalmMeta: pendingWalmMeta}},},
			result:           false,
			notReadyResource: WalmDaemonSet{WalmMeta: pendingWalmMeta},
		},
		{
			resourceSet:      &WalmResourceSet{Jobs: []WalmJob{{WalmMeta: pendingWalmMeta}},},
			result:           false,
			notReadyResource: WalmJob{WalmMeta: pendingWalmMeta},
		},
		{
			resourceSet:      &WalmResourceSet{Ingresses: []WalmIngress{{WalmMeta: pendingWalmMeta}},},
			result:           false,
			notReadyResource: WalmIngress{WalmMeta: pendingWalmMeta},
		},
		{
			resourceSet:      &WalmResourceSet{Services: []WalmService{{WalmMeta: pendingWalmMeta}},},
			result:           false,
			notReadyResource: WalmService{WalmMeta: pendingWalmMeta},
		},
		{
			resourceSet:      &WalmResourceSet{Secrets: []WalmSecret{{WalmMeta: pendingWalmMeta}},},
			result:           false,
			notReadyResource: WalmSecret{WalmMeta: pendingWalmMeta},
		},
		{
			resourceSet:      &WalmResourceSet{ConfigMaps: []WalmConfigMap{{WalmMeta: pendingWalmMeta}},},
			result:           false,
			notReadyResource: WalmConfigMap{WalmMeta: pendingWalmMeta},
		},
		{
			resourceSet:      &WalmResourceSet{},
			result:           true,
			notReadyResource: nil,
		},
		{
			resourceSet: &WalmResourceSet{
				ConfigMaps: []WalmConfigMap{{WalmMeta: readyWalmMeta}},
				Secrets: []WalmSecret{{WalmMeta: readyWalmMeta}},
				Services: []WalmService{{WalmMeta: readyWalmMeta}},
				Ingresses: []WalmIngress{{WalmMeta: readyWalmMeta}},
				Jobs: []WalmJob{{WalmMeta: readyWalmMeta}},
				DaemonSets: []WalmDaemonSet{{WalmMeta: readyWalmMeta}},
				Deployments: []WalmDeployment{{WalmMeta: readyWalmMeta}},
				StatefulSets: []WalmStatefulSet{{WalmMeta: readyWalmMeta}},
			},
			result:           true,
			notReadyResource: nil,
		},
	}

	for _, test := range tests {
		ready, notReadyResource := test.resourceSet.IsReady()
		assert.Equal(t, test.result, ready)
		assert.Equal(t, test.notReadyResource, notReadyResource)
	}
}
