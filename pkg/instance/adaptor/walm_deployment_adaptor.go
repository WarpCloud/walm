package adaptor

import (
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"walm/pkg/instance/lister"
	"fmt"
)

type WalmDeploymentAdaptor struct{
	Lister lister.K8sResourceLister
}

func(adaptor WalmDeploymentAdaptor) GetWalmModule(module v1beta1.ResourceReference) (WalmModule, error) {
	walmDeployment, err := adaptor.GetWalmDeployment(module.ResourceRef.Namespace, module.ResourceRef.Name)
	if err != nil {
		if isNotFoundErr(err) {
			return buildNotFoundWalmModule(module), nil
		}
		return WalmModule{}, err
	}

	return WalmModule{Kind: module.ResourceRef.Kind, Resource: walmDeployment, ModuleState: walmDeployment.DeploymentState}, nil
}

func (adaptor WalmDeploymentAdaptor) GetWalmDeployment(namespace string, name string) (WalmDeployment, error) {
	deployment, err := adaptor.Lister.GetDeployment(namespace, name)
	if err != nil {
		return WalmDeployment{}, err
	}

	return adaptor.BuildWalmDeployment(deployment)
}

func (adaptor WalmDeploymentAdaptor) BuildWalmDeployment(deployment *extv1beta1.Deployment) (walmDeployment WalmDeployment, err error){
	walmDeployment = WalmDeployment{
		WalmMeta: WalmMeta{Name: deployment.Name, Namespace: deployment.Namespace},
	}

	walmDeployment.Pods, err = WalmPodAdaptor{adaptor.Lister}.GetWalmPods(deployment.Namespace, deployment.Spec.Selector)
	walmDeployment.DeploymentState = BuildWalmDeploymentState(deployment, walmDeployment.Pods)
	return walmDeployment, err
}

func isDeploymentReady(deployment *extv1beta1.Deployment) bool {
	expectedReplicas, updatedReplicas, currentReplicas, availableReplicas := *deployment.Spec.Replicas, deployment.Status.UpdatedReplicas, deployment.Status.Replicas, deployment.Status.AvailableReplicas
	if expectedReplicas > 0 && updatedReplicas >= expectedReplicas && currentReplicas == updatedReplicas && availableReplicas == updatedReplicas {
		return true
	}
	return false
}

func BuildWalmDeploymentState(deployment *extv1beta1.Deployment, pods []*WalmPod) (walmState WalmState) {
	if isDeploymentReady(deployment) {
		walmState = BuildWalmState("Ready", "", "")
	} else {
		if len(pods) == 0 {
			walmState = BuildWalmState("Pending","PodNotCreated", "There is no pod created")
		} else {
			allPodsTerminating, unknownPod, pendingPod, runningPod := parsePods(pods)

			if allPodsTerminating {
				walmState = BuildWalmState("Terminating", "", "")
			} else {
				if unknownPod != nil {
					walmState = BuildWalmState("Pending", "PodUnknown", fmt.Sprintf("Pod %s/%s is in state Unknown", unknownPod.Namespace, unknownPod.Name))
				} else if pendingPod != nil {
					walmState = BuildWalmState("Pending", "PodPending", fmt.Sprintf("Pod %s/%s is in state Pending", pendingPod.Namespace, pendingPod.Name))
				} else if runningPod != nil {
					walmState = BuildWalmState("Pending", "PodRunning", fmt.Sprintf("Pod %s/%s is in state Running", runningPod.Namespace, runningPod.Name))
				} else {
					walmState = BuildWalmState("Pending", "DeploymentUpdating", "Deployment is updating")
				}
			}
		}
	}
	return walmState
}


