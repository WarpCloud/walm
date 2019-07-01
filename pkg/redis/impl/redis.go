package impl

import (
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	walmRedis "WarpCloud/walm/pkg/redis"
	errorModel "WarpCloud/walm/pkg/models/error"
)

type Redis struct {
	client *redis.Client
}

func (redis *Redis) GetFieldValue(key, namespace, name string) (value string,err error) {
	value, err = redis.client.HGet(key, walmRedis.BuildFieldName(namespace, name)).Result()
	if err != nil {
		if isKeyNotFoundError(err) {
			logrus.Warnf("filed %s/%s of key %s is not found in redis", namespace, name, key)
			err = errorModel.NotFoundError{}
			return
		}
		logrus.Errorf("failed to get filed %s/%s of key %s from redis: %s", namespace, name, key, err.Error())
		return
	}
	return
}

func (redis *Redis) GetFieldValues(key, namespace string) (values []string,err error) {
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

func (redis *Redis) GetFieldValuesByNames(key string, fieldNames ... string) (values []string,err error) {
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
	_, err := redis.client.HMSet(key, fieldValues).Result()
	if err != nil {
		logrus.Errorf("failed to set to redis : %s", err.Error())
		return err
	}
	return nil
}

func (redis *Redis) DeleteField(key, namespace, name string) error {
	_, err := redis.client.HDel(key, walmRedis.BuildFieldName(namespace, name)).Result()
	if err != nil {
		logrus.Errorf("failed to delete filed %s/%s of key %s from redis: %s", namespace, name, key, err.Error())
		return err
	}
	return nil
}

