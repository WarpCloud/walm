package helm

import (
	"github.com/sirupsen/logrus"
	"encoding/json"
	"walm/pkg/task"
)

const (
	deleteReleaseTaskName = "Delete-Release-Task"
)

func init() {
	task.RegisterTasks(deleteReleaseTaskName, deleteReleaseTask)
}

func deleteReleaseTask(releaseTaskArgsStr string) error {
	releaseTaskArgs := &DeleteReleaseTaskArgs{}
	err := json.Unmarshal([]byte(releaseTaskArgsStr), releaseTaskArgs)
	if err != nil {
		logrus.Errorf("%s args is not valid : %s", releaseTaskArgs.GetTaskName(), err.Error())
		return err
	}
	return releaseTaskArgs.Run()
}

type DeleteReleaseTaskArgs struct {
	Namespace   string
	ReleaseName string
	IsSystem    bool
	DeletePvcs  bool
}

func (deleteReleaseTaskArgs *DeleteReleaseTaskArgs) Run() error {
	return GetDefaultHelmClientV2().doDeleteReleaseV2(deleteReleaseTaskArgs.Namespace, deleteReleaseTaskArgs.ReleaseName, deleteReleaseTaskArgs.IsSystem, deleteReleaseTaskArgs.DeletePvcs)
}

func (deleteReleaseTaskArgs *DeleteReleaseTaskArgs) GetTaskName() string {
	return deleteReleaseTaskName
}
