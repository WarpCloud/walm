package release

import (
	"walm/pkg/k8s/adaptor"
	"k8s.io/helm/pkg/walm"
)

type ReleaseInfoList struct {
	Num   int            `json:"num" description:"release num"`
	Items []*ReleaseInfo `json:"items" description:"releases list"`
}

type ReleaseInfo struct {
	ReleaseSpec
	Ready   bool                     `json:"ready" description:"whether release is ready"`
	Message string                   `json:"message" description:"why release is not ready"`
	Status  *adaptor.WalmResourceSet `json:"releaseStatus" description:"status of release"`
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
	HelmValues
}

type ReleaseCache struct {
	ReleaseSpec
	ReleaseResourceMetas []ReleaseResourceMeta  `json:"releaseResourceMetas" description:"release resource metas"`
	ComputedValues       map[string]interface{} `json:"computedValues" description:"release computed values"`
}

type ReleaseResourceMeta struct {
	Kind      string `json:"kind" description:"resource kind"`
	Namespace string `json:"namespace" description:"resource namespace"`
	Name      string `json:"name" description:"resource name"`
}

type ReleaseRequest struct {
	Name                string                 `json:"name" description:"name of the release"`
	RepoName            string                 `json:"repoName" description:"chart name"`
	ChartName           string                 `json:"chartName" description:"chart name"`
	ChartVersion        string                 `json:"chartVersion" description:"chart repo"`
	ConfigValues        map[string]interface{} `json:"configValues" description:"extra values added to the chart"`
	Dependencies        map[string]string      `json:"dependencies" description:"map of dependency chart name and release"`
	ReleasePrettyParams PrettyChartParams      `json:"releasePrettyParams" description:"pretty chart params for market"`
}

type HelmExtraLabels struct {
	HelmLabels map[string]interface{} `json:"helmlabels"`
}

type HelmValues struct {
	HelmExtraLabels *HelmExtraLabels `json:"HelmExtraLabels"`
}

type RepoInfo struct {
	TenantRepoName string `json:"repoName"`
	TenantRepoURL  string `json:"repoUrl"`
}

type RepoInfoList struct {
	Items []*RepoInfo `json:"items" description:"chart repo list"`
}

type ChartDependencyInfo struct {
	ChartName          string  `json:"chartName"`
	MaxVersion         float32 `json:"maxVersion"`
	MinVersion         float32 `json:"minVersion"`
	DependencyOptional bool    `json:"dependencyOptional"`
}

type ChartInfo struct {
	ChartName         string                `json:"chartName"`
	ChartVersion      string                `json:"chartVersion"`
	ChartDescription  string                `json:"chartDescription"`
	ChartAppVersion   string                `json:"chartAppVersion"`
	ChartEngine       string                `json:"chartEngine"`
	DefaultValue      string                `json:"defaultValue" description:"default values.yaml defined by the chart"`
	DependencyCharts  []ChartDependencyInfo `json:"dependencyCharts" description:"dependency chart name"`
	ChartPrettyParams PrettyChartParams     `json:"chartPrettyParams" description:"pretty chart params for market"`
	Metainfo          *ChartMetaInfo        `json:"metainfo" description:"transwarp chart metainfo"`
}

type ChartDetailInfo struct {
	ChartInfo
	// additional info
	Advantage    []byte `json:"category" description:"chart production advantage description(rich text)"`
	Architecture []byte `json:"architecture" description:"chart production architecture description(rich text)"`
	Icon         []byte `json:"icon" description:"chart icon"`
}

type ChartInfoList struct {
	Items []*ChartInfo `json:"items" description:"chart list"`
}

type ReleaseConfigDeltaEventType string

const (
	CreateOrUpdate ReleaseConfigDeltaEventType = "CreateOrUpdate"
	Delete         ReleaseConfigDeltaEventType = "Delete"
)

type ReleaseConfigDeltaEvent struct {
	Type ReleaseConfigDeltaEventType `json:"type" description:"delta type: CreateOrUpdate, Delete"`
	Data ReleaseConfig               `json:"data" description:"release config data"`
}

type ReleaseConfig struct {
	AppName      string             `json:"appName" description:"release name"`
	Version      string             `json:"version" description:"chart version"`
	InstanceName string             `json:"instanceName" description:"release name"`
	ConfigSets   []ReleaseConfigSet `json:"configsets" description:"configsets"`
}

type ReleaseConfigSet struct {
	Name        string              `json:"name" description:"name"`
	CreatedBy   string              `json:"createdBy" description:"created by"`
	ConfigItems []ReleaseConfigItem `json:"configItems" description:"config items"`
	Format      string              `json:"format" description:"format"`
}

type ReleaseConfigItem struct {
	Name  string                 `json:"name" description:"name"`
	Value map[string]interface{} `json:"value" description:"value"`
	Type  string                 `json:"type" description:"value"`
}

type ReleaseInfoV2 struct {
	ReleaseInfo
	DependenciesConfigValues map[string]interface{} `json:"dependenciesConfigValues" description:"release's dependencies' config values"`
	ComputedValues           map[string]interface{} `json:"computedValues" description:"config values to render chart templates"`
	OutputConfigValues       map[string]interface{} `json:"outputConfigValues" description:"release's output config values'"`
	ReleaseLabels            map[string]string      `json:"releaseLabels" description:"release labels'"`
}

type ReleaseRequestV2 struct {
	ReleaseRequest
	ReleaseLabels map[string]string  `json:"releaseLabels" description:"release labels"`
	Plugins       []*walm.WalmPlugin `json:"plugins" description:"plugins"`
}

type ReleaseInfoV2List struct {
	Num   int              `json:"num" description:"release num"`
	Items []*ReleaseInfoV2 `json:"items" description:"release infos"`
}
