package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SecretHandler struct {
	client *kubernetes.Clientset
}

func (handler SecretHandler) GetSecret(namespace string, name string) (*v1.Secret, error) {
	return handler.client.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
}

func (handler SecretHandler) CreateSecret(namespace string, secret *v1.Secret) (*v1.Secret, error) {
	return handler.client.CoreV1().Secrets(namespace).Create(secret)
}

func (handler SecretHandler) UpdateSecret(namespace string, secret *v1.Secret) (*v1.Secret, error) {
	return handler.client.CoreV1().Secrets(namespace).Update(secret)
}

func (handler SecretHandler) DeleteSecret(namespace string, name string) (error) {
	return handler.client.CoreV1().Secrets(namespace).Delete(name, &metav1.DeleteOptions{})
}

func NewSecretHandler(client *kubernetes.Clientset) (SecretHandler) {
	return SecretHandler{client: client}
}
