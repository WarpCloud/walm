package project

import (
	"WarpCloud/walm/pkg/models/release"
	"WarpCloud/walm/pkg/models/task"
)

const (
	ProjectNameLabelKey = "Project-Name"
)

type ProjectParams struct {
	CommonValues map[string]interface{}      `json:"commonValues" description:"common values added to the chart"`
	Releases     []*release.ReleaseRequestV2 `json:"releases" description:"list of release of the project"`
}

type ProjectInfo struct {
	Name      string                   `json:"name" description:"project name"`
	Namespace string                   `json:"namespace" description:"project namespace"`
	Releases  []*release.ReleaseInfoV2 `json:"releases" description:"list of release of the project"`
	Ready     bool                     `json:"ready" description:"whether all the project releases are ready"`
	Message   string                   `json:"message" description:"why project is not ready"`
}

type ProjectInfoList struct {
	Num   int            `json:"num" description:"project number"`
	Items []*ProjectInfo `json:"items" description:"project info list"`
}

type ProjectTask struct {
	Name                string        `json:"name" description:"project name"`
	Namespace           string        `json:"namespace" description:"project namespace"`
	LatestTaskSignature *task.TaskSig `json:"latestTaskSignature" description:"latest task signature"`
	// compatible
	LatestTaskTimeoutSec int64 `json:"latestTaskTimeoutSec" description:"latest task timeout sec"`
}
