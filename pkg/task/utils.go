package task

import (
	"github.com/RichardKnop/machinery/v1/tasks"
	"time"
	"github.com/sirupsen/logrus"
)

func IsTaskFinishedOrTimeout(taskState *tasks.TaskState, taskTimeoutSec int64) bool {
	// task state has ttl, maybe task state can not be got
	if taskState == nil || taskState.TaskName == "" {
		return true
	} else if taskState.IsCompleted() {
		return true
	} else if time.Now().Sub(taskState.CreatedAt) > time.Duration(taskTimeoutSec)*time.Second {
		logrus.Warnf("task %s-%s time out", taskState.TaskName, taskState.TaskUUID)
		return true
	}
	return false
}
