package release

// Pretty Paramters
type ResourceStorageConfig struct {
	Name         string   `json:"name"`
	StorageType  string   `json:"type"`
	StorageClass string   `json:"storageClass"`
	Size         string   `json:"size"`
	AccessModes  []string `json:"accessModes"`
	AccessMode   string   `json:"accessMode"`
}

type ResourceConfig struct {
	CpuLimit            float64                 `json:"cpuLimit"`
	CpuRequest          float64                 `json:"cpuRequest"`
	MemoryLimit         float64                 `json:"memoryLimit"`
	MemoryRequest       float64                 `json:"memoryRequest"`
	GpuLimit            int                     `json:"gpuLimit"`
	GpuRequest          int                     `json:"gpuRequest"`
	ResourceStorageList []ResourceStorageConfig `json:"storage"`
}

type BaseConfig struct {
	ValueName        string      `json:"variable" description:"variable name"`
	DefaultValue     interface{} `json:"default" description:"variable default value"`
	ValueDescription string      `json:"description" description:"variable description"`
	ValueType        string      `json:"type" description:"variable type"`
}

type RoleConfig struct {
	Name               string          `json:"name"`
	Description        string          `json:"description"`
	Replicas           int             `json:"replicas"`
	RoleBaseConfig     []*BaseConfig   `json:"baseConfig"`
	RoleResourceConfig *ResourceConfig `json:"resouceConfig"`
}

type CommonConfig struct {
	Roles []*RoleConfig `json:"roles"`
}

type PrettyChartParams struct {
	CommonConfig        CommonConfig  `json:"commonConfig"`
	TranswarpBaseConfig []*BaseConfig `json:"transwarpBundleConfig"`
	AdvanceConfig       []*BaseConfig `json:"advanceConfig"`
}
