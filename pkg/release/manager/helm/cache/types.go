package cache

import (
	"walm/pkg/task"
	"time"
	"github.com/sirupsen/logrus"
	"github.com/RichardKnop/machinery/v1/tasks"
)

type ReleaseTask struct {
	Name                 string                `json:"name" description:"release name"`
	Namespace            string                `json:"namespace" description:"release namespace"`
	LatestReleaseTaskSig *task.WalmTaskSig     `json:"latest_release_task_signature" description:"latest release task signature"`
}

type ProjectCache struct {
	Name                 string                `json:"name" description:"project name"`
	Namespace            string                `json:"namespace" description:"project namespace"`
	//TODO refactor to use WalmTaskSig
	LatestTaskSignature  *ProjectTaskSignature `json:"latest_task_signature" description:"latest task signature"`
	LatestTaskTimeoutSec int64                 `json:"latest_task_timeout_sec" description:"latest task timeout sec"`
}

type ProjectTaskSignature struct {
	UUID string `json:"uuid" description:"task uuid"`
	Name string `json:"name" description:"task name"`
	Arg  string `json:"arg" description:"task arg"`
}

func (projectCache *ProjectCache) GetLatestTaskSignature() *tasks.Signature {
	if projectCache.LatestTaskSignature == nil {
		return nil
	}
	return &tasks.Signature{
		Name: projectCache.LatestTaskSignature.Name,
		UUID: projectCache.LatestTaskSignature.UUID,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: projectCache.LatestTaskSignature.Arg,
			},
		},
	}
}

func (projectCache *ProjectCache) GetLatestTaskState() *tasks.TaskState {
	if projectCache.LatestTaskSignature == nil {
		return nil
	}
	return task.GetDefaultTaskManager().NewAsyncResult(projectCache.GetLatestTaskSignature()).GetState()
}

func (projectCache *ProjectCache) IsLatestTaskFinishedOrTimeout() bool {
	taskState := projectCache.GetLatestTaskState()
	// task state has ttl, maybe task state can not be got
	if taskState == nil || taskState.TaskName == "" {
		return true
	} else if taskState.IsCompleted() {
		return true
	} else if time.Now().Sub(taskState.CreatedAt) > time.Duration(projectCache.LatestTaskTimeoutSec)*time.Second {
		logrus.Warnf("task %s-%s time out", projectCache.LatestTaskSignature.Name, projectCache.LatestTaskSignature.UUID)
		return true
	}
	return false
}