package project

import (
	"github.com/sirupsen/logrus"
	"encoding/json"
	"walm/pkg/task"
	"github.com/RichardKnop/machinery/v1/tasks"
	"fmt"
	"walm/pkg/release/manager/helm/cache"
)

const (
	removeReleaseTaskName = "Remove-Release-Task"
)

func init() {
	task.RegisterTasks(removeReleaseTaskName, RemoveReleaseTask)
}

func RemoveReleaseTask(removeReleaseTaskArgsStr string) error {
	removeReleaseTaskArgs := &RemoveReleaseTaskArgs{}
	err := json.Unmarshal([]byte(removeReleaseTaskArgsStr), removeReleaseTaskArgs)
	if err != nil {
		logrus.Errorf("remove release task arg is not valid : %s", err.Error())
		return err
	}
	return removeReleaseTaskArgs.removeRelease()
}

func SendRemoveReleaseTask(removeReleaseTaskArgs *RemoveReleaseTaskArgs) (*cache.ProjectTaskSignature, error) {
	removeReleaseTaskArgsStr, err := json.Marshal(removeReleaseTaskArgs)
	if err != nil {
		logrus.Errorf("failed to marshal remove release task args: %s", err.Error())
		return nil, err
	}
	removeReleaseTaskSig := &tasks.Signature{
		Name: removeReleaseTaskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: string(removeReleaseTaskArgsStr),
			},
		},
	}
	err = task.GetDefaultTaskManager().SendTask(removeReleaseTaskSig)
	if err != nil {
		logrus.Errorf("failed to send remove release task : %s", err.Error())
		return nil, err
	}
	return  &cache.ProjectTaskSignature{
		Name: removeReleaseTaskName,
		UUID: removeReleaseTaskSig.UUID,
		Arg:  string(removeReleaseTaskArgsStr),
	}, nil
}

type RemoveReleaseTaskArgs struct {
	Namespace     string
	Name          string
	ReleaseName   string
	DeletePvcs    bool
}

func (removeReleaseTaskArgs *RemoveReleaseTaskArgs) removeRelease() error {
	projectInfo, err := GetDefaultProjectManager().GetProjectInfo(removeReleaseTaskArgs.Namespace, removeReleaseTaskArgs.Name)
	if err != nil {
		logrus.Errorf("failed to get project info : %s", err.Error())
		return err
	}

	releaseParams := buildReleaseRequest(projectInfo, removeReleaseTaskArgs.ReleaseName)
	if releaseParams == nil {
		return fmt.Errorf("release is %s not found in project %s", removeReleaseTaskArgs.ReleaseName, removeReleaseTaskArgs.ReleaseName)
	}
	if projectInfo != nil {
		affectReleaseRequest, err2 := GetDefaultProjectManager().brainFuckRuntimeDepParse(projectInfo, releaseParams, true)
		if err2 != nil {
			logrus.Errorf("RuntimeDepParse install release %s error %v\n", releaseParams.Name, err)
			return err2
		}
		for _, affectReleaseParams := range affectReleaseRequest {
			logrus.Infof("Update BecauseOf Dependency Modified: %v", *affectReleaseParams)
			err = GetDefaultProjectManager().helmClient.InstallUpgradeReleaseWithRetry(removeReleaseTaskArgs.Namespace, affectReleaseParams, false, nil, false, 0)
			if err != nil {
				logrus.Errorf("RemoveReleaseInProject Other Affected Release install release %s error %v\n", releaseParams.Name, err)
				return err
			}
		}
	}

	err = GetDefaultProjectManager().helmClient.DeleteReleaseWithRetry(removeReleaseTaskArgs.Namespace, removeReleaseTaskArgs.ReleaseName, false, removeReleaseTaskArgs.DeletePvcs, false, 0)
	if err != nil {
		logrus.Errorf("RemoveReleaseInProject install release %s error %v\n", removeReleaseTaskArgs.ReleaseName, err)
		return err
	}
	return nil
}
