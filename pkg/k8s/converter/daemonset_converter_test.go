package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestConvertDaemonSetFromK8s(t *testing.T) {
	tests := []struct {
		oriDaemonSet  *extv1beta1.DaemonSet
		pods          []*corev1.Pod
		walmDaemonSet *k8s.DaemonSet
		err           error
	}{
		{
			oriDaemonSet: &extv1beta1.DaemonSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "DaemonSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-daemonSet",
					Namespace:   "test-namespace",
					Labels:      map[string]string{"test1": "test1"},
					Annotations: map[string]string{"test2": "test2"},
				},
				Status: extv1beta1.DaemonSetStatus{
					DesiredNumberScheduled: 1,
					UpdatedNumberScheduled: 2,
					NumberAvailable:        2,
				},
			},
			pods: []*corev1.Pod{
			},
			walmDaemonSet: &k8s.DaemonSet{
				Meta: k8s.Meta{
					Name: "test-daemonSet",
					Namespace: "test-namespace",
					Kind: "DaemonSet",
					State: k8s.State{
						Status:  "Ready",
						Reason:  "",
						Message: "",
					},
				},
				Labels: map[string]string{"test1": "test1"},
				Annotations: map[string]string{"test2": "test2"},
				DesiredNumberScheduled: 1,
				UpdatedNumberScheduled: 2,
				NumberAvailable:        2,
			},
			err: nil,
		},
		{
			oriDaemonSet: nil,
			pods: nil,
			walmDaemonSet: nil,
			err: nil,
		},
	}

	for _, test := range tests {
		walmDaemonSet, err := ConvertDaemonSetFromK8s(test.oriDaemonSet, test.pods)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.walmDaemonSet, walmDaemonSet)
	}
}
