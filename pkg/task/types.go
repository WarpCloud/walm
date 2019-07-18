package task

import (
	"time"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/RichardKnop/machinery/v1/backends/result"
)

type WalmTaskState struct {
	TaskUUID  string    `json:"taskUuid" description:"task uuid"`
	TaskName  string    `json:"taskName" description:"task name"`
	State     string    `json:"taskState" description:"task state"`
	Error     string    `json:"taskError" description:"task error"`
	CreatedAt time.Time `json:"createdAt" description:"task creation time"`
}

type WalmTaskSig struct {
	UUID       string `json:"uuid" description:"task uuid"`
	Name       string `json:"name" description:"task name"`
	Arg        string `json:"arg" description:"task arg"`
	TimeoutSec int64  `json:"timeout_sec" description:"task timeout(sec)"`
	taskSig *tasks.Signature `json: "-" description:"real task signature"`
	asyncResult *result.AsyncResult `json: "-" description:"task async result"`
}

func (walmTaskSig WalmTaskSig) GetTaskSignature() *tasks.Signature {
	if walmTaskSig.UUID == "" {
		return nil
	}
	if walmTaskSig.taskSig == nil {
		walmTaskSig.taskSig = &tasks.Signature{
			Name: walmTaskSig.Name,
			UUID: walmTaskSig.UUID,
			Args: []tasks.Arg{
				{
					Type:  "string",
					Value: walmTaskSig.Arg,
				},
			},
		}
	}
	return walmTaskSig.taskSig
}

func (walmTaskSig WalmTaskSig) GetAsyncResult() *result.AsyncResult {
	if walmTaskSig.GetTaskSignature() == nil {
		return nil
	}
	if walmTaskSig.asyncResult == nil {
		walmTaskSig.asyncResult = GetDefaultTaskManager().NewAsyncResult(walmTaskSig.GetTaskSignature())
	}
	return walmTaskSig.asyncResult
}

func (walmTaskSig WalmTaskSig) GetTaskState() *tasks.TaskState {
	asyncResult := walmTaskSig.GetAsyncResult()
	if asyncResult == nil {
		return nil
	}
	return asyncResult.GetState()
}

func (walmTaskSig WalmTaskSig) IsTaskFinishedOrTimeout() bool {
	taskState := walmTaskSig.GetTaskState()
	return IsTaskFinishedOrTimeout(taskState, walmTaskSig.TimeoutSec)
}
