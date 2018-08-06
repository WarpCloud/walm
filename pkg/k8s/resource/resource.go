package resource

import (
	"walm/pkg/k8s/resource/adaptor"
	"walm/pkg/k8s/handler"
)

var defaultResourceSet *ResourceSet

func GetDefaultResourceSet() *ResourceSet{
	if defaultResourceSet == nil {
		defaultResourceSet = &ResourceSet{handler.GetDefaultHandlerSet()}
	}
	return defaultResourceSet
}

type ResourceSet struct {
	handlerSet *handler.HandlerSet
}

func(set ResourceSet) GetResourceAdaptor(kind string) adaptor.ResourceAdaptor{
	return adaptor.GetAdaptor(kind, set.handlerSet)
}




