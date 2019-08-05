package release

import "WarpCloud/walm/pkg/models/task"

type ReleaseTask struct {
	Name                 string        `json:"name" description:"release name"`
	Namespace            string        `json:"namespace" description:"release namespace"`
	LatestReleaseTaskSig *task.TaskSig `json:"latestReleaseTaskSignature" description:"latest release task signature"`
}
