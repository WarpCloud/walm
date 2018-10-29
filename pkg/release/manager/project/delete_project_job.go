package project

import (
	"walm/pkg/job"
	"github.com/sirupsen/logrus"
)

const (
	deleteProjectJobType = "DeleteProject"
)

func init() {
	job.RegisterJobType(deleteProjectJobType, func() job.Job {
		return &DeleteProjectJob{}
	})
}

type DeleteProjectJob struct {
	Namespace string
	Name      string
	Async     bool
}

func (deleteProjectJob *DeleteProjectJob) Type() string {
	return deleteProjectJobType
}

func (deleteProjectJob *DeleteProjectJob) deleteProject() error {
	projectInfo, err := GetDefaultProjectManager().GetProjectInfo(deleteProjectJob.Namespace, deleteProjectJob.Name)
	if err != nil {
		logrus.Errorf("failed to get project info : %s", err.Error())
		return err
	}

	for _, releaseInfo := range projectInfo.Releases {
		releaseName := buildProjectReleaseName(projectInfo.Name, releaseInfo.Name)
		err = GetDefaultProjectManager().helmClient.DeleteRelease(deleteProjectJob.Namespace, releaseName, false)
		if err != nil {
			logrus.Errorf("failed to delete release %s : %s", releaseName, err.Error())
			return err
		}
	}
	return nil
}

func (deleteProjectJob *DeleteProjectJob) Do() error {
	logrus.Debugf("start to delete project %s/%s", deleteProjectJob.Namespace, deleteProjectJob.Name)

	projectCache := buildProjectCache(deleteProjectJob.Namespace, deleteProjectJob.Name, deleteProjectJob.Type(), "Running", deleteProjectJob.Async)
	setProjectCacheToRedisUntilSuccess(projectCache)

	err := deleteProjectJob.deleteProject()
	if err != nil {
		logrus.Errorf("failed to delete project %s/%s : %s", deleteProjectJob.Namespace, deleteProjectJob.Name, err.Error())
		projectCache.LatestProjectJobState.Status = "Failed"
		projectCache.LatestProjectJobState.Message = err.Error()
		setProjectCacheToRedisUntilSuccess(projectCache)
		return err
	}

	logrus.Infof("succeed to delete project %s/%s", deleteProjectJob.Namespace, deleteProjectJob.Name)
	deleteProjectCacheUntilSuccess(deleteProjectJob.Namespace, deleteProjectJob.Name)
	return nil
}
