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
	Status *adaptor.WalmResourceSet `json:"release_status" description:"status of release"`
}

type ReleaseSpec struct {
	Name            string                 `json:"name" description:"name of the release"`
	RepoName        string                 `json:"repo_name" description:"chart name"`
	ConfigValues    map[string]interface{} `json:"config_values" description:"extra values added to the chart"`
	Version         int32                  `json:"version" description:"version of the release"`
	Namespace       string                 `json:"namespace" description:"namespace of release"`
	Dependencies    map[string]string      `json:"dependencies" description:"map of dependency chart name and release"`
	ChartName       string                 `json:"chart_name" description:"chart name"`
	ChartVersion    string                 `json:"chart_version" description:"chart version"`
	ChartAppVersion string                 `json:"chart_app_version" description:"jsonnet app version"`
}

type ReleaseCache struct {
	ReleaseSpec
	ReleaseResourceMetas []ReleaseResourceMeta `json:"release_resource_metas" description:"release resource metas"`
}

type ReleaseResourceMeta struct {
	Kind      string `json:"kind" description:"resource kind"`
	Namespace string `json:"namespace" description:"resource namespace"`
	Name      string `json:"name" description:"resource name"`
}

type ChartValicationInfo struct {
	Name         string                 `json:"name" description:"name of the release"`
	ConfigValues map[string]interface{} `json:"configvalues" description:"extra values added to the chart"`
	Version      int32                  `json:"version" description:"version of the release"`
	Namespace    string                 `json:"namespace" description:"namespace of release"`
	Dependencies map[string]string      `json:"dependencies" description:"map of dependency chart name and release"`
	ChartName    string                 `json:"chartname" description:"chart name"`
	ChartVersion string                 `json:"chartversion" description:"chart version"`
	RenderStatus string                 `json:"render_status" description:"status of rending "`
	RenderResult map[string]string      `json:"render_result" description:"result of rending "`
	DryRunStatus string                 `json:"dryrun_status" description:"status of dry run "`
	DryRunResult map[string]string      `json:"dryrun_result" description:"result of dry run "`
	ErrorMessage string                 `json:"error_message" description:" error msg "`
}

type ReleaseRequest struct {
	Name         string                 `json:"name" description:"name of the release"`
	RepoName     string                 `json:"repo_name" description:"chart name"`
	ChartName    string                 `json:"chart_name" description:"chart name"`
	ChartVersion string                 `json:"chart_version" description:"chart repo"`
	ConfigValues map[string]interface{} `json:"config_values" description:"extra values added to the chart"`
	Dependencies map[string]string      `json:"dependencies" description:"map of dependency chart name and release"`
	//ChartURL string
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
	CommonValues map[string]interface{} `json:"common_values" description:"common values added to the chart"`
	Releases     []*ReleaseRequest      `json:"releases" description:"list of release of the project"`
}

type ProjectInfo struct {
	ProjectCache
	Releases         []*ReleaseInfo  `json:"releases" description:"list of release of the project"`
	Ready            bool            `json:"ready" description:"whether all the project releases are ready"`
}

type ProjectCache struct {
	Name             string          `json:"name" description:"project name"`
	Namespace        string          `json:"namespace" description:"project namespace"`
	LatestProjectJobState ProjectJobState `json:"latest_project_job_state" description:"latest project job state"`
}

func(projectCache *ProjectCache) IsProjectJobNotFinished() bool {
	return projectCache.LatestProjectJobState.Status == "Running" || projectCache.LatestProjectJobState.Status == "Pending"
}

type ProjectJobState struct {
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
	TenantRepoName string `json:"repo_name"`
	TenantRepoURL  string `json:"repo_url"`
}

type RepoInfoList struct {
	Items []*RepoInfo `json:"items" description:"chart repo list"`
}

type ChartInfo struct {
	ChartName        string   `json:"chart_name"`
	ChartVersion     string   `json:"chart_version"`
	ChartDescription string   `json:"chart_description"`
	ChartAppVersion  string   `json:"chart_appversion"`
	ChartEngine      string   `json:"chart_engine"`
	DefaultValue     string   `json:"default_value" description:"default values.yaml defined by the chart"`
	DependencyCharts []string `json:"dependency_charts" description:"dependency chart name"`
}

type ChartInfoList struct {
	Items []*ChartInfo `json:"items" description:"chart list"`
}
