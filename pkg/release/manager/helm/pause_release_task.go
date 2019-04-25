package helm

import (
	"github.com/sirupsen/logrus"
	"encoding/json"
	"walm/pkg/task"
)

const (
	pauseReleaseTaskName = "Pause-Release-Task"
)

func init() {
	task.RegisterTasks(pauseReleaseTaskName, pauseReleaseTask)
}

func pauseReleaseTask(releaseTaskArgsStr string) error {
	releaseTaskArgs := &PauseReleaseTaskArgs{}
	err := json.Unmarshal([]byte(releaseTaskArgsStr), releaseTaskArgs)
	if err != nil {
		logrus.Errorf("%s args is not valid : %s", releaseTaskArgs.GetTaskName(), err.Error())
		return err
	}
	return releaseTaskArgs.Run()
}

type PauseReleaseTaskArgs struct {
	Namespace   string
	ReleaseName string
}

func (pauseReleaseTaskArgs *PauseReleaseTaskArgs) Run() error {
	return GetDefaultHelmClient().doPauseRelease(pauseReleaseTaskArgs.Namespace, pauseReleaseTaskArgs.ReleaseName)
}

func (pauseReleaseTaskArgs *PauseReleaseTaskArgs) GetTaskName() string {
	return pauseReleaseTaskName
}
