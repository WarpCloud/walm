package project

import (
	"walm/pkg/release"
	"walm/pkg/job"
	"github.com/sirupsen/logrus"
	walmerr "walm/pkg/util/error"
)

const (
	addReleasesJobType = "AddReleasesInProject"
)

func init() {
	job.RegisterJobType(addReleasesJobType, func() job.Job {
		return &AddReleasesJob{}
	})
}

type AddReleasesJob struct {
	Async         bool
	Namespace     string
	Name          string
	ProjectParams *release.ProjectParams
}

func (addReleasesJob *AddReleasesJob) Type() string {
	return addReleasesJobType
}

func (addReleasesJob *AddReleasesJob) addReleases() error {
	projectInfo, err := GetDefaultProjectManager().GetProjectInfo(addReleasesJob.Namespace, addReleasesJob.Name)
	projectExists := true
	if err != nil {
		if !walmerr.IsNotFoundError(err) {
			projectExists = false
		} else {
			logrus.Errorf("failed to get project info : %s", err.Error())
			return err
		}
	}

	for _, releaseParams := range addReleasesJob.ProjectParams.Releases {
		releaseParams.Name = buildProjectReleaseName(addReleasesJob.Name, releaseParams.Name)
		releaseParams.ConfigValues = mergeValues(releaseParams.ConfigValues, addReleasesJob.ProjectParams.CommonValues)
	}
	releaseList, err := GetDefaultProjectManager().brainFuckChartDepParse(addReleasesJob.ProjectParams)
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
			err = GetDefaultProjectManager().helmClient.InstallUpgradeRealese(addReleasesJob.Namespace, releaseParams)
			if err != nil {
				logrus.Errorf("AddReleaseInProject install release %s error %v\n", releaseParams.Name, err)
				return err
			}
			for _, affectReleaseParams := range affectReleaseRequest {
				logrus.Infof("Update BecauseOf Dependency Modified: %v", *affectReleaseParams)
				err = GetDefaultProjectManager().helmClient.UpgradeRealese(addReleasesJob.Namespace, affectReleaseParams)
				if err != nil {
					logrus.Errorf("AddReleaseInProject Other Affected Release install release %s error %v\n", releaseParams.Name, err)
					return err
				}
			}
		} else {
			err = GetDefaultProjectManager().helmClient.InstallUpgradeRealese(addReleasesJob.Namespace, releaseParams)
			if err != nil {
				logrus.Errorf("AddReleaseInProject install release %s error %v\n", releaseParams.Name, err)
				return err
			}
		}
	}

	return nil
}

func (addReleasesJob *AddReleasesJob) Do() error {
	logrus.Debugf("start to add release in project %s/%s", addReleasesJob.Namespace, addReleasesJob.Name)

	projectCache := buildProjectCache(addReleasesJob.Namespace, addReleasesJob.Name, addReleasesJob.Type(), "Running", addReleasesJob.Async)
	setProjectCacheToRedisUntilSuccess(projectCache)

	err := addReleasesJob.addReleases()
	if err != nil {
		logrus.Errorf("failed to add release in project %s/%s : %s", addReleasesJob.Namespace, addReleasesJob.Name, err.Error())
		projectCache.LatestProjectJobState.Status = "Failed"
		projectCache.LatestProjectJobState.Message = err.Error()
		setProjectCacheToRedisUntilSuccess(projectCache)
		return err
	}

	logrus.Infof("succeed to add release in project %s/%s", addReleasesJob.Namespace, addReleasesJob.Name)
	projectCache.LatestProjectJobState.Status = "Succeed"
	setProjectCacheToRedisUntilSuccess(projectCache)
	return nil
}
