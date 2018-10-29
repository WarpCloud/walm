package release

import (
	"walm/pkg/k8s/adaptor"
	"k8s.io/helm/pkg/transwarp"
)

type ReleaseInfoList struct {
	Num   int            `json:"num" description:"release num"`
	Items []*ReleaseInfo `json:"items" description:"releases list"`
}

type ReleaseInfo struct {
	ReleaseSpec
	Ready  bool                     `json:"ready" description:"whether release is ready"`
	Status *adaptor.WalmResourceSet `json:"releaseStatus" description:"status of release"`
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
	ReleaseResourceMetas []ReleaseResourceMeta `json:"releaseResourceMetas" description:"release resource metas"`
}

type ReleaseResourceMeta struct {
	Kind      string `json:"kind" description:"resource kind"`
	Namespace string `json:"namespace" description:"resource namespace"`
	Name      string `json:"name" description:"resource name"`
}

type ChartValicationInfo struct {
	Name         string                 `json:"name" description:"name of the release"`
	ConfigValues map[string]interface{} `json:"configValues" description:"extra values added to the chart"`
	Version      int32                  `json:"version" description:"version of the release"`
	Namespace    string                 `json:"namespace" description:"namespace of release"`
	Dependencies map[string]string      `json:"dependencies" description:"map of dependency chart name and release"`
	ChartName    string                 `json:"chartName" description:"chart name"`
	ChartVersion string                 `json:"chartVersion" description:"chart version"`
	RenderStatus string                 `json:"renderStatus" description:"status of rending "`
	RenderResult map[string]string      `json:"renderResult" description:"result of rending "`
	DryRunStatus string                 `json:"dryRunStatus" description:"status of dry run "`
	DryRunResult map[string]string      `json:"dryRunResult" description:"result of dry run "`
	ErrorMessage string                 `json:"errorMessage" description:" error msg "`
}

type ReleaseRequest struct {
	Name         string                 `json:"name" description:"name of the release"`
	RepoName     string                 `json:"repoName" description:"chart name"`
	ChartName    string                 `json:"chartName" description:"chart name"`
	ChartVersion string                 `json:"chartVersion" description:"chart repo"`
	ConfigValues map[string]interface{} `json:"configValues" description:"extra values added to the chart"`
	Dependencies map[string]string      `json:"dependencies" description:"map of dependency chart name and release"`
}

type DependencyDeclare struct {
	// name of dependency declaration
	Name string `json:"name,omitempty"`
	// dependency variable mappings
	Requires map[string]string `json:"requires,omitempty"`
}

type AppDependency struct {
	Name         string               `json:"name,omitempty"`
	Dependencies []*DependencyDeclare `json:"dependencies"`
}

type HelmNativeValues struct {
	ChartName        string `json:"chartName"`
	ChartVersion     string `json:"chartVersion"`
	AppVersion       string `json:"appVersion"`
	ReleaseName      string `json:"releaseName"`
	ReleaseNamespace string `json:"releaseNamespace"`
}

type AppHelmValues struct {
	transwarp.AppHelmValues
}

type ProjectParams struct {
	CommonValues map[string]interface{} `json:"commonValues" description:"common values added to the chart"`
	Releases     []*ReleaseRequest      `json:"releases" description:"list of release of the project"`
}

type ProjectInfo struct {
	ProjectCache
	Releases []*ReleaseInfo `json:"releases" description:"list of release of the project"`
	Ready    bool           `json:"ready" description:"whether all the project releases are ready"`
}

type ProjectCache struct {
	Name                  string          `json:"name" description:"project name"`
	Namespace             string          `json:"namespace" description:"project namespace"`
	LatestProjectJobState ProjectJobState `json:"latestProjectJobState" description:"latest project job state"`
}

func (projectCache *ProjectCache) IsProjectJobNotFinished() bool {
	return projectCache.LatestProjectJobState.Status == "Running" || projectCache.LatestProjectJobState.Status == "Pending"
}

type ProjectJobState struct {
	Async   bool   `json:"async" description:"whether project job is async"`
	Type    string `json:"type" description:"project job type: create, add_releases, remove_releases, delete"`
	Status  string `json:"status" description:"project job status: pending, running, failed, succeed"`
	Message string `json:"message" description:"project job message"`
}

type ProjectInfoList struct {
	Num   int            `json:"num" description:"project number"`
	Items []*ProjectInfo `json:"items" description:"project info list"`
}

type HelmExtraLabels struct {
	HelmLabels map[string]interface{} `json:"helmlabels"`
}

type HelmValues struct {
	HelmExtraLabels *HelmExtraLabels `json:"HelmExtraLabels"`
	AppHelmValues   *AppHelmValues   `json:"HelmAdditionalValues"`
}

type RepoInfo struct {
	TenantRepoName string `json:"repoName"`
	TenantRepoURL  string `json:"repoUrl"`
}

type RepoInfoList struct {
	Items []*RepoInfo `json:"items" description:"chart repo list"`
}

type ChartInfo struct {
	ChartName        string   `json:"chartName"`
	ChartVersion     string   `json:"chartVersion"`
	ChartDescription string   `json:"chartDescription"`
	ChartAppVersion  string   `json:"chartAppVersion"`
	ChartEngine      string   `json:"chartEngine"`
	DefaultValue     string   `json:"defaultValue" description:"default values.yaml defined by the chart"`
	DependencyCharts []string `json:"dependencyCharts" description:"dependency chart name"`
}

type ChartInfoList struct {
	Items []*ChartInfo `json:"items" description:"chart list"`
}
