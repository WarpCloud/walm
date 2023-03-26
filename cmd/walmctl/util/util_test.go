package util

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"testing"
)

var configValuesStr = `metaInfoParams:
  params:
  - name: jenkinsUriPrefix
    type: string
    value: /chartv3-test/jenkins
  - name: ingressPath
    type: string
    value: /chartv3-test/jenkins
plugins:
- name: CustomIngress
  version: "1.0"
  args:
    ingressSkipAll: true
  disable: false
- name: CustomConfigMap
  version: "1.0"
  args:
    configMapSkipAll: true
  disable: false
dependencies: {}`

func Test_SmartConfigValues(t *testing.T) {
	configValues := make(map[string]interface{}, 0)

	err := yaml.Unmarshal([]byte(configValuesStr[:]), &configValues)
	if err != nil {
		panic(err)
	}
	configValuesStr, _ := json.Marshal(configValues)
	fmt.Printf("configValues %v\n", string(configValuesStr[:]))

	destConfigValues, metaInfoParamParams, releasePlugins, err := SmartConfigValues(configValues)
	if err != nil {
		panic(err)
	}
	destConfigValuesStr, _ := json.Marshal(destConfigValues)
	fmt.Printf("configValues %v\n", string(destConfigValuesStr[:]))

	fmt.Printf("metaInfoParamParams %v releasePlugins %v\n", metaInfoParamParams, releasePlugins)
}
