package adaptor

import (
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	"walm/pkg/instance/lister"
	corev1 "k8s.io/api/core/v1"
)

type WalmConfigMapAdaptor struct {
	Lister lister.K8sResourceLister
}

func (adaptor WalmConfigMapAdaptor) GetWalmModule(module v1beta1.ResourceReference) (WalmModule, error) {
	configMap, err := adaptor.GetWalmConfigMap(module.ResourceRef.Namespace, module.ResourceRef.Name)
	if err != nil {
		if isNotFoundErr(err) {
			return buildNotFoundWalmModule(module), nil
		}
		return WalmModule{}, err
	}

	return WalmModule{Kind: module.ResourceRef.Kind, Resource: configMap, ModuleState: WalmState{State: "Ready"}}, nil
}

func (adaptor WalmConfigMapAdaptor) GetWalmConfigMap(namespace string, name string) (WalmConfigMap, error) {
	configMap, err := adaptor.Lister.GetConfigMap(namespace, name)
	if err != nil {
		return WalmConfigMap{}, err
	}

	return adaptor.BuildWalmConfigMap(configMap)
}

func (adaptor WalmConfigMapAdaptor) BuildWalmConfigMap(configMap *corev1.ConfigMap) (walmConfigMap WalmConfigMap, err error) {
	walmConfigMap = WalmConfigMap{
		WalmMeta: WalmMeta{Name: configMap.Name, Namespace: configMap.Namespace},
		Data:     configMap.Data,
	}

	return
}
