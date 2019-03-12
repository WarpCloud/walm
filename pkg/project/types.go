package project

import (
	"walm/pkg/release/manager/helm/cache"
	"walm/pkg/task"
	"walm/pkg/release"
)

type ProjectParams struct {
	CommonValues map[string]interface{} `json:"commonValues" description:"common values added to the chart"`
	Releases     []*release.ReleaseRequestV2      `json:"releases" description:"list of release of the project"`
}

type ProjectInfo struct {
	cache.ProjectCache
	Releases        []*release.ReleaseInfoV2 `json:"releases" description:"list of release of the project"`
	Ready           bool                `json:"ready" description:"whether all the project releases are ready"`
	Message         string              `json:"message" description:"why project is not ready"`
	LatestTaskState *task.WalmTaskState      `json:"latestTaskState" description:"latest task state"`
}

type ProjectInfoList struct {
	Num   int            `json:"num" description:"project number"`
	Items []*ProjectInfo `json:"items" description:"project info list"`
}
