package adaptor

import (
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	"walm/pkg/instance/lister"
	batchv1 "k8s.io/api/batch/v1"
	"fmt"
)

type WalmJobAdaptor struct{
	Lister lister.K8sResourceLister
}

func(adaptor WalmJobAdaptor) GetWalmModule(module v1beta1.ResourceReference) (WalmModule, error) {
	walmJob, err := adaptor.GetWalmJob(module.ResourceRef.Namespace, module.ResourceRef.Name)
	if err != nil {
		if isNotFoundErr(err) {
			return buildNotFoundWalmModule(module), nil
		}
		return WalmModule{}, err
	}

	return WalmModule{Kind: module.ResourceRef.Kind, Resource: walmJob, ModuleState: walmJob.JobState}, nil
}

func (adaptor WalmJobAdaptor) GetWalmJob(namespace string, name string) (WalmJob, error) {
	job, err := adaptor.Lister.GetJob(namespace, name)
	if err != nil {
		return WalmJob{}, err
	}

	return adaptor.BuildWalmJob(job)
}

func (adaptor WalmJobAdaptor) BuildWalmJob(job *batchv1.Job) (walmJob WalmJob, err error){
	walmJob = WalmJob{
		WalmMeta: WalmMeta{Name: job.Name, Namespace: job.Namespace},
		JobState: BuildWalmJobState(job),
	}

	walmJob.Pods, err = WalmPodAdaptor{adaptor.Lister}.GetWalmPods(job.Namespace, job.Spec.Selector)
	walmJob.JobState = BuildWalmJobState(job)
	return walmJob, err
}

func BuildWalmJobState(job *batchv1.Job) (jobState WalmState) {
	if len(job.Status.Conditions) > 0 {
		for _, condition := range job.Status.Conditions {
			if condition.Type == "Complete" && condition.Status == "True"{
				jobState = BuildWalmState("Ready", "", "")
				break
			}
			if condition.Type == "Failed" && condition.Status == "True" {
				jobState = BuildWalmState("Pending", condition.Reason, condition.Message)
				break
			}
		}
	} else if job.Status.Active > 0{
		jobState = BuildWalmState("Pending", "JobActive", fmt.Sprintf("There are %d active pod", job.Status.Active))
	} else {
		jobState = BuildWalmState("Terminating", "", "")
	}

	return jobState
}
