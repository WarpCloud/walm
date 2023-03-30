package release

import (
	"WarpCloud/walm/pkg/models/k8s"
)

type ReleaseInfoList struct {
	Num   int            `json:"num" description:"release num"`
	Items []*ReleaseInfo `json:"items" description:"releases list"`
}

type ReleaseInfo struct {
	ReleaseSpec
	Ready   bool             `json:"ready" description:"whether release is ready"`
	Message string           `json:"message" description:"why release is not ready"`
	Status  *k8s.ResourceSet `json:"releaseStatus" description:"status of release"`
}

type ReleaseSpec struct {
	Name            string                 `json:"name" description:"name of the release"`
	RepoName        string                 `json:"repoName" description:"chart name"`
	ConfigValues    map[string]interface{} `json:"configValues" description:"extra values added to the chart"`
	Version         int32                  `json:"version" description:"version of the release"`
	Namespace       string                 `json:"namespace" description:"namespace of release"`
	Dependencies    map[string]string      `json:"dependencies" description:"map of dependency chart name and release"`
	ChartName       string                 `json:"chartName" description:"chart name"`
	ChartVersion    string                 `json:"chartVersion" description:"chart version"`
	ChartAppVersion string                 `json:"chartAppVersion" description:"jsonnet app version"`
}

type ReleaseCache struct {
	ReleaseSpec
	ReleaseResourceMetas []ReleaseResourceMeta  `json:"releaseResourceMetas" description:"release resource metas"`
	ComputedValues       map[string]interface{} `json:"computedValues" description:"release computed values"`
	MetaInfoValues       *MetaInfoParams        `json:"metaInfoValues" description:"meta info values"`
	Manifest             string                 `json:"manifest" description:"meta info values"`
}

type ReleaseResourceMeta struct {
	Kind      k8s.ResourceKind `json:"kind" description:"resource kind"`
	Namespace string           `json:"namespace" description:"resource namespace"`
	Name      string           `json:"name" description:"resource name"`
}

type ReleaseRequest struct {
	Name         string                 `json:"name" description:"name of the release"`
	RepoName     string                 `json:"repoName" description:"chart name"`
	ChartName    string                 `json:"chartName" description:"chart name"`
	ChartVersion string                 `json:"chartVersion" description:"chart repo"`
	ConfigValues map[string]interface{} `json:"configValues" description:"extra values added to the chart"`
	Dependencies map[string]string      `json:"dependencies" description:"map of dependency chart name and release"`
}

type ReleaseConfigDeltaEventType string

const (
	CreateOrUpdate ReleaseConfigDeltaEventType = "CreateOrUpdate"
	Delete         ReleaseConfigDeltaEventType = "Delete"
)

type ReleaseConfigDeltaEvent struct {
	Type ReleaseConfigDeltaEventType `json:"type" description:"delta type: CreateOrUpdate, Delete"`
	Data *k8s.ReleaseConfig           `json:"data" description:"release config data"`
}

type ReleaseInfoV2 struct {
	ReleaseInfo
	DependenciesConfigValues map[string]interface{} `json:"dependenciesConfigValues" description:"release's dependencies' config values"`
	ComputedValues           map[string]interface{} `json:"computedValues" description:"config values to render chart templates"`
	OutputConfigValues       map[string]interface{} `json:"outputConfigValues" description:"release's output config values'"`
	ReleaseLabels            map[string]string      `json:"releaseLabels" description:"release labels'"`
	Plugins                  []*ReleasePlugin       `json:"plugins" description:"plugins"`
	MetaInfoValues           *MetaInfoParams        `json:"metaInfoValues" description:"meta info values"`
	Paused                   bool                   `json:"paused" description:"whether release is paused"`
	ChartImage               string                 `json:"chartImage" description:"release chart image"`
}

type ReleasePlugin struct {
	Name    string `json:"name" description:"plugin name"`
	Args    string `json:"args" description:"plugin args"`
	Version string `json:"version" description:"plugin version"`
	Disable bool   `json:"disable" description:"disable plugin"`
}

func (releaseInfo *ReleaseInfoV2) BuildReleaseRequestV2() *ReleaseRequestV2 {
	return &ReleaseRequestV2{
		ReleaseRequest: ReleaseRequest{
			Name:         releaseInfo.Name,
			RepoName:     releaseInfo.RepoName,
			ChartVersion: releaseInfo.ChartVersion,
			ChartName:    releaseInfo.ChartName,
			Dependencies: releaseInfo.Dependencies,
			ConfigValues: releaseInfo.ConfigValues,
		},
		ReleaseLabels: releaseInfo.ReleaseLabels,
		Plugins:       releaseInfo.Plugins,
		ChartImage:    releaseInfo.ChartImage,
	}
}

type ReleaseRequestV2 struct {
	ReleaseRequest
	ReleaseLabels  map[string]string `json:"releaseLabels" description:"release labels"`
	Plugins        []*ReleasePlugin  `json:"plugins" description:"plugins"`
	MetaInfoParams *MetaInfoParams   `json:"metaInfoParams" description:"meta info parameters"`
	ChartImage     string            `json:"chartImage" description:"chart image url"`
}

type ReleaseInfoV2List struct {
	Num   int              `json:"num" description:"release num"`
	Items []*ReleaseInfoV2 `json:"items" description:"release infos"`
}

type ReleaseResources struct {
	Deployments  []*ReleaseResourceDeployment  `json:"deployments" description:"release resource deployments"`
	StatefulSets []*ReleaseResourceStatefulSet `json:"statefulSets" description:"release resource stateful sets"`
	DaemonSets   []*ReleaseResourceDaemonSet   `json:"daemonSets" description:"release resource daemon sets"`
	Jobs         []*ReleaseResourceJob         `json:"jobs" description:"release resource jobs"`
	Pvcs         []*ReleaseResourceStorage     `json:"pvcs" description:"release resource pvcs"`
}

type ReleaseResourcePodStorageType string

const (
	TosDiskPodStorageType ReleaseResourcePodStorageType = "tosDisk"
	PvcPodStorageType     ReleaseResourcePodStorageType = "pvc"
)

type ReleaseResourceStorage struct {
	Name         string                        `json:"name" description:"storage name"`
	Type         ReleaseResourcePodStorageType `json:"type" description:"storage type"`
	StorageClass string                        `json:"storageClass" description:"storage class"`
	Size         int64                         `json:"size" description:"size, unit: Gi"`
}

type ReleaseResourcePod struct {
	Cpu     float64                   `json:"cpu" description:"cpu, unit: 1"`
	Memory  int64                     `json:"memory" description:"memory, unit: Mi"`
	Storage []*ReleaseResourceStorage `json:"storage, omitempty" description:"storage"`
}

type ReleaseResourceBase struct {
	Name        string              `json:"name" description:"release resource name"`
	PodRequests *ReleaseResourcePod `json:"podRequests" description:"pod requests"`
	PodLimits   *ReleaseResourcePod `json:"podLimits" description:"pod limits"`
}

type ReleaseResourceDeployment struct {
	Replicas int32 `json:"replicas" description:"deployment expected replicas"`
	ReleaseResourceBase
}

type ReleaseResourceStatefulSet struct {
	Replicas int32 `json:"replicas" description:"stateful set expected replicas"`
	ReleaseResourceBase
}

type ReleaseResourceDaemonSet struct {
	NodeSelector map[string]string `json:"nodeSelector" description:"daemon set node selector"`
	ReleaseResourceBase
}

type ReleaseResourceJob struct {
	Parallelism int32 `json:"parallelism" description:"job parallelism"`
	Completions int32 `json:"completions" description:"job completions"`
	ReleaseResourceBase
}

type ReleaseDryRunUpdateInfo struct {
	Configmaps       []interface{}       `json:"configmaps" description:"dryrun configmap results"`
	DependedReleases []map[string]string `json:"dependedReleases" description:"releases depended on current release"`
}
