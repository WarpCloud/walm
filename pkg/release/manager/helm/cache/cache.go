package cache

import (
	"walm/pkg/redis"
	"k8s.io/helm/pkg/helm"
	"github.com/sirupsen/logrus"
	hapiRelease "k8s.io/helm/pkg/proto/hapi/release"
	"walm/pkg/release"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/chartutil"
	"github.com/ghodss/yaml"
	"bytes"
	"k8s.io/helm/pkg/kube"
	"encoding/json"
	goredis "github.com/go-redis/redis"
	"time"
	walmerr "walm/pkg/util/error"
	"strings"
		"fmt"
	"walm/pkg/k8s/handler"
	"walm/pkg/k8s/adaptor"
)

type HelmCache struct {
	redisClient *redis.RedisClient
	helmClient  *helm.Client
	kubeClient  *kube.Client
}

func (cache *HelmCache) CreateOrUpdateReleaseCache(helmRelease *hapiRelease.Release) error {
	releaseCache, err := cache.buildReleaseCaches([]*hapiRelease.Release{helmRelease})
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
			logrus.Errorf("release cache of %s is not found in redis", name)
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

func (cache *HelmCache) GetReleaseCachesByNames(namespace string, names ...string) (releaseCaches []*release.ReleaseCache, err error) {
	releaseCaches = []*release.ReleaseCache{}
	if len(names) == 0 {
		return
	}

	releaseCacheFieldNames := []string{}
	for _, name := range names {
		releaseCacheFieldNames = append(releaseCacheFieldNames, buildWalmReleaseFieldName(namespace, name))
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

func (cache *HelmCache) SingleTenantResync(client *helm.Client, tx *goredis.Tx, isSystemTiller bool) error {
	var currentHelmClient *helm.Client
	if isSystemTiller {
		currentHelmClient = cache.helmClient
	} else {
		currentHelmClient = client
		err := currentHelmClient.PingTiller()
		if err != nil {
			logrus.Errorf("RedisResync failed to ping tiller err: %s\n", err.Error())
			return err
		}
	}
	resp, err := currentHelmClient.ListReleases(helm.ReleaseListStatuses(
		[]hapiRelease.Status_Code{hapiRelease.Status_UNKNOWN, hapiRelease.Status_DEPLOYED,
			hapiRelease.Status_DELETED, hapiRelease.Status_FAILED,
			hapiRelease.Status_DELETING, hapiRelease.Status_PENDING_INSTALL, hapiRelease.Status_PENDING_UPGRADE,
			hapiRelease.Status_PENDING_ROLLBACK}))
	if err != nil {
		logrus.Errorf("failed to list helm releases: %s\n", err.Error())
		return err
	}
	releaseCachesFromHelm, err := cache.buildReleaseCaches(resp.Releases)
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

	projectCachesFromHelm := map[string]string{}
	for releaseCacheKey, releaseCacheStr := range releaseCachesFromHelm {
		releaseCache := &release.ReleaseCache{}
		err = json.Unmarshal(releaseCacheStr.([]byte), releaseCache)
		if err != nil {
			logrus.Errorf("failed to unmarshal release cache of %s: %s", releaseCacheKey, err.Error())
			return err
		}
		projectNameArray := strings.Split(releaseCache.Name, "--")
		if len(projectNameArray) == 2 {
			projectName := projectNameArray[0]
			_, ok := projectCachesFromHelm[buildWalmProjectFieldName(releaseCache.Namespace, projectName)]
			if !ok {
				projectCacheStr, err := json.Marshal(&release.ProjectCache{
					Namespace: releaseCache.Namespace,
					Name:      projectName,
					LatestProjectJobState: release.ProjectJobState{
						Type:    "NotKnown",
						Status:  "Succeed",
						Message: "This project is synced from helm",
					},
				})
				if err != nil {
					logrus.Errorf("failed to marshal project cache of %s/%s: %s", releaseCache.Namespace, projectName, err.Error())
					return err
				}
				projectCachesFromHelm[buildWalmProjectFieldName(releaseCache.Namespace, projectName)] = string(projectCacheStr)
			}
		}
	}

	projectCacheInRedis, err := tx.HGetAll(redis.WalmProjectsKey).Result()
	if err != nil {
		logrus.Errorf("failed to get project caches from redis: %s", err.Error())
		return err
	}

	projectCachesToSet := map[string]interface{}{}
	projectCachesToDel := []string{}
	for projectCacheKey, projectCacheStr := range projectCacheInRedis {
		if _, ok := projectCachesFromHelm[projectCacheKey] ; !ok {
			projectCache := &release.ProjectCache{}
			err = json.Unmarshal([]byte(projectCacheStr), projectCache)
			if err != nil {
				logrus.Errorf("failed to unmarshal projectCacheStr %s : %s", projectCacheStr, err.Error())
				return err
			}
			if !projectCache.IsProjectJobNotFinished() {
				projectCachesToDel = append(projectCachesToDel, projectCacheKey)
			}
		}
	}
	for projectCacheKey, projectCacheStr := range projectCachesFromHelm {
		if _, ok := projectCacheInRedis[projectCacheKey] ; !ok {
			projectCachesToSet[projectCacheKey] = projectCacheStr
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
		return nil
	})
	return err
}

func IsMultiTenant(tenantName string) (bool, error) {
	namespace, err := handler.GetDefaultHandlerSet().GetNamespaceHandler().GetNamespace(tenantName)
	if err != nil {
		if adaptor.IsNotFoundErr(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	_, ok := namespace.Labels["multi-tenant"]
	if ok {
		return true, nil
	} else {
		return false, nil
	}
}

func (cache *HelmCache) Resync() error {
	for {
		err := cache.redisClient.GetClient().Watch(func(tx *goredis.Tx) error {
			err := cache.SingleTenantResync(nil, tx, true)
			if err != nil {
				return err
			}
			namespaces, err := handler.GetDefaultHandlerSet().GetNamespaceHandler().ListNamespaces(nil)
			if err != nil {
				logrus.Errorf("ListNamespaces error %s\n", err.Error())
				return nil
			}
			for _, namespace := range namespaces {
				multiTenant, err := IsMultiTenant(namespace.Name)
				if err != nil {
					logrus.Errorf("IsMultiTenant error %s\n", err.Error())
					continue
				}
				if multiTenant {
					tillerHosts := fmt.Sprintf("tiller-tenant.%s.svc:44134", namespace.Name)
					tenantClient := helm.NewClient(helm.Host(tillerHosts))
					cache.SingleTenantResync(tenantClient, tx, true)
				}
			}

			return nil
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
			return err
		}
	}
}

func (cache *HelmCache) CreateOrUpdateProjectCache(projectCache *release.ProjectCache) (err error) {
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

func (cache *HelmCache) GetProjectCache(namespace, name string) (projectCache *release.ProjectCache, err error) {
	projectCacheStr, err := cache.redisClient.GetClient().HGet(redis.WalmProjectsKey, buildWalmProjectFieldName(namespace, name)).Result()
	if err != nil {
		if err.Error() == redis.KeyNotFoundErrMsg {
			logrus.Errorf("project cache of %s/%s is not found in redis", namespace, name)
			return nil, walmerr.NotFoundError{}
		}
		logrus.Errorf("failed to get project cache of %s/%s from redis : %s", namespace, name, err.Error())
		return nil, err
	}

	projectCache = &release.ProjectCache{}
	err = json.Unmarshal([]byte(projectCacheStr), projectCache)
	if err != nil {
		logrus.Errorf("failed to unmarshal projectCacheStr %s : %s", projectCacheStr, err.Error())
		return
	}
	return
}

func (cache *HelmCache) GetProjectCaches(namespace string) (projectCaches []*release.ProjectCache, err error) {
	filter := namespace + "/*"
	if namespace == "" {
		filter = "*/*"
	}
	scanResult, _, err := cache.redisClient.GetClient().HScan(redis.WalmProjectsKey, 0, filter, 1000).Result()
	if err != nil {
		logrus.Errorf("failed to scan the release caches from redis in namespace %s : %s", namespace, err.Error())
		return nil, err
	}

	projectCacheStrs := []string{}
	for i := 1; i < len(scanResult); i += 2 {
		projectCacheStrs = append(projectCacheStrs, scanResult[i])
	}

	projectCaches = []*release.ProjectCache{}
	for _, projectCacheStr := range projectCacheStrs {
		projectCache := &release.ProjectCache{}
		err = json.Unmarshal([]byte(projectCacheStr), projectCache)
		if err != nil {
			logrus.Errorf("failed to unmarshal projectCacheStr %s : %s", projectCacheStr, err.Error())
			return
		}
		projectCaches = append(projectCaches, projectCache)
	}

	return
}

func (cache *HelmCache) buildReleaseCaches(releases []*hapiRelease.Release) (releaseCaches map[string]interface{}, err error) {
	releaseCaches = map[string]interface{}{}
	for _, helmRelease := range releases {
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

		fieldName := buildWalmReleaseFieldName(releaseCache.Namespace, releaseCache.Name)
		releaseCaches[fieldName] = releaseCacheStr
	}
	return
}

func (cache *HelmCache) buildReleaseCache(helmRelease *hapiRelease.Release) (releaseCache *release.ReleaseCache, err error) {
	emptyChart := chart.Chart{}
	helmVals := release.HelmValues{}
	releaseSpec := release.ReleaseSpec{}
	releaseSpec.Name = helmRelease.Name
	releaseSpec.Namespace = helmRelease.Namespace
	releaseSpec.Dependencies = make(map[string]string)
	releaseSpec.Version = helmRelease.Version
	releaseSpec.ChartVersion = helmRelease.Chart.Metadata.Version
	releaseSpec.ChartName = helmRelease.Chart.Metadata.Name
	releaseSpec.ChartAppVersion = helmRelease.Chart.Metadata.AppVersion
	cvals, err := chartutil.CoalesceValues(&emptyChart, helmRelease.Config)
	if err != nil {
		logrus.Errorf("parse raw values error %s\n", helmRelease.Config.Raw)
		return
	}
	releaseSpec.ConfigValues = cvals
	err = yaml.Unmarshal([]byte(helmRelease.GetChart().GetValues().GetRaw()), &helmVals)
	if err == nil {
		if helmVals.AppHelmValues != nil && helmVals.AppHelmValues.Dependencies != nil {
			releaseSpec.Dependencies = helmVals.AppHelmValues.Dependencies
		}
		releaseSpec.HelmValues = helmVals
	}

	releaseCache = &release.ReleaseCache{
		ReleaseSpec: releaseSpec,
	}

	releaseCache.ReleaseResourceMetas, err = cache.getReleaseResourceMetas(helmRelease)
	return
}

func (cache *HelmCache) getReleaseResourceMetas(helmRelease *hapiRelease.Release) (resources []release.ReleaseResourceMeta, err error) {
	resources = []release.ReleaseResourceMeta{}
	results, err := cache.kubeClient.BuildUnstructured(helmRelease.Namespace, bytes.NewBufferString(helmRelease.Manifest))
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

func NewHelmCache(redisClient *redis.RedisClient, helmClient *helm.Client, kubeClient *kube.Client) *HelmCache {
	return &HelmCache{
		redisClient: redisClient,
		helmClient:  helmClient,
		kubeClient:  kubeClient,
	}
}
