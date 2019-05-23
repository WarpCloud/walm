package adaptor

import (
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"WarpCloud/walm/pkg/k8s/handler"
)

type WalmIngressAdaptor struct {
	handler *handler.IngressHandler
}

func (adaptor *WalmIngressAdaptor) GetResource(namespace string, name string) (WalmResource, error) {
	ingress, err := adaptor.handler.GetIngress(namespace, name)
	if err != nil {
		if IsNotFoundErr(err) {
			return WalmIngress{
				WalmMeta: buildNotFoundWalmMeta("Ingress", namespace, name),
			}, nil
		}
		return WalmIngress{}, err
	}

	return adaptor.BuildWalmIngress(ingress)
}

func (adaptor *WalmIngressAdaptor) BuildWalmIngress(ingress *extv1beta1.Ingress) (walmIngress WalmIngress, err error) {
	walmIngress = WalmIngress{
		WalmMeta: buildWalmMeta("Ingress", ingress.Namespace, ingress.Name, buildWalmState("Ready", "", "")),
	}

	if len(ingress.Spec.Rules) > 0 {
		rule := ingress.Spec.Rules[0]
		walmIngress.Host = rule.Host
		if rule.HTTP != nil && len(rule.HTTP.Paths) > 0 {
			path := rule.HTTP.Paths[0]
			walmIngress.Path = path.Path
			walmIngress.ServiceName = path.Backend.ServiceName
			walmIngress.ServicePort = path.Backend.ServicePort.String()
		}
	}

	return
}
