package converter

import (
	corev1 "k8s.io/api/core/v1"
	"WarpCloud/walm/pkg/models/k8s"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
	"k8s.io/apimachinery/pkg/util/sets"
)

func ConvertServiceFromK8s(oriService *corev1.Service, endpoints *corev1.Endpoints) (walmService *k8s.Service,err error) {
	if oriService == nil {
		return
	}
	service := oriService.DeepCopy()

	walmService = &k8s.Service{
		Meta:    k8s.NewMeta(k8s.ServiceKind, service.Namespace, service.Name, k8s.NewState("Ready", "", "")),
		ClusterIp:   service.Spec.ClusterIP,
		ServiceType: string(service.Spec.Type),
	}

	walmService.Ports, err = buildWalmServicePorts(service, endpoints)
	if err != nil {
		logrus.Errorf("failed to build walm service ports: %s", err.Error())
		return
	}

	return
}

func buildWalmServicePorts(service *corev1.Service, endpoints *corev1.Endpoints) ([]k8s.ServicePort, error) {
	ports := []k8s.ServicePort{}
	for _, port := range service.Spec.Ports {
		walmServicePort := k8s.ServicePort{
			Name:       port.Name,
			Port:       port.Port,
			NodePort:   port.NodePort,
			Protocol:   string(port.Protocol),
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