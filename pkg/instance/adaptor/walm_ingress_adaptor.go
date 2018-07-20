package adaptor

import (
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	"walm/pkg/instance/walmlister"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
)

type WalmIngressAdaptor struct{
	Lister walmlister.K8sResourceLister
}

func(adaptor WalmIngressAdaptor) GetWalmModule(module v1beta1.ResourceReference) (WalmModule, error) {
	ingress, err := adaptor.GetWalmIngress(module.ResourceRef.Namespace, module.ResourceRef.Name)
	if err != nil {
		return WalmModule{}, err
	}

	return WalmModule{Kind: module.ResourceRef.Kind, Object: ingress}, nil
}

func (adaptor WalmIngressAdaptor) GetWalmIngress(namespace string, name string) (WalmIngress, error) {
	ingress, err := adaptor.Lister.GetIngress(namespace, name)
	if err != nil {
		return WalmIngress{}, err
	}

	return adaptor.BuildWalmIngress(ingress)
}

func (adaptor WalmIngressAdaptor) BuildWalmIngress(ingress *extv1beta1.Ingress) (walmIngress WalmIngress, err error){
	walmIngress = WalmIngress{
		WalmMeta: WalmMeta{Name: ingress.Name, Namespace: ingress.Namespace},
	}

	if len(ingress.Spec.Rules) > 0 {
		rule := ingress.Spec.Rules[0]
		walmIngress.Host = rule.Host
		if rule.HTTP != nil && len(rule.HTTP.Paths) > 0{
			path := rule.HTTP.Paths[0]
			walmIngress.Path = path.Path
			walmIngress.ServiceName = path.Backend.ServiceName
			walmIngress.ServicePort = path.Backend.ServicePort.String()
		}
	}

	return
}
