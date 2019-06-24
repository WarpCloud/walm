package metainfo

import (
	"k8s.io/helm/pkg/walm"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
)

// chart metainfo
type ChartMetaInfo struct {
	FriendlyName          string                     `json:"friendlyName" description:"friendlyName"`
	Categories            []string                   `json:"categories" description:"categories"`
	ChartDependenciesInfo []*ChartDependencyMetaInfo `json:"dependencies" description:"dependency metainfo"`
	ChartRoles            []*MetaRoleConfig          `json:"roles"`
	ChartParams           []*MetaCommonConfig        `json:"params"`
	Plugins               []*walm.WalmPlugin         `json:"plugins"`
	CustomChartParams     map[string]string          `json:"customParams"`
}

func (chartMetaInfo *ChartMetaInfo) CheckMetainfoValidate(valuesStr string) ([]*MetaConfigTestSet, error) {

	var err error
	// friendlyName
	if !(len(chartMetaInfo.FriendlyName) > 0) {
		err = errors.Errorf("field friendlyName required")
		return nil, err
	}

	// dependencies
	if len(chartMetaInfo.ChartDependenciesInfo) > 0 {
		for _, dependency := range chartMetaInfo.ChartDependenciesInfo {
			if dependency.Name == "" || dependency.MinVersion == "" || dependency.MaxVersion == "" || dependency.AliasConfigVar == "" ||
				reflect.TypeOf(dependency.DependencyOptional).String() != "bool" {
				err = errors.Errorf("Name, MinVersion, MaxVersion, [aliasConfigVar,omitempty], dependencyOptional all required in field dependencies")
				return nil, err
			}
		}
	}

	var configSets []*MetaConfigTestSet

	// roles
	if len(chartMetaInfo.ChartRoles) > 0 {
		for index, chartRole := range chartMetaInfo.ChartRoles {

			if chartRole.Name == "" {
				err = errors.Errorf("name required in field roles[%d]", index)
				return nil, err
			}

			if chartRole.Type == "" {
				err = errors.Errorf("type required in field roles[%d]", index)
				return nil, err
			}

			if chartRole.RoleBaseConfig != nil {

				baseConfig := chartRole.RoleBaseConfig
				if baseConfig.Image != nil {
					if baseConfig.Image.MapKey == "" {
						err = errors.Errorf("mapKey required in field roles[%d].baseConfig.image", index)
						return nil, err
					}

					configSet := &MetaConfigTestSet{
						MapKey:   baseConfig.Image.MapKey,
						Required: baseConfig.Image.Required,
						Type:     "string",
					}
					configSets = append(configSets, configSet)
				}

				if baseConfig.Priority != nil {
					if baseConfig.Priority.MapKey == "" {
						err = errors.Errorf("mapKey required in field roles[%d].baseConfig.priority", index)
						return nil, err
					}
					configSet := &MetaConfigTestSet{
						MapKey:   baseConfig.Priority.MapKey,
						Required: baseConfig.Priority.Required,
						Type:     "int",
					}
					configSets = append(configSets, configSet)
				}

				if baseConfig.Replicas != nil {
					if baseConfig.Replicas.MapKey == "" {
						err = errors.Errorf("mapKey required in field roles[%d].baseConfig.replicas", index)
						return nil, err
					}
					configSet := &MetaConfigTestSet{
						MapKey:   baseConfig.Replicas.MapKey,
						Required: baseConfig.Replicas.Required,
						Type:     "int",
					}
					configSets = append(configSets, configSet)
				}

				if baseConfig.Env != nil {
					if baseConfig.Env.MapKey == "" {
						err = errors.Errorf("mapKey required in field roles[%d].baseConfig.env", index)
						return nil, err
					}
					configSet := &MetaConfigTestSet{
						MapKey:   baseConfig.Env.MapKey,
						Required: baseConfig.Env.Required,
						Type:     "env",
					}
					configSets = append(configSets, configSet)
				}

				if baseConfig.UseHostNetwork != nil {
					if baseConfig.UseHostNetwork.MapKey == "" {
						err = errors.Errorf("mapKey required in field roles[%d].baseConfig.useHostNetwork", index)
						return nil, err
					}
					configSet := &MetaConfigTestSet{
						MapKey:   baseConfig.UseHostNetwork.MapKey,
						Required: baseConfig.UseHostNetwork.Required,
						Type:     "boolean",
					}
					configSets = append(configSets, configSet)
				}

				if len(baseConfig.Others) > 0 {
					for otherIndex, otherConfig := range baseConfig.Others {
						if otherConfig.Name == "" {
							err = errors.Errorf("name required in field roles[%d].baseConfig.others[%d]", index, otherIndex)
							return nil, err
						}
						if otherConfig.MapKey == "" {
							err = errors.Errorf("mapKey required in field roles[%d].baseConfig.others[%d]", index, otherIndex)
							return nil, err
						}
						switch otherConfig.Type {
						case "boolean", "int", "float", "string", "yaml", "json", "kvPair", "text":
						default:
							err = errors.Errorf("type <%s> not support in field roles[%d].baseConfig.others[%d]", otherConfig.Type, index, otherIndex)
							return nil, err
						}

						configSet := &MetaConfigTestSet{
							MapKey:   otherConfig.MapKey,
							Required: otherConfig.Required,
							Type:     otherConfig.Type,
						}
						configSets = append(configSets, configSet)
					}
				}
			}
			resourceConfig := chartRole.RoleResourceConfig
			if resourceConfig != nil {

				if resourceConfig.LimitsCpu != nil {
					if resourceConfig.LimitsCpu.MapKey == "" {
						err = errors.Errorf("mapKey required in field roles[%d].resources.limitsCpu", index)
						return nil, err
					}
					configSet := &MetaConfigTestSet{
						MapKey:   resourceConfig.LimitsCpu.MapKey,
						Required: resourceConfig.LimitsCpu.Required,
						Type:     "float",
					}
					configSets = append(configSets, configSet)
				}

				if resourceConfig.LimitsMemory != nil {
					if resourceConfig.LimitsMemory.MapKey == "" {
						err = errors.Errorf("mapKey required in field roles[%d].resources.LimitsMemory", index)
						return nil, err
					}

					configSet := &MetaConfigTestSet{
						MapKey:   resourceConfig.LimitsMemory.MapKey,
						Required: resourceConfig.LimitsMemory.Required,
						Type:     "string",
					}
					configSets = append(configSets, configSet)
				}
				if resourceConfig.LimitsGpu != nil {
					if resourceConfig.LimitsGpu.MapKey == "" {
						err = errors.Errorf("mapKey required in field roles[%d].resources.LimitsGpu", index)
						return nil, err
					}
					configSet := &MetaConfigTestSet{
						MapKey:   resourceConfig.LimitsGpu.MapKey,
						Required: resourceConfig.LimitsGpu.Required,
						Type:     "float",
					}
					configSets = append(configSets, configSet)
				}
				if resourceConfig.RequestsMemory != nil {
					if resourceConfig.RequestsMemory.MapKey == "" {
						err = errors.Errorf("mapKey required in field roles[%d].resources.RequestsMemory", index)
						return nil, err
					}
					configSet := &MetaConfigTestSet{
						MapKey:   resourceConfig.RequestsMemory.MapKey,
						Required: resourceConfig.RequestsMemory.Required,
						Type:     "string",
					}
					configSets = append(configSets, configSet)
				}
				if resourceConfig.RequestsCpu != nil {
					if resourceConfig.RequestsCpu.MapKey == "" {
						err = errors.Errorf("mapKey required in field roles[%d].resources.RequestsCpu", index)
						return nil, err
					}
					configSet := &MetaConfigTestSet{
						MapKey:   resourceConfig.RequestsCpu.MapKey,
						Required: resourceConfig.RequestsCpu.Required,
						Type:     "float",
					}
					configSets = append(configSets, configSet)
				}
				if resourceConfig.RequestsGpu != nil {
					if resourceConfig.RequestsGpu.MapKey == "" {
						err = errors.Errorf("mapKey required in field roles[%d].resources.RequestsGpu", index)
						return nil, err
					}
					configSet := &MetaConfigTestSet{
						MapKey:   resourceConfig.RequestsGpu.MapKey,
						Required: resourceConfig.RequestsGpu.Required,
						Type:     "float",
					}
					configSets = append(configSets, configSet)
				}

				if len(resourceConfig.StorageResources) > 0 {
					for storageIndex, storageResource := range resourceConfig.StorageResources {
						if storageResource.Name == "" {
							err = errors.Errorf("name required in field roles[%d].resources.storageResources[%d]", index, storageIndex)
							return nil, err
						}
						if storageResource.MapKey == "" {
							err = errors.Errorf("mapKey required in field roles[%d].resources.storageResources[%d]", index, storageIndex)
							return nil, err
						}
						configSet := &MetaConfigTestSet{
							MapKey:   storageResource.MapKey,
							Required: storageResource.Required,
							Type:     storageResource.Type,
						}
						configSets = append(configSets, configSet)
					}
				}
			}
		}
	}

	// params
	params := chartMetaInfo.ChartParams
	if len(params) > 0 {
		for paramIndex, param := range params {
			if param.Name == "" {
				err = errors.Errorf("name required in field params[%d]", paramIndex)
				return nil, err
			}
			if param.MapKey == "" {
				err = errors.Errorf("mapKey required in field params[%d]", paramIndex)
				return nil, err
			}
			switch param.Type {
			case "boolean", "int", "float", "string", "yaml", "json", "kvPair", "text":
			default:
				err = errors.Errorf("type <%s> not support in field params[%d]", param.Type, paramIndex)
				return nil, err
			}
			configSet := &MetaConfigTestSet{
				MapKey:   param.MapKey,
				Required: param.Required,
				Type:     param.Type,
			}
			configSets = append(configSets, configSet)
		}
	}

	// plugins
	plugins := chartMetaInfo.Plugins
	if len(plugins) > 0 {

		for pluginIndex, plugin := range plugins {
			if plugin.Name == "" {
				err = errors.Errorf("name required in field plugins[%d]", pluginIndex)
				return nil, err
			}
			if plugin.Version == "" {
				err = errors.Errorf("version required in field plugins[%d]", pluginIndex)
				return nil, err
			}
			if plugin.Args == "" {
				err = errors.Errorf("args required in field plugins[%d]", pluginIndex)
				return nil, err
			}
		}
	}
	return configSets, err
}

func (chartMetaInfo *ChartMetaInfo) CheckParamsInValues(valuesStr string, configSets []*MetaConfigTestSet) error {

	var err error
	for _, configSet := range configSets {
		/*
		boolean --> True, False
		string --> String
		int, float --> Number
		interface, {}, [], Null --> JSON, Null
		*/
		result := gjson.Get(valuesStr, configSet.MapKey)
		if result.Exists() {

			switch configSet.Type {
			case "boolean":
				if !(result.Type.String() == "True" || result.Type.String() == "False") {
					return errors.Errorf("%s Type error in values.yaml, %s expected", configSet.MapKey, configSet.Type)
				}
			case "string":
				if result.Type.String() != "String" {
					return errors.Errorf("%s Type error in values.yaml, %s expected", configSet.MapKey, configSet.Type)
				}
				if strings.HasSuffix(configSet.MapKey, ".memory") || strings.HasSuffix(configSet.MapKey, ".memory_request") ||
					strings.HasSuffix(configSet.MapKey, ".memory_limit") {

					if !(strings.HasSuffix(result.Str, "Mi") || strings.HasSuffix(result.Str, "Gi")) {
						return errors.Errorf("%s Format error in values.yaml, eg: 4Gi, 400Mi expected", configSet.MapKey)
					}
				}

			case "int":
				_, err = strconv.Atoi(result.Raw)
				if err != nil {
					return errors.Errorf("%s Type error in values.yaml, %s expected", configSet.MapKey, configSet.Type)
				}
			case "float":
				if result.Type.String() != "Number" {
					return errors.Errorf("%s Type error in values.yaml, %s expected", configSet.MapKey, configSet.Type)
				}
			case "number":
				if result.Type.String() != "Number" {
					return errors.Errorf("%s Type error in values.yaml, %s expected", configSet.MapKey, configSet.Type)
				}
			case "text":
				if result.Type.String() != "String" {
					return errors.Errorf("%s Type error in values.yaml, %s expected", configSet.MapKey, configSet.Type)
				}
			case "kvPair":
				if result.Type.String() != "JSON" {
					return errors.Errorf("%s Type error in values.yaml, %s expected", configSet.MapKey, configSet.Type)
				}
			case "":
			default:
				if result.Type.String() == "Null" {
					break
				}
				if result.Type.String() != "JSON" {
					return errors.Errorf("%s Type error in values.yaml, %s expected", configSet.MapKey, configSet.Type)
				}
			}
		} else {
			if configSet.Required {
				return errors.Errorf("%s not exist in values.yaml", configSet.MapKey)
			}
		}
	}

	return nil
}

func (chartMetaInfo *ChartMetaInfo) BuildDefaultValue(jsonStr string) {
	if jsonStr != "" {
		for _, chartParam := range chartMetaInfo.ChartParams {
			if chartParam != nil {
				chartParam.BuildDefaultValue(jsonStr)
			}
		}
		for _, chartRole := range chartMetaInfo.ChartRoles {
			if chartRole != nil {
				chartRole.BuildDefaultValue(jsonStr)
			}
		}
	}
}

func (chartMetaInfo *ChartMetaInfo) BuildMetaInfoParams(configValues map[string]interface{}) (*MetaInfoParams, error) {
	if len(configValues) > 0 {
		jsonBytes, err := json.Marshal(configValues)
		if err != nil {
			logrus.Errorf("failed to marshal computed values : %s", err.Error())
			return nil, err
		}
		jsonStr := string(jsonBytes)
		metaInfoValues := &MetaInfoParams{}

		for _, chartParam := range chartMetaInfo.ChartParams {
			metaInfoValues.Params = append(metaInfoValues.Params, chartParam.BuildCommonConfigValue(jsonStr))
		}

		for _, chartRole := range chartMetaInfo.ChartRoles {
			if chartRole != nil {
				metaInfoValues.Roles = append(metaInfoValues.Roles, chartRole.BuildRoleConfigValue(jsonStr))
			}
		}
		metaInfoValues.CustomChartParams = chartMetaInfo.CustomChartParams
		return metaInfoValues, nil
	}
	return nil, nil
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

type MetaRoleConfig struct {
	Name                  string                 `json:"name"`
	Description           string                 `json:"description"`
	Type                  string                 `json:"type"`
	RoleBaseConfig        *MetaRoleBaseConfig    `json:"baseConfig"`
	RoleResourceConfig    *MetaResourceConfig    `json:"resources"`
	RoleHealthCheckConfig *MetaHealthCheckConfig `json:"healthChecks"`
}

func (roleConfig *MetaRoleConfig) BuildDefaultValue(jsonStr string) {
	if roleConfig.RoleBaseConfig != nil {
		roleConfig.RoleBaseConfig.BuildDefaultValue(jsonStr)
	}
	if roleConfig.RoleResourceConfig != nil {
		roleConfig.RoleResourceConfig.BuildDefaultValue(jsonStr)
	}
}

func (roleConfig *MetaRoleConfig) BuildRoleConfigValue(jsonStr string) *MetaRoleConfigValue {
	roleConfigValue := &MetaRoleConfigValue{Name: roleConfig.Name}
	if roleConfig.RoleBaseConfig != nil {
		roleConfigValue.RoleBaseConfigValue = roleConfig.RoleBaseConfig.BuildRoleBaseConfigValue(jsonStr)
	}
	if roleConfig.RoleResourceConfig != nil {
		roleConfigValue.RoleResourceConfigValue = roleConfig.RoleResourceConfig.BuildResourceConfigValue(jsonStr)
	}
	return roleConfigValue
}

type MetaRoleBaseConfig struct {
	Image          *MetaStringConfig   `json:"image" description:"role image"`
	Priority       *MetaIntConfig      `json:"priority" description:"role priority"`
	Replicas       *MetaIntConfig      `json:"replicas" description:"role replicas"`
	Env            *MetaEnvConfig      `json:"env" description:"role env list"`
	UseHostNetwork *MetaBoolConfig     `json:"useHostNetwork" description:"whether role use host network"`
	Others         []*MetaCommonConfig `json:"others" description:"role other configs"`
}

func (roleBaseConfig *MetaRoleBaseConfig) BuildDefaultValue(jsonStr string) {
	if roleBaseConfig.Image != nil {
		roleBaseConfig.Image.BuildDefaultValue(jsonStr)
	}
	if roleBaseConfig.Replicas != nil {
		roleBaseConfig.Replicas.BuildDefaultValue(jsonStr)
	}
	if roleBaseConfig.Env != nil {
		roleBaseConfig.Env.BuildDefaultValue(jsonStr)
	}
	if roleBaseConfig.UseHostNetwork != nil {
		roleBaseConfig.UseHostNetwork.BuildDefaultValue(jsonStr)
	}
	if roleBaseConfig.Priority != nil {
		roleBaseConfig.Priority.BuildDefaultValue(jsonStr)
	}
	for _, config := range roleBaseConfig.Others {
		if config != nil {
			config.BuildDefaultValue(jsonStr)
		}
	}
}

func (roleBaseConfig *MetaRoleBaseConfig) BuildRoleBaseConfigValue(jsonStr string) *MetaRoleBaseConfigValue {
	roleBaseConfigValue := &MetaRoleBaseConfigValue{}
	if roleBaseConfig.Image != nil {
		image := roleBaseConfig.Image.BuildStringConfigValue(jsonStr)
		roleBaseConfigValue.Image = &image
	}
	if roleBaseConfig.Replicas != nil {
		replicas := roleBaseConfig.Replicas.BuildIntConfigValue(jsonStr)
		roleBaseConfigValue.Replicas = &replicas
	}
	if roleBaseConfig.Env != nil {
		roleBaseConfigValue.Env = roleBaseConfig.Env.BuildEnvConfigValue(jsonStr)
	}
	if roleBaseConfig.UseHostNetwork != nil {
		useHostNetwork := roleBaseConfig.UseHostNetwork.BuildBoolConfigValue(jsonStr)
		roleBaseConfigValue.UseHostNetwork = &useHostNetwork
	}
	if roleBaseConfig.Priority != nil {
		priority := roleBaseConfig.Priority.BuildIntConfigValue(jsonStr)
		roleBaseConfigValue.Priority = &priority
	}
	for _, config := range roleBaseConfig.Others {
		if config != nil {
			roleBaseConfigValue.Others = append(roleBaseConfigValue.Others, config.BuildCommonConfigValue(jsonStr))
		}
	}
	return roleBaseConfigValue
}

type MetaResourceConfig struct {
	LimitsMemory     *MetaResourceIntConfig       `json:"limitsMemory" description:"resource memory limit"`
	LimitsCpu        *MetaResourceFloatConfig     `json:"limitsCpu" description:"resource cpu limit"`
	LimitsGpu        *MetaResourceFloatConfig     `json:"limitsGpu" description:"resource gpu limit"`
	RequestsMemory   *MetaResourceIntConfig       `json:"requestsMemory" description:"resource memory request"`
	RequestsCpu      *MetaResourceFloatConfig     `json:"requestsCpu" description:"resource cpu request"`
	RequestsGpu      *MetaResourceFloatConfig     `json:"requestsGpu" description:"resource gpu request"`
	StorageResources []*MetaResourceStorageConfig `json:"storageResources" description:"resource storage request"`
}

func (config *MetaResourceConfig) BuildDefaultValue(jsonStr string) {
	if config.LimitsMemory != nil {
		config.LimitsMemory.BuildDefaultValue(jsonStr)
	}
	if config.LimitsGpu != nil {
		config.LimitsGpu.BuildDefaultValue(jsonStr)
	}
	if config.LimitsCpu != nil {
		config.LimitsCpu.BuildDefaultValue(jsonStr)
	}
	if config.RequestsMemory != nil {
		config.RequestsMemory.BuildDefaultValue(jsonStr)
	}
	if config.RequestsGpu != nil {
		config.RequestsGpu.BuildDefaultValue(jsonStr)
	}
	if config.RequestsCpu != nil {
		config.RequestsCpu.BuildDefaultValue(jsonStr)
	}

	for _, storageConfig := range config.StorageResources {
		if storageConfig != nil {
			storageConfig.BuildDefaultValue(jsonStr)
		}
	}
}

func (config *MetaResourceConfig) BuildResourceConfigValue(jsonStr string) *MetaResourceConfigValue {
	resourceConfigValue := &MetaResourceConfigValue{}
	if config.LimitsMemory != nil {
		limitsMemory := config.LimitsMemory.BuildIntConfigValue(jsonStr)
		resourceConfigValue.LimitsMemory = &limitsMemory
	}
	if config.LimitsGpu != nil {
		limitsGpu := config.LimitsGpu.BuildFloatConfigValue(jsonStr)
		resourceConfigValue.LimitsGpu = &limitsGpu
	}
	if config.LimitsCpu != nil {
		limitsCpu := config.LimitsCpu.BuildFloatConfigValue(jsonStr)
		resourceConfigValue.LimitsCpu = &limitsCpu
	}
	if config.RequestsMemory != nil {
		requestsMemory := config.RequestsMemory.BuildIntConfigValue(jsonStr)
		resourceConfigValue.RequestsMemory = &requestsMemory
	}
	if config.RequestsGpu != nil {
		requestsGpu := config.RequestsGpu.BuildFloatConfigValue(jsonStr)
		resourceConfigValue.RequestsGpu = &requestsGpu
	}
	if config.RequestsCpu != nil {
		requestsCpu := config.RequestsCpu.BuildFloatConfigValue(jsonStr)
		resourceConfigValue.RequestsCpu = &requestsCpu
	}

	for _, storageConfig := range config.StorageResources {
		if storageConfig != nil {
			resourceConfigValue.StorageResources = append(resourceConfigValue.StorageResources, storageConfig.BuildStorageConfigValue(jsonStr))
		}
	}
	return resourceConfigValue
}

type MetaCommonConfig struct {
	Name         string `json:"name" description:"config name"`
	MapKey       string `json:"mapKey" description:"config map values.yaml key"`
	DefaultValue string `json:"defaultValue" description:"default value of mapKey"`
	Description  string `json:"description" description:"config description"`
	Type         string `json:"type" description:"config type"`
	Required     bool   `json:"required" description:"required"`
}

func (config *MetaCommonConfig) BuildDefaultValue(jsonStr string) {
	config.DefaultValue = config.BuildCommonConfigValue(jsonStr).Value
}

func (config *MetaCommonConfig) BuildCommonConfigValue(jsonStr string) *MetaCommonConfigValue {
	return &MetaCommonConfigValue{
		Name:  config.Name,
		Type:  config.Type,
		Value: gjson.Get(jsonStr, config.MapKey).Raw,
	}
}

type MetaHealthProbeConfig struct {
	Defined bool `json:"defined" description:"health check is defined"`
	Enable  bool `json:"enable" description:"enable health check"`
}

type MetaHealthCheckConfig struct {
	ReadinessProbe *MetaHealthProbeConfig `json:"readinessProbe"`
	LivenessProbe  *MetaHealthProbeConfig `json:"livenessProbe"`
}
