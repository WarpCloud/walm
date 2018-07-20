package adaptor

import (
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	"walm/pkg/instance/walmlister"
	corev1 "k8s.io/api/core/v1"
)

type WalmConfigMapAdaptor struct{
	Lister walmlister.K8sResourceLister
}

func(adaptor WalmConfigMapAdaptor) GetWalmModule(module v1beta1.ResourceReference) (WalmModule, error) {
	configMap, err := adaptor.GetWalmConfigMap(module.ResourceRef.Namespace, module.ResourceRef.Name)
	if err != nil {
		return WalmModule{}, err
	}

	return WalmModule{Kind: module.ResourceRef.Kind, Object: configMap}, nil
}

func (adaptor WalmConfigMapAdaptor) GetWalmConfigMap(namespace string, name string) (WalmConfigMap, error) {
	configMap, err := adaptor.Lister.GetConfigMap(namespace, name)
	if err != nil {
		return WalmConfigMap{}, err
	}

	return adaptor.BuildWalmConfigMap(configMap)
}

func (adaptor WalmConfigMapAdaptor) BuildWalmConfigMap(configMap *corev1.ConfigMap) (walmConfigMap WalmConfigMap, err error){
	walmConfigMap = WalmConfigMap{
		WalmMeta: WalmMeta{Name: configMap.Name, Namespace: configMap.Namespace},
		Data: configMap.Data,
	}

	return
}
