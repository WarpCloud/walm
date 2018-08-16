package redis

import (
	"time"

	"walm/pkg/setting"

	"github.com/go-redis/redis"
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
		Password:	  setting.Config.RedisConfig.Password,
		DB:	          setting.Config.RedisConfig.DB,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})
	client.FlushDB()
	redisClient = &RedisClient{client: client}
}
