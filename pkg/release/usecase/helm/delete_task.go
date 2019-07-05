package helm

import (
	"github.com/sirupsen/logrus"
	"encoding/json"
)

const (
	deleteReleaseTaskName = "Delete-Release-Task"
)

type DeleteReleaseTaskArgs struct {
	Namespace   string
	ReleaseName string
	DeletePvcs  bool
}

func (helm *Helm)registerDeleteReleaseTask() error{
	return helm.task.RegisterTask(deleteReleaseTaskName, helm.deleteReleaseTask)
}

func (helm *Helm)deleteReleaseTask(releaseTaskArgsStr string) error {
	releaseTaskArgs := &DeleteReleaseTaskArgs{}
	err := json.Unmarshal([]byte(releaseTaskArgsStr), releaseTaskArgs)
	if err != nil {
		logrus.Errorf("%s args is not valid : %s", deleteReleaseTaskName, err.Error())
		return err
	}
	err = helm.doDeleteRelease(releaseTaskArgs.Namespace, releaseTaskArgs.ReleaseName, releaseTaskArgs.DeletePvcs)
	if err != nil {
		logrus.Errorf("failed to delete release %s/%s: %s", releaseTaskArgs.Namespace, releaseTaskArgs.ReleaseName, err.Error())
		return err
	}

	err = helm.releaseCache.DeleteReleaseTask(releaseTaskArgs.Namespace, releaseTaskArgs.ReleaseName)
	if err != nil {
		logrus.Warnf("failed to delete release task %s/%s : %s", releaseTaskArgs.Namespace, releaseTaskArgs.ReleaseName, err.Error())
	}

	return nil
}