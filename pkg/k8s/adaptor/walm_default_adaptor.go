package adaptor

type WalmDefaultAdaptor struct{
	Kind string
}

func(adaptor *WalmDefaultAdaptor) GetResource(namespace string, name string) (walmresource WalmResource, err error) {
	return	WalmDefaultResource{buildWalmMeta(adaptor.Kind, namespace, name, buildWalmState("Unknown", "NotSupportedKind", "Can not get this resource"))}, nil
}
