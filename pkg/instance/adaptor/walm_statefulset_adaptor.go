package adaptor

import (
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	"walm/pkg/instance/lister"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
)

type WalmStatefulSetAdaptor struct{
	Lister lister.K8sResourceLister
}

func(adaptor WalmStatefulSetAdaptor) GetWalmModule(module v1beta1.ResourceReference) (WalmModule, error) {
	statefulSet, err := adaptor.GetWalmStatefulSet(module.ResourceRef.Namespace, module.ResourceRef.Name)
	if err != nil {
		return WalmModule{}, err
	}

	return WalmModule{Kind: module.ResourceRef.Kind, Object: statefulSet}, nil
}

func (adaptor WalmStatefulSetAdaptor) GetWalmStatefulSet(namespace string, name string) (WalmStatefulSet, error) {
	statefulSet, err := adaptor.Lister.GetStatefulSet(namespace, name)
	if err != nil {
		return WalmStatefulSet{}, err
	}

	return adaptor.BuildWalmStatefulSet(statefulSet)
}

func (adaptor WalmStatefulSetAdaptor) BuildWalmStatefulSet(statefulSet *appsv1beta1.StatefulSet) (walmStatefulSet WalmStatefulSet, err error){
	walmStatefulSet = WalmStatefulSet{
		WalmMeta: WalmMeta{Name: statefulSet.Name, Namespace: statefulSet.Namespace},
	}
	selector, err := metav1.LabelSelectorAsSelector(statefulSet.Spec.Selector)
	if err != nil {
		return walmStatefulSet, err
	}
	walmStatefulSet.Pods, err = WalmPodAdaptor{adaptor.Lister}.GetWalmPods(statefulSet.Namespace, selector.String())

	return walmStatefulSet, err
}
