package project

import (
	"github.com/sirupsen/logrus"
	"encoding/json"
	"walm/pkg/task"
	"github.com/RichardKnop/machinery/v1/tasks"
	"walm/pkg/release"
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

func SendDeleteProjectTask(deleteProjectTaskArgs *DeleteProjectTaskArgs) (*release.ProjectTaskSignature, error) {
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
	return  &release.ProjectTaskSignature{
		Name: deleteProjectTaskName,
		UUID: deleteProjectTaskSig.UUID,
		Arg:  string(deleteProjectTaskArgsStr),
	}, nil
}

type DeleteProjectTaskArgs struct {
	Namespace     string
	Name          string
}

func (deleteProjectTaskArgs *DeleteProjectTaskArgs) deleteProject() error {
	projectInfo, err := GetDefaultProjectManager().GetProjectInfo(deleteProjectTaskArgs.Namespace, deleteProjectTaskArgs.Name)
	if err != nil {
		logrus.Errorf("failed to get project info : %s", err.Error())
		return err
	}

	for _, releaseInfo := range projectInfo.Releases {
		releaseName := buildProjectReleaseName(projectInfo.Name, releaseInfo.Name)
		err = GetDefaultProjectManager().helmClient.DeleteRelease(deleteProjectTaskArgs.Namespace, releaseName, false)
		if err != nil {
			logrus.Errorf("failed to delete release %s : %s", releaseName, err.Error())
			return err
		}
	}
	return nil
}
