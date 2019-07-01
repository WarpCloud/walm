package k8s

const (
	DeploymentKind            ResourceKind = "Deployment"
	ServiceKind               ResourceKind = "Service"
	StatefulSetKind           ResourceKind = "StatefulSet"
	DaemonSetKind             ResourceKind = "DaemonSet"
	JobKind                   ResourceKind = "Job"
	ConfigMapKind             ResourceKind = "ConfigMap"
	IngressKind               ResourceKind = "Ingress"
	SecretKind                ResourceKind = "Secret"
	PodKind                   ResourceKind = "Pod"
	NodeKind                  ResourceKind = "Node"
	ResourceQuotaKind         ResourceKind = "ResourceQuota"
	PersistentVolumeClaimKind ResourceKind = "PersistentVolumeClaim"
	StorageClassKind          ResourceKind = "StorageClass"
	NamespaceKind             ResourceKind = "Namespace"
	ReleaseConfigKind         ResourceKind = "ReleaseConfig"
)

type ResourceKind string

type Resource interface {
	GetKind() ResourceKind
	GetName() string
	GetNamespace() string
	GetState() State
	AddToResourceSet(resourceSet *ResourceSet)
}

type DefaultResource struct {
	Meta
}

func (resource *DefaultResource) AddToResourceSet(resourceSet *ResourceSet) {
}

func NewMeta(kind ResourceKind, namespace string, name string, state State) Meta {
	return Meta{
		Kind:      kind,
		Namespace: namespace,
		Name:      name,
		State:     state,
	}
}

func NewEmptyStateMeta(kind ResourceKind, namespace string, name string) Meta {
	return NewMeta(kind, namespace, name, State{})
}

func NewNotFoundMeta(kind ResourceKind, namespace string, name string) Meta {
	return NewMeta(kind, namespace, name, NewState("NotFound", "", ""))
}

type Meta struct {
	Name      string       `json:"name" description:"resource name"`
	Namespace string       `json:"namespace" description:"resource namespace"`
	Kind      ResourceKind `json:"kind" description:"resource kind"`
	State     State        `json:"state" description:"resource state"`
}

func (meta Meta) GetName() string {
	return meta.Name
}

func (meta Meta) GetNamespace() string {
	return meta.Namespace
}

func (meta Meta) GetKind() ResourceKind {
	return meta.Kind
}

func (meta Meta) GetState() State {
	return meta.State
}

func NewState(state string, reason string, message string) State {
	return State{
		Status:  state,
		Reason:  reason,
		Message: message,
	}
}

type State struct {
	Status  string `json:"status" description:"resource state status"`
	Reason  string `json:"reason" description:"resource state reason"`
	Message string `json:"message" description:"resource state message"`
}

type Event struct {
	Type           string `json:"type" description:"event type"`
	Reason         string `json:"reason" description:"event reason"`
	Message        string `json:"message" description:"event message"`
	From           string `json:"from" description:"component reporting this event"`
	Count          int32  `json:"count" description:"the number of times this event has occurred"`
	FirstTimestamp string `json:"firstTimestamp" description:"The time at which the event was first recorded"`
	LastTimestamp  string `json:"lastTimestamp" description:"The time at which the most recent occurrence of this event was recorded"`
}

type EventList struct {
	Events []Event `json:"events" description:"events"`
}

type Deployment struct {
	Meta
	Labels            map[string]string `json:"labels" description:"deployment labels"`
	Annotations       map[string]string `json:"annotations" description:"deployment annotations"`
	ExpectedReplicas  int32             `json:"expectedReplicas" description:"expected replicas"`
	UpdatedReplicas   int32             `json:"updatedReplicas" description:"updated replicas"`
	CurrentReplicas   int32             `json:"currentReplicas" description:"current replicas"`
	AvailableReplicas int32             `json:"availableReplicas" description:"available replicas"`
	Pods              []*Pod            `json:"pods" description:"deployment pods"`
}

func (resource *Deployment) AddToResourceSet(resourceSet *ResourceSet) {
	resourceSet.Deployments = append(resourceSet.Deployments, resource)
}

type Pod struct {
	Meta
	Labels      map[string]string `json:"labels" description:"pod labels"`
	Annotations map[string]string `json:"annotations" description:"pod annotations"`
	HostIp      string            `json:"hostIp" description:"host ip where pod is on"`
	PodIp       string            `json:"podIp" description:"pod ip"`
	Containers  []Container       `json:"containers" description:"pod containers"`
	Age         string            `json:"age" description:"pod age"`
}

type Container struct {
	Name         string `json:"name" description:"container name"`
	Image        string `json:"image" description:"container image"`
	Ready        bool   `json:"ready" description:"container ready"`
	RestartCount int32  `json:"restartCount" description:"container restart count"`
	State        State  `json:"state" description:"container state"`
}

func (resource *Pod) AddToResourceSet(resourceSet *ResourceSet) {
}

type Service struct {
	Meta
	Ports       []ServicePort `json:"ports" description:"service ports"`
	ClusterIp   string        `json:"clusterIp" description:"service cluster ip"`
	ServiceType string        `json:"serviceType" description:"service type"`
}

func (resource *Service) AddToResourceSet(resourceSet *ResourceSet) {
	resourceSet.Services = append(resourceSet.Services, resource)
}

type ServicePort struct {
	Name       string   `json:"name" description:"service port name"`
	Protocol   string   `json:"protocol" description:"service port protocol"`
	Port       int32    `json:"port" description:"service port"`
	TargetPort string   `json:"targetPort" description:"backend pod port"`
	NodePort   int32    `json:"nodePort" description:"node port"`
	Endpoints  []string `json:"endpoints" description:"service endpoints"`
}

type StatefulSet struct {
	Meta
	Labels           map[string]string `json:"labels" description:"stateful set labels"`
	Annotations      map[string]string `json:"annotations" description:"stateful set annotations"`
	ExpectedReplicas int32             `json:"expectedReplicas" description:"expected replicas"`
	ReadyReplicas    int32             `json:"readyReplicas" description:"ready replicas"`
	CurrentVersion   string            `json:"currentVersion" description:"stateful set pods"`
	UpdateVersion    string            `json:"updateVersion" description:"stateful set pods"`
	Pods             []*Pod            `json:"pods" description:"stateful set pods"`
	Selector         string            `json:"selector" description:"stateful set label selector"`
}

func (resource *StatefulSet) AddToResourceSet(resourceSet *ResourceSet) {
	resourceSet.StatefulSets = append(resourceSet.StatefulSets, resource)
}

type DaemonSet struct {
	Meta
	Labels                 map[string]string `json:"labels" description:"daemon set labels"`
	Annotations            map[string]string `json:"annotations" description:"daemon set annotations"`
	DesiredNumberScheduled int32             `json:"desiredNumberScheduled" description:"desired number scheduled"`
	UpdatedNumberScheduled int32             `json:"updatedNumberScheduled" description:"updated number scheduled"`
	NumberAvailable        int32             `json:"numberAvailable" description:"number available"`
	Pods                   []*Pod            `json:"pods" description:"daemon set pods"`
}

func (resource *DaemonSet) AddToResourceSet(resourceSet *ResourceSet) {
	resourceSet.DaemonSets = append(resourceSet.DaemonSets, resource)
}

type Job struct {
	Meta
	Labels             map[string]string `json:"labels" description:"job labels"`
	Annotations        map[string]string `json:"annotations" description:"job annotations"`
	ExpectedCompletion int32             `json:"expectedCompletion" description:"expected num which is succeeded"`
	Succeeded          int32             `json:"succeeded" description:"succeeded pods"`
	Failed             int32             `json:"failed" description:"failed pods"`
	Active             int32             `json:"active" description:"active pods"`
	Pods               []*Pod            `json:"pods" description:"job pods"`
}

func (resource *Job) AddToResourceSet(resourceSet *ResourceSet) {
	resourceSet.Jobs = append(resourceSet.Jobs, resource)
}

type ConfigMap struct {
	Meta
	Data map[string]string `json:"data" description:"config map data"`
}

func (resource *ConfigMap) AddToResourceSet(resourceSet *ResourceSet) {
	resourceSet.ConfigMaps = append(resourceSet.ConfigMaps, resource)
}

type Ingress struct {
	Meta
	Host        string `json:"host" description:"ingress host"`
	Path        string `json:"path" description:"ingress path"`
	ServiceName string `json:"serviceName" description:"ingress backend service name"`
	ServicePort string `json:"servicePort" description:"ingress backend service port"`
}

func (resource *Ingress) AddToResourceSet(resourceSet *ResourceSet) {
	resourceSet.Ingresses = append(resourceSet.Ingresses, resource)
}

type Secret struct {
	Meta
	Data map[string]string `json:"data" description:"secret data"`
	Type string            `json:"type" description:"secret type"`
}

func (resource *Secret) AddToResourceSet(resourceSet *ResourceSet) {
	resourceSet.Secrets = append(resourceSet.Secrets, resource)
}

type SecretList struct {
	Num   int       `json:"num" description:"secret num"`
	Items []*Secret `json:"items" description:"secrets"`
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

type Node struct {
	Meta
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

func (resource *Node) AddToResourceSet(resourceSet *ResourceSet) {
}

type NodeList struct {
	Items []*Node `json:"items" description:"node list info"`
}

type ResourceQuota struct {
	Meta
	ResourceLimits map[string]string `json:"limits" description:"resource quota hard limits"`
	ResourceUsed   map[string]string `json:"used" description:"resource quota used"`
}

func (resource *ResourceQuota) AddToResourceSet(resourceSet *ResourceSet) {
}

type PersistentVolumeClaim struct {
	Meta
	StorageClass string   `json:"storageClass" description:"storage class"`
	VolumeName   string   `json:"volumeName" description:"volume name"`
	Capacity     string   `json:"capacity" description:"capacity"`
	AccessModes  []string `json:"accessModes" description:"access modes"`
	VolumeMode   string   `json:"volumeMode" description:"volume mode"`
}

func (resource *PersistentVolumeClaim) AddToResourceSet(resourceSet *ResourceSet) {
}

type PersistentVolumeClaimList struct {
	Num   int                      `json:"num" description:"pvc num"`
	Items []*PersistentVolumeClaim `json:"items" description:"pvcs"`
}

type StorageClass struct {
	Meta
	Provisioner          string `json:"provisioner"description:"sc provisioner"`
	ReclaimPolicy        string `json:"reclaimPolicy" description:"sc reclaim policy"`
	AllowVolumeExpansion bool   `json:"allowVolumeExpansion" description:"sc allow volume expansion"`
	VolumeBindingMode    string `json:"volumeBindingMode" description:"sc volume binding mode"`
}

func (resource *StorageClass) AddToResourceSet(resourceSet *ResourceSet) {
}

type StorageClassList struct {
	Num   int             `json:"num" description:"storage class num"`
	Items []*StorageClass `json:"items" description:"storage classes"`
}

type ReleaseConfig struct {
	Meta
	Labels                   map[string]string      `json:"labels" description:"release labels"`
	ConfigValues             map[string]interface{} `json:"configValues" description:"user config values added to the chart"`
	DependenciesConfigValues map[string]interface{} `json:"dependenciesConfigValues" description:"dependencies' config values added to the chart"`
	Dependencies             map[string]string      `json:"dependencies" description:"map of dependency chart name and release"`
	ChartName                string                 `json:"chartName" description:"chart name"`
	ChartVersion             string                 `json:"chartVersion" description:"chart version"`
	ChartAppVersion          string                 `json:"chartAppVersion" description:"jsonnet app version"`
	OutputConfig             map[string]interface{} `json:"outputConfig"`
	Repo                     string                 `json:"repo" description:"chart repo"`
	ChartImage               string                 `json:"chartImage" description:"chart image"`
}

func (resource *ReleaseConfig) AddToResourceSet(resourceSet *ResourceSet) {
}
