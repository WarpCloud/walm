package adaptor

import (
	corev1 "k8s.io/api/core/v1"
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
}

type WalmResourceSet struct {
	Services     []WalmService     `json:"services" description:"release services"`
	ConfigMaps   []WalmConfigMap   `json:"configmaps" description:"release configmaps"`
	DaemonSets   []WalmDaemonSet   `json:"daemonsets" description:"release daemonsets"`
	Deployments  []WalmDeployment  `json:"deployments" description:"release deployments"`
	Ingresses    []WalmIngress     `json:"ingresses" description:"release ingresses"`
	Jobs         []WalmJob         `json:"jobs" description:"release jobs"`
	Secrets      []WalmSecret      `json:"secrets" description:"release secrets"`
	StatefulSets []WalmStatefulSet `json:"statefulsets" description:"release statefulsets"`
}

func (resourceSet *WalmResourceSet) GetPodsNeedRestart() []*WalmPod {
	walmPods := []*WalmPod{}
	for _, ds := range resourceSet.DaemonSets {
		if len(ds.Pods) > 0 {
			walmPods = append(walmPods, ds.Pods...)
		}
	}
	for _, ss := range resourceSet.StatefulSets {
		if len(ss.Pods) > 0 {
			walmPods = append(walmPods, ss.Pods...)
		}
	}
	for _, dp := range resourceSet.Deployments {
		if len(dp.Pods) > 0 {
			walmPods = append(walmPods, dp.Pods...)
		}
	}
	return walmPods
}

func (resourceSet *WalmResourceSet) IsReady() (bool, WalmResource) {
	for _, secret := range resourceSet.Secrets {
		if secret.State.Status != "Ready" {
			return false, secret
		}
	}

	for _, job := range resourceSet.Jobs {
		if job.State.Status != "Ready" {
			return false, job
		}
	}

	for _, statefulSet := range resourceSet.StatefulSets {
		if statefulSet.State.Status != "Ready" {
			return false, statefulSet
		}
	}

	for _, service := range resourceSet.Services {
		if service.State.Status != "Ready" {
			return false, service
		}
	}

	for _, ingress := range resourceSet.Ingresses {
		if ingress.State.Status != "Ready" {
			return false, ingress
		}
	}

	for _, deployment := range resourceSet.Deployments {
		if deployment.State.Status != "Ready" {
			return false, deployment
		}
	}

	for _, daemonSet := range resourceSet.DaemonSets {
		if daemonSet.State.Status != "Ready" {
			return false, daemonSet
		}
	}

	for _, configMap := range resourceSet.ConfigMaps {
		if configMap.State.Status != "Ready" {
			return false, configMap
		}
	}

	return true, nil
}

func NewWalmResourceSet() *WalmResourceSet {
	return &WalmResourceSet{
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

type WalmEvent struct {
	Type           string      `json:"type" description:"event type"`
	Reason         string      `json:"reason" description:"event reason"`
	Message        string      `json:"message" description:"event message"`
	From           string      `json:"from" description:"component reporting this event"`
	Count          int32       `json:"count" description:"the number of times this event has occurred"`
	FirstTimestamp metav1.Time `json:"firstTimestamp" description:"The time at which the event was first recorded"`
	LastTimestamp  metav1.Time `json:"lastTimestamp" description:"The time at which the most recent occurrence of this event was recorded"`
}

type WalmEventList struct {
	Events []WalmEvent `json:"events" description:"events"`
}

type WalmDeployment struct {
	WalmMeta
	Labels            map[string]string `json:"labels" description:"deployment labels"`
	Annotations       map[string]string `json:"annotations" description:"deployment annotations"`
	ExpectedReplicas  int32             `json:"expectedReplicas" description:"expected replicas"`
	UpdatedReplicas   int32             `json:"updatedReplicas" description:"updated replicas"`
	CurrentReplicas   int32             `json:"currentReplicas" description:"current replicas"`
	AvailableReplicas int32             `json:"availableReplicas" description:"available replicas"`
	Pods              []*WalmPod        `json:"pods" description:"deployment pods"`
}

func (resource WalmDeployment) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
	resourceSet.Deployments = append(resourceSet.Deployments, resource)
}

type WalmPod struct {
	WalmMeta
	Labels      map[string]string `json:"labels" description:"pod labels"`
	Annotations map[string]string `json:"annotations" description:"pod annotations"`
	HostIp      string            `json:"hostIp" description:"host ip where pod is on"`
	PodIp       string            `json:"podIp" description:"pod ip"`
	Containers  []WalmContainer   `json:"containers" description:"pod containers"`
}

type WalmContainer struct {
	Name         string    `json:"name" description:"container name"`
	Image        string    `json:"image" description:"container image"`
	Ready        bool      `json:"ready" description:"container ready"`
	RestartCount int32     `json:"restartCount" description:"container restart count"`
	State        WalmState `json:"state" description:"container state"`
}

func (resource WalmPod) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
}

type WalmService struct {
	WalmMeta
	Ports       []WalmServicePort  `json:"ports" description:"service ports"`
	ClusterIp   string             `json:"clusterIp" description:"service cluster ip"`
	ServiceType corev1.ServiceType `json:"serviceType" description:"service type"`
}

func (resource WalmService) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
	resourceSet.Services = append(resourceSet.Services, resource)
}

type WalmServicePort struct {
	Name       string          `json:"name" description:"service port name"`
	Protocol   corev1.Protocol `json:"protocol" description:"service port protocol"`
	Port       int32           `json:"port" description:"service port"`
	TargetPort string          `json:"targetPort" description:"backend pod port"`
	NodePort   int32           `json:"nodePort" description:"node port"`
	Endpoints  []string        `json:"endpoints" description:"service endpoints"`
}

type WalmStatefulSet struct {
	WalmMeta
	Labels           map[string]string     `json:"labels" description:"stateful set labels"`
	Annotations      map[string]string     `json:"annotations" description:"stateful set annotations"`
	ExpectedReplicas int32                 `json:"expectedReplicas" description:"expected replicas"`
	ReadyReplicas    int32                 `json:"readyReplicas" description:"ready replicas"`
	CurrentVersion   string                `json:"currentVersion" description:"stateful set pods"`
	UpdateVersion    string                `json:"updateVersion" description:"stateful set pods"`
	Pods             []*WalmPod            `json:"pods" description:"stateful set pods"`
	Selector         *metav1.LabelSelector `json:"-" description:"stateful set label selector"`
}

func (resource WalmStatefulSet) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
	resourceSet.StatefulSets = append(resourceSet.StatefulSets, resource)
}

type WalmDaemonSet struct {
	WalmMeta
	Labels                 map[string]string `json:"labels" description:"daemon set labels"`
	Annotations            map[string]string `json:"annotations" description:"daemon set annotations"`
	DesiredNumberScheduled int32             `json:"desiredNumberScheduled" description:"desired number scheduled"`
	UpdatedNumberScheduled int32             `json:"updatedNumberScheduled" description:"updated number scheduled"`
	NumberAvailable        int32             `json:"numberAvailable" description:"number available"`
	Pods                   []*WalmPod        `json:"pods" description:"daemon set pods"`
}

func (resource WalmDaemonSet) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
	resourceSet.DaemonSets = append(resourceSet.DaemonSets, resource)
}

type WalmJob struct {
	WalmMeta
	Labels             map[string]string `json:"labels" description:"job labels"`
	Annotations        map[string]string `json:"annotations" description:"job annotations"`
	ExpectedCompletion int32             `json:"expectedCompletion" description:"expected num which is succeeded"`
	Succeeded          int32             `json:"succeeded" description:"succeeded pods"`
	Failed             int32             `json:"failed" description:"failed pods"`
	Active             int32             `json:"active" description:"active pods"`
	Pods               []*WalmPod        `json:"pods" description:"job pods"`
}

func (resource WalmJob) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
	resourceSet.Jobs = append(resourceSet.Jobs, resource)
}

type WalmConfigMap struct {
	WalmMeta
	Data map[string]string `json:"data" description:"config map data"`
}

func (resource WalmConfigMap) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
	resourceSet.ConfigMaps = append(resourceSet.ConfigMaps, resource)
}

type WalmIngress struct {
	WalmMeta
	Host        string `json:"host" description:"ingress host"`
	Path        string `json:"path" description:"ingress path"`
	ServiceName string `json:"serviceName" description:"ingress backend service name"`
	ServicePort string `json:"servicePort" description:"ingress backend service port"`
}

func (resource WalmIngress) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
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

type WalmSecretList struct {
	Num   int           `json:"num" description:"secret num"`
	Items []*WalmSecret `json:"items" description:"secrets"`
}

type NodeResourceInfo struct {
	Cpu    float64 `json:"cpu" description:"cpu with unit 1"`
	Memory int64   `json:"memory" description:"memory with unit Mi"`
}

type UnifyUnitNodeResourceInfo struct {
	Capacity          NodeResourceInfo `json:"capacity" description:"node capacity info"`
	Allocatable       NodeResourceInfo `json:"allocatable" description:"node allocatable info"`
	RequestsAllocated NodeResourceInfo `json:"requestsAllocated" description:"node requests allocated info"`
	LimitsAllocated   NodeResourceInfo `json:"limitsAllocated" description:"node limits allocated info"`
}

type WalmNode struct {
	WalmMeta
	Labels                map[string]string         `json:"labels" description:"node labels"`
	Annotations           map[string]string         `json:"annotations" description:"node annotations"`
	NodeIp                string                    `json:"nodeIp" description:"ip of node"`
	Capacity              map[string]string         `json:"capacity" description:"resource capacity"`
	Allocatable           map[string]string         `json:"allocatable" description:"resource allocatable"`
	RequestsAllocated     map[string]string         `json:"requestsAllocated" description:"requests resource allocated"`
	LimitsAllocated       map[string]string         `json:"limitsAllocated" description:"limits resource allocated"`
	WarpDriveStorageList  []WarpDriveStorage        `json:"warpDriveStorageList" description:"warp drive storage list"`
	UnifyUnitResourceInfo UnifyUnitNodeResourceInfo `json:"unifyUnitResourceInfo" description:"resource info with unified unit"`
}

type WarpDriveStorage struct {
	PoolName     string `json:"poolName" description:"pool name"`
	StorageLeft  int64  `json:"storageLeft" description:"storage left, unit: kb"`
	StorageTotal int64  `json:"storageTotal" description:"storage total, unit: kb"`
}

type PoolResource struct {
	PoolName string
	SubPools map[string]SubPoolInfo
}

type SubPoolInfo struct {
	Phase      string `json:"-"`
	Name       string `json:"name,omitempty"`
	DriverName string `json:"driverName,omitempty"`
	// This 'Parent' just use for create and delete from api.
	// Can't get parent from pool.Info()
	Parent            string `json:"parent,omitempty"`
	Size              int64  `json:"size,omitempty"`
	Throughput        int64  `json:"throughput,omitempty"`
	UsedSize          int64  `json:"usedSize,omitempty"`
	RequestSize       int64  `json:"requestSize,omitempty"`
	RequestThroughput int64  `json:"requestThroughput,omitempty"`
}

func (resource WalmNode) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
}

type WalmNodeList struct {
	Items []*WalmNode `json:"items" description:"node list info"`
}

type WalmResourceQuota struct {
	WalmMeta
	ResourceLimits map[corev1.ResourceName]string `json:"limits" description:"resource quota hard limits"`
	ResourceUsed   map[corev1.ResourceName]string `json:"used" description:"resource quota used"`
}

func (resource WalmResourceQuota) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
}

type WalmPersistentVolumeClaim struct {
	WalmMeta
	StorageClass string                              `json:"storageClass" description:"storage class"`
	VolumeName   string                              `json:"volumeName" description:"volume name"`
	Capacity     string                              `json:"capacity" description:"capacity"`
	AccessModes  []corev1.PersistentVolumeAccessMode `json:"accessModes" description:"access modes"`
	VolumeMode   string                              `json:"volumeMode" description:"volume mode"`
}

func (resource WalmPersistentVolumeClaim) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
}

type WalmPersistentVolumeClaimList struct {
	Num   int                          `json:"num" description:"pvc num"`
	Items []*WalmPersistentVolumeClaim `json:"items" description:"pvcs"`
}

type WalmStorageClass struct {
	WalmMeta
	Provisioner          string `json:"provisioner"description:"sc provisioner"`
	ReclaimPolicy        string `json:"reclaimPolicy" description:"sc reclaim policy"`
	AllowVolumeExpansion bool   `json:"allowVolumeExpansion" description:"sc allow volume expansion"`
	VolumeBindingMode    string `json:"volumeBindingMode" description:"sc volume binding mode"`
}

func (resource WalmStorageClass) AddToWalmResourceSet(resourceSet *WalmResourceSet) {
}

type WalmStorageClassList struct {
	Num   int                 `json:"num" description:"storage class num"`
	Items []*WalmStorageClass `json:"items" description:"storage classes"`
}
