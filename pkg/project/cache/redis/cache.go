package redis

import (
	"WarpCloud/walm/pkg/redis"
	"github.com/sirupsen/logrus"
	"WarpCloud/walm/pkg/models/project"
	"encoding/json"
)

type Cache struct {
	redis redis.Redis
}

func (cache *Cache) GetProjectTask(namespace, name string) (projectTask *project.ProjectTask, err error) {
	projectTaskStr, err := cache.redis.GetFieldValue(redis.WalmProjectsKey, namespace, name)
	if err != nil {
		return nil, err
	}

	projectTask = &project.ProjectTask{}
	err = json.Unmarshal([]byte(projectTaskStr), projectTask)
	if err != nil {
		logrus.Errorf("failed to unmarshal projectTaskStr %s : %s", projectTaskStr, err.Error())
		return
	}
	projectTask.CompatiblePreviousProjectTask()
	return
}

func (cache *Cache) GetProjectTasks(namespace string) (projectTasks []*project.ProjectTask, err error) {
	projectTaskStrs, err := cache.redis.GetFieldValues(redis.WalmProjectsKey, namespace)
	if err != nil {
		return nil, err
	}

	projectTasks = []*project.ProjectTask{}
	for _, projectTaskStr := range projectTaskStrs {
		projectTask := &project.ProjectTask{}

		err = json.Unmarshal([]byte(projectTaskStr), projectTask)
		if err != nil {
			logrus.Errorf("failed to unmarshal project task of %s: %s", projectTaskStr, err.Error())
			return
		}
		projectTask.CompatiblePreviousProjectTask()
		projectTasks = append(projectTasks, projectTask)
	}

	return
}

func (cache *Cache) CreateOrUpdateProjectTask(projectTask *project.ProjectTask) error {
	if projectTask == nil {
		logrus.Warn("failed to create or update project task as it is nil")
		return nil
	}

	err := cache.redis.SetFieldValues(redis.WalmProjectsKey, map[string]interface{}{redis.BuildFieldName(projectTask.Namespace, projectTask.Name): projectTask})
	if err != nil {
		return err
	}
	logrus.Debugf("succeed to set project task of %s/%s to redis", projectTask.Namespace, projectTask.Name)
	return nil
}

