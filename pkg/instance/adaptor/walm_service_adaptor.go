package adaptor

import (
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	"walm/pkg/instance/lister"
	corev1 "k8s.io/api/core/v1"
)

type WalmServiceAdaptor struct{
	Lister lister.K8sResourceLister
}

func(adaptor WalmServiceAdaptor) GetWalmModule(module v1beta1.ResourceReference) (WalmModule, error) {
	service, err := adaptor.GetWalmService(module.ResourceRef.Namespace, module.ResourceRef.Name)
	if err != nil {
		return WalmModule{}, err
	}

	return WalmModule{Kind: module.ResourceRef.Kind, Object: service}, nil
}

func (adaptor WalmServiceAdaptor) GetWalmService(namespace string, name string) (WalmService, error) {
	service, err := adaptor.Lister.GetService(namespace, name)
	if err != nil {
		return WalmService{}, err
	}

	return adaptor.BuildWalmService(service)
}

func (adaptor WalmServiceAdaptor) BuildWalmService(service *corev1.Service) (walmService WalmService, err error){
	walmService = WalmService{
		WalmMeta: WalmMeta{Name: service.Name, Namespace: service.Namespace},
		ServiceType: service.Spec.Type,
	}

	return
}
