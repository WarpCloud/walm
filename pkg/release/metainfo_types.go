package release

import (
	"k8s.io/helm/pkg/walm"
	"github.com/tidwall/sjson"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/api/resource"
	"strconv"
	"fmt"
)

const (
	// k8s resource memory unit
	k8s_resource_memory_unit        = "Mi"
	k8s_resource_memory_scale int64 = 1024 * 1024

	// k8s resource storage unit
	k8s_resource_storage_unit        = "Gi"
	k8s_resource_storage_scale int64 = 1024 * 1024 * 1024

	k8s_resource_cpu_scale float64 = 1000
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

type MetaInfoParams struct {
	Params []*MetaCommonConfigValue `json:"params"`
	Roles  []*MetaRoleConfigValue   `json:"roles"`
}

func (metaInfoParams *MetaInfoParams) BuildConfigValues(metaInfo *ChartMetaInfo) (configValues map[string]interface{}, err error) {
	configValues = map[string]interface{}{}
	if metaInfo == nil {
		return
	}

	mapping := map[string]interface{}{}
	buildCommonConfigArrayValues(mapping, metaInfoParams.Params, metaInfo.ChartParams)
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

type MetaStringConfig struct {
	MapKey       string `json:"mapKey" description:"config map values.yaml key"`
	DefaultValue string `json:"defaultValue" description:"default value of mapKey"`
	Description  string `json:"description" description:"config description"`
	Type         string `json:"type" description:"config type"`
	Required     bool   `json:"required" description:"required"`
}

func (config *MetaStringConfig) BuildDefaultValue(jsonStr string) {
	config.DefaultValue = config.BuildStringConfigValue(jsonStr)
}

func (config *MetaStringConfig) BuildStringConfigValue(jsonStr string) string {
	return gjson.Get(jsonStr, config.MapKey).String()
}

type IntConfig struct {
	MapKey       string `json:"mapKey" description:"config map values.yaml key"`
	DefaultValue int64  `json:"defaultValue" description:"default value of mapKey"`
	Description  string `json:"description" description:"config description"`
	Type         string `json:"type" description:"config type"`
	Required     bool   `json:"required" description:"required"`
}

type MetaIntConfig struct {
	IntConfig
}

func (config *MetaIntConfig) BuildDefaultValue(jsonStr string) {
	config.DefaultValue = config.BuildIntConfigValue(jsonStr)
}

func (config *MetaIntConfig) BuildIntConfigValue(jsonStr string) int64 {
	return gjson.Get(jsonStr, config.MapKey).Int()
}

type MetaEnvConfig struct {
	MapKey       string    `json:"mapKey" description:"config map values.yaml key"`
	DefaultValue []MetaEnv `json:"defaultValue" description:"default value of mapKey"`
	Description  string    `json:"description" description:"config description"`
	Type         string    `json:"type" description:"config type"`
	Required     bool      `json:"required" description:"required"`
}

func (config *MetaEnvConfig) BuildDefaultValue(jsonStr string) {
	config.DefaultValue = config.BuildEnvConfigValue(jsonStr)
}

func (config *MetaEnvConfig) BuildEnvConfigValue(jsonStr string) []MetaEnv {
	metaEnv := []MetaEnv{}
	rawMsg := gjson.Get(jsonStr, config.MapKey).Raw
	if rawMsg == "" {
		return metaEnv
	}
	err := json.Unmarshal([]byte(rawMsg), &metaEnv)
	if err != nil {
		logrus.Warnf("failed to unmarshal %s : %s", rawMsg, err.Error())
	}
	return metaEnv
}

type MetaEnv struct {
	Name  string `json:"name" description:"env name"`
	Value string `json:"value" description:"env value"`
}

type MetaBoolConfig struct {
	MapKey       string `json:"mapKey" description:"config map values.yaml key"`
	DefaultValue bool   `json:"defaultValue" description:"default value of mapKey"`
	Description  string `json:"description" description:"config description"`
	Type         string `json:"type" description:"config type"`
	Required     bool   `json:"required" description:"required"`
}

func (config *MetaBoolConfig) BuildDefaultValue(jsonStr string) {
	config.DefaultValue = config.BuildBoolConfigValue(jsonStr)
}

func (config *MetaBoolConfig) BuildBoolConfigValue(jsonStr string) bool {
	return gjson.Get(jsonStr, config.MapKey).Bool()
}

type FloatConfig struct {
	MapKey       string  `json:"mapKey" description:"config map values.yaml key"`
	DefaultValue float64 `json:"defaultValue" description:"default value of mapKey"`
	Description  string  `json:"description" description:"config description"`
	Type         string  `json:"type" description:"config type"`
	Required     bool    `json:"required" description:"required"`
}

type MetaFloatConfig struct {
	FloatConfig
}

func (config *MetaFloatConfig) BuildDefaultValue(jsonStr string) {
	config.DefaultValue = config.BuildFloatConfigValue(jsonStr)
}

func (config *MetaFloatConfig) BuildFloatConfigValue(jsonStr string) float64 {
	return gjson.Get(jsonStr, config.MapKey).Float()
}

type MetaCommonConfigValue struct {
	Name  string      `json:"name" description:"config name"`
	Type  string      `json:"type" description:"config type"`
	Value interface{} `json:"value" description:"config value"`
}

type MetaResourceConfig struct {
	LimitsMemoryKey   *MetaResourceIntConfig       `json:"limitsMemoryKey" description:"resource memory limit"`
	LimitsCpuKey      *MetaResourceFloatConfig     `json:"limitsCpuKey" description:"resource cpu limit"`
	LimitsGpuKey      *MetaResourceFloatConfig     `json:"limitsGpuKey" description:"resource gpu limit"`
	RequestsMemoryKey *MetaResourceIntConfig       `json:"requestsMemoryKey" description:"resource memory request"`
	RequestsCpuKey    *MetaResourceFloatConfig     `json:"requestsCpuKey" description:"resource cpu request"`
	RequestsGpuKey    *MetaResourceFloatConfig     `json:"requestsGpuKey" description:"resource gpu request"`
	StorageResources  []*MetaResourceStorageConfig `json:"storageResources" description:"resource storage request"`
}

func (config *MetaResourceConfig) BuildDefaultValue(jsonStr string) {
	if config.LimitsMemoryKey != nil {
		config.LimitsMemoryKey.BuildDefaultValue(jsonStr)
	}
	if config.LimitsGpuKey != nil {
		config.LimitsGpuKey.BuildDefaultValue(jsonStr)
	}
	if config.LimitsCpuKey != nil {
		config.LimitsCpuKey.BuildDefaultValue(jsonStr)
	}
	if config.RequestsMemoryKey != nil {
		config.RequestsMemoryKey.BuildDefaultValue(jsonStr)
	}
	if config.RequestsGpuKey != nil {
		config.RequestsGpuKey.BuildDefaultValue(jsonStr)
	}
	if config.RequestsCpuKey != nil {
		config.RequestsCpuKey.BuildDefaultValue(jsonStr)
	}

	for _, storageConfig := range config.StorageResources {
		if storageConfig != nil {
			storageConfig.BuildDefaultValue(jsonStr)
		}
	}
}

func (config *MetaResourceConfig) BuildResourceConfigValue(jsonStr string) *MetaResourceConfigValue {
	resourceConfigValue := &MetaResourceConfigValue{}
	if config.LimitsMemoryKey != nil {
		resourceConfigValue.LimitsMemoryKey = config.LimitsMemoryKey.BuildIntConfigValue(jsonStr)
	}
	if config.LimitsGpuKey != nil {
		resourceConfigValue.LimitsGpuKey = config.LimitsGpuKey.BuildFloatConfigValue(jsonStr)
	}
	if config.LimitsCpuKey != nil {
		resourceConfigValue.LimitsCpuKey = config.LimitsCpuKey.BuildFloatConfigValue(jsonStr)
	}
	if config.RequestsMemoryKey != nil {
		resourceConfigValue.RequestsMemoryKey = config.RequestsMemoryKey.BuildIntConfigValue(jsonStr)
	}
	if config.RequestsGpuKey != nil {
		resourceConfigValue.RequestsGpuKey = config.RequestsGpuKey.BuildFloatConfigValue(jsonStr)
	}
	if config.RequestsCpuKey != nil {
		resourceConfigValue.RequestsCpuKey = config.RequestsCpuKey.BuildFloatConfigValue(jsonStr)
	}

	for _, storageConfig := range config.StorageResources {
		if storageConfig != nil {
			resourceConfigValue.StorageResources = append(resourceConfigValue.StorageResources, storageConfig.BuildStorageConfigValue(jsonStr))
		}
	}
	return resourceConfigValue
}

type MetaResourceIntConfig struct {
	IntConfig
}

func (config *MetaResourceIntConfig) BuildDefaultValue(jsonStr string) {
	config.DefaultValue = config.BuildIntConfigValue(jsonStr)
}

func (config *MetaResourceIntConfig) BuildIntConfigValue(jsonStr string) int64 {
	quantity := parseK8sResourceQuantity(jsonStr, config.MapKey)
	if quantity != nil {
		return quantity.Value() / k8s_resource_memory_scale
	}
	return 0
}

func parseK8sResourceQuantity(jsonStr, mapKey string) *resource.Quantity {
	if jsonStr == "" || mapKey == "" {
		return nil
	}
	strValue := gjson.Get(jsonStr, mapKey).String()
	if strValue == "" {
		return nil
	}
	quantity, err := resource.ParseQuantity(strValue)
	if err != nil {
		logrus.Warnf("failed to parse quantity %s : %s", strValue, err.Error())
		return nil
	}
	return &quantity
}

type MetaResourceFloatConfig struct {
	FloatConfig
}

func (config *MetaResourceFloatConfig) BuildDefaultValue(jsonStr string) {
	config.DefaultValue = config.BuildFloatConfigValue(jsonStr)
}

func (config *MetaResourceFloatConfig) BuildFloatConfigValue(jsonStr string) float64 {
	quantity := parseK8sResourceQuantity(jsonStr, config.MapKey)
	if quantity != nil {
		return float64(quantity.MilliValue()) / k8s_resource_cpu_scale
	}
	return 0
}

type ResourceStorage struct {
	AccessModes  []string `json:"accessModes, omitempty" description:"storage access modes"`
	AccessMode   string   `json:"accessMode, omitempty" description:"storage access mode"`
	StorageClass string   `json:"storageClass" description:"storage class"`
}

type MetaResourceStorage struct {
	ResourceStorage
	Size int64 `json:"size" description:"storage size"`
}

type MetaResourceStorageWithStringSize struct {
	ResourceStorage
	Size string `json:"size" description:"storage size"`
}

type MetaResourceStorageConfig struct {
	Name         string               `json:"name" description:"config name"`
	MapKey       string               `json:"mapKey" description:"config map values.yaml key"`
	DefaultValue *MetaResourceStorage `json:"defaultValue" description:"default value of mapKey"`
	Description  string               `json:"description" description:"config description"`
	Type         string               `json:"type" description:"config type"`
	Required     bool                 `json:"required" description:"required"`
}

func (config *MetaResourceStorageConfig) BuildDefaultValue(jsonStr string) {
	config.DefaultValue = config.BuildStorageConfigValue(jsonStr).Value
}

func (config *MetaResourceStorageConfig) BuildStorageConfigValue(jsonStr string) *MetaResourceStorageConfigValue {
	resourceStorageConfigValue := &MetaResourceStorageConfigValue{
		Name: config.Name,
	}
	resourceStorageWithStringSize := parseResourceStorageWithStringSize(jsonStr, config.MapKey)
	if resourceStorageWithStringSize != nil {
		resourceStorageConfigValue.Value = &MetaResourceStorage{
			ResourceStorage: resourceStorageWithStringSize.ResourceStorage,
		}

		if resourceStorageWithStringSize.Size != "" {
			quantity, err := resource.ParseQuantity(resourceStorageWithStringSize.Size)
			if err != nil {
				logrus.Warnf("failed to parse quantity %s : %s", resourceStorageWithStringSize.Size, err.Error())
				return nil
			}
			resourceStorageConfigValue.Value.Size = quantity.Value() / k8s_resource_storage_scale
		}
	}
	return resourceStorageConfigValue
}

func parseResourceStorageWithStringSize(jsonStr, mapKey string) *MetaResourceStorageWithStringSize {
	rawMsg := gjson.Get(jsonStr, mapKey).Raw
	if rawMsg == "" {
		return nil
	}
	resourceStorage := &MetaResourceStorageWithStringSize{}
	err := json.Unmarshal([]byte(rawMsg), resourceStorage)
	if err != nil {
		logrus.Warnf("failed to unmarshal %s : %s", rawMsg, err.Error())
		return nil
	}
	return resourceStorage
}

type MetaResourceConfigValue struct {
	LimitsMemoryKey   int64                             `json:"limitsMemoryKey" description:"resource memory limit"`
	LimitsCpuKey      float64                           `json:"limitsCpuKey" description:"resource cpu limit"`
	LimitsGpuKey      float64                           `json:"limitsGpuKey" description:"resource gpu limit"`
	RequestsMemoryKey int64                             `json:"requestsMemoryKey" description:"resource memory request"`
	RequestsCpuKey    float64                           `json:"requestsCpuKey" description:"resource cpu request"`
	RequestsGpuKey    float64                           `json:"requestsGpuKey" description:"resource gpu request"`
	StorageResources  []*MetaResourceStorageConfigValue `json:"storageResources" description:"resource storage request"`
}

type MetaResourceStorageConfigValue struct {
	Name  string               `json:"name" description:"config name"`
	Value *MetaResourceStorage `json:"value" description:"config value"`
}

func (resourceConfigValue *MetaResourceConfigValue) BuildConfigValue(mapping map[string]interface{}, resourceConfig *MetaResourceConfig) {
	if resourceConfig == nil {
		return
	}

	if resourceConfigValue.LimitsMemoryKey != 0 && resourceConfig.LimitsMemoryKey != nil {
		mapping[resourceConfig.LimitsMemoryKey.MapKey] = convertResourceBinaryIntByUnit(resourceConfigValue.LimitsMemoryKey, k8s_resource_memory_unit)
	}
	if resourceConfigValue.LimitsGpuKey != 0 && resourceConfig.LimitsGpuKey != nil {
		mapping[resourceConfig.LimitsGpuKey.MapKey] = convertResourceDecimalFloat(resourceConfigValue.LimitsGpuKey)
	}
	if resourceConfigValue.LimitsCpuKey != 0 && resourceConfig.LimitsCpuKey != nil {
		mapping[resourceConfig.LimitsCpuKey.MapKey] = convertResourceDecimalFloat(resourceConfigValue.LimitsCpuKey)
	}
	if resourceConfigValue.RequestsMemoryKey != 0 && resourceConfig.RequestsMemoryKey != nil {
		mapping[resourceConfig.RequestsMemoryKey.MapKey] = convertResourceBinaryIntByUnit(resourceConfigValue.RequestsMemoryKey, k8s_resource_memory_unit)
	}
	if resourceConfigValue.RequestsGpuKey != 0 && resourceConfig.RequestsGpuKey != nil {
		mapping[resourceConfig.RequestsGpuKey.MapKey] = convertResourceDecimalFloat(resourceConfigValue.RequestsGpuKey)
	}
	if resourceConfigValue.RequestsCpuKey != 0 && resourceConfig.RequestsCpuKey != nil {
		mapping[resourceConfig.RequestsCpuKey.MapKey] = convertResourceDecimalFloat(resourceConfigValue.RequestsCpuKey)
	}

	buildResourceStorageArrayValues(mapping, resourceConfigValue.StorageResources, resourceConfig.StorageResources)
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
				Size: convertResourceBinaryIntByUnit(resourceStorageConfigValue.Value.Size, k8s_resource_storage_unit),
			}
			mapping[resourceStorageConfig.MapKey] = resourceStorageWithStringSize
		}
	}
}

func convertResourceBinaryIntByUnit(i int64, unit string) string {
	return strconv.FormatInt(i, 10) + unit
}

func convertResourceDecimalFloat(f float64) string {
	return fmt.Sprintf("%g", f)
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
		roleBaseConfigValue.Image = roleBaseConfig.Image.BuildStringConfigValue(jsonStr)
	}
	if roleBaseConfig.Replicas != nil {
		roleBaseConfigValue.Replicas = roleBaseConfig.Replicas.BuildIntConfigValue(jsonStr)
	}
	if roleBaseConfig.Env != nil {
		roleBaseConfigValue.Env = roleBaseConfig.Env.BuildEnvConfigValue(jsonStr)
	}
	if roleBaseConfig.UseHostNetwork != nil {
		roleBaseConfigValue.UseHostNetwork = roleBaseConfig.UseHostNetwork.BuildBoolConfigValue(jsonStr)
	}
	if roleBaseConfig.Priority != nil {
		roleBaseConfigValue.Priority = roleBaseConfig.Priority.BuildIntConfigValue(jsonStr)
	}
	for _, config := range roleBaseConfig.Others {
		if config != nil {
			roleBaseConfigValue.Others = append(roleBaseConfigValue.Others, config.BuildCommonConfigValue(jsonStr))
		}
	}
	return roleBaseConfigValue
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
	Image          string                   `json:"image" description:"role image"`
	Priority       int64                    `json:"priority" description:"role priority"`
	Replicas       int64                    `json:"replicas" description:"role replicas"`
	Env            []MetaEnv                `json:"env" description:"role env list"`
	UseHostNetwork bool                     `json:"useHostNetwork" description:"whether role use host network"`
	Others         []*MetaCommonConfigValue `json:"others" description:"role other configs"`
}

func (roleBaseConfigValue *MetaRoleBaseConfigValue) BuildConfigValue(mapping map[string]interface{}, roleBaseConfig *MetaRoleBaseConfig) {
	if roleBaseConfig == nil {
		return
	}

	if roleBaseConfig.Image != nil && roleBaseConfigValue.Image != "" {
		mapping[roleBaseConfig.Image.MapKey] = roleBaseConfigValue.Image
	}
	// TODO bool use pointer?
	if roleBaseConfig.UseHostNetwork != nil {
		mapping[roleBaseConfig.UseHostNetwork.MapKey] = roleBaseConfigValue.UseHostNetwork
	}
	if roleBaseConfig.Priority != nil && roleBaseConfigValue.Priority != 0 {
		mapping[roleBaseConfig.Priority.MapKey] = roleBaseConfigValue.Priority
	}
	if roleBaseConfig.Env != nil && len(roleBaseConfigValue.Env) > 0 {
		mapping[roleBaseConfig.Env.MapKey] = roleBaseConfigValue.Env
	}
	if roleBaseConfig.Replicas != nil && roleBaseConfigValue.Replicas != 0 {
		mapping[roleBaseConfig.Replicas.MapKey] = roleBaseConfigValue.Replicas
	}

	buildCommonConfigArrayValues(mapping, roleBaseConfigValue.Others, roleBaseConfig.Others)
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

func buildCommonConfigArrayValues(mapping map[string]interface{}, commonConfigValues []*MetaCommonConfigValue, commonConfigs []*MetaCommonConfig) {
	commonConfigsMap := convertCommonConfigArrayToMap(commonConfigs)

	for _, commonConfigValue := range commonConfigValues {
		if commonConfigValue.Value == nil {
			continue
		}
		if commonConfig, ok := commonConfigsMap[commonConfigValue.Name]; ok {
			if commonConfig.MapKey != "" {
				mapping[commonConfig.MapKey] = commonConfigValue.Value
			}
		}
	}
}
