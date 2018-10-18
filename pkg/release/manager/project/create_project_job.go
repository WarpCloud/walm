package project

import (
	"walm/pkg/release"
	"walm/pkg/job"
	"github.com/sirupsen/logrus"
)

const (
	createProjectJobType = "CreateProject"
)

func init() {
	job.RegisterJobType(createProjectJobType, func() job.Job {
		return &CreateProjectJob{}
	})
}

type CreateProjectJob struct {
	Namespace     string
	Name          string
	ProjectParams *release.ProjectParams
}

func (createProjectJob *CreateProjectJob)Type() string {
	return createProjectJobType
}

func (createProjectJob *CreateProjectJob) createProject() error {
	helmExtraLabelsBase := map[string]interface{}{}
	helmExtraLabelsVals := release.HelmExtraLabels{}
	helmExtraLabelsVals.HelmLabels = make(map[string]interface{})
	helmExtraLabelsVals.HelmLabels["project_name"] = createProjectJob.Name
	helmExtraLabelsBase["HelmExtraLabels"] = helmExtraLabelsVals

	rawValsBase := map[string]interface{}{}
	rawValsBase = mergeValues(rawValsBase, createProjectJob.ProjectParams.CommonValues)
	rawValsBase = mergeValues(helmExtraLabelsBase, rawValsBase)

	for _, releaseParams := range createProjectJob.ProjectParams.Releases {
		releaseParams.Name = buildProjectReleaseName(createProjectJob.Name, releaseParams.Name)
		releaseParams.ConfigValues = mergeValues(releaseParams.ConfigValues, rawValsBase)
	}

	releaseList, err := GetDefaultProjectManager().brainFuckChartDepParse(createProjectJob.ProjectParams)
	if err != nil {
		logrus.Errorf("failed to parse project charts dependency relation  : %s", err.Error())
		return err
	}
	for _, releaseParams := range releaseList {
		err = GetDefaultProjectManager().helmClient.InstallUpgradeRealese(createProjectJob.Namespace, releaseParams)
		if err != nil {
			logrus.Errorf("failed to create project release %s/%s : %s", createProjectJob.Namespace, releaseParams.Name, err)
			return err
		}
		logrus.Debugf("succeed to create project release %s/%s", createProjectJob.Namespace, releaseParams.Name)
	}
	return nil
}

func (createProjectJob *CreateProjectJob) Do() error {
	logrus.Debugf("start to create project %s/%s", createProjectJob.Namespace, createProjectJob.Name)

	projectCache := buildProjectCache(createProjectJob.Namespace, createProjectJob.Name, createProjectJob.Type(), "Running")
	setProjectCacheToRedisUntilSuccess(projectCache)

	err := createProjectJob.createProject()
	if err != nil {
		logrus.Errorf("failed to create project %s/%s : %s", createProjectJob.Namespace, createProjectJob.Name, err.Error())
		projectCache.LatestProjectJobState.Status = "Failed"
		projectCache.LatestProjectJobState.Message = err.Error()
		setProjectCacheToRedisUntilSuccess(projectCache)
		return err
	}

	logrus.Infof("succeed to create project %s/%s", createProjectJob.Namespace, createProjectJob.Name)
	projectCache.LatestProjectJobState.Status = "Succeed"
	setProjectCacheToRedisUntilSuccess(projectCache)
	return nil
}
