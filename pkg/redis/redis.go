package redis

import (
	"time"
	"walm/pkg/setting"
	"github.com/go-redis/redis"
)

const (
	KeyNotFoundErrMsg = "redis: nil"
	WalmJobsKey       = "walm-jobs"
	WalmReleasesKey   = "walm-releases"
)

type RedisClient struct {
	client *redis.Client
}

var redisClient *RedisClient

func GetDefaultRedisClient() *RedisClient {
	return redisClient
}

func InitRedisClient() {
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

func (redisClient *RedisClient) GetClient() *redis.Client {
	return redisClient.client
}

func CreateFakeRedisClient() *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:         "172.16.1.45:6379",
		Password:     "walmtest",
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})
	return &RedisClient{client: client}
}
