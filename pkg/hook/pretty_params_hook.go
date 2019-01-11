package hook

import (
	"bytes"
	"fmt"
	"github.com/imdario/mergo"
	"github.com/sirupsen/logrus"
	"io"
	"walm/pkg/release"
)

func ProcessPrettyParams(releaseRequest *release.ReleaseRequest) {
	defaultConfigValue := releaseRequest.ConfigValues
	commonAppValues := make(map[string]interface{}, 0)
	for _, roleConfig := range releaseRequest.ReleasePrettyParams.CommonConfig.Roles {
		commonAppRoleValues := make(map[string]interface{}, 0)
		if roleConfig.Replicas == -1 {
			mergo.Merge(&commonAppRoleValues, map[string]interface{}{
				"replicas": roleConfig.Replicas,
			})
		}

		mergo.Merge(&commonAppRoleValues, map[string]interface{}{
			"resources": map[string]interface{}{
				"memory_request": roleConfig.RoleResourceConfig.MemoryRequest,
			},
		})
		mergo.Merge(&commonAppRoleValues, map[string]interface{}{
			"resources": map[string]interface{}{
				"memory_limit": roleConfig.RoleResourceConfig.MemoryLimit,
			},
		})
		mergo.Merge(&commonAppRoleValues, map[string]interface{}{
			"resources": map[string]interface{}{
				"cpu_request": roleConfig.RoleResourceConfig.CpuRequest,
			},
		})
		mergo.Merge(&commonAppRoleValues, map[string]interface{}{
			"resources": map[string]interface{}{
				"cpu_limit": roleConfig.RoleResourceConfig.CpuLimit,
			},
		})
		mergo.Merge(&commonAppRoleValues, map[string]interface{}{
			"resources": map[string]interface{}{
				"gpu_request": roleConfig.RoleResourceConfig.GpuRequest,
			},
		})
		mergo.Merge(&commonAppRoleValues, map[string]interface{}{
			"resources": map[string]interface{}{
				"gpu_limit": roleConfig.RoleResourceConfig.GpuLimit,
			},
		})

		commonAppRoleStorage := make(map[string]interface{}, 0)
		for _, storageConfig := range roleConfig.RoleResourceConfig.ResourceStorageList {
			storageConfigValues := make(map[string]interface{}, 0)
			if storageConfig.StorageType == "tosDisk" {
				storageConfigValues["storageClass"] = storageConfig.StorageClass
				storageConfigValues["size"] = storageConfig.Size
				storageConfigValues["accessMode"] = storageConfig.AccessMode
				commonAppRoleStorage[storageConfig.Name] = storageConfigValues
			}
			if storageConfig.StorageType == "pvc" {
				storageConfigValues["storageClass"] = storageConfig.StorageClass
				storageConfigValues["size"] = storageConfig.Size
				storageConfigValues["accessModes"] = storageConfig.AccessModes
				commonAppRoleStorage[storageConfig.Name] = storageConfigValues
			}
		}
		if len(commonAppRoleStorage) > 0 {
			mergo.Merge(&commonAppRoleValues, map[string]interface{}{
				"resources": map[string]interface{}{
					"storage": commonAppRoleStorage,
				},
			})
		}

		for _, roleBaseConfig := range roleConfig.RoleBaseConfig {
			commonAppRoleValues[roleBaseConfig.ValueName] = roleBaseConfig.DefaultValue
		}

		commonAppValues[roleConfig.Name] = commonAppRoleValues
	}

	logrus.Debugf("commonAppValues %+v\n", commonAppValues)
	if len(commonAppValues) > 0 {
		err := mergo.Merge(&defaultConfigValue, map[string]interface{}{
			"App": commonAppValues,
		}, mergo.WithOverride)
		if err != nil {
			logrus.Errorf("mergo.Merge error src %+v, dest %+v, err %+v\n", commonAppValues, defaultConfigValue, err)
		}
	}

	if releaseRequest.ReleasePrettyParams.AdvanceConfig != nil {
		for _, baseConfig := range releaseRequest.ReleasePrettyParams.AdvanceConfig {
			logrus.Infof("### %v", baseConfig)
			configValues := make(map[string]interface{}, 0)
			mapKey(baseConfig.ValueName, baseConfig.DefaultValue, configValues)

			err := mergo.Merge(&defaultConfigValue, configValues, mergo.WithOverride)
			if err != nil {
				logrus.Errorf("mergo.Merge error src %+v, dest %+v, err %+v\n", configValues, defaultConfigValue, err)
			}
		}
	}

	if releaseRequest.ReleasePrettyParams.TranswarpBaseConfig != nil {
		for _, baseConfig := range releaseRequest.ReleasePrettyParams.TranswarpBaseConfig {
			configValues := make(map[string]interface{}, 0)
			mapKey(baseConfig.ValueName, baseConfig.DefaultValue, configValues)

			err := mergo.Merge(&defaultConfigValue, configValues, mergo.WithOverride)
			if err != nil {
				logrus.Errorf("mergo.Merge error src %+v, dest %+v, err %+v\n", configValues, defaultConfigValue, err)
			}
		}
	}
}

func mapKey(key string, value interface{}, data map[string]interface{}) error {
	scanner := bytes.NewBufferString(key)
	keyName := ""
	var pMap map[string]interface{}
	pMap = data
	for {
		switch r, _, e := scanner.ReadRune(); {
		case e != nil:
			if e == io.EOF {
				pMap[keyName] = value
				keyName = ""
				return nil
			}
			return e
		case r == '[':
			if len(keyName) > 0 {
				pMap[keyName] = make(map[string]interface{}, 0)
				pMap = pMap[keyName].(map[string]interface{})
			}
			keyName = ""
			next, _, e := scanner.ReadRune()
			if next != '"' || e != nil {
				return fmt.Errorf("invalid key %s err %v", key, e)
			}
			for {
				next, _, e = scanner.ReadRune()
				if next == '"' || e != nil {
					next, _, e := scanner.ReadRune()
					if next != ']' || e != nil {
						return fmt.Errorf("invalid key %s err %v", key, e)
					} else {
						_, _, e = scanner.ReadRune();
						if e == io.EOF {
							pMap[keyName] = value
							return nil
						} else if len(keyName) > 0 {
							pMap[keyName] = make(map[string]interface{}, 0)
							pMap = pMap[keyName].(map[string]interface{})
							scanner.UnreadRune()
						}
						keyName = ""
						break
					}
				}
				keyName += string(next)
			}
		case r == '.':
			if len(keyName) > 0 {
				pMap[keyName] = make(map[string]interface{}, 0)
				pMap = pMap[keyName].(map[string]interface{})
			}
			keyName = ""
		default:
			keyName += string(r)
		}
	}
}
