package release

type RepoInfo struct {
	TenantRepoName string `json:"repoName"`
	TenantRepoURL  string `json:"repoUrl"`
}

type RepoInfoList struct {
	Items []*RepoInfo `json:"items" description:"chart repo list"`
}

type ChartInfo struct {
	ChartName        string                  `json:"chartName"`
	ChartVersion     string                  `json:"chartVersion"`
	ChartDescription string                  `json:"chartDescription"`
	ChartAppVersion  string                  `json:"chartAppVersion"`
	ChartEngine      string                  `json:"chartEngine"`
	DefaultValue     string                  `json:"defaultValue" description:"default values.yaml defined by the chart"`
	MetaInfo         *ChartMetaInfo `json:"metaInfo" description:"transwarp chart meta info"`
}

type ChartDetailInfo struct {
	ChartInfo
	// additional info
	Advantage    string `json:"advantage" description:"chart production advantage description(rich text)"`
	Architecture string `json:"architecture" description:"chart production architecture description(rich text)"`
	Icon         string `json:"icon" description:"chart icon"`
}

type ChartInfoList struct {
	Items []*ChartInfo `json:"items" description:"chart list"`
}
