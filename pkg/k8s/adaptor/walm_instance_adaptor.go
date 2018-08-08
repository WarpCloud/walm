package adaptor

import (
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	"fmt"
	"k8s.io/api/core/v1"
)

type WalmInstanceAdaptor struct {
	adaptorSet      *AdaptorSet
}

func (adaptor *WalmInstanceAdaptor) GetResource(namespace string, name string) (WalmResource, error) {
	instance, err := adaptor.adaptorSet.GetHandlerSet().GetInstanceHandler().GetInstance(namespace, name)
	if err != nil {
		if isNotFoundErr(err) {
			return WalmApplicationInstance{
				WalmMeta: buildNotFoundWalmMeta("ApplicationInstance", namespace, name),
			}, nil
		}
		return WalmApplicationInstance{}, err
	}

	return adaptor.BuildWalmInstance(instance)
}

func (adaptor *WalmInstanceAdaptor) BuildWalmInstance(instance *v1beta1.ApplicationInstance) (walmInstance WalmApplicationInstance, err error) {
	walmInstance = WalmApplicationInstance{
		WalmMeta: buildWalmMetaWithoutState("ApplicationInstance", instance.Namespace, instance.Name),
	}

	walmInstance.Modules, err = adaptor.getWalmInstanceModules(instance)
	if err != nil {
		return
	}
	walmInstance.State = adaptor.buildWalmInstanceState(walmInstance.Modules)
	walmInstance.Events, err = adaptor.getInstanceEvents(instance)
	return
}

func (adaptor *WalmInstanceAdaptor) getWalmInstanceModules(instance *v1beta1.ApplicationInstance) ([]WalmModule, error) {
	walmModules := []WalmModule{}
	for _, module := range instance.Status.Modules {
		resource, err := adaptor.adaptorSet.GetAdaptor(module.ResourceRef.Kind).
			GetResource(module.ResourceRef.Namespace, module.ResourceRef.Name)
		if err != nil {
			return walmModules, err
		}
		walmModules = append(walmModules, WalmModule{module.ResourceRef.Kind, resource})
	}
	return walmModules, nil
}
func (adaptor *WalmInstanceAdaptor) buildWalmInstanceState(modules []WalmModule) (instanceState WalmState) {
	instanceState = buildWalmState("Ready", "", "")
	for _, module := range modules {
		if module.Resource.GetState().Status != "Ready" {
			instanceState = buildWalmState("Pending", "ModulePending", fmt.Sprintf("%s %s/%s is in state %s", module.Kind, module.Resource.GetNamespace(), module.Resource.GetName(), module.Resource.GetState().Status))
			return
		}
	}

	return
}
func (adaptor *WalmInstanceAdaptor) getInstanceEvents(inst *v1beta1.ApplicationInstance) ([]v1.Event, error) {
	ref := v1.ObjectReference{
		Namespace:       inst.Namespace,
		Name:            inst.Name,
		Kind:            inst.Kind,
		ResourceVersion: inst.ResourceVersion,
		UID:             inst.UID,
		APIVersion:      inst.APIVersion,
	}
	events, err := adaptor.adaptorSet.GetHandlerSet().GetEventHandler().SearchEvents(inst.Namespace, &ref)
	if err != nil {
		return nil, err
	}
	return events.Items, nil
}
