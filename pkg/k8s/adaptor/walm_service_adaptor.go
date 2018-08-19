package adaptor

import (
	corev1 "k8s.io/api/core/v1"
	"walm/pkg/k8s/handler"
)

type WalmServiceAdaptor struct {
	handler *handler.ServiceHandler
}

func (adaptor *WalmServiceAdaptor) GetResource(namespace string, name string) (WalmResource, error) {
	service, err := adaptor.handler.GetService(namespace, name)
	if err != nil {
		if isNotFoundErr(err) {
			return WalmService{
				WalmMeta: buildNotFoundWalmMeta("Service", namespace, name),
			}, nil
		}
		return WalmService{}, err
	}

	return adaptor.BuildWalmService(service)
}

func (adaptor *WalmServiceAdaptor) BuildWalmService(service *corev1.Service) (walmService WalmService, err error) {
	walmService = WalmService{
		WalmMeta:    buildWalmMeta("Service", service.Namespace, service.Name, buildWalmState("Ready", "", "")),
		Ports:       buildWalmServicePorts(service),
		ClusterIp:   service.Spec.ClusterIP,
		ServiceType: service.Spec.Type,
	}

	return
}

func buildWalmServicePorts(service *corev1.Service) []WalmServicePort {
	ports := []WalmServicePort{}
	for _, port := range service.Spec.Ports {
		ports = append(ports, WalmServicePort{
			Name:       port.Name,
			Port:       port.Port,
			NodePort:   port.NodePort,
			Protocol:   port.Protocol,
			TargetPort: port.TargetPort,
		})
	}
	return ports
}
