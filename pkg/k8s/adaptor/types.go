package adaptor

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Name      string    `json:"name" description:"resource name"`
	Namespace string    `json:"namespace" description:"resource namespace"`
	Kind      string    `json:"kind" description:"resource kind"`
	State     WalmState `json:"state" description:"resource state"`
}

func (meta WalmMeta) GetName() string {
	return meta.Name
}

func (meta WalmMeta) GetNamespace() string {
	return meta.Namespace
}

func (meta WalmMeta) GetKind() string {
	return meta.Kind
}

func (meta WalmMeta) GetState() WalmState {
	return meta.State
}

type WalmState struct {
	Status  string `json:"status" description:"resource state status"`
	Reason  string `json:"reason" description:"resource state reason"`
	Message string `json:"message" description:"resource state message"`
}

type WalmApplicationInstance struct {
	WalmMeta
	Modules []WalmModule   `json:"modules" description:"instance modules"`
	Events  []WalmEvent `json:"events" description:"instance events"`
}

type WalmEvent struct {
	Type           string      `json:"type" description:"event type"`
	Reason         string      `json:"reason" description:"event reason"`
	Message        string      `json:"message" description:"event message"`
	From           string      `json:"from" description:"component reporting this event"`
	Count          int32       `json:"count" description:"the number of times this event has occurred"`
	FirstTimestamp metav1.Time `json:"first_timestamp" description:"The time at which the event was first recorded"`
	LastTimestamp  metav1.Time `json:"last_timestamp" description:"The time at which the most recent occurrence of this event was recorded"`
}

type WalmModule struct {
	Kind     string       `json:"kind" description:"module kind"`
	Resource WalmResource `json:"resource" description:"module object"`
}

type WalmDeployment struct {
	WalmMeta
	ExpectedReplicas  int32      `json:"expected_replicas" description:"expected replicas"`
	UpdatedReplicas   int32      `json:"updated_replicas" description:"updated replicas"`
	CurrentReplicas   int32      `json:"current_replicas" description:"current replicas"`
	AvailableReplicas int32      `json:"available_replicas" description:"available replicas"`
	Pods              []*WalmPod `json:"pods" description:"deployment pods"`
}

type WalmPod struct {
	WalmMeta
	HostIp string `json:"host_ip" description:"host ip where pod is on"`
	PodIp  string `json:"pod_ip" description:"pod ip"`
}

type WalmService struct {
	WalmMeta
	Ports       []WalmServicePort  `json:"ports" description:"service ports"`
	ClusterIp   string             `json:"cluster_ip" description:"service cluster ip"`
	ServiceType corev1.ServiceType `json:"service_type" description:"service type"`
}

type WalmServicePort struct {
	Name       string             `json:"name" description:"service port name"`
	Protocol   corev1.Protocol    `json:"protocol" description:"service port protocol"`
	Port       int32              `json:"port" description:"service port"`
	TargetPort intstr.IntOrString `json:"target_port" description:"backend pod port"`
	NodePort   int32              `json:"node_port" description:"node port"`
}

type WalmStatefulSet struct {
	WalmMeta
	ExpectedReplicas int32      `json:"expected_replicas" description:"expected replicas"`
	ReadyReplicas    int32      `json:"ready_replicas" description:"ready replicas"`
	CurrentVersion   string     `json:"current_version" description:"stateful set pods"`
	UpdateVersion    string     `json:"update_version" description:"stateful set pods"`
	Pods             []*WalmPod `json:"pods" description:"stateful set pods"`
}

type WalmDaemonSet struct {
	WalmMeta
	DesiredNumberScheduled int32      `json:"desired_number_scheduled" description:"desired number scheduled"`
	UpdatedNumberScheduled int32      `json:"updated_number_scheduled" description:"updated number scheduled"`
	NumberAvailable        int32      `json:"number_available" description:"number available"`
	Pods                   []*WalmPod `json:"pods" description:"daemon set pods"`
}

type WalmJob struct {
	WalmMeta
	ExpectedCompletion int32      `json:"expected_completion" description:"expected num which is succeeded"`
	Succeeded          int32      `json:"succeeded" description:"succeeded pods"`
	Failed             int32      `json:"failed" description:"failed pods"`
	Active             int32      `json:"active" description:"active pods"`
	Pods               []*WalmPod `json:"pods" description:"job pods"`
}

type WalmConfigMap struct {
	WalmMeta
	Data map[string]string `json:"data" description:"config map data"`
}

type WalmIngress struct {
	WalmMeta
	Host        string `json:"host" description:"ingress host"`
	Path        string `json:"path" description:"ingress path"`
	ServiceName string `json:"service_name" description:"ingress backend service name"`
	ServicePort string `json:"service_port" description:"ingress backend service port"`
}

type WalmSecret struct {
	WalmMeta
	Data map[string][]byte `json:"data" description:"secret data"`
	Type corev1.SecretType `json:"type" description:"secret type"`
}

type WalmNode struct {
	WalmMeta
	Labels map[string]string `json:"labels" description:"node labels"`
	NodeIp string            `json:"node_ip" description:"ip of node"`
}

type WalmNodeList struct {
	Items []*WalmNode `json:"items" description:"node list info"`
}

type WalmResourceQuota struct {
	WalmMeta
	ResourceLimits map[corev1.ResourceName]string `json:"limits" description:"resource quota hard limits"`
}
