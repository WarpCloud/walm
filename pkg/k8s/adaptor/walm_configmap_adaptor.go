package adaptor

import (
	corev1 "k8s.io/api/core/v1"
	"WarpCloud/walm/pkg/k8s/handler"
)

type WalmConfigMapAdaptor struct {
	handler *handler.ConfigMapHandler
}

func (adaptor *WalmConfigMapAdaptor) GetResource(namespace string, name string) (WalmResource, error) {
	configMap, err := adaptor.handler.GetConfigMap(namespace, name)
	if err != nil {
		if IsNotFoundErr(err) {
			return WalmConfigMap{
				WalmMeta: buildNotFoundWalmMeta("ConfigMap", namespace, name),
			}, nil
		}
		return WalmConfigMap{}, err
	}

	return adaptor.BuildWalmConfigMap(configMap)
}

func (adaptor *WalmConfigMapAdaptor) BuildWalmConfigMap(configMap *corev1.ConfigMap) (walmConfigMap WalmConfigMap, err error) {
	walmConfigMap = WalmConfigMap{
		WalmMeta: buildWalmMeta("ConfigMap", configMap.Namespace, configMap.Name, buildWalmState("Ready", "", "") ),
		Data:     configMap.Data,
	}

	return
}
