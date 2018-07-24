package adaptor

import (
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	"walm/pkg/instance/lister"
	batchv1 "k8s.io/api/batch/v1"
)

type WalmJobAdaptor struct{
	Lister lister.K8sResourceLister
}

func(adaptor WalmJobAdaptor) GetWalmModule(module v1beta1.ResourceReference) (WalmModule, error) {
	walmJob, err := adaptor.GetWalmJob(module.ResourceRef.Namespace, module.ResourceRef.Name)
	if err != nil {
		return WalmModule{}, err
	}

	return WalmModule{Kind: module.ResourceRef.Kind, Object: walmJob}, nil
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
	}

	walmJob.Pods, err = WalmPodAdaptor{adaptor.Lister}.GetWalmPods(job.Namespace, job.Spec.Selector)

	return walmJob, err
}
