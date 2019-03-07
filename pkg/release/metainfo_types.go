package release

import (
	"k8s.io/helm/pkg/walm"
	"github.com/tidwall/sjson"
	"encoding/json"
	"github.com/sirupsen/logrus"
)

// chart metainfo
type ChartMetaInfo struct {
	FriendlyName          string                     `json:"friendlyName" description:"friendlyName"`
	Categories            []string                   `json:"categories" description:"categories"`
	ChartDependenciesInfo []*ChartDependencyMetaInfo `json:"dependencies" description:"dependency metainfo"`
	ChartRoles            []*MetaRoleConfig          `json:"roles"`
	ChartParams           []*MetaCommonConfig        `json:"params"`
	Plugins               []*walm.WalmPlugin         `json:"plugins"`
}

type MetaInfoParams struct {
	Params []*MetaCommonConfigValue `json:"params"`
	Roles  []*MetaRoleConfigValue   `json:"roles"`
}

func commonConfigMapping(params []*MetaCommonConfigValue, chartParams []*MetaCommonConfig) map[string]interface{} {
	mapping := map[string]interface{}{}

	chartParamsMap := map[string]*MetaCommonConfig{}
	for _, chartParam := range chartParams {
		chartParamsMap[chartParam.Name] = chartParam
	}

	for _, param := range params {
		if chartParam, ok := chartParamsMap[param.Name]; ok {
			mapping[chartParam.MapKey] = param.Value
		}
	}

	return mapping
}

func roleConfigMapping(roles []*MetaRoleConfigValue, chartRoles []*MetaRoleConfig) map[string]interface{} {
	mapping := map[string]interface{}{}
	chartRoleMap := map[string]*MetaRoleConfig{}
	for _, chartRole := range chartRoles {
		chartRoleMap[chartRole.Name] = chartRole
	}

	for _, role := range roles {
		if chartRole, ok := chartRoleMap[role.Name]; ok {
			chartRoleBaseConfigMap := map[string]*MetaCommonConfig{}
			for _, chartRoleBaseConfig := range chartRole.RoleBaseConfig {
				chartRoleBaseConfigMap[chartRoleBaseConfig.Name] = chartRoleBaseConfig
			}

			for _, roleBaseConfig := range role.RoleBaseConfig {
				if chartRoleBaseConfig, ok := chartRoleBaseConfigMap[roleBaseConfig.Name]; ok {
					mapping[chartRoleBaseConfig.MapKey] = roleBaseConfig.Value
				}
			}

			mapping[chartRole.RoleResourceConfig.LimitsCpuKey.MapKey] = role.RoleResourceConfig.LimitsCpuKey
			mapping[chartRole.RoleResourceConfig.LimitsGpuKey.MapKey] = role.RoleResourceConfig.LimitsGpuKey
			mapping[chartRole.RoleResourceConfig.RequestsCpuKey.MapKey] = role.RoleResourceConfig.RequestsCpuKey
			mapping[chartRole.RoleResourceConfig.RequestsGpuKey.MapKey] = role.RoleResourceConfig.RequestsGpuKey
			mapping[chartRole.RoleResourceConfig.RequestsMemoryKey.MapKey] = role.RoleResourceConfig.RequestsMemoryKey
			mapping[chartRole.RoleResourceConfig.LimitsMemoryKey.MapKey] = role.RoleResourceConfig.LimitsMemoryKey

			chartStorageConfigMap := map[string]*MetaCommonConfig{}
			for _, chartStorageConfig := range chartRole.RoleResourceConfig.StorageResources {
				chartStorageConfigMap[chartStorageConfig.Name] = chartStorageConfig
			}

			for _, storageConfig := range role.RoleResourceConfig.StorageResources {
				if chartStorageConfig, ok := chartStorageConfigMap[storageConfig.Name]; ok {
					mapping[chartStorageConfig.MapKey] = storageConfig.Value
				}
			}
		}
	}

	return mapping
}

func (metaInfoParams *MetaInfoParams) ToConfigValues(metaInfo *ChartMetaInfo) (configValues map[string]interface{}, err error) {
	jsonStr := "{}"
	mapping := commonConfigMapping(metaInfoParams.Params, metaInfo.ChartParams)
	for key, value := range mapping {
		jsonStr, err = sjson.Set(jsonStr, key, value)
		if err != nil {
			logrus.Errorf("failed to set json : %s", err.Error())
			return
		}
	}

	mapping = roleConfigMapping(metaInfoParams.Roles, metaInfo.ChartRoles)
	for key, value := range mapping {
		jsonStr, err = sjson.Set(jsonStr, key, value)
		if err != nil {
			logrus.Errorf("failed to set json : %s", err.Error())
			return
		}
	}

	configValues = map[string]interface{}{}
	err = json.Unmarshal([]byte(jsonStr), &configValues)
	if err != nil {
		logrus.Errorf("failed to unmarshal config values : %s", err.Error())
		return nil, err
	}

	return
}

type ChartDependencyMetaInfo struct {
	Name               string `json:"name,omitempty"`
	MinVersion         string `json:"minVersion"`
	MaxVersion         string `json:"maxVersion"`
	DependencyOptional bool   `json:"dependencyOptional"`
	AliasConfigVar     string `json:"aliasConfigVar,omitempty"`
	ChartName          string `json:"chartName"`
	DependencyType     string `json:"type"`
}

func (chartDependencyMetaInfo *ChartDependencyMetaInfo) AutoDependency() bool {
	if chartDependencyMetaInfo.Name == "" {
		return false
	}
	if chartDependencyMetaInfo.ChartName == "" {
		// 默认chartName = name
		return true
	}
	if chartDependencyMetaInfo.Name == chartDependencyMetaInfo.ChartName {
		return true
	}
	return false
}

type MetaCommonConfig struct {
	Name         string      `json:"name" description:"config name"`
	MapKey       string      `json:"mapKey" description:"config map values.yaml key"`
	DefaultValue interface{} `json:"defaultValue" description:"default value of mapKey"`
	Description  string      `json:"description" description:"config description"`
	Type         string      `json:"type" description:"config type"`
	Required     bool        `json:"required" description:"required"`
}

type MetaCommonConfigValue struct {
	Name  string      `json:"name" description:"config name"`
	Value interface{} `json:"value" description:"config value"`
}

type MetaResourceConfig struct {
	LimitsMemoryKey   *MetaCommonConfig    `json:"limitsMemoryKey" description:"resource memory limit"`
	LimitsCpuKey      *MetaCommonConfig    `json:"limitsCpuKey" description:"resource cpu limit"`
	LimitsGpuKey      *MetaCommonConfig    `json:"limitsGpuKey" description:"resource gpu limit"`
	RequestsMemoryKey *MetaCommonConfig    `json:"requestsMemoryKey" description:"resource memory request"`
	RequestsCpuKey    *MetaCommonConfig    `json:"requestsCpuKey" description:"resource cpu request"`
	RequestsGpuKey    *MetaCommonConfig    `json:"requestsGpuKey" description:"resource gpu request"`
	StorageResources  []*MetaCommonConfig `json:"storageResources" description:"resource storage request"`
}

type MetaResourceConfigValue struct {
	LimitsMemoryKey   interface{}              `json:"limitsMemoryKey" description:"resource memory limit"`
	LimitsCpuKey      interface{}              `json:"limitsCpuKey" description:"resource cpu limit"`
	LimitsGpuKey      interface{}              `json:"limitsGpuKey" description:"resource gpu limit"`
	RequestsMemoryKey interface{}              `json:"requestsMemoryKey" description:"resource memory request"`
	RequestsCpuKey    interface{}              `json:"requestsCpuKey" description:"resource cpu request"`
	RequestsGpuKey    interface{}              `json:"requestsGpuKey" description:"resource gpu request"`
	StorageResources  []*MetaCommonConfigValue `json:"storageResources" description:"resource storage request"`
}

type MetaHealthProbeConfig struct {
	Defined bool `json:"defined" description:"health check is defined"`
	Enable  bool `json:"enable" description:"enable health check"`
}

type MetaHealthCheckConfig struct {
	ReadinessProbe *MetaHealthProbeConfig `json:"readinessProbe"`
	LivenessProbe  *MetaHealthProbeConfig `json:"livenessProbe"`
}

type MetaRoleConfig struct {
	Name                  string                 `json:"name"`
	Description           string                 `json:"description"`
	RoleBaseConfig        []*MetaCommonConfig    `json:"baseConfig"`
	RoleResourceConfig    *MetaResourceConfig    `json:"resources"`
	RoleHealthCheckConfig *MetaHealthCheckConfig `json:"healthChecks"`
}

type MetaRoleConfigValue struct {
	Name               string                   `json:"name"`
	RoleBaseConfig     []*MetaCommonConfigValue `json:"baseConfig"`
	RoleResourceConfig *MetaResourceConfigValue `json:"resources"`
	// TODO healthChecks
}
