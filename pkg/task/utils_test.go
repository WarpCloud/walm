package task

import (
	"testing"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/stretchr/testify/assert"
	"time"
)

func Test_IsTaskFinishedOrTimeout(t *testing.T) {
	tests := []struct {
		taskState      *tasks.TaskState
		taskTimeoutSec int64
		result         bool
	}{
		{
			taskState: nil,
			taskTimeoutSec: 60,
			result: true,
		},
		{
			taskState: &tasks.TaskState{
				TaskName: "",
			},
			taskTimeoutSec: 60,
			result: true,
		},
		{
			taskState: &tasks.TaskState{
				TaskName: "test_name",
				TaskUUID: "test_uuid",
				State:    tasks.StateSuccess,
			},
			taskTimeoutSec: 60,
			result: true,
		},
		{
			taskState: &tasks.TaskState{
				TaskName: "test_name",
				TaskUUID: "test_uuid",
				State:    tasks.StateFailure,
			},
			taskTimeoutSec: 60,
			result: true,
		},
		{
			taskState: &tasks.TaskState{
				TaskName:  "test_name",
				TaskUUID: "test_uuid",
				State:     tasks.StatePending,
				CreatedAt: time.Now().Add(time.Second * -70),
			},
			taskTimeoutSec: 60,
			result: true,
		},
		{
			taskState: &tasks.TaskState{
				TaskName:  "test_name",
				TaskUUID: "test_uuid",
				State:     tasks.StatePending,
				CreatedAt: time.Now().Add(time.Second * -50),
			},
			taskTimeoutSec: 60,
			result: false,
		},
	}

	for _, test := range tests {
		result := IsTaskFinishedOrTimeout(test.taskState, test.taskTimeoutSec)
		assert.Equal(t, test.result, result)
	}
}
