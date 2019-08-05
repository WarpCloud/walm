package converter

import (
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"WarpCloud/walm/pkg/models/k8s"
	"k8s.io/api/core/v1"
)

func ConvertDeploymentFromK8s(oriDeployment *extv1beta1.Deployment, pods []*v1.Pod) (walmDeployment *k8s.Deployment, err error) {
	if oriDeployment == nil {
		return
	}
	deployment := oriDeployment.DeepCopy()

	walmDeployment = &k8s.Deployment{
		Meta:              k8s.NewEmptyStateMeta(k8s.DeploymentKind, deployment.Namespace, deployment.Name),
		Labels:            deployment.Labels,
		Annotations:       deployment.Annotations,
		UpdatedReplicas:   deployment.Status.UpdatedReplicas,
		CurrentReplicas:   deployment.Status.Replicas,
		AvailableReplicas: deployment.Status.AvailableReplicas,
	}

	if deployment.Spec.Replicas == nil {
		walmDeployment.ExpectedReplicas = 1
	} else {
		walmDeployment.ExpectedReplicas = *deployment.Spec.Replicas
	}

	for _, pod := range pods {
		walmPod, err := ConvertPodFromK8s(pod)
		if err != nil {
			return nil, err
		}
		walmDeployment.Pods = append(walmDeployment.Pods, walmPod)
	}
	walmDeployment.State = buildWalmDeploymentState(deployment, walmDeployment.Pods)
	return walmDeployment, nil
}

func isDeploymentReady(deployment *extv1beta1.Deployment) bool {
	expectedReplicas, updatedReplicas, currentReplicas, availableReplicas := deployment.Spec.Replicas, deployment.Status.UpdatedReplicas, deployment.Status.Replicas, deployment.Status.AvailableReplicas
	if expectedReplicas != nil && updatedReplicas < *expectedReplicas {
		return false
	}
	if currentReplicas > updatedReplicas {
		return false
	}
	if availableReplicas < updatedReplicas {
		return false
	}
	return true
}

func buildWalmDeploymentState(deployment *extv1beta1.Deployment, pods []*k8s.Pod) (walmState k8s.State) {
	if isDeploymentReady(deployment) {
		walmState = k8s.NewState("Ready", "", "")
	} else {
		walmState = buildWalmStateByPods(pods, "Deployment")
	}
	return walmState
}

