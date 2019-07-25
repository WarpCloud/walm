package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestConvertJobFromK8s(t *testing.T) {
	tests := []struct{
		oriJob  *batchv1.Job
		pods    []*corev1.Pod
		walmJob *k8s.Job
		err     error
	}{
		{
			oriJob: &batchv1.Job{
				TypeMeta:   metav1.TypeMeta{
					Kind: "Job",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-job",
					Namespace: "test-namespace",
					Labels: map[string]string{"test1": "test1"},
					Annotations: map[string]string{"test2": "test2"},
				},
				Status:     batchv1.JobStatus{
					Conditions: []batchv1.JobCondition{
						{
							Type:  "Complete",
							Status: "True",
						},
					},
					Succeeded: 1,
				},
			},
			pods: []*corev1.Pod{},
			walmJob: &k8s.Job{
				Meta:               k8s.Meta{
					Name: "test-job",
					Namespace: "test-namespace",
					Kind: "Job",
					State: k8s.State{
						Status:  "Ready",
						Reason:  "",
						Message: "",
					},
				},
				Labels: map[string]string{"test1": "test1"},
				Annotations: map[string]string{"test2": "test2"},
				ExpectedCompletion: 1,
				Succeeded:          1,
			},
			err: nil,
		},
		{
			oriJob: nil,
			pods: nil,
			walmJob: nil,
			err:     nil,
		},
	}

	for _,  test := range tests {
		walmJob, err := ConvertJobFromK8s(test.oriJob, test.pods)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.walmJob, walmJob)
	}
}