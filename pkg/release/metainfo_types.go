package release

import "k8s.io/helm/pkg/walm"

// chart metainfo
type ChartMetaInfo struct {
	FriendlyName          string                     `json:"friendlyName" description:"friendlyName"`
	Categories            []string                   `json:"categories" description:"categories"`
	ChartDependenciesInfo []*ChartDependencyMetaInfo `json:"dependencies" description:"dependency metainfo"`
	ChartRoles            []*MetaRoleConfig          `json:"roles"`
	ChartParams           []*MetaCommonConfig        `json:"params"`
	Plugins               []*walm.WalmPlugin         `json:"plugins"`
}

type ChartDependencyMetaInfo struct {
	Name               string `json:"name,omitempty"`
	MinVersion         string `json:"minVersion"`
	MaxVersion         string `json:"maxVersion"`
	DependencyOptional bool   `json:"dependencyOptional"`
	AliasConfigVar     string `json:"aliasConfigVar,omitempty"`
	ChartName          string `json:"chartName"`
	DependencyType     string `json:"type"`
	AutoDependency     bool   `json:"autoDependency"`
}

type MetaCommonConfig struct {
	Name         string      `json:"name" description:"config name"`
	MapKey       string      `json:"mapKey" description:"config map values.yaml key"`
	DefaultValue interface{} `json:"defaultValue" description:"default value of mapKey"`
	Description  string      `json:"description" description:"config description"`
	Type         string      `json:"type" description:"config type"`
	Required     bool        `json:"required" description:"required"`
}

type MetaResourceConfig struct {
	LimitsMemoryKey   string `json:"limitsMemoryKey" description:"resource memory limit"`
	LimitsCpuKey      string `json:"limitsCpuKey" description:"resource cpu limit"`
	LimitsGpuKey      string `json:"limitsGpuKey" description:"resource gpu limit"`
	RequestsMemoryKey string `json:"requestsMemoryKey" description:"resource memory request"`
	RequestsCpuKey    string `json:"limitsCpuKey" description:"resource cpu request"`
	RequestsGpuKey    string `json:"limitsGpuKey" description:"resource gpu request"`
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
