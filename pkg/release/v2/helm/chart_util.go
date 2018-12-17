package helm

import (
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/storage/driver"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"path/filepath"
	"os"
	"io/ioutil"
	"github.com/golang/glog"
	"github.com/ghodss/yaml"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
)

func isJsonnetChart(chart *chart.Chart) (isJsonnetChart bool, jsonnetChart *chart.Chart, appYaml []byte, err error) {
	jsonnetTemplatesFound, appYamlFound := false, false
	for _, file := range chart.Files {
		if jsonnetTemplatesFound && appYamlFound {
			break
		}
		if file.TypeUrl == "transwarp-configmap-reserved" {
			releaseStr := string(file.Value[:])
			release, err := driver.DecodeRelease(releaseStr)
			if err != nil {
				logrus.Errorf("failed to decode release : %s", err.Error())
				return false, nil, nil, err
			}
			jsonnetChart = release.Chart
			isJsonnetChart = true
			jsonnetTemplatesFound = true
		} else if file.TypeUrl == "transwarp-app-yaml" {
			appYaml = file.Value
			appYamlFound = true
		}
	}
	return
}

// convert jsonnet chart to native chart
// 1. load jsonnet template files to render
//     a. load common jsonnet lib
//     b. load jsonnet chart template files
// 2. build config values to render jsonnet template files
//     a. merge values from value.yaml
//     b. merge system values
//     c. merge dependency release output configs
//     d. merge configs user provided
// 3. render jsonnet template files to generate native chart templates
func convertJsonnetChart(namespace string, jsonnetChart *chart.Chart, userConfigs map[string]interface{}, dependencyConfigs map[string]interface{}) (nativeChart *chart.Chart, err error) {
	nativeChart = &chart.Chart{
		Metadata: jsonnetChart.Metadata,
		Files:    jsonnetChart.Files,
		Values:   jsonnetChart.Values,
	}
	//TODO build nativeChart.Values: config.jsonnet + jsonnetChart.Values

	templateFiles, err := loadJsonnetFilesToRender(jsonnetChart)
	if err != nil {
		logrus.Errorf("failed to load jsonnet template files to render : %s", err.Error())
		return nil, err
	}

	configValues, err := buildConfigValuesToRender(namespace, jsonnetChart, userConfigs, dependencyConfigs)
	if err != nil {
		logrus.Errorf("failed to build config values to render jsonnet template files : %s", err.Error())
		return nil, err
	}

	nativeChart.Templates, err = renderJsonnetFiles(templateFiles, configValues)
	if err != nil {
		logrus.Errorf("failed to render jsonnet files : %s", err.Error())
		return nil, err
	}

	return
}

func buildConfigValuesToRender(namespace string, jsonnetChart *chart.Chart, userConfigs map[string]interface{}, dependencyConfigs map[string]interface{}) (configValues map[string]interface{}, err error) {
	configValues = map[string]interface{}{}
	defaultValue := map[string]interface{}{}
	if jsonnetChart.Values != nil {
		err = yaml.Unmarshal([]byte(jsonnetChart.Values.Raw), &defaultValue)
		if err != nil {
			logrus.Errorf("failed to unmarshal jsonnet chart value.yaml : %s", err.Error())
			return nil, err
		}
	}
	mergeValues(configValues, defaultValue)
	//TODO merge system values
	mergeValues(configValues, dependencyConfigs)
	mergeValues(configValues, userConfigs)
	configValues["Transwarp_Install_ID"] = utilrand.String(5)
	configValues["Transwarp_Install_Namespace"] = namespace
	return
}

func loadJsonnetFilesToRender(jsonnetChart *chart.Chart) (templateFiles map[string]string, err error) {
	templateFiles = map[string]string{}
	err = loadCommonJsonnetLib(templateFiles)
	if err != nil {
		logrus.Errorf("failed to load common jsonnet lib : %s", err.Error())
		return nil, err
	}

	err = loadJsonnetFilesFromJsonnetChart(jsonnetChart, templateFiles)
	if err != nil {
		logrus.Errorf("failed to load jsonnet files from jsonnet chart : %s", err.Error())
		return nil, err
	}
	return
}

func loadJsonnetFilesFromJsonnetChart(jsonnetChart *chart.Chart, templateFiles map[string]string) error {
	for _, template := range jsonnetChart.Templates {
		templateFiles[template.Name] = string(template.Data)
	}
	return nil
}

func loadCommonJsonnetLib(templates map[string]string) error {
	//TODO
	return nil
}

// LoadFilesFromDisk loads all files inside baseDir directory and its subdirectory recursively,
// mapping each file's path/content as a key/value into a map.
func LoadFilesFromDisk(baseDir string) (map[string]string, error) {
	cacheFiles := make(map[string]string)
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			b, err := ioutil.ReadFile(path)
			if err != nil {
				glog.Errorf("Read file \"%s\", err: %v", path, err)
				return err
			}
			cacheFiles[path] = string(b)
		}
		return nil
	})
	if err != nil {
		return cacheFiles, err
	}
	return cacheFiles, nil
}