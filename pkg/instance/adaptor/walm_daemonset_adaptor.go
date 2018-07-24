package adaptor

import (
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"walm/pkg/instance/lister"
)

type WalmDaemonSetAdaptor struct{
	Lister lister.K8sResourceLister
}

func(adaptor WalmDaemonSetAdaptor) GetWalmModule(module v1beta1.ResourceReference) (WalmModule, error) {
	walmDaemonSet, err := adaptor.GetWalmDaemonSet(module.ResourceRef.Namespace, module.ResourceRef.Name)
	if err != nil {
		return WalmModule{}, err
	}

	return WalmModule{Kind: module.ResourceRef.Kind, Object: walmDaemonSet}, nil
}

func (adaptor WalmDaemonSetAdaptor) GetWalmDaemonSet(namespace string, name string) (WalmDaemonSet, error) {
	daemonSet, err := adaptor.Lister.GetDaemonSet(namespace, name)
	if err != nil {
		return WalmDaemonSet{}, err
	}

	return adaptor.BuildWalmDaemonSet(daemonSet)
}

func (adaptor WalmDaemonSetAdaptor) BuildWalmDaemonSet(daemonSet *extv1beta1.DaemonSet) (walmDaemonSet WalmDaemonSet, err error){
	walmDaemonSet = WalmDaemonSet{
		WalmMeta: WalmMeta{Name: daemonSet.Name, Namespace: daemonSet.Namespace},
	}

	walmDaemonSet.Pods, err = WalmPodAdaptor{adaptor.Lister}.GetWalmPods(daemonSet.Namespace, daemonSet.Spec.Selector)

	return walmDaemonSet, err
}
