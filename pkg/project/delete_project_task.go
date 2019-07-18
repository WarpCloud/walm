package project

import (
	"github.com/sirupsen/logrus"
	"encoding/json"
	"WarpCloud/walm/pkg/task"
	"github.com/RichardKnop/machinery/v1/tasks"
	"WarpCloud/walm/pkg/release/manager/helm/cache"
)

const (
	deleteProjectTaskName = "Delete-Project-Task"
)

func init() {
	task.RegisterTasks(deleteProjectTaskName, DeleteProjectTask)
}

func DeleteProjectTask(deleteProjectTaskArgsStr string) error {
	deleteProjectTaskArgs := &DeleteProjectTaskArgs{}
	err := json.Unmarshal([]byte(deleteProjectTaskArgsStr), deleteProjectTaskArgs)
	if err != nil {
		logrus.Errorf("delete project task arg is not valid : %s", err.Error())
		return err
	}
	return deleteProjectTaskArgs.deleteProject()
}

func SendDeleteProjectTask(deleteProjectTaskArgs *DeleteProjectTaskArgs) (*cache.ProjectTaskSignature, error) {
	deleteProjectTaskArgsStr, err := json.Marshal(deleteProjectTaskArgs)
	if err != nil {
		logrus.Errorf("failed to marshal delete project job : %s", err.Error())
		return nil, err
	}
	deleteProjectTaskSig := &tasks.Signature{
		Name: deleteProjectTaskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: string(deleteProjectTaskArgsStr),
			},
		},
	}
	err = task.GetDefaultTaskManager().SendTask(deleteProjectTaskSig)
	if err != nil {
		logrus.Errorf("failed to send delete project task : %s", err.Error())
		return nil, err
	}

	return  &cache.ProjectTaskSignature{
		Name: deleteProjectTaskName,
		UUID: deleteProjectTaskSig.UUID,
		Arg:  string(deleteProjectTaskArgsStr),
	}, nil
}

type DeleteProjectTaskArgs struct {
	Namespace     string
	Name          string
	DeletePvcs    bool
}

func (deleteProjectTaskArgs *DeleteProjectTaskArgs) deleteProject() error {
	projectInfo, err := GetDefaultProjectManager().GetProjectInfo(deleteProjectTaskArgs.Namespace, deleteProjectTaskArgs.Name)
	if err != nil {
		logrus.Errorf("failed to get project info : %s", err.Error())
		return err
	}

	for _, releaseInfo := range projectInfo.Releases {
		err = GetDefaultProjectManager().helmClient.DeleteReleaseWithRetry(deleteProjectTaskArgs.Namespace, releaseInfo.Name, false, deleteProjectTaskArgs.DeletePvcs, false, 0)
		if err != nil {
			logrus.Errorf("failed to delete release %s : %s", releaseInfo.Name, err.Error())
			return err
		}
	}

	err = GetDefaultProjectManager().helmClient.GetHelmCache().DeleteProjectCache(deleteProjectTaskArgs.Namespace, deleteProjectTaskArgs.Name)
	if err != nil {
		logrus.Warnf("failed to delete project task of %s : %s", deleteProjectTaskArgs.Name, err.Error())
	}

	return nil
}
