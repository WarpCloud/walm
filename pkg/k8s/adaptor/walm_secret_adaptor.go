package adaptor

import (
	corev1 "k8s.io/api/core/v1"
	"walm/pkg/k8s/handler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/sirupsen/logrus"
)

type WalmSecretAdaptor struct {
	handler *handler.SecretHandler
}

func (adaptor *WalmSecretAdaptor) GetResource(namespace string, name string) (WalmResource, error) {
	secret, err := adaptor.handler.GetSecret(namespace, name)
	if err != nil {
		if IsNotFoundErr(err) {
			return WalmSecret{
				WalmMeta: buildNotFoundWalmMeta("Secret", namespace, name),
			}, nil
		}
		return WalmSecret{}, err
	}

	return adaptor.BuildWalmSecret(secret), nil
}

func (adaptor *WalmSecretAdaptor) ListSecrets(namespace string, labelSelector *metav1.LabelSelector) (walmSecrets *WalmSecretList, err error) {
	secrets, err := adaptor.handler.ListSecrets(namespace, labelSelector)
	if err != nil {
		return nil, err
	}

	walmSecrets = &WalmSecretList{
		Items: []*WalmSecret{},
	}
	for _, secret := range secrets {
		walmSecret := adaptor.BuildWalmSecret(secret)
		walmSecrets.Items = append(walmSecrets.Items, &walmSecret)
	}
	walmSecrets.Num = len(walmSecrets.Items)
	return
}

func (adaptor *WalmSecretAdaptor) CreateSecret(walmSecret *WalmSecret) (err error) {
	_, err = adaptor.handler.CreateSecret(walmSecret.Namespace, BuildSecret(walmSecret))
	if err != nil {
		logrus.Errorf("failed to create secret : %s", err.Error())
	}
	return
}

func (adaptor *WalmSecretAdaptor) BuildWalmSecret(secret *corev1.Secret) (walmSecret WalmSecret) {
	walmSecret = WalmSecret{
		WalmMeta: buildWalmMeta("Secret", secret.Namespace, secret.Name, buildWalmState("Ready", "", "")),
		Data:     convertDataToStringData(secret.Data),
		Type:     secret.Type,
	}

	return
}

func BuildSecret(walmSecret *WalmSecret) (secret *corev1.Secret) {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: walmSecret.Namespace,
			Name: walmSecret.Name,
		},
		StringData: walmSecret.Data,
		Type: walmSecret.Type,
	}
}

func convertDataToStringData(data map[string][]byte) (stringData map[string]string) {
	stringData = map[string]string{}
	for key, value := range data {
		stringData[key] = string(value)
	}
	return
}
