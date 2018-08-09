package adaptor

import (
	corev1 "k8s.io/api/core/v1"
)

type ResourceAdaptor interface {
	GetResource(namespace string, name string) (WalmResource, error)
}

type WalmResource interface {
	GetKind() string
	GetName() string
	GetNamespace() string
	GetState() WalmState
}

type WalmMeta struct {
	Name string
	Namespace string
	Kind string
	State WalmState
}

func(meta WalmMeta) GetName() string {
	return meta.Name
}

func(meta WalmMeta) GetNamespace() string {
	return meta.Namespace
}

func(meta WalmMeta) GetKind() string {
	return meta.Kind
}

func(meta WalmMeta) GetState() WalmState {
	return meta.State
}

type WalmState struct {
	Status  string
	Reason  string
	Message string
}

type WalmApplicationInstance struct {
	WalmMeta
	Modules []WalmModule `json:"walmModules" protobuf:"bytes,8,rep,name=walmModules"`
	Events []corev1.Event
}

type WalmModule struct{
	Kind        string
	Resource    WalmResource
}

type WalmDeployment struct {
	WalmMeta
	Pods []*WalmPod
}

type WalmPod struct {
	WalmMeta
	PodIp string
}

type WalmService struct {
	WalmMeta
	ServiceType corev1.ServiceType
}

type WalmStatefulSet struct {
	WalmMeta
	Pods []*WalmPod
}

type WalmDaemonSet struct {
	WalmMeta
	Pods []*WalmPod
}

type WalmJob struct {
	WalmMeta
	Pods []*WalmPod
}

type WalmConfigMap struct {
	WalmMeta
	Data map[string]string
}

type WalmIngress struct {
	WalmMeta
	Host string
	Path string
	ServiceName string
	ServicePort string
}

type WalmSecret struct {
	WalmMeta
	Data map[string][]byte
	Type corev1.SecretType
}

type WalmNode struct {
	WalmMeta
	Labels map[string]string `json:"labels" description:"node labels"`
	NodeIp string `json:"nodeip" description:"ip of node"`
}

type WalmNodeList struct {
	Items *[]WalmNode `json:"items" description:"node list info"`
}
