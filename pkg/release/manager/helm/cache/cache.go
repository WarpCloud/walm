package cache

import (
	"walm/pkg/redis"
	"k8s.io/helm/pkg/helm"
)

type HelmCache struct {
	redisClient *redis.RedisClient
	helmClient       *helm.Client
}

func (cache *HelmCache) Resync() error {
	//resp, err := cache.helmClient.ListReleases()
	//if err != nil {
	//	logrus.Errorf("failed to list helm releases: %s\n", err.Error())
	//	return err
	//}

	//releaseInfoCaches, err := buildReleaseInfoCaches(resp.Releases)
	//if err != nil {
	//	logrus.Errorf("failed to build release info caches: %s", err.Error())
	//	return err
	//}
	//
	//cache.redisClient.GetClient().HScan()
	return nil
}
