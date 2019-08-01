package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"fmt"
)

func ConvertJobFromK8s(oriJob *batchv1.Job, pods []*v1.Pod) (walmJob *k8s.Job, err error) {
	if oriJob == nil {
		return
	}
	job := oriJob.DeepCopy()

	walmJob = &k8s.Job{
		Meta:        k8s.NewEmptyStateMeta(k8s.JobKind, job.Namespace, job.Name),
		Labels:      job.Labels,
		Annotations: job.Annotations,
		Succeeded:   job.Status.Succeeded,
		Failed:      job.Status.Failed,
		Active:      job.Status.Active,
	}

	if job.Spec.Completions == nil {
		walmJob.ExpectedCompletion = 1
	} else {
		walmJob.ExpectedCompletion = *job.Spec.Completions
	}

	for _, pod := range pods {
		walmPod, err := ConvertPodFromK8s(pod)
		if err != nil {
			return nil, err
		}
		walmJob.Pods = append(walmJob.Pods, walmPod)
	}
	walmJob.State = buildWalmJobState(job)
	return walmJob, nil
}

func buildWalmJobState(job *batchv1.Job) (jobState k8s.State) {
	if len(job.Status.Conditions) > 0 {
		for _, condition := range job.Status.Conditions {
			if condition.Type == "Complete" && condition.Status == "True" {
				jobState = k8s.NewState("Ready", "", "")
				break
			}
			if condition.Type == "Failed" && condition.Status == "True" {
				jobState = k8s.NewState("Pending", condition.Reason, condition.Message)
				break
			}
		}
	} else if job.Status.Active > 0 {
		jobState = k8s.NewState("Pending", "JobActive", fmt.Sprintf("There are %d active pod", job.Status.Active))
	} else {
		jobState = k8s.NewState("Terminating", "", "")
	}

	return jobState
}
