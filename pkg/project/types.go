package project

import (
	"walm/pkg/release/v2"
	"walm/pkg/release/manager/helm/cache"
	"walm/pkg/task"
)

type ProjectParams struct {
	CommonValues map[string]interface{} `json:"common_values" description:"common values added to the chart"`
	Releases     []*v2.ReleaseRequestV2      `json:"releases" description:"list of release of the project"`
}

type ProjectInfo struct {
	cache.ProjectCache
	Releases        []*v2.ReleaseInfoV2 `json:"releases" description:"list of release of the project"`
	Ready           bool                `json:"ready" description:"whether all the project releases are ready"`
	Message         string              `json:"message" description:"why project is not ready"`
	LatestTaskState *task.WalmTaskState      `json:"latest_task_state" description:"latest task state"`
}

type ProjectInfoList struct {
	Num   int            `json:"num" description:"project number"`
	Items []*ProjectInfo `json:"items" description:"project info list"`
}
