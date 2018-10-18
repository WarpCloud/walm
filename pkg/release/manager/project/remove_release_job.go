package project

import (
	"walm/pkg/job"
	"github.com/sirupsen/logrus"
	"fmt"
)

const (
	removeReleaseJobType = "RemoveReleaseInProject"
)

func init() {
	job.RegisterJobType(removeReleaseJobType, func() job.Job {
		return &RemoveReleaseJob{}
	})
}

type RemoveReleaseJob struct {
	Async       bool
	Namespace   string
	Name        string
	ReleaseName string
}

func (removeReleaseJob *RemoveReleaseJob) Type() string {
	return removeReleaseJobType
}

func (removeReleaseJob *RemoveReleaseJob) removeRelease() error {
	projectInfo, err := GetDefaultProjectManager().GetProjectInfo(removeReleaseJob.Namespace, removeReleaseJob.Name)
	if err != nil {
		logrus.Errorf("failed to get project info : %s", err.Error())
		return err
	}

	releaseParams := buildReleaseRequest(projectInfo, removeReleaseJob.ReleaseName)
	if releaseParams == nil {
		return fmt.Errorf("release is %s not found in project %s", removeReleaseJob.ReleaseName, removeReleaseJob.ReleaseName)
	}
	if projectInfo != nil {
		affectReleaseRequest, err2 := GetDefaultProjectManager().brainFuckRuntimeDepParse(projectInfo, releaseParams, true)
		if err2 != nil {
			logrus.Errorf("RuntimeDepParse install release %s error %v\n", releaseParams.Name, err)
			return err2
		}
		for _, affectReleaseParams := range affectReleaseRequest {
			logrus.Infof("Update BecauseOf Dependency Modified: %v", *affectReleaseParams)
			err = GetDefaultProjectManager().helmClient.UpgradeRealese(removeReleaseJob.Namespace, affectReleaseParams)
			if err != nil {
				logrus.Errorf("RemoveReleaseInProject Other Affected Release install release %s error %v\n", releaseParams.Name, err)
				return err
			}
		}
	}

	releaseProjectName := buildProjectReleaseName(removeReleaseJob.Name, removeReleaseJob.ReleaseName)
	err = GetDefaultProjectManager().helmClient.DeleteRelease(removeReleaseJob.Namespace, releaseProjectName)
	if err != nil {
		logrus.Errorf("RemoveReleaseInProject install release %s error %v\n", releaseProjectName, err)
		return err
	}
	return nil
}

func (removeReleaseJob *RemoveReleaseJob) Do() error {
	logrus.Debugf("start to remove release in project %s/%s", removeReleaseJob.Namespace, removeReleaseJob.Name)

	projectCache := buildProjectCache(removeReleaseJob.Namespace, removeReleaseJob.Name, removeReleaseJob.Type(), "Running", removeReleaseJob.Async)
	setProjectCacheToRedisUntilSuccess(projectCache)

	err := removeReleaseJob.removeRelease()
	if err != nil {
		logrus.Errorf("failed to remove release in project %s/%s : %s", removeReleaseJob.Namespace, removeReleaseJob.Name, err.Error())
		projectCache.LatestProjectJobState.Status = "Failed"
		projectCache.LatestProjectJobState.Message = err.Error()
		setProjectCacheToRedisUntilSuccess(projectCache)
		return err
	}

	logrus.Infof("succeed to remove release in project %s/%s", removeReleaseJob.Namespace, removeReleaseJob.Name)
	projectCache.LatestProjectJobState.Status = "Succeed"
	setProjectCacheToRedisUntilSuccess(projectCache)
	return nil
}
