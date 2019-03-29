package metainfo

import (
	"k8s.io/helm/pkg/walm"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
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
	Name         string      `json:"name" description:"config name"`
	MapKey       string      `json:"mapKey" description:"config map values.yaml key"`
	DefaultValue interface{} `json:"defaultValue" description:"default value of mapKey"`
	Description  string      `json:"description" description:"config description"`
	Type         string      `json:"type" description:"config type"`
	Required     bool        `json:"required" description:"required"`
}

func (config *MetaCommonConfig) BuildDefaultValue(jsonStr string) {
	config.DefaultValue = config.BuildCommonConfigValue(jsonStr).Value
}

func (config *MetaCommonConfig) BuildCommonConfigValue(jsonStr string) *MetaCommonConfigValue {
	return &MetaCommonConfigValue{
		Name:  config.Name,
		Type:  config.Type,
		Value: gjson.Get(jsonStr, config.MapKey).Value(),
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
