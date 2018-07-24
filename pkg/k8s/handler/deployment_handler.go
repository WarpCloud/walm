package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeploymentHandler struct {
	client *kubernetes.Clientset
}

func (handler DeploymentHandler) GetDeployment(namespace string, name string) (*v1beta1.Deployment, error) {
	return handler.client.ExtensionsV1beta1().Deployments(namespace).Get(name, v1.GetOptions{})
}

func (handler DeploymentHandler) CreateDeployment(namespace string, deployment *v1beta1.Deployment) (*v1beta1.Deployment, error) {
	return handler.client.ExtensionsV1beta1().Deployments(namespace).Create(deployment)
}

func (handler DeploymentHandler) UpdateDeployment(namespace string, deployment *v1beta1.Deployment) (*v1beta1.Deployment, error) {
	return handler.client.ExtensionsV1beta1().Deployments(namespace).Update(deployment)
}

func (handler DeploymentHandler) DeleteDeployment(namespace string, name string) (error) {
	return handler.client.ExtensionsV1beta1().Deployments(namespace).Delete(name, &v1.DeleteOptions{})
}

func (handler DeploymentHandler) ScaleDeployment(namespace string, name string, replicas int32) (*v1beta1.Scale, error) {
	return handler.client.ExtensionsV1beta1().Deployments(namespace).UpdateScale(name, &v1beta1.Scale{Spec: v1beta1.ScaleSpec{Replicas: replicas}})
}

func (handler DeploymentHandler) RollbackDeployment(namespace string, name string, revision int64) (error) {
	return handler.client.ExtensionsV1beta1().Deployments(namespace).Rollback(&v1beta1.DeploymentRollback{Name: name, RollbackTo:v1beta1.RollbackConfig{Revision:revision}})
}

func NewDeploymentHandler(client *kubernetes.Clientset) (DeploymentHandler) {
	return DeploymentHandler{client:client}
}
