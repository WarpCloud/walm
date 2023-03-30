package plugins

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_convertK8SConfigMap(t *testing.T) {
	releaseName := "test-release"
	releaseNamespace := "test"
	configMapName := "test-cm"
	//addObj := &AddConfigmapObject{
	//	ApplyAllResources: true,
	//	Kind:              "",
	//	ResourceName:      "",
	//	ContainerName:     "",
	//	Items: []*AddConfigItem{
	//		{
	//			ConfigMapData:                  "test data\n",
	//			ConfigMapVolumeMountsMountPath: "/aa/bb/c",
	//			ConfigMapVolumeMountsSubPath:   "path-name",
	//		},
	//	},
	//}

	//configMap, _ := convertK8SConfigMap(releaseName, releaseNamespace, configMapName, addObj)
	assert.Equal(t, releaseName+releaseNamespace+configMapName, "ConfigMap")
}

