package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigMapHandler struct {
	client *kubernetes.Clientset
}

func (handler ConfigMapHandler) GetConfigMap(namespace string, name string) (*v1.ConfigMap, error) {
	return handler.client.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
}

func (handler ConfigMapHandler) CreateConfigMap(namespace string, configMap *v1.ConfigMap) (*v1.ConfigMap, error) {
	return handler.client.CoreV1().ConfigMaps(namespace).Create(configMap)
}

func (handler ConfigMapHandler) UpdateConfigMap(namespace string, configMap *v1.ConfigMap) (*v1.ConfigMap, error) {
	return handler.client.CoreV1().ConfigMaps(namespace).Update(configMap)
}

func (handler ConfigMapHandler) DeleteConfigMap(namespace string, name string) (error) {
	return handler.client.CoreV1().ConfigMaps(namespace).Delete(name, &metav1.DeleteOptions{})
}

func NewConfigMapHandler(client *kubernetes.Clientset) (ConfigMapHandler) {
	return ConfigMapHandler{client: client}
}