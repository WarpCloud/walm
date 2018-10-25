package adaptor

import (
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	"fmt"
	"k8s.io/api/core/v1"
	"sync"
	"strings"
	"sort"
	"walm/pkg/k8s/utils"
)

type WalmInstanceAdaptor struct {
	adaptorSet *AdaptorSet
}

func (adaptor *WalmInstanceAdaptor) GetResource(namespace string, name string) (WalmResource, error) {
	instance, err := adaptor.adaptorSet.GetHandlerSet().GetInstanceHandler().GetInstance(namespace, name)
	if err != nil {
		if IsNotFoundErr(err) {
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
		InstanceId: instance.Spec.InstanceId,
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		walmInstance.Events, err = adaptor.getInstanceEvents(instance)
	}()

	walmInstance.Modules, err = adaptor.getWalmInstanceModules(instance)
	if err != nil {
		return
	}

	walmInstance.State = adaptor.buildWalmInstanceState(walmInstance.Modules, instance.Status.Ready)
	wg.Wait()
	return
}

func (adaptor *WalmInstanceAdaptor) getWalmInstanceModules(instance *v1beta1.ApplicationInstance) (instanceResourceSet *WalmInstanceResourceSet, err error) {
	instanceResourceSet = NewWalmInstanceResourceSet()
	for _, module := range instance.Status.Modules {
		resource, err := adaptor.adaptorSet.GetAdaptor(module.ResourceRef.Kind).
			GetResource(module.ResourceRef.Namespace, module.ResourceRef.Name)
		if err != nil {
			return nil, err
		}
		resource.AddToWalmInstanceResourceSet(instanceResourceSet)
	}
	return
}
func (adaptor *WalmInstanceAdaptor) buildWalmInstanceState(instanceResourceSet *WalmInstanceResourceSet, ready bool) (instanceState WalmState) {
	if ready {
		instanceState = buildWalmState("Ready", "", "")
	} else {
		instanceState = buildWalmState("Pending", "ModuleNotEnough", "there is module still not created")
	}

	if ready, notReadyResource := instanceResourceSet.IsReady(); !ready {
		instanceState = buildWalmState("Pending", "ModulePending", fmt.Sprintf("%s %s/%s is in state %s", notReadyResource.GetKind(), notReadyResource.GetNamespace(), notReadyResource.GetName(), notReadyResource.GetState().Status))
	}

	return
}
func (adaptor *WalmInstanceAdaptor) getInstanceEvents(inst *v1beta1.ApplicationInstance) ([]WalmEvent, error) {
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
	sort.Sort(utils.SortableEvents(events.Items))

	walmEvents := []WalmEvent{}
	for _, event := range events.Items {
		walmEvent := WalmEvent{
			Type:           event.Type,
			Reason:         event.Reason,
			Message:        event.Message,
			Count:          event.Count,
			FirstTimestamp: event.FirstTimestamp,
			LastTimestamp:  event.LastTimestamp,
			From:           formatEventSource(event.Source),
		}
		walmEvents = append(walmEvents, walmEvent)
	}
	return walmEvents, nil
}

func formatEventSource(es v1.EventSource) string {
	EventSourceString := []string{es.Component}
	if len(es.Host) > 0 {
		EventSourceString = append(EventSourceString, es.Host)
	}
	return strings.Join(EventSourceString, ", ")
}
