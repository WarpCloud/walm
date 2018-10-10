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

func (cache *HelmCache) Resync() error {
	for {
		err := cache.redisClient.GetClient().Watch(func(tx *goredis.Tx) error {
			resp, err := cache.helmClient.ListReleases(helm.ReleaseListStatuses(
				[]hapiRelease.Status_Code{hapiRelease.Status_UNKNOWN, hapiRelease.Status_DEPLOYED,
					hapiRelease.Status_DELETED, hapiRelease.Status_SUPERSEDED, hapiRelease.Status_FAILED,
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

			releaseCacheKeys, err := tx.HKeys(redis.WalmReleasesKey).Result()
			if err != nil {
				logrus.Errorf("failed to get release cache keys from redis: %s", err.Error())
				return err
			}

			releaseCacheKeysToDel := []string{}
			for _, releaseCacheKey := range releaseCacheKeys {
				if _, ok := releaseCachesFromHelm[releaseCacheKey]; !ok {
					releaseCacheKeysToDel = append(releaseCacheKeysToDel, releaseCacheKey)
				}
			}

			_, err = tx.Pipelined(func(pipe goredis.Pipeliner) error {
				pipe.HMSet(redis.WalmReleasesKey, releaseCachesFromHelm)
				if len(releaseCacheKeysToDel) > 0 {
					pipe.HDel(redis.WalmReleasesKey, releaseCacheKeysToDel...)
				}
				return nil
			})

			return err
		}, redis.WalmReleasesKey)

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

func NewHelmCache(redisClient *redis.RedisClient, helmClient *helm.Client, kubeClient *kube.Client) *HelmCache {
	return &HelmCache{
		redisClient: redisClient,
		helmClient:  helmClient,
		kubeClient:  kubeClient,
	}
}
