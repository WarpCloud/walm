package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	listv1 "k8s.io/client-go/listers/core/v1"
	k8sutils "WarpCloud/walm/pkg/k8s/utils"
	"github.com/sirupsen/logrus"
)

type SecretHandler struct {
	client *kubernetes.Clientset
	lister listv1.SecretLister
}

func (handler *SecretHandler) GetSecret(namespace string, name string) (*v1.Secret, error) {
	return handler.lister.Secrets(namespace).Get(name)
}

func (handler *SecretHandler) ListSecrets(namespace string, labelSelector *metav1.LabelSelector) (secrets []*v1.Secret, err error) {
	selector, err := k8sutils.ConvertLabelSelectorToSelector(labelSelector)
	if err != nil {
		logrus.Errorf("failed to parse label selector : %v", *labelSelector)
		return nil, err
	}
	secrets, err = handler.lister.Secrets(namespace).List(selector)
	if err != nil {
		logrus.Errorf("failed to list secrets under namespace %s", namespace)
	}
	return
}

func (handler *SecretHandler) CreateSecret(namespace string, secret *v1.Secret) (*v1.Secret, error) {
	return handler.client.CoreV1().Secrets(namespace).Create(secret)
}

func (handler *SecretHandler) UpdateSecret(namespace string, secret *v1.Secret) (*v1.Secret, error) {
	return handler.client.CoreV1().Secrets(namespace).Update(secret)
}

func (handler *SecretHandler) DeleteSecret(namespace string, name string) (error) {
	return handler.client.CoreV1().Secrets(namespace).Delete(name, &metav1.DeleteOptions{})
}
