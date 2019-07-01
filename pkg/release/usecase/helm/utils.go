package helm

import (
	"reflect"
)


func ConfigValuesDiff(configValue1 map[string]interface{}, configValue2 map[string]interface{}) bool {
	if len(configValue1) == 0 && len(configValue2) == 0 {
		return false
	}
	return !reflect.DeepEqual(configValue1, configValue2)
}

func DeleteReleaseDependency(dependencies map[string]string, dependencyKey string) {
	if _, ok := dependencies[dependencyKey]; ok {
		dependencies[dependencyKey] = ""
	}
}
