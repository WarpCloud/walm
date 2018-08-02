package instance

import (
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	"walm/pkg/instance/lister"
	"walm/pkg/instance/adaptor"
	"fmt"
)

type InstanceManager struct {
	lister lister.K8sResourceLister
}

func (instManager InstanceManager) BuildWalmApplicationInstance(inst v1beta1.ApplicationInstance) (walmInst adaptor.WalmApplicationInstance, err error) {
	walmInst.ObjectMeta = *inst.ObjectMeta.DeepCopy()
	walmInst.TypeMeta = inst.TypeMeta
	walmInst.Spec = *inst.Spec.DeepCopy()
	walmInst.Status.ApplicationInstanceStatus = *inst.Status.DeepCopy()
	walmInst.Status.WalmModules, err = instManager.buildWalmModules(inst.Status.Modules)
	if err != nil {
		return
	}
	walmInst.Status.Events, err = instManager.lister.GetInstanceEvents(inst)
	walmInst.Status.InstanceState = buildInstanceState(walmInst.Status.WalmModules)
	return
}
func buildInstanceState(modules []adaptor.WalmModule) (instanceState adaptor.WalmState) {
	instanceState = adaptor.BuildWalmState("Ready", "", "")
	for _, module := range modules {
		if module.ModuleState.State != "Ready" {
			instanceState = adaptor.BuildWalmState("Pending", "ModulePending", fmt.Sprintf("%s %s/%s is in state %s", module.Kind, module.Resource.GetNamespace(), module.Resource.GetName(), module.ModuleState.State))
			return
		}
	}

	return
}

func (instManager InstanceManager) buildWalmModules(modules []v1beta1.ResourceReference) (walmModules []adaptor.WalmModule, err error) {
	walmModules = []adaptor.WalmModule{}
	for _, module := range modules {
		walmModule, err := instManager.buildWalmModule(module)
		if err != nil {
			return walmModules, err
		}
		walmModules = append(walmModules, walmModule)
	}
	return
}

func (instManager InstanceManager) buildWalmModule(module v1beta1.ResourceReference) (adaptor.WalmModule, error) {
	walmModuleAdaptor := instManager.getWalmModuleAdaptor(module.ResourceRef.Kind)
	return walmModuleAdaptor.GetWalmModule(module)
}

func (instManager InstanceManager) getWalmModuleAdaptor(kind string) adaptor.WalmModuleAdaptor {
	switch kind {
	case "Deployment":
		return adaptor.WalmDeploymentAdaptor{instManager.lister}
	case "Service":
		return adaptor.WalmServiceAdaptor{instManager.lister}
	case "StatefulSet":
		return adaptor.WalmStatefulSetAdaptor{instManager.lister}
	case "DaemonSet":
		return adaptor.WalmDaemonSetAdaptor{instManager.lister}
	case "Job":
		return adaptor.WalmJobAdaptor{instManager.lister}
	case "ConfigMap":
		return adaptor.WalmConfigMapAdaptor{instManager.lister}
	case "Ingress":
		return adaptor.WalmIngressAdaptor{instManager.lister}
	case "Secret":
		return adaptor.WalmSecretAdaptor{instManager.lister}
	default:
		return adaptor.WalmDefaultAdaptor{}
	}
}
