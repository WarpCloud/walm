package adaptor

import (
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"fmt"
	"walm/pkg/k8s/handler"
)

type WalmDeploymentAdaptor struct {
	deploymentHandler *handler.DeploymentHandler
	podAdaptor        *WalmPodAdaptor
}

func (adaptor *WalmDeploymentAdaptor) GetResource(namespace string, name string) (WalmResource, error) {
	deployment, err := adaptor.deploymentHandler.GetDeployment(namespace, name)
	if err != nil {
		if isNotFoundErr(err) {
			return WalmDeployment{
				WalmMeta: buildNotFoundWalmMeta("Deployment", namespace, name),
			}, nil
		}
		return WalmDeployment{}, err
	}

	return adaptor.BuildWalmDeployment(deployment)
}

func (adaptor *WalmDeploymentAdaptor) BuildWalmDeployment(deployment *extv1beta1.Deployment) (walmDeployment WalmDeployment, err error) {
	walmDeployment = WalmDeployment{
		WalmMeta:          buildWalmMetaWithoutState("Deployment", deployment.Namespace, deployment.Name),
		UpdatedReplicas:   deployment.Status.UpdatedReplicas,
		CurrentReplicas:   deployment.Status.Replicas,
		AvailableReplicas: deployment.Status.AvailableReplicas,
	}

	if deployment.Spec.Replicas == nil {
		walmDeployment.ExpectedReplicas = 1
	} else {
		walmDeployment.ExpectedReplicas = *deployment.Spec.Replicas
	}

	walmDeployment.Pods, err = adaptor.podAdaptor.GetWalmPods(deployment.Namespace, deployment.Spec.Selector)
	walmDeployment.State = BuildWalmDeploymentState(deployment, walmDeployment.Pods)
	return walmDeployment, err
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

func BuildWalmDeploymentState(deployment *extv1beta1.Deployment, pods []*WalmPod) (walmState WalmState) {
	if isDeploymentReady(deployment) {
		walmState = buildWalmState("Ready", "", "")
	} else {
		if len(pods) == 0 {
			walmState = buildWalmState("Pending", "PodNotCreated", "There is no pod created")
		} else {
			allPodsTerminating, unknownPod, pendingPod, runningPod := parsePods(pods)

			if allPodsTerminating {
				walmState = buildWalmState("Terminating", "", "")
			} else {
				if unknownPod != nil {
					walmState = buildWalmState("Pending", "PodUnknown", fmt.Sprintf("Pod %s/%s is in state Unknown", unknownPod.Namespace, unknownPod.Name))
				} else if pendingPod != nil {
					walmState = buildWalmState("Pending", "PodPending", fmt.Sprintf("Pod %s/%s is in state Pending", pendingPod.Namespace, pendingPod.Name))
				} else if runningPod != nil {
					walmState = buildWalmState("Pending", "PodRunning", fmt.Sprintf("Pod %s/%s is in state Running", runningPod.Namespace, runningPod.Name))
				} else {
					walmState = buildWalmState("Pending", "DeploymentUpdating", "Deployment is updating")
				}
			}
		}
	}
	return walmState
}
