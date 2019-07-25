package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestConvertNodeFromK8s(t *testing.T) {
	tests := []struct {
		oriNode    *corev1.Node
		podsOnNode *corev1.PodList
		walmNode   *k8s.Node
		err        error
	}{
		{
			oriNode: &corev1.Node{
				TypeMeta: metav1.TypeMeta{
					Kind: "Node",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-node",
					Namespace:   "test-namespace",
					Labels:      map[string]string{"test1": "test1"},
					Annotations: map[string]string{"test2": "test2"},
				},
				Spec: corev1.NodeSpec{},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{
							Type:    corev1.NodeInternalIP,
							Address: "172.26.1.128",
						},
					},
					Conditions: []corev1.NodeCondition{
						{
							Type:   "Ready",
							Status: "True",
						},
					},
					Capacity: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceCPU:              resource.MustParse("4"),
						corev1.ResourceEphemeralStorage: resource.MustParse("78477888Ki"),
						corev1.ResourceMemory:           resource.MustParse("24950868Ki"),
					},
					Allocatable: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceCPU:              resource.MustParse("4"),
						corev1.ResourceEphemeralStorage: resource.MustParse("72325221462"),
						corev1.ResourceMemory:           resource.MustParse("24848468Ki"),
					},
				},
			},
			podsOnNode: &corev1.PodList{
				Items: []corev1.Pod{
					{
						TypeMeta: metav1.TypeMeta{
							Kind: "Pod",
						},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "test-pod",
									Resources: corev1.ResourceRequirements{
										Limits: map[corev1.ResourceName]resource.Quantity{
											corev1.ResourceCPU:    resource.MustParse("100m"),
											corev1.ResourceMemory: resource.MustParse("390Mi"),
										},
										Requests: map[corev1.ResourceName]resource.Quantity{
											corev1.ResourceCPU:    resource.MustParse("850m"),
											corev1.ResourceMemory: resource.MustParse("190Mi"),
										},
									},
								},
							},
						},
					},
				},
			},
			walmNode: &k8s.Node{
				Meta: k8s.Meta{
					Name: "test-node",
					Namespace: "test-namespace",
					Kind: "Node",
					State: k8s.State{
						Status:  "Ready",
						Reason:  "",
						Message: "",
					},
				},
				Labels:      map[string]string{"test1": "test1"},
				Annotations: map[string]string{"test2": "test2"},
				NodeIp:      "172.26.1.128",
				Capacity: map[string]string{
					"cpu":               "4",
					"ephemeral-storage": "78477888Ki",
					"memory":            "24950868Ki",
				},
				Allocatable: map[string]string{
					"cpu":               "4",
					"ephemeral-storage": "72325221462",
					"memory":            "24848468Ki",
				},
				RequestsAllocated: map[string]string{
					"cpu":    "850m",
					"memory": "190Mi",
				},
				LimitsAllocated: map[string]string{
					"cpu":    "100m",
					"memory": "390Mi",
				},
				WarpDriveStorageList: []k8s.WarpDriveStorage{},
				UnifyUnitResourceInfo: k8s.UnifyUnitNodeResourceInfo{
					Capacity: k8s.NodeResourceInfo{
						Cpu:    4,
						Memory: 24366,
					},
					Allocatable: k8s.NodeResourceInfo{
						Cpu:    4,
						Memory: 24266,
					},
					RequestsAllocated: k8s.NodeResourceInfo{
						Cpu:    0.85,
						Memory: 190,
					},
					LimitsAllocated: k8s.NodeResourceInfo{
						Cpu:    0.1,
						Memory: 390,
					},
				},
			},
			err: nil,
		},
		{
			oriNode:    nil,
			podsOnNode: nil,
			walmNode:   nil,
			err:        nil,
		},
	}

	for _, test := range tests {
		walmNode, err := ConvertNodeFromK8s(test.oriNode, test.podsOnNode)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.walmNode, walmNode)
	}
}
