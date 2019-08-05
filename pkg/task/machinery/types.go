package machinery

import (
	"github.com/RichardKnop/machinery/v1/tasks"
	"time"
)

type TaskStateAdaptor struct {
	taskState *tasks.TaskState
	taskTimeoutSec int64
}

func (adaptor *TaskStateAdaptor) IsFinished() bool {
	return adaptor.taskState.IsCompleted()
}

func (adaptor *TaskStateAdaptor) IsSuccess() bool {
	return adaptor.taskState.IsSuccess()
}

func (adaptor *TaskStateAdaptor) GetErrorMsg() string {
	return adaptor.taskState.Error
}

func (adaptor *TaskStateAdaptor) IsTimeout() bool {
	return time.Now().Sub(adaptor.taskState.CreatedAt) > time.Duration(adaptor.taskTimeoutSec)*time.Second
}