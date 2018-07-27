package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	listv1 "k8s.io/client-go/listers/batch/v1"
)

type JobHandler struct {
	client *kubernetes.Clientset
	lister listv1.JobLister
}

func (handler JobHandler) GetJob(namespace string, name string) (*v1.Job, error) {
	return handler.lister.Jobs(namespace).Get(name)
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

func NewJobHandler(client *kubernetes.Clientset, lister listv1.JobLister) (JobHandler) {
	return JobHandler{client: client, lister: lister}
}