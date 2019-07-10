package impl

import (
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	walmRedis "WarpCloud/walm/pkg/redis"
	errorModel "WarpCloud/walm/pkg/models/error"
	"time"
	"WarpCloud/walm/pkg/setting"
	"encoding/json"
)

type Redis struct {
	client *redis.Client
}

func (redis *Redis) GetFieldValue(key, namespace, name string) (value string, err error) {
	value, err = redis.client.HGet(key, walmRedis.BuildFieldName(namespace, name)).Result()
	if err != nil {
		if isKeyNotFoundError(err) {
			logrus.Warnf("field %s/%s of key %s is not found in redis", namespace, name, key)
			err = errorModel.NotFoundError{}
			return
		}
		logrus.Errorf("failed to get field %s/%s of key %s from redis: %s", namespace, name, key, err.Error())
		return
	}
	return
}

func (redis *Redis) GetFieldValues(key, namespace string) (values []string, err error) {
	values = []string{}
	if namespace == "" {
		releaseCacheMap, err := redis.client.HGetAll(key).Result()
		if err != nil {
			logrus.Errorf("failed to get all the fields of key %s from redis: %s", key, err.Error())
			return nil, err
		}
		for _, releaseCacheStr := range releaseCacheMap {
			values = append(values, releaseCacheStr)
		}
	} else {
		filter := buildHScanFilter(namespace)
		// ridiculous logic: scan result contains both key and value
		scanResult, _, err := redis.client.HScan(key, 0, filter, 10000).Result()
		if err != nil {
			logrus.Errorf("failed to scan the redis with filter=%s : %s", filter, err.Error())
			return nil, err
		}

		for i := 1; i < len(scanResult); i += 2 {
			values = append(values, scanResult[i])
		}
	}
	return
}

func (redis *Redis) GetFieldValuesByNames(key string, fieldNames ... string) (values []string, err error) {
	objects, err := redis.client.HMGet(key, fieldNames...).Result()
	if err != nil {
		logrus.Errorf("failed to get fields %v of key %s from redis : %s", fieldNames, key, err.Error())
		return nil, err
	}
	values = []string{}
	for _, object := range objects {
		values = append(values, object.(string))
	}
	return
}

func (redis *Redis) SetFieldValues(key string, fieldValues map[string]interface{}) error {
	if len(fieldValues) == 0 {
		return nil
	}
	marshaledFieldValues := map[string]interface{}{}
	for k, value := range fieldValues {
		valueStr, err := json.Marshal(value)
		if err != nil {
			logrus.Errorf("failed to marshal value : %s", err.Error())
			return err
		}
		marshaledFieldValues[k] = string(valueStr)
	}
	_, err := redis.client.HMSet(key, marshaledFieldValues).Result()
	if err != nil {
		logrus.Errorf("failed to set to redis : %s", err.Error())
		return err
	}
	return nil
}

//TODO delete a filed which does not exist to check whether an error is returned
func (redis *Redis) DeleteField(key, namespace, name string) error {
	_, err := redis.client.HDel(key, walmRedis.BuildFieldName(namespace, name)).Result()
	if err != nil {
		logrus.Errorf("failed to delete filed %s/%s of key %s from redis: %s", namespace, name, key, err.Error())
		return err
	}
	return nil
}

func NewRedisClient(redisConfig *setting.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         redisConfig.Addr,
		Password:     redisConfig.Password,
		DB:           redisConfig.DB,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})
}

func NewRedis(redisClient *redis.Client) *Redis {
	return &Redis{
		client: redisClient,
	}
}
