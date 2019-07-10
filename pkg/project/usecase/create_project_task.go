package usecase

import (
	"github.com/sirupsen/logrus"
	"encoding/json"
	"WarpCloud/walm/pkg/release"
	"WarpCloud/walm/pkg/release/manager/helm/cache"
	"WarpCloud/walm/pkg/models/project"
	"WarpCloud/walm/pkg/util"
)

const (
	createProjectTaskName = "Create-Project-Task"
)

type CreateProjectTaskArgs struct {
	Namespace     string
	Name          string
	ProjectParams *project.ProjectParams
}

func (projectImpl *Project) registerCreateProjectTask() {
	projectImpl.task.RegisterTask(createProjectTaskName, projectImpl.CreateProjectTask)
}

func (projectImpl *Project)CreateProjectTask(createProjectTaskArgsStr string) error {
	createProjectTaskArgs := &CreateProjectTaskArgs{}
	err := json.Unmarshal([]byte(createProjectTaskArgsStr), createProjectTaskArgs)
	if err != nil {
		logrus.Errorf("create project task arg is not valid : %s", err.Error())
		return err
	}
	err = projectImpl.doCreateProject(createProjectTaskArgs.Namespace, createProjectTaskArgs.Name, createProjectTaskArgs.ProjectParams)
	if err != nil {
		logrus.Errorf("failed to create project %s/%s : %s", createProjectTaskArgs.Namespace, createProjectTaskArgs.Name, err.Error())
		return err
	}
	return nil
}

func (projectImpl *Project) doCreateProject(namespace string, name string, projectParams *project.ProjectParams) error {
	helmExtraLabelsBase := map[string]interface{}{}
	helmExtraLabelsVals := release.HelmExtraLabels{}
	helmExtraLabelsVals.HelmLabels = make(map[string]interface{})
	helmExtraLabelsVals.HelmLabels["project_name"] = name
	helmExtraLabelsBase["HelmExtraLabels"] = helmExtraLabelsVals

	rawValsBase := map[string]interface{}{}
	rawValsBase = util.MergeValues(rawValsBase, projectParams.CommonValues, false)
	rawValsBase = util.MergeValues(helmExtraLabelsBase, rawValsBase, false)

	for _, releaseParams := range projectParams.Releases {
		releaseParams.ConfigValues = util.MergeValues(releaseParams.ConfigValues, rawValsBase, false)
		if releaseParams.ReleaseLabels == nil {
			releaseParams.ReleaseLabels = map[string]string{}
		}
		releaseParams.ReleaseLabels[cache.ProjectNameLabelKey] = name
	}

	releaseList, err := projectImpl.autoCreateReleaseDependencies(projectParams)
	if err != nil {
		logrus.Errorf("failed to parse project charts dependency relation  : %s", err.Error())
		return err
	}
	for _, releaseParams := range releaseList {
		err = projectImpl.releaseUseCase.InstallUpgradeReleaseWithRetry(namespace, releaseParams,  nil, false, 0, nil)
		if err != nil {
			logrus.Errorf("failed to create project release %s/%s : %s", namespace, releaseParams.Name, err)
			return err
		}
		logrus.Debugf("succeed to create project release %s/%s", namespace, releaseParams.Name)
	}
	return nil
}