package helm

import (
	"walm/pkg/task"
	"github.com/sirupsen/logrus"
	"encoding/json"
	"github.com/RichardKnop/machinery/v1/tasks"
)

type ReleaseTaskArgs interface {
	Run() error
	GetTaskName() string
}

func SendReleaseTask(releaseTaskArgs ReleaseTaskArgs) (*task.WalmTaskSig, error) {
	releaseTaskArgsStr, err := json.Marshal(releaseTaskArgs)
	if err != nil {
		logrus.Errorf("failed to marshal %s args : %s", releaseTaskArgs.GetTaskName(), err.Error())
		return nil, err
	}
	releaseTaskSig := &tasks.Signature{
		Name: releaseTaskArgs.GetTaskName(),
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: string(releaseTaskArgsStr),
			},
		},
	}
	err = task.GetDefaultTaskManager().SendTask(releaseTaskSig)
	if err != nil {
		logrus.Errorf("failed to send %s : %s", releaseTaskArgs.GetTaskName(), err.Error())
		return nil, err
	}

	return &task.WalmTaskSig{
		Name: releaseTaskArgs.GetTaskName(),
		UUID: releaseTaskSig.UUID,
		Arg:  string(releaseTaskArgsStr),
	}, nil
}