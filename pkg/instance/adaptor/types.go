package adaptor

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"transwarp/application-instance/pkg/apis/transwarp/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

type WalmApplicationInstance struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired behavior of the ApplicationInstance.
	// +optional
	Spec v1beta1.ApplicationInstanceSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Most recently observed status of the ApplicationInstance. This data
	// may be out of date by some window of time.
	// +optional
	Status WalmApplicationInstanceStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type WalmApplicationInstanceStatus struct {
	v1beta1.ApplicationInstanceStatus

	WalmModules []WalmModule `json:"walmModules,omitempty" protobuf:"bytes,8,rep,name=walmModules"`
}

type WalmModule struct{
	Kind string
	Object interface{}
}

type WalmMeta struct {
	Name string
	Namespace string
}

type WalmDeployment struct {
	WalmMeta
	Pods []WalmPod
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
	Pods []WalmPod
}

type WalmDaemonSet struct {
	WalmMeta
	Pods []WalmPod
}

type WalmJob struct {
	WalmMeta
	Pods []WalmPod
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
