package adaptor

import (
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestIsDeploymentReady(t *testing.T) {
	expectedReplica := int32(3)
	tests := []struct {
		deployment *extv1beta1.Deployment
		result     bool
	} {
		{
			deployment: &extv1beta1.Deployment{
				Spec: extv1beta1.DeploymentSpec{
					Replicas: &expectedReplica,
				},
				Status: extv1beta1.DeploymentStatus{
					Replicas: 3,
					UpdatedReplicas: 3,
					AvailableReplicas: 3,
				},
			},
			result: true,
		},
		{
			deployment: &extv1beta1.Deployment{
				Spec: extv1beta1.DeploymentSpec{
					Replicas: &expectedReplica,
				},
				Status: extv1beta1.DeploymentStatus{
					Replicas: 3,
					UpdatedReplicas: 2,
					AvailableReplicas: 3,
				},
			},
			result: false,
		},
		{
			deployment: &extv1beta1.Deployment{
				Spec: extv1beta1.DeploymentSpec{
					Replicas: &expectedReplica,
				},
				Status: extv1beta1.DeploymentStatus{
					Replicas: 4,
					UpdatedReplicas: 3,
					AvailableReplicas: 3,
				},
			},
			result: false,
		},
		{
			deployment: &extv1beta1.Deployment{
				Spec: extv1beta1.DeploymentSpec{
					Replicas: &expectedReplica,
				},
				Status: extv1beta1.DeploymentStatus{
					Replicas: 3,
					UpdatedReplicas: 3,
					AvailableReplicas: 2,
				},
			},
			result: false,
		},
	}

	for _, test := range tests {
		result := isDeploymentReady(test.deployment)
		assert.Equal(t, test.result, result)
	}
}
