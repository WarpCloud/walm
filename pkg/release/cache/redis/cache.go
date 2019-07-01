package redis

import (
	"WarpCloud/walm/pkg/models/release"
	"WarpCloud/walm/pkg/models/k8s"
	"github.com/sirupsen/logrus"
	"encoding/json"
	"WarpCloud/walm/pkg/redis"
)

type Cache struct {
	redis redis.Redis
}

func (cache *Cache) GetReleaseCache(namespace, name string) (releaseCache *release.ReleaseCache, err error) {
	releaseCacheStr, err := cache.redis.GetFieldValue(redis.WalmReleasesKey, namespace, name)
	if err != nil {
		return
	}

	releaseCache = &release.ReleaseCache{}
	err = json.Unmarshal([]byte(releaseCacheStr), releaseCache)
	if err != nil {
		logrus.Errorf("failed to unmarshal release cache of %s: %s", name, err.Error())
		return
	}
	return
}

func (cache *Cache) GetReleaseCaches(namespace string) (releaseCaches []*release.ReleaseCache, err error) {
	releaseCacheStrs, err := cache.redis.GetFieldValues(redis.WalmReleasesKey, namespace)
	if err != nil {
		return nil, err
	}

	releaseCaches = []*release.ReleaseCache{}
	for _, releaseCacheStr := range releaseCacheStrs {
		releaseCache := &release.ReleaseCache{}

		err = json.Unmarshal([]byte(releaseCacheStr), releaseCache)
		if err != nil {
			logrus.Errorf("failed to unmarshal release cache of %s: %s", releaseCacheStr, err.Error())
			return
		}
		releaseCaches = append(releaseCaches, releaseCache)
	}

	return
}

func (cache *Cache) GetReleaseCachesByReleaseConfigs(releaseConfigs []*k8s.ReleaseConfig) (releaseCaches []*release.ReleaseCache, error error) {
	releaseCaches = []*release.ReleaseCache{}
	if len(releaseConfigs) == 0 {
		return
	}

	releaseCacheFieldNames := []string{}
	for _, releaseConfig := range releaseConfigs {
		releaseCacheFieldNames = append(releaseCacheFieldNames, redis.BuildFieldName(releaseConfig.Namespace, releaseConfig.Name))
	}

	releaseCacheStrs, err := cache.redis.GetFieldValuesByNames(redis.WalmReleasesKey, releaseCacheFieldNames...)
	if err != nil {
		return nil, err
	}

	for index, releaseCacheStr := range releaseCacheStrs {
		if releaseCacheStr == "" {
			logrus.Warnf("release cache %s is not found", releaseCacheFieldNames[index])
			continue
		}

		releaseCache := &release.ReleaseCache{}
		err = json.Unmarshal([]byte(releaseCacheStr), releaseCache)
		if err != nil {
			logrus.Errorf("failed to unmarshal release cache of %s: %s", releaseCacheStr, err.Error())
			return
		}
		releaseCaches = append(releaseCaches, releaseCache)
	}

	return
}

func (cache *Cache) CreateOrUpdateReleaseCache(releaseCache *release.ReleaseCache) error {
	if releaseCache == nil {
		logrus.Warn("failed to create or update cache as release cache is nil")
		return nil
	}

	err := cache.redis.SetFieldValues(redis.WalmReleasesKey, map[string]interface{}{redis.BuildFieldName(releaseCache.Namespace, releaseCache.Name): releaseCache})
	if err != nil {
		return err
	}
	logrus.Debugf("succeed to set release cache of %s/%s to redis", releaseCache.Namespace, releaseCache.Name)
	return nil
}

func (cache *Cache) DeleteReleaseCache(namespace string, name string) error {
	err := cache.redis.DeleteField(redis.WalmReleasesKey, namespace, name)
	if err != nil {
		return err
	}
	logrus.Debugf("succeed to delete release cache of %s from redis", name)
	return nil
}

func (cache *Cache) GetReleaseTask(namespace, name string) (releaseTask *release.ReleaseTask, err error) {
	releaseTaskStr, err := cache.redis.GetFieldValue(redis.WalmReleaseTasksKey, namespace, name)
	if err != nil {
		return nil, err
	}

	releaseTask = &release.ReleaseTask{}
	err = json.Unmarshal([]byte(releaseTaskStr), releaseTask)
	if err != nil {
		logrus.Errorf("failed to unmarshal releaseTaskStr %s : %s", releaseTaskStr, err.Error())
		return
	}
	return
}

func (cache *Cache) GetReleaseTasks(namespace string) (releaseTasks []*release.ReleaseTask, err error) {
	releaseTaskStrs, err := cache.redis.GetFieldValues(redis.WalmReleaseTasksKey, namespace)
	if err != nil {
		return nil, err
	}

	releaseTasks = []*release.ReleaseTask{}
	for _, releaseTaskStr := range releaseTaskStrs {
		releaseTask := &release.ReleaseTask{}

		err = json.Unmarshal([]byte(releaseTaskStr), releaseTask)
		if err != nil {
			logrus.Errorf("failed to unmarshal release task of %s: %s", releaseTaskStr, err.Error())
			return
		}
		releaseTasks = append(releaseTasks, releaseTask)
	}

	return
}

func (cache *Cache) GetReleaseTasksByReleaseConfigs(releaseConfigs []*k8s.ReleaseConfig) (releaseTasks []*release.ReleaseTask, err error) {
	releaseTasks = []*release.ReleaseTask{}
	if len(releaseConfigs) == 0 {
		return
	}

	releaseTaskFieldNames := []string{}
	for _, releaseConfig := range releaseConfigs {
		releaseTaskFieldNames = append(releaseTaskFieldNames, redis.BuildFieldName(releaseConfig.Namespace, releaseConfig.Name))
	}

	releaseTaskStrs, err := cache.redis.GetFieldValuesByNames(redis.WalmReleaseTasksKey, releaseTaskFieldNames...)
	if err != nil {
		return nil, err
	}

	for index, releaseTaskStr := range releaseTaskStrs {
		if releaseTaskStr == "" {
			logrus.Warnf("release task %s is not found", releaseTaskFieldNames[index])
			continue
		}

		releaseTask := &release.ReleaseTask{}

		err = json.Unmarshal([]byte(releaseTaskStr), releaseTask)
		if err != nil {
			logrus.Errorf("failed to unmarshal release task of %s: %s", releaseTaskStr, err.Error())
			return
		}
		releaseTasks = append(releaseTasks, releaseTask)
	}

	return
}

func (cache *Cache) CreateOrUpdateReleaseTask(releaseTask *release.ReleaseTask) error {
	if releaseTask == nil {
		logrus.Warn("failed to create or update release task as it is nil")
		return nil
	}

	err := cache.redis.SetFieldValues(redis.WalmReleaseTasksKey, map[string]interface{}{redis.BuildFieldName(releaseTask.Namespace, releaseTask.Name): releaseTask})
	if err != nil {
		return err
	}
	logrus.Debugf("succeed to set release task of %s/%s to redis", releaseTask.Namespace, releaseTask.Name)
	return nil
}
