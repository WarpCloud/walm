package redis

import (
	"time"
	"WarpCloud/walm/pkg/setting"
	"github.com/go-redis/redis"
)

const (
	KeyNotFoundErrMsg = "redis: nil"

	WalmJobsKey       = "walm-jobs"
	//WalmReleasesKey   = "walm-releases"
	//WalmProjectsKey   = "walm-project-tasks"
	//WalmReleaseTasksKey   = "walm-release-tasks"
)

type RedisClient struct {
	client *redis.Client
}

var redisClient *RedisClient

func GetDefaultRedisClient() *RedisClient {
	if redisClient == nil {
		client := redis.NewClient(&redis.Options{
			Addr:         setting.Config.RedisConfig.Addr,
			Password:     setting.Config.RedisConfig.Password,
			DB:           setting.Config.RedisConfig.DB,
			DialTimeout:  10 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			PoolSize:     10,
			PoolTimeout:  30 * time.Second,
		})
		redisClient = &RedisClient{client: client}
	}
	return redisClient
}

func (redisClient *RedisClient) GetClient() *redis.Client {
	return redisClient.client
}