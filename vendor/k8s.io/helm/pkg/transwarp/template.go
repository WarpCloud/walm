package transwarp

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/storage/driver"
	"math/rand"
	"path/filepath"
	"strings"
	"time"
	"encoding/json"
	"github.com/ghodss/yaml"
)

const (
	SystemNamespace            = "kube-system"
	TemplateFolderName         = "templates"
	ValuesFileName             = "values.yaml"
	TemplateAppYaml            = "app.yaml"
	DefaultDirectoryPermission = 0755
)

func Render(chartRequested *chart.Chart, namespace string, kubeContext string, depLinks map[string]interface{}) (map[string]string, error) {

	// get AppName And Jsonnet Template Path
	appName, jsonnetTemplatePath, err := getJsonnetAppNameAndTemplatePath(chartRequested)
	if err != nil {
		return nil, err
	}

	// handle dependencies
	depConfig, err := handleDependencies(chartRequested, namespace, kubeContext, depLinks)
	if err != nil {
		return nil, err
	}

	// get templates/instance-crd.yaml config
	templateConfig, err := getInstanceCrdConfig(chartRequested)
	if err != nil {
		return nil, err
	}

	for _, file := range chartRequested.Templates {

		if file.Name == "templates/transwarp-configmap-reserved.yaml" {

			if len(file.Data) == 0 {
				return nil, fmt.Errorf("file transwarp-configmap-reserved is null")
			}

			type T struct {
				ApiVersion string `json:"apiVersion"`
				Data   struct {
					Release string `json:"release"`
				}
			}

			t := &T{}
			err := yaml.Unmarshal(file.Data, &t)
			if err != nil {
				return nil, err
			}

			releaseData := t.Data.Release
			if releaseData == "" {
				return nil, fmt.Errorf("releaseData is null")
			}

			rls, err := driver.DecodeRelease(releaseData)
			if err != nil {
				return nil, err
			}

			template := make(map[string]string)
			if err := generateFilesFromConfigMap(rls.Chart, &template); err != nil {
				return nil, err
			}

			entrance, err := getTemplateEntrance(template, appName, jsonnetTemplatePath)
			if err != nil {
				return nil, err
			}

			// add extra config
			//userConfig["TosVersion"] = interface{}("1.5")
			templateConfig["Customized_Namespace"] = interface{}(namespace)
			templateConfig["Transwarp_Install_ID"] = interface{}(getRandomString(5))
			templateConfig["Transwarp_Install_Namespace"] = interface{}(namespace)
			templateConfig["Transwarp_Cni_Network"] = interface{}("overlay")

			mergedConfig := mergeValues(depConfig, templateConfig)
			configArray, err := json.Marshal(mergedConfig)
			if err != nil {
				glog.Errorf("Fail to marsh configs %+v", err)
				return nil, err
			}

			//templateConfig = string(configArray)

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

func getInstanceCrdConfig(chartRequested *chart.Chart) (map[string]interface{}, error) {

	config := make(map[string]interface{})
	var err error
	for _, file := range chartRequested.Templates {

		if file.Name == "templates/instance-crd.yaml" {

			if len(file.Data) == 0 {
				return config, fmt.Errorf("file transwarp-configmap-reserved is null")
			}

			type T struct {
				Kind string `json:"kind"`
				Spec   struct {
					Configs map[string]interface{} `json:"configs"`
				}
			}

			t := T{}
			err = yaml.Unmarshal(file.Data, &t)
			if err != nil {
				return config, err
			}

			config = t.Spec.Configs
		}
	}

	if config == nil || len(config) == 0 {
		return config, fmt.Errorf("fail to find config in templates/instance-crd.yaml ")
	}

	return config, nil
}

func getJsonnetAppNameAndTemplatePath(chartRequested *chart.Chart) (string, string, error) {

	appName := ""
	jsonnetTemplatePath := ""

	for _, file := range chartRequested.Files {

		if file.TypeUrl == "transwarp-app-yaml" {
			data := file.Value
			if len(data) == 0 {
				return appName, jsonnetTemplatePath, fmt.Errorf("file transwarp-app-yaml is null")
			}

			type T struct {
				Name string `json:"name"`
				JsonnetTemplatePath string `json:"jsonnetTemplatePath"`
			}

			t := T{}
			err := yaml.Unmarshal(data, &t)
			if err != nil {
				return "", "", err
			}

			appName = t.Name
			jsonnetTemplatePath = t.JsonnetTemplatePath
		}

	}

	if appName == "" {
		return appName, jsonnetTemplatePath, fmt.Errorf("fail to find appName in transwarp-app-yaml ")
	}

	if jsonnetTemplatePath == "" {
		return appName, jsonnetTemplatePath, fmt.Errorf("fail to find jsonnetTemplatePath in transwarp-app-yaml ")
	}

	return appName, jsonnetTemplatePath, nil

}

func handleDependencies(chartRequested *chart.Chart, namespace string, kubeContext string, depLinks map[string]interface{}) (map[string]interface{}, error) {

	err := CheckDepencies(chartRequested, depLinks)
	if err != nil {
		return nil, err
	}

	// init k8s transwarp client
	k8sTranswarpClient, err := GetTranswarpKubeClient(kubeContext)
	if err != nil {
		return nil, err
	}

	// init k8s client
	k8sClient, err := GetK8sKubeClient(kubeContext)
	if err != nil {
		return nil, err
	}

	depVals, err := GetDepenciesConfig(k8sTranswarpClient, k8sClient, namespace, depLinks)
	if err != nil {
		return nil, err
	}

 	return depVals, nil
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
	err := yaml.Unmarshal([]byte(resultStr), &resourcesMap)
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

		configArray, err := yaml.Marshal(value)
		if err != nil {
			return fmt.Errorf("Fail to marsh configs %+v", err)
		}

		filename := strings.Replace(string(key), ".json", ".yaml", -1)
		(*resultMap)[filename] = string(configArray)

	}

	return nil

}
