package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type JobHandler struct {
	client *kubernetes.Clientset
}

func (handler JobHandler) GetJob(namespace string, name string) (*v1.Job, error) {
	return handler.client.BatchV1().Jobs(namespace).Get(name, metav1.GetOptions{})
}

func (handler JobHandler) CreateJob(namespace string, job *v1.Job) (*v1.Job, error) {
	return handler.client.BatchV1().Jobs(namespace).Create(job)
}

func (handler JobHandler) UpdateJob(namespace string, job *v1.Job) (*v1.Job, error) {
	return handler.client.BatchV1().Jobs(namespace).Update(job)
}

func (handler JobHandler) DeleteJob(namespace string, name string) (error) {
	return handler.client.BatchV1().Jobs(namespace).Delete(name, &metav1.DeleteOptions{})
}

func NewJobHandler(client *kubernetes.Clientset) (JobHandler) {
	return JobHandler{client: client}
}