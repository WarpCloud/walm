package project

import (
	"github.com/sirupsen/logrus"
	"encoding/json"
	"walm/pkg/task"
	"github.com/RichardKnop/machinery/v1/tasks"
	"walm/pkg/release"
)

const (
	createProjectTaskName = "Create-Project-Task"
)

func init() {
	task.RegisterTasks(createProjectTaskName, CreateProjectTask)
}

func CreateProjectTask(createProjectTaskArgsStr string) error {
	createProjectTaskArgs := &CreateProjectTaskArgs{}
	err := json.Unmarshal([]byte(createProjectTaskArgsStr), createProjectTaskArgs)
	if err != nil {
		logrus.Errorf("create project task arg is not valid : %s", err.Error())
		return err
	}
	return createProjectTaskArgs.createProject()
}

func SendCreateProjectTask(createProjectTaskArgs *CreateProjectTaskArgs) (*release.ProjectTaskSignature, error) {
	createProjectTaskArgsStr, err := json.Marshal(createProjectTaskArgs)
	if err != nil {
		logrus.Errorf("failed to marshal create project job : %s", err.Error())
		return nil, err
	}
	createProjectTaskSig := &tasks.Signature{
		Name: createProjectTaskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: string(createProjectTaskArgsStr),
			},
		},
	}
	err = task.GetDefaultTaskManager().SendTask(createProjectTaskSig)
	if err != nil {
		logrus.Errorf("failed to send create project task : %s", err.Error())
		return nil, err
	}
	return  &release.ProjectTaskSignature{
		Name: createProjectTaskName,
		UUID: createProjectTaskSig.UUID,
		Arg:  string(createProjectTaskArgsStr),
	}, nil
}

type CreateProjectTaskArgs struct {
	Namespace     string
	Name          string
	ProjectParams *release.ProjectParams
}

func (createProjectTaskArgs *CreateProjectTaskArgs) createProject() error {
	helmExtraLabelsBase := map[string]interface{}{}
	helmExtraLabelsVals := release.HelmExtraLabels{}
	helmExtraLabelsVals.HelmLabels = make(map[string]interface{})
	helmExtraLabelsVals.HelmLabels["project_name"] = createProjectTaskArgs.Name
	helmExtraLabelsBase["HelmExtraLabels"] = helmExtraLabelsVals

	rawValsBase := map[string]interface{}{}
	rawValsBase = mergeValues(rawValsBase, createProjectTaskArgs.ProjectParams.CommonValues)
	rawValsBase = mergeValues(helmExtraLabelsBase, rawValsBase)

	for _, releaseParams := range createProjectTaskArgs.ProjectParams.Releases {
		releaseParams.Name = buildProjectReleaseName(createProjectTaskArgs.Name, releaseParams.Name)
		releaseParams.ConfigValues = mergeValues(releaseParams.ConfigValues, rawValsBase)
	}

	releaseList, err := GetDefaultProjectManager().brainFuckChartDepParse(createProjectTaskArgs.ProjectParams)
	if err != nil {
		logrus.Errorf("failed to parse project charts dependency relation  : %s", err.Error())
		return err
	}
	for _, releaseParams := range releaseList {
		err = GetDefaultProjectManager().helmClient.InstallUpgradeRealese(createProjectTaskArgs.Namespace, releaseParams, false)
		if err != nil {
			logrus.Errorf("failed to create project release %s/%s : %s", createProjectTaskArgs.Namespace, releaseParams.Name, err)
			return err
		}
		logrus.Debugf("succeed to create project release %s/%s", createProjectTaskArgs.Namespace, releaseParams.Name)
	}
	return nil
}
