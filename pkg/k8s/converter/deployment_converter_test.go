package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestConvertDeploymentFromK8s(t *testing.T) {
	tests := []struct{
		oriDeployment *extv1beta1.Deployment
		pods []*corev1.Pod
		walmDeployment *k8s.Deployment
		err error
	}{
		{
			oriDeployment: &extv1beta1.Deployment{
				TypeMeta:   metav1.TypeMeta{
					Kind: "Deployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-deployment",
					Namespace: "test-namespace",
					Labels: map[string]string{"test1": "test1"},
					Annotations: map[string]string{"test2": "test2"},
				},
				Status:     extv1beta1.DeploymentStatus{
					UpdatedReplicas: 1,
					Replicas: 1,
					AvailableReplicas: 1,
					ReadyReplicas: 1,
				},
			},
			walmDeployment: &k8s.Deployment{
				Meta:              k8s.Meta{
					Name: "test-deployment",
					Namespace: "test-namespace",
					Kind: "Deployment",
					State: k8s.State{
						Status:  "Ready",
						Reason:  "",
						Message: "",
					},
				},
				Labels: map[string]string{"test1": "test1"},
				Annotations: map[string]string{"test2": "test2"},
				ExpectedReplicas:  1,
				UpdatedReplicas:   1,
				CurrentReplicas:   1,
				AvailableReplicas: 1,
			},
			err: nil,
		},
		{
			oriDeployment: nil,
			pods: nil,
			walmDeployment: nil,
			err: nil,
		},
	}


	for _, test := range tests {
		walmDeployment, err := ConvertDeploymentFromK8s(test.oriDeployment, test.pods)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.walmDeployment, walmDeployment)
	}
}