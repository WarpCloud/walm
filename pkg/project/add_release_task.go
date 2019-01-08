package project

import (
	"github.com/sirupsen/logrus"
	"encoding/json"
	"walm/pkg/task"
	"github.com/RichardKnop/machinery/v1/tasks"
	walmerr "walm/pkg/util/error"
	"walm/pkg/release/manager/helm/cache"
)

const (
	addReleaseTaskName = "Add-Release-Task"
)

func init() {
	task.RegisterTasks(addReleaseTaskName, AddReleaseTask)
}

func AddReleaseTask(addReleaseTaskArgsStr string) error {
	addReleaseTaskArgs := &AddReleaseTaskArgs{}
	err := json.Unmarshal([]byte(addReleaseTaskArgsStr), addReleaseTaskArgs)
	if err != nil {
		logrus.Errorf("add release task arg is not valid : %s", err.Error())
		return err
	}
	return addReleaseTaskArgs.addRelease()
}

func SendAddReleaseTask(addReleaseTaskArgs *AddReleaseTaskArgs) (*cache.ProjectTaskSignature, error) {
	addReleaseTaskArgsStr, err := json.Marshal(addReleaseTaskArgs)
	if err != nil {
		logrus.Errorf("failed to marshal add release task args : %s", err.Error())
		return nil, err
	}
	addReleaseTaskSig := &tasks.Signature{
		Name: addReleaseTaskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: string(addReleaseTaskArgsStr),
			},
		},
	}
	err = task.GetDefaultTaskManager().SendTask(addReleaseTaskSig)
	if err != nil {
		logrus.Errorf("failed to send add release task : %s", err.Error())
		return nil, err
	}

	return &cache.ProjectTaskSignature{
		Name: addReleaseTaskName,
		UUID: addReleaseTaskSig.UUID,
		Arg:  string(addReleaseTaskArgsStr),
	}, nil
}

type AddReleaseTaskArgs struct {
	Namespace     string
	Name          string
	ProjectParams *ProjectParams
}

func (addReleaseTaskArgs *AddReleaseTaskArgs) addRelease() error {
	projectInfo, err := GetDefaultProjectManager().GetProjectInfo(addReleaseTaskArgs.Namespace, addReleaseTaskArgs.Name)
	projectExists := true
	if err != nil {
		if !walmerr.IsNotFoundError(err) {
			projectExists = false
		} else {
			logrus.Errorf("failed to get project info : %s", err.Error())
			return err
		}
	}

	for _, releaseParams := range addReleaseTaskArgs.ProjectParams.Releases {
		releaseParams.Name = buildProjectReleaseName(addReleaseTaskArgs.Name, releaseParams.Name)
		releaseParams.ConfigValues = mergeValues(releaseParams.ConfigValues, addReleaseTaskArgs.ProjectParams.CommonValues)
	}
	releaseList, err := GetDefaultProjectManager().brainFuckChartDepParse(addReleaseTaskArgs.ProjectParams)
	if err != nil {
		logrus.Errorf("failed to parse project charts dependency relation  : %s", err.Error())
		return err
	}

	for _, releaseParams := range releaseList {
		if projectExists {
			affectReleaseRequest, err2 := GetDefaultProjectManager().brainFuckRuntimeDepParse(projectInfo, releaseParams, false)
			if err2 != nil {
				logrus.Errorf("RuntimeDepParse install release %s error %v\n", releaseParams.Name, err)
				return err2
			}
			err = GetDefaultProjectManager().helmClient.InstallUpgradeReleaseV2(addReleaseTaskArgs.Namespace, releaseParams, false, nil)
			if err != nil {
				logrus.Errorf("AddReleaseInProject install release %s error %v\n", releaseParams.Name, err)
				return err
			}
			for _, affectReleaseParams := range affectReleaseRequest {
				logrus.Infof("Update BecauseOf Dependency Modified: %v", *affectReleaseParams)
				err = GetDefaultProjectManager().helmClient.InstallUpgradeReleaseV2(addReleaseTaskArgs.Namespace, affectReleaseParams, false, nil)
				if err != nil {
					logrus.Errorf("AddReleaseInProject Other Affected Release install release %s error %v\n", releaseParams.Name, err)
					return err
				}
			}
		} else {
			err = GetDefaultProjectManager().helmClient.InstallUpgradeReleaseV2(addReleaseTaskArgs.Namespace, releaseParams, false, nil)
			if err != nil {
				logrus.Errorf("AddReleaseInProject install release %s error %v\n", releaseParams.Name, err)
				return err
			}
		}
	}

	return nil
}
