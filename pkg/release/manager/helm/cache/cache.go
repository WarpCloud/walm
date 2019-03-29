package cache

import (
	hapirelease "k8s.io/helm/pkg/hapi/release"
	"walm/pkg/redis"
	"github.com/sirupsen/logrus"
	"walm/pkg/release"
	"k8s.io/helm/pkg/chartutil"
	"bytes"
	"encoding/json"
	goredis "github.com/go-redis/redis"
	"time"
	walmerr "walm/pkg/util/error"
	"k8s.io/helm/pkg/action"
	"walm/pkg/k8s/client"
	"k8s.io/helm/pkg/storage/driver"
	"k8s.io/helm/pkg/storage"
	"walm/pkg/k8s/handler"
	"k8s.io/helm/pkg/chart"
	"walm/pkg/util/transwarpjsonnet"
	"walm/pkg/release/manager/metainfo"
)

const (
	ProjectNameLabelKey = "Project-Name"
)

type HelmCache struct {
	redisClient *redis.RedisClient
	list        *action.List
}

func (cache *HelmCache) CreateOrUpdateReleaseCache(helmRelease *hapirelease.Release) error {
	if helmRelease == nil {
		logrus.Warn("failed to create or update cache as helm release is nil")
		return nil
	}
	releaseCache, err := cache.buildReleaseCaches(map[string]*hapirelease.Release{
		buildWalmReleaseFieldName(helmRelease.Namespace, helmRelease.Name): helmRelease})
	if err != nil {
		logrus.Errorf("failed to build release cache of %s : %s", helmRelease.Name, err.Error())
		return err
	}

	_, err = cache.redisClient.GetClient().HMSet(redis.WalmReleasesKey, releaseCache).Result()
	if err != nil {
		logrus.Errorf("failed to set release cache of %s to redis: %s", helmRelease.Name, err.Error())
		return err
	}
	logrus.Debugf("succeed to set release cache of %s to redis", helmRelease.Name)
	return nil
}

func (cache *HelmCache) DeleteReleaseCache(namespace, name string) error {
	_, err := cache.redisClient.GetClient().HDel(redis.WalmReleasesKey, buildWalmReleaseFieldName(namespace, name)).Result()
	if err != nil {
		logrus.Errorf("failed to delete release cache of %s from redis: %s", name, err.Error())
		return err
	}
	logrus.Debugf("succeed to delete release cache of %s from redis", name)
	return nil
}

func (cache *HelmCache) GetReleaseCache(namespace, name string) (releaseCache *release.ReleaseCache, err error) {
	releaseCacheStr, err := cache.redisClient.GetClient().HGet(redis.WalmReleasesKey, buildWalmReleaseFieldName(namespace, name)).Result()
	if err != nil {
		if err.Error() == redis.KeyNotFoundErrMsg {
			logrus.Warnf("release cache of %s is not found in redis", name)
			return nil, walmerr.NotFoundError{}
		}
		logrus.Errorf("failed to get release cache of %s from redis: %s", name, err.Error())
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

//TODO count is not available every time
func (cache *HelmCache) GetReleaseCaches(namespace, filter string, count int64) (releaseCaches []*release.ReleaseCache, err error) {
	var releaseCacheStrs []string
	if namespace == "" && filter == "" && count == 0 {
		releaseCacheMap, err := cache.redisClient.GetClient().HGetAll(redis.WalmReleasesKey).Result()
		if err != nil {
			logrus.Errorf("failed to get all the release caches from redis: %s", err.Error())
			return nil, err
		}
		for _, releaseCacheStr := range releaseCacheMap {
			releaseCacheStrs = append(releaseCacheStrs, releaseCacheStr)
		}
	} else {
		newFilter := buildHScanFilter(namespace, filter)
		if count == 0 {
			count = 1000
		}

		// ridiculous logic: scan result contains both key and value
		// TODO deal with cursor
		scanResult, _, err := cache.redisClient.GetClient().HScan(redis.WalmReleasesKey, 0, newFilter, count).Result()
		if err != nil {
			logrus.Errorf("failed to scan the release caches from redis with namespace=%s filter=%s count=%d: %s", namespace, filter, count, err.Error())
			return nil, err
		}

		for i := 1; i < len(scanResult); i += 2 {
			releaseCacheStrs = append(releaseCacheStrs, scanResult[i])
		}
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

func (cache *HelmCache) GetReleaseCachesByNames(names []ReleaseFieldName) (releaseCaches []*release.ReleaseCache, err error) {
	releaseCaches = []*release.ReleaseCache{}
	if len(names) == 0 {
		return
	}

	releaseCacheFieldNames := []string{}
	for _, name := range names {
		releaseCacheFieldNames = append(releaseCacheFieldNames, buildWalmReleaseFieldName(name.Namespace, name.Name))
	}

	releaseCacheStrs, err := cache.redisClient.GetClient().HMGet(redis.WalmReleasesKey, releaseCacheFieldNames...).Result()
	if err != nil {
		logrus.Errorf("failed to get release caches from redis : %s", err.Error())
		return nil, err
	}

	for index, releaseCacheStr := range releaseCacheStrs {
		if releaseCacheStr == nil {
			logrus.Warnf("release cache %s is not found", releaseCacheFieldNames[index])
			continue
		}

		releaseCache := &release.ReleaseCache{}

		err = json.Unmarshal([]byte(releaseCacheStr.(string)), releaseCache)
		if err != nil {
			logrus.Errorf("failed to unmarshal release cache of %s: %s", releaseCacheStr, err.Error())
			return
		}
		releaseCaches = append(releaseCaches, releaseCache)
	}

	return
}

func buildHScanFilter(namespace string, filter string) string {
	newFilter := namespace
	if newFilter == "" {
		newFilter = "*"
	}
	newFilter += "/"
	if filter == "" {
		newFilter += "*"
	} else {
		newFilter += filter
	}
	return newFilter
}

func (cache *HelmCache) Resync() {
	for {
		err := cache.redisClient.GetClient().Watch(func(tx *goredis.Tx) error {
			helmReleases, err := cache.list.Run()
			if err != nil {
				logrus.Errorf("failed to list helm releases: %s\n", err.Error())
				return err
			}

			helmReleasesMap := map[string]*hapirelease.Release{}
			for _, helmRelease := range helmReleases {
				filedName := buildWalmReleaseFieldName(helmRelease.Namespace, helmRelease.Name)
				if existedHelmRelease, ok := helmReleasesMap[filedName] ; ok {
					if existedHelmRelease.Version < helmRelease.Version {
						helmReleasesMap[filedName] = helmRelease
					}
				} else {
					helmReleasesMap[filedName] = helmRelease
				}

			}

			releaseCachesFromHelm, err := cache.buildReleaseCaches(helmReleasesMap)
			if err != nil {
				logrus.Errorf("failed to build release caches: %s", err.Error())
				return err
			}
			releaseCacheKeysFromRedis, err := tx.HKeys(redis.WalmReleasesKey).Result()
			if err != nil {
				logrus.Errorf("failed to get release cache keys from redis: %s", err.Error())
				return err
			}

			releaseCacheKeysToDel := []string{}
			for _, releaseCacheKey := range releaseCacheKeysFromRedis {
				if _, ok := releaseCachesFromHelm[releaseCacheKey]; !ok {
					releaseCacheKeysToDel = append(releaseCacheKeysToDel, releaseCacheKey)
				}
			}

			releaseConfigs, err := handler.GetDefaultHandlerSet().GetReleaseConfigHandler().ListReleaseConfigs("", nil)
			if err != nil {
				logrus.Errorf("failed to list release configs : %s", err.Error())
				return err
			}

			projectCachesFromReleaseConfig := map[string]string{}
			for _, releaseConfig := range releaseConfigs {
				if projectName, ok1 := releaseConfig.Labels[ProjectNameLabelKey]; ok1 {
					_, ok := projectCachesFromReleaseConfig[buildWalmProjectFieldName(releaseConfig.Namespace, projectName)]
					if !ok {
						projectCacheStr, err := json.Marshal(&ProjectCache{
							Namespace: releaseConfig.Namespace,
							Name:      projectName,
						})
						if err != nil {
							logrus.Errorf("failed to marshal project cache of %s/%s: %s", releaseConfig.Namespace, projectName, err.Error())
							return err
						}
						projectCachesFromReleaseConfig[buildWalmProjectFieldName(releaseConfig.Namespace, projectName)] = string(projectCacheStr)
					}
				}
			}

			releaseTasksFromHelm := map[string]string{}
			for releaseCacheKey, releaseCacheStr := range releaseCachesFromHelm {
				releaseCache := &release.ReleaseCache{}
				err = json.Unmarshal(releaseCacheStr.([]byte), releaseCache)
				if err != nil {
					logrus.Errorf("failed to unmarshal release cache of %s: %s", releaseCacheKey, err.Error())
					return err
				}

				releaseTaskStr, err := json.Marshal(&ReleaseTask{
					Namespace: releaseCache.Namespace,
					Name:      releaseCache.Name,
				})
				if err != nil {
					logrus.Errorf("failed to marshal release task of %s/%s: %s", releaseCache.Namespace, releaseCache.Name, err.Error())
					return err
				}
				releaseTasksFromHelm[buildWalmReleaseFieldName(releaseCache.Namespace, releaseCache.Name)] = string(releaseTaskStr)
			}

			projectCacheInRedis, err := tx.HGetAll(redis.WalmProjectsKey).Result()
			if err != nil {
				logrus.Errorf("failed to get project caches from redis: %s", err.Error())
				return err
			}

			releaseTaskInRedis, err := tx.HGetAll(redis.WalmReleaseTasksKey).Result()
			if err != nil {
				logrus.Errorf("failed to get release tasks from redis: %s", err.Error())
				return err
			}

			projectCachesToSet := map[string]interface{}{}
			projectCachesToDel := []string{}
			for projectCacheKey, projectCacheStr := range projectCacheInRedis {
				if _, ok := projectCachesFromReleaseConfig[projectCacheKey]; !ok {
					projectCache := &ProjectCache{}
					err = json.Unmarshal([]byte(projectCacheStr), projectCache)
					if err != nil {
						logrus.Errorf("failed to unmarshal projectCacheStr %s : %s", projectCacheStr, err.Error())
						return err
					}
					if projectCache.IsLatestTaskFinishedOrTimeout() {
						projectCachesToDel = append(projectCachesToDel, projectCacheKey)
					}
				}
			}
			for projectCacheKey, projectCacheStr := range projectCachesFromReleaseConfig {
				if _, ok := projectCacheInRedis[projectCacheKey]; !ok {
					projectCachesToSet[projectCacheKey] = projectCacheStr
				}
			}

			releaseTasksToSet := map[string]interface{}{}
			releaseTasksToDel := []string{}
			for releaseTaskKey, releaseTaskStr := range releaseTaskInRedis {
				if _, ok := releaseTasksFromHelm[releaseTaskKey]; !ok {
					releaseTask := &ReleaseTask{}
					err = json.Unmarshal([]byte(releaseTaskStr), releaseTask)
					if err != nil {
						logrus.Errorf("failed to unmarshal release task string %s : %s", releaseTaskStr, err.Error())
						return err
					}
					if releaseTask.LatestReleaseTaskSig == nil || releaseTask.LatestReleaseTaskSig.IsTaskFinishedOrTimeout() {
						releaseTasksToDel = append(releaseTasksToDel, releaseTaskKey)
					}
				}
			}
			for releaseTaskKey, releaseTaskStr := range releaseTasksFromHelm {
				if _, ok := releaseTaskInRedis[releaseTaskKey]; !ok {
					releaseTasksToSet[releaseTaskKey] = releaseTaskStr
				}
			}

			_, err = tx.Pipelined(func(pipe goredis.Pipeliner) error {
				if len(releaseCachesFromHelm) > 0 {
					pipe.HMSet(redis.WalmReleasesKey, releaseCachesFromHelm)
				}
				if len(releaseCacheKeysToDel) > 0 {
					pipe.HDel(redis.WalmReleasesKey, releaseCacheKeysToDel...)
				}
				if len(projectCachesToSet) > 0 {
					pipe.HMSet(redis.WalmProjectsKey, projectCachesToSet)
				}
				if len(projectCachesToDel) > 0 {
					pipe.HDel(redis.WalmProjectsKey, projectCachesToDel...)
				}
				if len(releaseTasksToSet) > 0 {
					pipe.HMSet(redis.WalmReleaseTasksKey, releaseTasksToSet)
				}
				if len(releaseTasksToDel) > 0 {
					pipe.HDel(redis.WalmReleaseTasksKey, releaseTasksToDel...)
				}
				return nil
			})
			return err
		}, redis.WalmReleasesKey, redis.WalmProjectsKey)

		if err == goredis.TxFailedErr {
			logrus.Warn("resync release cache transaction failed, will retry after 5 seconds")
			time.Sleep(5 * time.Second)
		} else {
			if err != nil {
				logrus.Errorf("failed to resync release caches: %s", err.Error())
			} else {
				logrus.Info("succeed to resync release caches")
			}
			return
		}
	}
}

func (cache *HelmCache) CreateOrUpdateProjectCache(projectCache *ProjectCache) (err error) {
	projectCacheStr, err := json.Marshal(projectCache)
	if err != nil {
		logrus.Errorf("failed to marshal project cache of %s/%s: %s", projectCache.Namespace, projectCache.Name, err.Error())
		return err
	}
	_, err = cache.redisClient.GetClient().HSet(redis.WalmProjectsKey, buildWalmProjectFieldName(projectCache.Namespace, projectCache.Name), projectCacheStr).Result()
	if err != nil {
		logrus.Errorf("failed to set project cache of  %s/%s: %s", projectCache.Namespace, projectCache.Name, err.Error())
		return
	}
	return
}

func (cache *HelmCache) DeleteProjectCache(namespace, name string) (err error) {
	_, err = cache.redisClient.GetClient().HDel(redis.WalmProjectsKey, buildWalmProjectFieldName(namespace, name)).Result()
	if err != nil {
		logrus.Errorf("failed to delete project cache of %s/%s from redis : %s", namespace, name, err.Error())
		return
	}

	return
}

func (cache *HelmCache) GetProjectCache(namespace, name string) (projectCache *ProjectCache, err error) {
	projectCacheStr, err := cache.redisClient.GetClient().HGet(redis.WalmProjectsKey, buildWalmProjectFieldName(namespace, name)).Result()
	if err != nil {
		if err.Error() == redis.KeyNotFoundErrMsg {
			logrus.Warnf("project cache of %s/%s is not found in redis", namespace, name)
			return nil, walmerr.NotFoundError{}
		}
		logrus.Errorf("failed to get project cache of %s/%s from redis : %s", namespace, name, err.Error())
		return nil, err
	}

	projectCache = &ProjectCache{}
	err = json.Unmarshal([]byte(projectCacheStr), projectCache)
	if err != nil {
		logrus.Errorf("failed to unmarshal projectCacheStr %s : %s", projectCacheStr, err.Error())
		return
	}
	return
}

func (cache *HelmCache) GetProjectCaches(namespace string) (projectCaches []*ProjectCache, err error) {
	filter := namespace + "/*"
	if namespace == "" {
		filter = "*/*"
	}
	scanResult, _, err := cache.redisClient.GetClient().HScan(redis.WalmProjectsKey, 0, filter, 1000).Result()
	if err != nil {
		logrus.Errorf("failed to scan the project caches from redis in namespace %s : %s", namespace, err.Error())
		return nil, err
	}

	projectCacheStrs := []string{}
	for i := 1; i < len(scanResult); i += 2 {
		projectCacheStrs = append(projectCacheStrs, scanResult[i])
	}

	projectCaches = []*ProjectCache{}
	for _, projectCacheStr := range projectCacheStrs {
		projectCache := &ProjectCache{}
		err = json.Unmarshal([]byte(projectCacheStr), projectCache)
		if err != nil {
			logrus.Errorf("failed to unmarshal projectCacheStr %s : %s", projectCacheStr, err.Error())
			return
		}
		projectCaches = append(projectCaches, projectCache)
	}

	return
}

func (cache *HelmCache) CreateOrUpdateReleaseTask(releaseTask *ReleaseTask) (err error) {
	releaseTaskStr, err := json.Marshal(releaseTask)
	if err != nil {
		logrus.Errorf("failed to marshal release task of %s/%s: %s", releaseTask.Namespace, releaseTask.Name, err.Error())
		return err
	}
	_, err = cache.redisClient.GetClient().HSet(redis.WalmReleaseTasksKey, buildWalmReleaseFieldName(releaseTask.Namespace, releaseTask.Name), releaseTaskStr).Result()
	if err != nil {
		logrus.Errorf("failed to set release task of  %s/%s: %s", releaseTask.Namespace, releaseTask.Name, err.Error())
		return
	}
	return
}

func (cache *HelmCache) DeleteReleaseTask(namespace, name string) (err error) {
	_, err = cache.redisClient.GetClient().HDel(redis.WalmReleaseTasksKey, buildWalmProjectFieldName(namespace, name)).Result()
	if err != nil {
		logrus.Errorf("failed to delete release task of %s/%s from redis : %s", namespace, name, err.Error())
		return
	}

	return
}

func (cache *HelmCache) GetReleaseTask(namespace, name string) (releaseTask *ReleaseTask, err error) {
	releaseTaskStr, err := cache.redisClient.GetClient().HGet(redis.WalmReleaseTasksKey, buildWalmProjectFieldName(namespace, name)).Result()
	if err != nil {
		if err.Error() == redis.KeyNotFoundErrMsg {
			logrus.Warnf("release task of %s/%s is not found in redis", namespace, name)
			return nil, walmerr.NotFoundError{}
		}
		logrus.Errorf("failed to get release task of %s/%s from redis : %s", namespace, name, err.Error())
		return nil, err
	}

	releaseTask = &ReleaseTask{}
	err = json.Unmarshal([]byte(releaseTaskStr), releaseTask)
	if err != nil {
		logrus.Errorf("failed to unmarshal releaseTaskStr %s : %s", releaseTaskStr, err.Error())
		return
	}
	return
}

func (cache *HelmCache) GetReleaseTasks(namespace, filter string, count int64) (releaseTasks []*ReleaseTask, err error) {
	var releaseTaskStrs []string
	if namespace == "" && filter == "" && count == 0 {
		releaseTaskMap, err := cache.redisClient.GetClient().HGetAll(redis.WalmReleaseTasksKey).Result()
		if err != nil {
			logrus.Errorf("failed to get all the release tasks from redis: %s", err.Error())
			return nil, err
		}
		for _, releaseTaskStr := range releaseTaskMap {
			releaseTaskStrs = append(releaseTaskStrs, releaseTaskStr)
		}
	} else {
		newFilter := buildHScanFilter(namespace, filter)
		if count == 0 {
			count = 1000
		}

		// ridiculous logic: scan result contains both key and value
		// TODO deal with cursor
		scanResult, _, err := cache.redisClient.GetClient().HScan(redis.WalmReleaseTasksKey, 0, newFilter, count).Result()
		if err != nil {
			logrus.Errorf("failed to scan the release tasks from redis with namespace=%s filter=%s count=%d: %s", namespace, filter, count, err.Error())
			return nil, err
		}

		for i := 1; i < len(scanResult); i += 2 {
			releaseTaskStrs = append(releaseTaskStrs, scanResult[i])
		}
	}

	releaseTasks = []*ReleaseTask{}
	for _, releaseTaskStr := range releaseTaskStrs {
		releaseTask := &ReleaseTask{}

		err = json.Unmarshal([]byte(releaseTaskStr), releaseTask)
		if err != nil {
			logrus.Errorf("failed to unmarshal release task of %s: %s", releaseTaskStr, err.Error())
			return
		}
		releaseTasks = append(releaseTasks, releaseTask)
	}

	return
}

type ReleaseFieldName struct {
	Namespace string
	Name      string
}

func (cache *HelmCache) GetReleaseTasksByNames(names []ReleaseFieldName) (releaseTasks []*ReleaseTask, err error) {
	releaseTasks = []*ReleaseTask{}
	if len(names) == 0 {
		return
	}

	releaseTaskFieldNames := []string{}
	for _, name := range names {
		releaseTaskFieldNames = append(releaseTaskFieldNames, buildWalmReleaseFieldName(name.Namespace, name.Name))
	}

	releaseTaskStrs, err := cache.redisClient.GetClient().HMGet(redis.WalmReleaseTasksKey, releaseTaskFieldNames...).Result()
	if err != nil {
		logrus.Errorf("failed to get release caches from redis : %s", err.Error())
		return nil, err
	}

	for index, releaseTaskStr := range releaseTaskStrs {
		if releaseTaskStr == nil {
			logrus.Warnf("release task %s is not found", releaseTaskFieldNames[index])
			continue
		}

		releaseTask := &ReleaseTask{}

		err = json.Unmarshal([]byte(releaseTaskStr.(string)), releaseTask)
		if err != nil {
			logrus.Errorf("failed to unmarshal release task of %s: %s", releaseTaskStr, err.Error())
			return
		}
		releaseTasks = append(releaseTasks, releaseTask)
	}

	return
}

func (cache *HelmCache) buildReleaseCaches(releases map[string]*hapirelease.Release) (releaseCaches map[string]interface{}, err error) {
	releaseCaches = map[string]interface{}{}
	for fieldName, helmRelease := range releases {
		releaseCache, err := cache.buildReleaseCache(helmRelease)
		if err != nil {
			logrus.Errorf("failed to build release cache of %s: %s", helmRelease.Name, err.Error())
			return nil, err
		}

		releaseCacheStr, err := json.Marshal(releaseCache)
		if err != nil {
			logrus.Errorf("failed to marshal release cache of %s: %s", helmRelease.Name, err.Error())
			return nil, err
		}

		releaseCaches[fieldName] = releaseCacheStr
	}
	return
}

func (cache *HelmCache) buildReleaseCache(helmRelease *hapirelease.Release) (releaseCache *release.ReleaseCache, err error) {
	releaseSpec := release.ReleaseSpec{}
	releaseSpec.Name = helmRelease.Name
	releaseSpec.Namespace = helmRelease.Namespace
	releaseSpec.Dependencies = make(map[string]string)
	releaseSpec.Version = int32(helmRelease.Version)
	releaseSpec.ChartVersion = helmRelease.Chart.Metadata.Version
	releaseSpec.ChartName = helmRelease.Chart.Metadata.Name
	releaseSpec.ChartAppVersion = helmRelease.Chart.Metadata.AppVersion
	releaseSpec.ConfigValues = helmRelease.Config
	releaseCache = &release.ReleaseCache{
		ReleaseSpec: releaseSpec,
	}

	releaseCache.ComputedValues, err = chartutil.CoalesceValues(helmRelease.Chart, helmRelease.Config)
	if err != nil {
		logrus.Errorf("failed to get computed values : %s", err.Error())
		return nil, err
	}

	releaseCache.ReleaseResourceMetas, err = cache.getReleaseResourceMetas(helmRelease)
	releaseCache.MetaInfoValues = buildMetaInfoValues(helmRelease.Chart, releaseCache.ComputedValues)
	return
}

func buildMetaInfoValues(chart *chart.Chart, computedValues map[string]interface{}) (*metainfo.MetaInfoParams) {
	chartMetaInfo, err := transwarpjsonnet.GetChartMetaInfo(chart)
	if err != nil {
		return nil
	}
	if chartMetaInfo != nil {
		metaInfoParams, err := chartMetaInfo.BuildMetaInfoParams(computedValues)
		if err != nil {
			return nil
		}
		return metaInfoParams
	}

	return nil
}

func (cache *HelmCache) getReleaseResourceMetas(helmRelease *hapirelease.Release) (resources []release.ReleaseResourceMeta, err error) {
	resources = []release.ReleaseResourceMeta{}
	results, err := client.GetKubeClient(helmRelease.Namespace).BuildUnstructured(helmRelease.Namespace, bytes.NewBufferString(helmRelease.Manifest))
	if err != nil {
		logrus.Errorf("failed to get release resource metas of %s", helmRelease.Name)
		return resources, err
	}
	for _, result := range results {
		resource := release.ReleaseResourceMeta{
			Kind:      result.Object.GetObjectKind().GroupVersionKind().Kind,
			Namespace: result.Namespace,
			Name:      result.Name,
		}
		resources = append(resources, resource)
	}
	return
}

func buildWalmReleaseFieldName(namespace, name string) string {
	return namespace + "/" + name
}

func buildWalmProjectFieldName(namespace, name string) string {
	return namespace + "/" + name
}

func NewHelmCache(redisClient *redis.RedisClient) *HelmCache {
	kc := client.GetKubeClient("")
	clientset, err := kc.KubernetesClientSet()
	if err != nil {
		logrus.Fatal("failed to get clientset, error %v", err)
	}

	d := driver.NewConfigMaps(clientset.CoreV1().ConfigMaps(""))
	store := storage.Init(d)
	config := &action.Configuration{
		KubeClient: kc,
		Releases:   store,
		Discovery:  clientset.Discovery(),
	}

	result := &HelmCache{
		redisClient: redisClient,
		list:        action.NewList(config),
	}

	result.list.AllNamespaces = true
	result.list.All = true
	result.list.StateMask = action.ListDeployed | action.ListFailed | action.ListPendingInstall | action.ListPendingRollback |
		action.ListPendingUpgrade | action.ListUninstalled | action.ListUninstalling | action.ListUnknown

	return result
}
