package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"testing"
	"time"
)

func TestConvertPodFromK8s(t *testing.T) {
	testCreationTimestamp := metav1.Now()
	tests := []struct {
		oriPod *corev1.Pod
		pod    *k8s.Pod
		err    error
	}{
		{
			oriPod: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind: "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-namespace",
					Labels: map[string]string{"test1": "test1"},
					Annotations: map[string]string{"test2": "test2"},
					CreationTimestamp: testCreationTimestamp,
				},
				Status: corev1.PodStatus{
					Phase: "Running",
					Conditions: []corev1.PodCondition{
						{
							Type:   "Ready",
							Status: "True",
						},
					},
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name: "walm",
							Ready: true,
							Image: "docker.io/warpcloud/walm:dev",
							RestartCount: 2,
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{
									StartedAt: testCreationTimestamp,
								},
							},
						},
					},
				},
			},
			pod: &k8s.Pod{
				Meta:        k8s.Meta{
					Name: "test-pod",
					Namespace: "test-namespace",
					Kind: "Pod",
					State: k8s.State{
						Status:  "Ready",
					},
				},
				Labels: map[string]string{"test1": "test1"},
				Annotations: map[string]string{"test2": "test2"},
				Containers:  []k8s.Container{
					{
						Name: "walm",
						Image: "docker.io/warpcloud/walm:dev",
						Ready: true,
						RestartCount: 2,
						State: k8s.State{
							Status:  "Running",
							Reason:  "",
							Message: "",
						},
					},
				},
				Age: duration.ShortHumanDuration(time.Since(testCreationTimestamp.Time)),
			},
			err: nil,
		},
		{
			oriPod: nil,
			pod:    nil,
			err:    nil,
		},
	}

	for _, test := range tests {
		pod, err := ConvertPodFromK8s(test.oriPod)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.pod, pod)
	}
}
