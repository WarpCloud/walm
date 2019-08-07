package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
)

func ConvertIngressFromK8s(oriIngress *extv1beta1.Ingress) (*k8s.Ingress, error) {
	if oriIngress == nil {
		return nil, nil
	}
	ingress := oriIngress.DeepCopy()

	walmIngress := k8s.Ingress{
		Meta: k8s.NewMeta(k8s.IngressKind, ingress.Namespace, ingress.Name, k8s.NewState("Ready", "", "")),
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
	return &walmIngress, nil
}