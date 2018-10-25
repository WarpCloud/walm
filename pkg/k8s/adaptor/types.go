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
	AddToWalmResourceSet(resourceSet *WalmResourceSet)
	AddToWalmInstanceResourceSet(resourceSet *WalmInstanceResourceSet)
}

type WalmResourceSet struct {
	WalmInstanceResourceSet
	Instances []WalmApplicationInstance `json:"instances" description:"release instances"`
}

func (resourceSet *WalmResourceSet) IsReady() bool {
	if ready, _ := resourceSet.WalmInstanceResourceSet.IsReady(); !ready {
		return false
	}

	for _, instance := range resourceSet.Instances {
		if instance.State.Status != "Ready" {
			return false
		}
	}

	return true
}

func NewWalmResourceSet() *WalmResourceSet {
	return &WalmResourceSet{
		WalmInstanceResourceSet: *NewWalmInstanceResourceSet(),
		Instances:               []WalmApplicationInstance{},
	}
}

type WalmInstanceResourceSet struct {
	Services     []WalmService     `json:"services" description:"release services"`
	ConfigMaps   []WalmConfigMap   `json:"configmaps" description:"release configmaps"`
	DaemonSets   []WalmDaemonSet   `json:"daemonsets" description:"release daemonsets"`
	Deployments  []WalmDeployment  `json:"deployments" description:"release deployments"`
	Ingresses    []WalmIngress     `json:"ingresses" description:"release ingresses"`
	Jobs         []WalmJob         `json:"jobs" description:"release jobs"`
	Secrets      []WalmSecret      `json:"secrets" description:"release secrets"`
	StatefulSets []WalmStatefulSet `json:"statefulsets" description:"release statefulsets"`
}

func (instanceResourceSet *WalmInstanceResourceSet) IsReady() (bool, WalmResource) {
	for _, secret := range instanceResourceSet.Secrets {
		if secret.State.Status != "Ready" {
			return false, secret
		}
	}

	for _, job := range instanceResourceSet.Jobs {
		if job.State.Status != "Ready" {
			return false, job
		}
	}

	for _, statefulSet := range instanceResourceSet.StatefulSets {
		if statefulSet.State.Status != "Ready" {
			return false, statefulSet
		}
	}

	for _, service := range instanceResourceSet.Services {
		if service.State.Status != "Ready" {
			return false, service
		}
	}

	for _, ingress := range instanceResourceSet.Ingresses {
		if ingress.State.Status != "Ready" {
			return false, ingress
		}
	}

	for _, deployment := range instanceResourceSet.Deployments {
		if deployment.State.Status != "Ready" {
			return false, deployment
		}
	}

	for _, daemonSet := range instanceResourceSet.DaemonSets {
		if daemonSet.State.Status != "Ready" {
			return false, daemonSet
		}
	}

	for _, configMap := range instanceResourceSet.ConfigMaps {
		if configMap.State.Status != "Ready" {
			return false, configMap
		}
	}

	return true, nil
}

func NewWalmInstanceResourceSet() *WalmInstanceResourceSet {
	return &WalmInstanceResourceSet{
		StatefulSets: []WalmStatefulSet{},
		Services:     []WalmService{},
		Jobs:         []WalmJob{},
		Ingresses:    []WalmIngress{},
		Deployments:  []WalmDeployment{},
		DaemonSets:   []WalmDaemonSet{},
		ConfigMaps:   []WalmConfigMap{},
		Secrets:      []WalmSecret{},
	}
}

type WalmDefaultResource struct {
	WalmMeta
}

func (resource WalmDefaultResource) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
}

func (resource WalmDefaultResource) AddToWalmInstanceResourceSet(resourceSet *WalmInstanceResourceSet) {
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
	InstanceId string                   `json:"instance_id" description:"instance id"`
	Modules    *WalmInstanceResourceSet `json:"modules" description:"instance modules"`
	Events     []WalmEvent              `json:"events" description:"instance events"`
}

func (resource WalmApplicationInstance) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
	resourceSet.Instances = append(resourceSet.Instances, resource)
}

func (resource WalmApplicationInstance) AddToWalmInstanceResourceSet(resourceSet *WalmInstanceResourceSet) {
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

type WalmEventList struct {
	Events []WalmEvent `json:"events" description:"events"`
}

type WalmDeployment struct {
	WalmMeta
	ExpectedReplicas  int32      `json:"expected_replicas" description:"expected replicas"`
	UpdatedReplicas   int32      `json:"updated_replicas" description:"updated replicas"`
	CurrentReplicas   int32      `json:"current_replicas" description:"current replicas"`
	AvailableReplicas int32      `json:"available_replicas" description:"available replicas"`
	Pods              []*WalmPod `json:"pods" description:"deployment pods"`
}

func (resource WalmDeployment) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
	resourceSet.Deployments = append(resourceSet.Deployments, resource)
}

func (resource WalmDeployment) AddToWalmInstanceResourceSet(resourceSet *WalmInstanceResourceSet) {
	resourceSet.Deployments = append(resourceSet.Deployments, resource)
}

type WalmPod struct {
	WalmMeta
	HostIp string `json:"host_ip" description:"host ip where pod is on"`
	PodIp  string `json:"pod_ip" description:"pod ip"`
}

func (resource WalmPod) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
}

func (resource WalmPod) AddToWalmInstanceResourceSet(resourceSet *WalmInstanceResourceSet) {
}

type WalmService struct {
	WalmMeta
	Ports       []WalmServicePort  `json:"ports" description:"service ports"`
	ClusterIp   string             `json:"cluster_ip" description:"service cluster ip"`
	ServiceType corev1.ServiceType `json:"service_type" description:"service type"`
}

func (resource WalmService) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
	resourceSet.Services = append(resourceSet.Services, resource)
}

func (resource WalmService) AddToWalmInstanceResourceSet(resourceSet *WalmInstanceResourceSet) {
	resourceSet.Services = append(resourceSet.Services, resource)
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

func (resource WalmStatefulSet) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
	resourceSet.StatefulSets = append(resourceSet.StatefulSets, resource)
}

func (resource WalmStatefulSet) AddToWalmInstanceResourceSet(resourceSet *WalmInstanceResourceSet) {
	resourceSet.StatefulSets = append(resourceSet.StatefulSets, resource)
}

type WalmDaemonSet struct {
	WalmMeta
	DesiredNumberScheduled int32      `json:"desired_number_scheduled" description:"desired number scheduled"`
	UpdatedNumberScheduled int32      `json:"updated_number_scheduled" description:"updated number scheduled"`
	NumberAvailable        int32      `json:"number_available" description:"number available"`
	Pods                   []*WalmPod `json:"pods" description:"daemon set pods"`
}

func (resource WalmDaemonSet) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
	resourceSet.DaemonSets = append(resourceSet.DaemonSets, resource)
}

func (resource WalmDaemonSet) AddToWalmInstanceResourceSet(resourceSet *WalmInstanceResourceSet) {
	resourceSet.DaemonSets = append(resourceSet.DaemonSets, resource)
}

type WalmJob struct {
	WalmMeta
	ExpectedCompletion int32      `json:"expected_completion" description:"expected num which is succeeded"`
	Succeeded          int32      `json:"succeeded" description:"succeeded pods"`
	Failed             int32      `json:"failed" description:"failed pods"`
	Active             int32      `json:"active" description:"active pods"`
	Pods               []*WalmPod `json:"pods" description:"job pods"`
}

func (resource WalmJob) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
	resourceSet.Jobs = append(resourceSet.Jobs, resource)
}

func (resource WalmJob) AddToWalmInstanceResourceSet(resourceSet *WalmInstanceResourceSet) {
	resourceSet.Jobs = append(resourceSet.Jobs, resource)
}

type WalmConfigMap struct {
	WalmMeta
	Data map[string]string `json:"data" description:"config map data"`
}

func (resource WalmConfigMap) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
	resourceSet.ConfigMaps = append(resourceSet.ConfigMaps, resource)
}

func (resource WalmConfigMap) AddToWalmInstanceResourceSet(resourceSet *WalmInstanceResourceSet) {
	resourceSet.ConfigMaps = append(resourceSet.ConfigMaps, resource)
}

type WalmIngress struct {
	WalmMeta
	Host        string `json:"host" description:"ingress host"`
	Path        string `json:"path" description:"ingress path"`
	ServiceName string `json:"service_name" description:"ingress backend service name"`
	ServicePort string `json:"service_port" description:"ingress backend service port"`
}

func (resource WalmIngress) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
	resourceSet.Ingresses = append(resourceSet.Ingresses, resource)
}

func (resource WalmIngress) AddToWalmInstanceResourceSet(resourceSet *WalmInstanceResourceSet) {
	resourceSet.Ingresses = append(resourceSet.Ingresses, resource)
}

type WalmSecret struct {
	WalmMeta
	Data map[string]string `json:"data" description:"secret data"`
	Type corev1.SecretType `json:"type" description:"secret type"`
}

func (resource WalmSecret) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
	resourceSet.Secrets = append(resourceSet.Secrets, resource)
}

func (resource WalmSecret) AddToWalmInstanceResourceSet(resourceSet *WalmInstanceResourceSet) {
	resourceSet.Secrets = append(resourceSet.Secrets, resource)
}

type WalmSecretList struct {
	Num   int           `json:"num" description:"secret num"`
	Items []*WalmSecret `json:"items" description:"secrets"`
}

type WalmNode struct {
	WalmMeta
	Labels            map[string]string   `json:"labels" description:"node labels"`
	Annotations       map[string]string   `json:"annotations" description:"node annotations"`
	NodeIp            string              `json:"node_ip" description:"ip of node"`
	Capacity          corev1.ResourceList `json:"capacity" description:"resource capacity"`
	Allocatable       corev1.ResourceList `json:"allocatable" description:"resource allocatable"`
	RequestsAllocated corev1.ResourceList `json:"requests_allocated" description:"requests resource allocated"`
	LimitsAllocated   corev1.ResourceList `json:"limits_allocated" description:"limits resource allocated"`
}

func (resource WalmNode) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
}

func (resource WalmNode) AddToWalmInstanceResourceSet(resourceSet *WalmInstanceResourceSet) {
}

type WalmNodeList struct {
	Items []*WalmNode `json:"items" description:"node list info"`
}

type WalmResourceQuota struct {
	WalmMeta
	ResourceLimits map[corev1.ResourceName]string `json:"limits" description:"resource quota hard limits"`
}

func (resource WalmResourceQuota) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
}

func (resource WalmResourceQuota) AddToWalmInstanceResourceSet(resourceSet *WalmInstanceResourceSet) {
}
