package adaptor

import (
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	"walm/pkg/instance/lister"
	corev1 "k8s.io/api/core/v1"
)

type WalmSecretAdaptor struct{
	Lister lister.K8sResourceLister
}

func(adaptor WalmSecretAdaptor) GetWalmModule(module v1beta1.ResourceReference) (WalmModule, error) {
	secret, err := adaptor.GetWalmSecret(module.ResourceRef.Namespace, module.ResourceRef.Name)
	if err != nil {
		if isNotFoundErr(err) {
			return buildNotFoundWalmModule(module), nil
		}
		return WalmModule{}, err
	}

	return WalmModule{Kind: module.ResourceRef.Kind, Resource: secret, ModuleState: WalmState{State: "Ready"}}, nil
}

func (adaptor WalmSecretAdaptor) GetWalmSecret(namespace string, name string) (WalmSecret, error) {
	secret, err := adaptor.Lister.GetSecret(namespace, name)
	if err != nil {
		return WalmSecret{}, err
	}

	return adaptor.BuildWalmSecret(secret)
}

func (adaptor WalmSecretAdaptor) BuildWalmSecret(secret *corev1.Secret) (walmSecret WalmSecret, err error){
	walmSecret = WalmSecret{
		WalmMeta: WalmMeta{Name: secret.Name, Namespace: secret.Namespace},
		Data: secret.Data,
		Type: secret.Type,
	}

	return
}
