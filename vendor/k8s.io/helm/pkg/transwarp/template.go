package transwarp

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	yaml2 "gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/storage/driver"
	"math/rand"
	"path/filepath"
	"strings"
	"time"
)

const (
	SystemNamespace            = "kube-system"
	TemplateFolderName         = "templates"
	ValuesFileName             = "values.yaml"
	TemplateAppYaml            = "app.yaml"
	DefaultDirectoryPermission = 0755
)

func Render(chartRequested *chart.Chart, namespace string, userVals []byte, kubeVersion string) (map[string]string, error) {

	appName := ""
	jsonnetTemplatePath := ""
	for _, file := range chartRequested.Files {

		if file.TypeUrl == "transwarp-app-yaml" {
			data := file.Value
			if len(data) == 0 {
				return nil, fmt.Errorf("file transwarp-app-yaml is null")
			}

			appYaml := make(map[string]interface{})
			err := yaml2.Unmarshal(data, &appYaml)
			if err != nil {
				return nil, err
			}
			appName = appYaml["name"].(string)
			jsonnetTemplatePath = appYaml["jsonnetTemplatePath"].(string)
		}

	}

	if appName == "" {
		return nil, fmt.Errorf("fail to find appName in transwarp-app-yaml ")
	}

	if jsonnetTemplatePath == "" {
		return nil, fmt.Errorf("fail to find jsonnetTemplatePath in transwarp-app-yaml ")
	}

	for _, file := range chartRequested.Files {

		if file.TypeUrl == "transwarp-configmap-reserved" {

			if len(file.Value) == 0 {
				return nil, fmt.Errorf("file transwarp-configmap-reserved is null")
			}

			appChart := string(file.Value[:])
			rls, err := driver.DecodeRelease(appChart)
			if err != nil {
				return nil, err
			}

			template := make(map[string]string)
			if err := generateFilesFromConfigMap(rls.Chart, &template); err != nil {
				return nil, err
			}

			defaultConfigs := make(map[string]interface{})

			if values, ok := template[filepath.Join(appName, ValuesFileName)]; ok {
				result := make(map[string]interface{})
				err = yaml2.Unmarshal([]byte(values), &result)
				if err != nil {
					return nil, err
				}
				for key, value := range result {
					defaultConfigs[key] = value
				}
			}

			userConfigs := make(map[string]interface{})
			err = yaml.Unmarshal([]byte(userVals), &userConfigs)
			if err != nil {
				return nil, err
			}

			// TosVersion Transwarp_Install_ID  Transwarp_Install_Namespace Customized_Namespace
			userConfigs["TosVersion"] = interface{}(kubeVersion)
			userConfigs["Transwarp_Install_Namespace"] = interface{}(namespace)
			userConfigs["Customized_Namespace"] = interface{}(namespace)
			userConfigs["Transwarp_Install_ID"] = interface{}(getRandomString(5))

			combinedConfigs := mergeValues(defaultConfigs, userConfigs)
			for k, v := range combinedConfigs {
				combinedConfigs[k] = yamlJsonConvert(v)
			}

			//// parse the template with configs
			configArray, err := json.Marshal(combinedConfigs)
			if err != nil {
				glog.Errorf("Fail to marsh configs %+v", err)
				return nil, err
			}
			entrance, err := getTemplateEntrance(template, appName, jsonnetTemplatePath)
			if err != nil {
				return nil, err
			}

			resultStr, err := parseTemplateWithTLAString(entrance, "config", string(configArray), template)
			if err != nil {
				return nil, err
			}

			//format return value, which is same with opensource helm
			resultMap := make(map[string]string)
			if err := changeResultStrToMap(resultStr, &resultMap); err != nil {
				return nil, err
			}

			return resultMap, nil

		}
	}

	return nil, fmt.Errorf("Fail to find file transwarp-configmap-reserved ")
}

func yamlJsonConvert(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = yamlJsonConvert(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = yamlJsonConvert(v)
		}
	}
	return i
}

func getTemplateEntrance(templates map[string]string, appName string, jsonnetTemplatePath string) (string, error) {
	if templates == nil {
		return "", fmt.Errorf("unexpected nil templates cache for %s", appName)
	}
	for k := range templates {
		if strings.HasPrefix(k, filepath.Join(appName, TemplateFolderName)) && strings.HasSuffix(k, jsonnetTemplatePath) {
			return k, nil
		}
	}
	return "", fmt.Errorf("unable to find template for entrance %s or %s", appName, jsonnetTemplatePath)
}

func generateFilesFromConfigMap(chart *chart.Chart, cacheTemplate *map[string]string) error {

	if cacheTemplate == nil {
		return fmt.Errorf("nil cache template")
	}
	appName := chart.Metadata.Name
	(*cacheTemplate)[filepath.Join(appName, ValuesFileName)] = chart.Values.Raw

	templates := chart.Templates
	for _, template := range templates {
		path := filepath.Join(appName, template.Name)
		(*cacheTemplate)[path] = string(template.Data)
	}

	dependencies := chart.Dependencies
	for _, dependency := range dependencies {
		if err := generateFilesFromConfigMap(dependency, cacheTemplate); err != nil {
			return err
		}
	}
	return nil
}

func parseTemplateWithTLAString(templatePath string, tlaVar string, tlaValue string, templateData map[string]string) (string, error) {
	vm := MakeMemoryVM(templateData)
	vm.TLACode(tlaVar, tlaValue)
	if _, ok := templateData[templatePath]; !ok {
		glog.Errorf("Fail to find entrance of template %s", templatePath)
		return "", fmt.Errorf("Fail to find entrance of template %s", templatePath)
	}
	output, err := vm.EvaluateSnippet(templatePath, templateData[templatePath])
	if err != nil {
		glog.Errorf("Fail to parse template %s, %s=%s, error: %+v", "", tlaVar, tlaValue, err)
		return "", err
	}
	return string(output), nil
}

func mergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {

	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = mergeValues(destMap, nextMap)
	}
	return dest

}

func getRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func changeResultStrToMap(resultStr string, resultMap *map[string]string) error {

	// key: resource.json, value: resource template(map)
	resourcesMap := make(map[string]map[string]interface{})
	err := yaml2.Unmarshal([]byte(resultStr), &resourcesMap)
	if err != nil {
		return err
	}

	// every yaml must has apiVersion and kind
	for key, value := range resourcesMap {

		resourceMap := resourcesMap[key]
		if _, ok := resourceMap["apiVersion"]; !ok {
			return fmt.Errorf("%s: invalid format to decode, lack of apiVersion fields.", key)
		}

		if _, ok := resourceMap["kind"]; !ok {
			return fmt.Errorf("%s: invalid format to decode, lack of kind fields.", key)
		}

		configArray, err := yaml2.Marshal(value)
		if err != nil {
			return fmt.Errorf("Fail to marsh configs %+v", err)
		}

		filename := strings.Replace(string(key), ".json", ".yaml", -1)
		(*resultMap)[filename] = string(configArray)

	}

	return nil

}