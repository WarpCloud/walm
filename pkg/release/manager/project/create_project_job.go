package project

import (
	"walm/pkg/release"
	"walm/pkg/job"
	"fmt"
	"github.com/sirupsen/logrus"
	"walm/pkg/redis"
	"time"
)

func init() {
	job.RegisterJobType("CreateProject", &CreateProjectJob{})
}

type CreateProjectJob struct {
	Namespace     string
	Name          string
	ProjectParams *release.ProjectParams
}

func (createProjectJob *CreateProjectJob) createProject(projectCache *release.ProjectCache) error {
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
		projectCache.InstalledReleases = append(projectCache.InstalledReleases, releaseParams.Name)
		setProjectCacheToRedisUntilSuccess(projectCache, createProjectJob)
		logrus.Debugf("succeed to create project release %s/%s", createProjectJob.Namespace, releaseParams.Name)
	}
	return nil
}

func (createProjectJob *CreateProjectJob) Do() error {
	logrus.Debugf("start to create project %s/%s", createProjectJob.Namespace, createProjectJob.Name)

	projectCache := buildProjectCache(createProjectJob.Namespace, createProjectJob.Name, "Running", createProjectJob.ProjectParams)
	setProjectCacheToRedisUntilSuccess(projectCache, createProjectJob)

	err := createProjectJob.createProject(projectCache)
	if err != nil {
		logrus.Errorf("failed to create project %s/%s : %s", createProjectJob.Namespace, createProjectJob.Name, err.Error())
		projectCache.CreateProjectJobState.CreateProjectJobStatus = "Failed"
		projectCache.CreateProjectJobState.Message = err.Error()
		setProjectCacheToRedisUntilSuccess(projectCache, createProjectJob)
		return err
	}

	logrus.Infof("succeed to create project %s/%s", createProjectJob.Namespace, createProjectJob.Name)
	projectCache.CreateProjectJobState.CreateProjectJobStatus = "Succeed"
	setProjectCacheToRedisUntilSuccess(projectCache, createProjectJob)
	return nil
}

func setProjectCacheToRedisUntilSuccess(projectCache *release.ProjectCache, createProjectJob *CreateProjectJob) {
	for {
		err := setProjectCacheToRedis(redis.GetDefaultRedisClient(), projectCache)
		if err != nil {
			logrus.Errorf("failed to set project cache of %s/%s to redis: %s", createProjectJob.Namespace, createProjectJob.Name, err.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
}

func buildProjectReleaseName(projectName, releaseName string) string {
	return fmt.Sprintf("%s--%s", projectName, releaseName)
}
