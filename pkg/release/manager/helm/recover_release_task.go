package helm

import (
	"github.com/sirupsen/logrus"
	"encoding/json"
	"walm/pkg/task"
)

const (
	recoverReleaseTaskName = "Recover-Release-Task"
)

func init() {
	task.RegisterTasks(recoverReleaseTaskName, recoverReleaseTask)
}

func recoverReleaseTask(releaseTaskArgsStr string) error {
	releaseTaskArgs := &RecoverReleaseTaskArgs{}
	err := json.Unmarshal([]byte(releaseTaskArgsStr), releaseTaskArgs)
	if err != nil {
		logrus.Errorf("%s args is not valid : %s", releaseTaskArgs.GetTaskName(), err.Error())
		return err
	}
	return releaseTaskArgs.Run()
}

type RecoverReleaseTaskArgs struct {
	Namespace   string
	ReleaseName string
}

func (recoverReleaseTaskArgs *RecoverReleaseTaskArgs) Run() error {
	return GetDefaultHelmClient().doRecoverRelease(recoverReleaseTaskArgs.Namespace, recoverReleaseTaskArgs.ReleaseName)
}

func (recoverReleaseTaskArgs *RecoverReleaseTaskArgs) GetTaskName() string {
	return recoverReleaseTaskName
}
