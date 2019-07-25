package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"github.com/stretchr/testify/assert"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"testing"
	"time"
)

func TestConvertStatefulSetFromK8s(t *testing.T) {
	testReplicas := int32(3)
	testCreationTimestamp := metav1.Now()

	tests := []struct {
		oriStatefulSet  *appsv1beta1.StatefulSet
		pods            []*corev1.Pod
		walmStatefulSet *k8s.StatefulSet
		err             error
	}{
		{
			oriStatefulSet: &appsv1beta1.StatefulSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-statefulset",
					Namespace: "test-namespace",
					Labels: map[string]string{"test1": "test1"},
					Annotations: map[string]string{"test2": "test2"},
				},
				Spec: appsv1beta1.StatefulSetSpec{
					Replicas: &testReplicas,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"test1": "test1", "test2": "test2"},
					},
				},
				Status: appsv1beta1.StatefulSetStatus{
					Replicas:      3,
					ReadyReplicas: 3,
				},
			},

			pods: []*corev1.Pod{
				{
					TypeMeta: metav1.TypeMeta{
						Kind: "Pod",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-namespace",
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
								Ready: true,
								State: corev1.ContainerState{
									Running: &corev1.ContainerStateRunning{
										StartedAt: testCreationTimestamp,
									},
								},
							},
						},
					},

				},
			},

			walmStatefulSet: &k8s.StatefulSet{
				Meta: k8s.Meta{
					Name:      "test-statefulset",
					Namespace: "test-namespace",
					Kind:      "StatefulSet",
					State: k8s.State{
						Status:  "Ready",
						Reason:  "",
						Message: "",
					},
				},
				Labels: map[string]string{"test1": "test1"},
				Annotations: map[string]string{"test2": "test2"},
				ExpectedReplicas: 3,
				ReadyReplicas:    3,
				Pods: []*k8s.Pod{
					{
						Meta: k8s.Meta{
							Name:      "test-pod",
							Namespace: "test-namespace",
							Kind:      "Pod",
							State: k8s.State{
								Status:  "Ready",
								Reason:  "",
								Message: "",
							},
						},
						Labels: map[string]string{},
						Annotations: map[string]string{},
						Containers: []k8s.Container{
							{
								Ready: true,
								State: k8s.State{
									Status: "Running",
									Reason: "",
									Message: "",
								},
							},
						},
						Age: duration.ShortHumanDuration(time.Since(testCreationTimestamp.Time)),

					},
				},
				Selector: "test1=test1,test2=test2",
			},
			err: nil,
		},
		{
			oriStatefulSet:  nil,
			pods:            nil,
			walmStatefulSet: nil,
			err:             nil,
		},
	}

	for _, test := range tests {
		walmStatefulSet, err := ConvertStatefulSetFromK8s(test.oriStatefulSet, test.pods)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.walmStatefulSet, walmStatefulSet)
	}
}
