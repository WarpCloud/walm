package plugins

const (
	CustomConfigmapPluginName = "CustomConfigmap"
)

type AddConfigmapObject struct {
	ApplyAllResources bool             `json:"applyAllResources"`
	Kind              string           `json:"kind"`
	ResourceName      string           `json:"resourceName"`
	ContainerName     string           `json:"containerName"`
	Items             []*AddConfigItem `json:"items"`
}

type AddConfigItem struct {
	ConfigMapData                  string `json:"configMapData"`
	ConfigMapVolumeMountsMountPath string `json:"configMapVolumeMountsMountPath"`
	ConfigMapVolumeMountsSubPath   string `json:"configMapVolumeMountsSubPath"`
	ConfigMapMode                  int32  `json:"configMapMode"`
}

type CustomConfigmapArgs struct {
	ConfigmapToAdd       map[string]*AddConfigmapObject `json:"configmapToAdd" description:"add extra configmap"`
	ConfigmapToSkipNames []string                       `json:"configmapToSkipNames" description:"upgrade skip to render configmap"`
	ConfigmapSkipAll     bool                           `json:"configmapSkipAll" description:"upgrade skip all configmap resources"`
}

func isSkippedConfigMap(name string, args *CustomConfigmapArgs) bool{
	if args.ConfigmapSkipAll == true {
		return true
	} else {
		for _, skipConfigmapName := range args.ConfigmapToSkipNames {
			if skipConfigmapName == name {
				return true
			}
		}
		return false
	}
}




