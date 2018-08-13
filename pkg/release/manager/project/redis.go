package project

import (
	"time"

	"walm/pkg/setting"

	"github.com/go-redis/redis"
)

var redisClient *redis.Client

func InitRedisClient() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:         setting.Config.RedisConfig.Addr,
		Password:	  setting.Config.RedisConfig.Password,
		DB:	          setting.Config.RedisConfig.DB,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})
	redisClient.FlushDB()
}
