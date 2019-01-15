package adaptor

import (
	corev1 "k8s.io/api/core/v1"
	"walm/pkg/k8s/handler"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
	"k8s.io/apimachinery/pkg/util/sets"
)

type WalmServiceAdaptor struct {
	handler          *handler.ServiceHandler
	endpointsHandler *handler.EndpointsHandler
}

func (adaptor *WalmServiceAdaptor) GetResource(namespace string, name string) (WalmResource, error) {
	service, err := adaptor.handler.GetService(namespace, name)
	if err != nil {
		if IsNotFoundErr(err) {
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
		ClusterIp:   service.Spec.ClusterIP,
		ServiceType: service.Spec.Type,
	}

	walmService.Ports, err = adaptor.buildWalmServicePorts(service)
	if err != nil {
		logrus.Errorf("failed to build walm service ports: %s", err.Error())
		return
	}

	return
}

func (adaptor *WalmServiceAdaptor) buildWalmServicePorts(service *corev1.Service) ([]WalmServicePort, error) {
	endpoints, err := adaptor.endpointsHandler.GetEndpoints(service.Namespace, service.Name)
	if err != nil {
		if IsNotFoundErr(err) {
			logrus.Warnf("endpoints %s/%s is not found", service.Namespace, service.Name)
		} else {
			logrus.Errorf("failed to get service endpoints %s/%s : %s", service.Namespace, service.Name, err.Error())
			return nil, err
		}
	}

	ports := []WalmServicePort{}
	for _, port := range service.Spec.Ports {
		walmServicePort := WalmServicePort{
			Name:       port.Name,
			Port:       port.Port,
			NodePort:   port.NodePort,
			Protocol:   port.Protocol,
			TargetPort: port.TargetPort.String(),
		}
		if endpoints != nil {
			walmServicePort.Endpoints = formatEndpoints(endpoints, sets.NewString(port.Name))
		} else {
			walmServicePort.Endpoints = []string{}
		}
		ports = append(ports, walmServicePort)
	}

	return ports, nil
}

func formatEndpoints(endpoints *corev1.Endpoints, ports sets.String) (list []string) {
	list = []string{}
	for i := range endpoints.Subsets {
		ss := &endpoints.Subsets[i]
		for i := range ss.Ports {
			port := &ss.Ports[i]
			if ports == nil || ports.Has(port.Name) {
				for i := range ss.Addresses {
					addr := &ss.Addresses[i]
					hostPort := net.JoinHostPort(addr.IP, strconv.Itoa(int(port.Port)))
					list = append(list, hostPort)
				}
			}
		}
	}

	return
}
