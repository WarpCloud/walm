package release

import (
	"github.com/tidwall/sjson"
	"github.com/sirupsen/logrus"
	"encoding/json"
	"WarpCloud/walm/pkg/util"
	"WarpCloud/walm/pkg/k8s/utils"
)

type MetaInfoParams struct {
	Params            []*MetaCommonConfigValue `json:"params"`
	Roles             []*MetaRoleConfigValue   `json:"roles"`
	CustomChartParams map[string]string        `json:"customParams"`
}

func (metaInfoParams *MetaInfoParams) BuildConfigValues(metaInfo *ChartMetaInfo) (configValues map[string]interface{}, err error) {
	configValues = map[string]interface{}{}
	if metaInfo == nil {
		return
	}

	mapping := map[string]interface{}{}
	err = buildCommonConfigArrayValues(mapping, metaInfoParams.Params, metaInfo.ChartParams)
	if err != nil {
		return
	}
	buildRoleConfigArrayValues(mapping, metaInfoParams.Roles, metaInfo.ChartRoles)

	jsonStr := "{}"
	for key, value := range mapping {
		jsonStr, err = sjson.Set(jsonStr, key, value)
		if err != nil {
			logrus.Errorf("failed to set json : %s", err.Error())
			return
		}
	}

	err = json.Unmarshal([]byte(jsonStr), &configValues)
	if err != nil {
		logrus.Errorf("failed to unmarshal config values : %s", err.Error())
		return nil, err
	}

	return
}

func buildCommonConfigArrayValues(mapping map[string]interface{}, commonConfigValues []*MetaCommonConfigValue, commonConfigs []*MetaCommonConfig) (err error) {
	commonConfigsMap := convertCommonConfigArrayToMap(commonConfigs)
	jsonStr := "{}"
	for _, commonConfigValue := range commonConfigValues {
		if commonConfig, ok := commonConfigsMap[commonConfigValue.Name]; ok {
			if commonConfig.MapKey != "" {
				if commonConfigValue.Value == "" {
					jsonStr, err = sjson.Set(jsonStr, commonConfig.MapKey, nil)
					if err != nil {
						return err
					}
				} else {
					jsonStr, err = sjson.SetRaw(jsonStr, commonConfig.MapKey, commonConfigValue.Value)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	if jsonStr != "{}" {
		m := map[string]interface{}{}
		err = json.Unmarshal([]byte(jsonStr), &m)
		if err != nil {
			return err
		}
		util.MergeValues(mapping, m, false)
	}

	return nil
}

func convertCommonConfigArrayToMap(configs []*MetaCommonConfig) map[string]*MetaCommonConfig {
	configMap := map[string]*MetaCommonConfig{}
	for _, config := range configs {
		if config != nil {
			configMap[config.Name] = config
		}
	}
	return configMap
}

func buildRoleConfigArrayValues(mapping map[string]interface{}, roleConfigValues []*MetaRoleConfigValue, roleConfigs []*MetaRoleConfig) {
	roleConfigMap := map[string]*MetaRoleConfig{}
	for _, roleConfig := range roleConfigs {
		if roleConfig != nil {
			roleConfigMap[roleConfig.Name] = roleConfig
		}
	}

	for _, roleConfigValue := range roleConfigValues {
		if roleConfigValue == nil {
			continue
		}
		if roleConfig, ok := roleConfigMap[roleConfigValue.Name]; ok {
			roleConfigValue.BuildConfigValue(mapping, roleConfig)
		}
	}
}

type MetaCommonConfigValue struct {
	Name  string `json:"name" description:"config name"`
	Type  string `json:"type" description:"config type"`
	Value string `json:"value" description:"config value : json raw message"`
}

type MetaRoleConfigValue struct {
	Name                    string                   `json:"name"`
	RoleBaseConfigValue     *MetaRoleBaseConfigValue `json:"baseConfig"`
	RoleResourceConfigValue *MetaResourceConfigValue `json:"resources"`
	// TODO healthChecks
}

func (roleConfigValue *MetaRoleConfigValue) BuildConfigValue(mapping map[string]interface{}, roleConfig *MetaRoleConfig) {
	if roleConfig == nil {
		return
	}
	if roleConfigValue.RoleBaseConfigValue != nil {
		roleConfigValue.RoleBaseConfigValue.BuildConfigValue(mapping, roleConfig.RoleBaseConfig)
	}
	if roleConfigValue.RoleResourceConfigValue != nil {
		roleConfigValue.RoleResourceConfigValue.BuildConfigValue(mapping, roleConfig.RoleResourceConfig)
	}
}

type MetaRoleBaseConfigValue struct {
	Image          *string                  `json:"image" description:"role image"`
	Priority       *int64                   `json:"priority" description:"role priority"`
	Replicas       *int64                   `json:"replicas" description:"role replicas"`
	Env            []MetaEnv                `json:"env" description:"role env list"`
	UseHostNetwork *bool                    `json:"useHostNetwork" description:"whether role use host network"`
	Others         []*MetaCommonConfigValue `json:"others" description:"role other configs"`
}

func (roleBaseConfigValue *MetaRoleBaseConfigValue) BuildConfigValue(mapping map[string]interface{}, roleBaseConfig *MetaRoleBaseConfig) error {
	if roleBaseConfig == nil {
		return nil
	}

	if roleBaseConfig.Image != nil && roleBaseConfigValue.Image != nil {
		mapping[roleBaseConfig.Image.MapKey] = *roleBaseConfigValue.Image
	}
	if roleBaseConfig.UseHostNetwork != nil && roleBaseConfigValue.UseHostNetwork != nil {
		mapping[roleBaseConfig.UseHostNetwork.MapKey] = *roleBaseConfigValue.UseHostNetwork
	}
	if roleBaseConfig.Priority != nil && roleBaseConfigValue.Priority != nil {
		mapping[roleBaseConfig.Priority.MapKey] = *roleBaseConfigValue.Priority
	}
	if roleBaseConfig.Env != nil && len(roleBaseConfigValue.Env) > 0 {
		mapping[roleBaseConfig.Env.MapKey] = roleBaseConfigValue.Env
	}
	if roleBaseConfig.Replicas != nil && roleBaseConfigValue.Replicas != nil {
		mapping[roleBaseConfig.Replicas.MapKey] = *roleBaseConfigValue.Replicas
	}

	return buildCommonConfigArrayValues(mapping, roleBaseConfigValue.Others, roleBaseConfig.Others)
}

type MetaResourceConfigValue struct {
	LimitsMemory     *int64                            `json:"limitsMemory" description:"resource memory limit"`
	LimitsCpu        *float64                          `json:"limitsCpu" description:"resource cpu limit"`
	LimitsGpu        *float64                          `json:"limitsGpu" description:"resource gpu limit"`
	RequestsMemory   *int64                            `json:"requestsMemory" description:"resource memory request"`
	RequestsCpu      *float64                          `json:"requestsCpu" description:"resource cpu request"`
	RequestsGpu      *float64                          `json:"requestsGpu" description:"resource gpu request"`
	StorageResources []*MetaResourceStorageConfigValue `json:"storageResources" description:"resource storage request"`
}

func (resourceConfigValue *MetaResourceConfigValue) BuildConfigValue(mapping map[string]interface{}, resourceConfig *MetaResourceConfig) {
	if resourceConfig == nil {
		return
	}

	if resourceConfigValue.LimitsMemory != nil && resourceConfig.LimitsMemory != nil {
		mapping[resourceConfig.LimitsMemory.MapKey] = convertResourceBinaryIntByUnit(resourceConfigValue.LimitsMemory, utils.K8sResourceMemoryUnit)
	}
	if resourceConfigValue.LimitsGpu != nil && resourceConfig.LimitsGpu != nil {
		mapping[resourceConfig.LimitsGpu.MapKey] = convertResourceDecimalFloat(resourceConfigValue.LimitsGpu)
	}
	if resourceConfigValue.LimitsCpu != nil && resourceConfig.LimitsCpu != nil {
		mapping[resourceConfig.LimitsCpu.MapKey] = convertResourceDecimalFloat(resourceConfigValue.LimitsCpu)
	}
	if resourceConfigValue.RequestsMemory != nil && resourceConfig.RequestsMemory != nil {
		mapping[resourceConfig.RequestsMemory.MapKey] = convertResourceBinaryIntByUnit(resourceConfigValue.RequestsMemory, utils.K8sResourceMemoryUnit)
	}
	if resourceConfigValue.RequestsGpu != nil && resourceConfig.RequestsGpu != nil {
		mapping[resourceConfig.RequestsGpu.MapKey] = convertResourceDecimalFloat(resourceConfigValue.RequestsGpu)
	}
	if resourceConfigValue.RequestsCpu != nil && resourceConfig.RequestsCpu != nil {
		mapping[resourceConfig.RequestsCpu.MapKey] = convertResourceDecimalFloat(resourceConfigValue.RequestsCpu)
	}

	buildResourceStorageArrayValues(mapping, resourceConfigValue.StorageResources, resourceConfig.StorageResources)
}

type MetaResourceStorageConfigValue struct {
	Name  string               `json:"name" description:"config name"`
	Value *MetaResourceStorage `json:"value" description:"config value"`
}

func buildResourceStorageArrayValues(mapping map[string]interface{}, resourceStorageConfigValues []*MetaResourceStorageConfigValue, resourceStorageConfigs []*MetaResourceStorageConfig) {
	resourceStorageConfigMap := map[string]*MetaResourceStorageConfig{}
	for _, resourceStorageConfig := range resourceStorageConfigs {
		if resourceStorageConfig != nil {
			resourceStorageConfigMap[resourceStorageConfig.Name] = resourceStorageConfig
		}
	}

	for _, resourceStorageConfigValue := range resourceStorageConfigValues {
		if resourceStorageConfigValue.Value == nil {
			continue
		}
		if resourceStorageConfig, ok := resourceStorageConfigMap[resourceStorageConfigValue.Name]; ok {
			resourceStorageWithStringSize := MetaResourceStorageWithStringSize{
				ResourceStorage: resourceStorageConfigValue.Value.ResourceStorage,
				Size:            convertResourceBinaryIntByUnit(&resourceStorageConfigValue.Value.Size, utils.K8sResourceStorageUnit),
			}
			mapping[resourceStorageConfig.MapKey] = resourceStorageWithStringSize
		}
	}
}
