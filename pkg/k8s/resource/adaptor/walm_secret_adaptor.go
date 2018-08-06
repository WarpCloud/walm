package adaptor

import (
	corev1 "k8s.io/api/core/v1"
	"walm/pkg/k8s/handler"
)

type WalmSecretAdaptor struct {
	handler *handler.SecretHandler
}

func (adaptor WalmSecretAdaptor) GetResource(namespace string, name string) (WalmResource, error) {
	secret, err := adaptor.handler.GetSecret(namespace, name)
	if err != nil {
		if isNotFoundErr(err) {
			return WalmSecret{
				WalmMeta: buildNotFoundWalmMeta("Secret", namespace, name),
			}, nil
		}
		return WalmSecret{}, err
	}

	return adaptor.BuildWalmSecret(secret)
}

func (adaptor WalmSecretAdaptor) BuildWalmSecret(secret *corev1.Secret) (walmSecret WalmSecret, err error) {
	walmSecret = WalmSecret{
		WalmMeta: buildWalmMeta("Secret", secret.Namespace, secret.Name, buildWalmState("Ready", "", "")),
		Data:     secret.Data,
		Type:     secret.Type,
	}

	return
}
