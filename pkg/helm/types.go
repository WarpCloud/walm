package helm

type ReleaseInfo struct {
	Name            string                 `json:"name" description:"name of the release"`
	ConfigValues    map[string]interface{} `json:"configvalues" description:"extra values added to the chart"`
	Version         int32                  `json:"version" description:"version of the release"`
	Namespace       string                 `json:"namespace" description:"namespace of release"`
	Statuscode      int32                  `json:"statuscode" description:"status of release"`
	Dependencies    map[string]string      `json:"dependencies" description:"map of dependency chart name and release"`
	ChartName       string                 `json:"chartname" description:"chart name"`
	ChartVersion    string                 `json:"chartversion" description:"chart version"`
	ChartAppVersion string                 `json:"chartappversion" description:"jsonnet app version"`
}

type ReleaseRequest struct {
	Name         string                 `json:"name" description:"name of the release"`
	Namespace    string                 `json:"namespace" description:"namespace of release"`
	ChartName    string                 `json:"chartname" description:"chart name"`
	ChartVersion string                 `json:"chartversion" description:"chart repo"`
	ConfigValues map[string]interface{} `json:"configvalues" description:"extra values added to the chart"`
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
	Name string `json:"name,omitempty"`
	Dependencies []*DependencyDeclare `json:"dependencies"`
}

