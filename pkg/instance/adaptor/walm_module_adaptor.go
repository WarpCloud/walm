package adaptor

import (
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	"fmt"
)

type WalmModuleAdaptor interface{
	GetWalmModule(v1beta1.ResourceReference) (WalmModule, error)
}

type WalmDefaultAdaptor struct{}

func(adaptor WalmDefaultAdaptor) GetWalmModule(module v1beta1.ResourceReference) (walmModule WalmModule, err error) {
	walmModule = WalmModule{
		Kind:        module.ResourceRef.Kind,
		ModuleState: BuildWalmState("Unknown", "NotSupportKind", fmt.Sprintf("%s/%s is a kind not supported", module.ResourceRef.Namespace, module.ResourceRef.Name)),
	}
	return
}
