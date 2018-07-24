package adaptor

import (
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"walm/pkg/instance/lister"
)

type WalmDeploymentAdaptor struct{
	Lister lister.K8sResourceLister
}

func(adaptor WalmDeploymentAdaptor) GetWalmModule(module v1beta1.ResourceReference) (WalmModule, error) {
	walmDeployment, err := adaptor.GetWalmDeployment(module.ResourceRef.Namespace, module.ResourceRef.Name)
	if err != nil {
		return WalmModule{}, err
	}

	return WalmModule{Kind: module.ResourceRef.Kind, Object: walmDeployment}, nil
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

	return walmDeployment, err
}
