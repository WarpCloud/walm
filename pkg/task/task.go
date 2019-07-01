package task

import (
	"WarpCloud/walm/pkg/models/task"
)

type Task interface {
	RegisterTask(taskName string, task func(taskArgs string) error) error
	SendTask(taskName, taskArgs string, timeoutSec int64) (*task.TaskSig, error)
	GetTaskState(sig *task.TaskSig) (TaskState, error)
	TouchTask(sig *task.TaskSig, pollingIntervalSec int64) (error)
	PurgeTaskState(sig *task.TaskSig) (error)
}

type TaskState interface {
	IsFinished() bool
	IsSuccess() bool
	GetErrorMsg() string
	IsTimeout() bool
}
