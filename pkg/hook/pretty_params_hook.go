package hook

import (
	"github.com/imdario/mergo"
	"github.com/sirupsen/logrus"
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
			"resources": map[string]interface{} {
				"memory_request": roleConfig.RoleResourceConfig.MemoryRequest,
			},
		})
		mergo.Merge(&commonAppRoleValues, map[string]interface{}{
			"resources": map[string]interface{} {
				"memory_limit": roleConfig.RoleResourceConfig.MemoryLimit,
			},
		})
		mergo.Merge(&commonAppRoleValues, map[string]interface{}{
			"resources": map[string]interface{} {
				"cpu_request": roleConfig.RoleResourceConfig.CpuRequest,
			},
		})
		mergo.Merge(&commonAppRoleValues, map[string]interface{}{
			"resources": map[string]interface{} {
				"cpu_limit": roleConfig.RoleResourceConfig.CpuLimit,
			},
		})
		mergo.Merge(&commonAppRoleValues, map[string]interface{}{
			"resources": map[string]interface{} {
				"gpu_request": roleConfig.RoleResourceConfig.GpuRequest,
			},
		})
		mergo.Merge(&commonAppRoleValues, map[string]interface{}{
			"resources": map[string]interface{} {
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
				"resources": map[string]interface{} {
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
		err := mergo.Merge(&defaultConfigValue, map[string]interface{} {
			"App": commonAppValues,
		}, mergo.WithOverride)
		if err != nil {
			logrus.Errorf("mergo.Merge error src %+v, dest %+v, err %+v\n", commonAppValues, defaultConfigValue, err)
		}
	}

//		roleConfig.Name; "App.<name>"
//		roleConfig.Replicas; "App.<name>.replicas"
//		for _, roleBaseConfig := range roleConfig.RoleBaseConfig {
//			roleBaseConfig.DefaultValue
//			roleBaseConfig.ValueDescription
//			roleBaseConfig.ValueName
//			roleBaseConfig.ValueType
//			"App.<name>.<valuename>: <defaultvalue>"
//		}
//		roleConfig.RoleResourceConfig.MemoryRequest "App.<name>.resources.memory_request"
//		roleConfig.RoleResourceConfig.MemoryLimit "App.<name>.resources.memory_limit"
//		roleConfig.RoleResourceConfig.CpuRequest "App.<name>.resources.cpu_request"
//		roleConfig.RoleResourceConfig.CpuLimit "App.<name>.resources.cpu_limit"
//		roleConfig.RoleResourceConfig.GpuRequest "App.<name>.resources.gpu_request"
//		roleConfig.RoleResourceConfig.GpuLimit "App.<name>.resources.gpu_limit"
//
//		for _, storageConfig := range roleConfig.RoleResourceConfig.ResourceStorageList {
//			storageConfig.StorageType == "tosDisk"
//			storageConfig.Name =
//			storageConfig.StorageClass = "App.<name>.resources.storage.<storageConfig.Name>.storageClass"
//			storageConfig.Size = "App.<name>.resources.storage.<storageConfig.Name>.size"
//			storageConfig.AccessMode = ""
//
//			storageConfig.StorageType == "tosPVC"
//			storageConfig.Name =
//			storageConfig.StorageClass = "App.<name>.resources.storage.<storageConfig.Name>.storageClass"
//			storageConfig.Size = "App.<name>.resources.storage.<storageConfig.Name>.size"
//			storageConfig.AccessModes = ""
//		}
//	}
}
