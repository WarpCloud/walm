package helm

import (
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/storage/driver"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"path/filepath"
	"os"
	"io/ioutil"
	"github.com/ghodss/yaml"
	"path"
	"walm/pkg/setting"
	"transwarp/release-config/pkg/apis/transwarp/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	corev1 "k8s.io/api/core/v1"
	"encoding/json"
)

const (
	commonTemplateDir = "templates/applib/ksonnet-lib"
)

var commonTemplateFilesPath string
var commonTemplateFiles map[string]string

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
func convertJsonnetChart(releaseNamespace, releaseName string, dependencies map[string]string, jsonnetChart *chart.Chart, userConfigs map[string]interface{}, dependencyConfigs map[string]interface{}) (nativeChart *chart.Chart, err error) {
	nativeChart = &chart.Chart{
		Metadata: jsonnetChart.Metadata,
		Files:    jsonnetChart.Files,
	}

	templateFiles, err := loadJsonnetFilesToRender(jsonnetChart)
	if err != nil {
		logrus.Errorf("failed to load jsonnet template files to render : %s", err.Error())
		return nil, err
	}

	defaultValues := map[string]interface{}{}
	configJsonStr, _ := renderConfigJsonnetFile(templateFiles)
	if configJsonStr != "" {
		err = json.Unmarshal([]byte(configJsonStr), &defaultValues)
		if err != nil {
			logrus.Errorf("failed to unmarshal config json string : %s", err.Error())
			return nil, err
		}
	}

	configValues, err := buildConfigValuesToRender(releaseNamespace, releaseName, jsonnetChart, userConfigs, dependencyConfigs, defaultValues)
	if err != nil {
		logrus.Errorf("failed to build config values to render jsonnet template files : %s", err.Error())
		return nil, err
	}

	defaultValuesBytes, err := yaml.Marshal(defaultValues)
	if err != nil {
		logrus.Errorf("failed to marshal default config values : %s", err.Error())
		return nil, err
	}
	nativeChart.Values = &chart.Config{Raw: string(defaultValuesBytes)}

	jsonStr, err := renderMainJsonnetFile(templateFiles, configValues)
	if err != nil {
		logrus.Errorf("failed to render jsonnet files : %s", err.Error())
		return nil, err
	}

	k8sResources, err := buildK8sResourcesByJsonStr(jsonStr)
	if err != nil {
		logrus.Errorf("failed to build native chart templates : %s", err.Error())
		return nil, err
	}

	//TODO walm pre hook : do something after rendering k8s resources, before making them into native chart templates and install them

	nativeChart.Templates = []*chart.Template{}
	for fileName, k8sResource := range k8sResources {
		ok, outputConfig, err := isAppDummyService(k8sResource)
		if err != nil {
			logrus.Errorf("failed to check whether %s is app dummy service : %s", fileName, err.Error())
			return nil, err
		}
		if ok {
			releaseConfig := &v1beta1.ReleaseConfig{
				TypeMeta: metav1.TypeMeta{
					Kind: "ReleaseConfig",
					APIVersion: "apiextensions.transwarp.io/v1beta1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: releaseNamespace,
					Name:      releaseName,
				},
				Spec: v1beta1.ReleaseConfigSpec{
					DependenciesConfigValues: dependencyConfigs,
					ChartVersion:             nativeChart.Metadata.Version,
					ChartName:                nativeChart.Metadata.Name,
					ChartAppVersion:          nativeChart.Metadata.AppVersion,
					ConfigValues:             userConfigs,
					Dependencies:             dependencies,
					OutputConfig:             outputConfig,
				},
			}
			k8sResource = releaseConfig
		}

		k8sResourceBytes, err := yaml.Marshal(k8sResource)
		if err != nil {
			logrus.Errorf("failed to marshal k8s resource : %s", err.Error())
			return nil, err
		}
		nativeChart.Templates = append(nativeChart.Templates, &chart.Template{
			Name: buildNotRenderedFileName(fileName),
			Data: k8sResourceBytes,
		})
	}

	return
}

func isAppDummyService(k8sResource runtime.Object) (is bool, outputConfig map[string]interface{}, err error) {
	if k8sResource.GetObjectKind().GroupVersionKind().Kind == "Service" {
		service := k8sResource.(*corev1.Service)
		if len(service.Labels) > 0 && service.Labels["transwarp.meta"] == "true" {
			is = true
			if len(service.Annotations) > 0 {
				transwarpMetaStr := service.Annotations["transwarp.meta"]
				outputConfig = map[string]interface{}{}
				err = json.Unmarshal([]byte(transwarpMetaStr), &outputConfig)
				if err != nil {
					logrus.Errorf("failed to unmarshal transwarp meta string : %s", err.Error())
					return
				}
			}
		}
	}
	return
}

func parseSvc(svc *corev1.Service) (isDummyService bool, transwarpMetaStr, releaseName, namespace string) {
	if len(svc.Labels) > 0 && svc.Labels["transwarp.meta"] == "true" {
		isDummyService = true
		releaseName = svc.Labels["release"]
		namespace = svc.Namespace
		if len(svc.Annotations) > 0 {
			transwarpMetaStr = svc.Annotations["transwarp.meta"]
		}
	}
	return
}

func buildConfigValuesToRender(namespace string, name string, jsonnetChart *chart.Chart, userConfigs map[string]interface{}, dependencyConfigs map[string]interface{}, jsonDefaultValues map[string]interface{}) (configValues map[string]interface{}, err error) {
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

	configValues["Transwarp_Install_ID"] = name
	configValues["Transwarp_Install_Namespace"] = namespace
	configValues["TosVersion"] = "1.9"
	configValues["Customized_Namespace"] = namespace
	mergeValues(configValues, userConfigs)

	mergeValues(jsonDefaultValues, configValues)
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

func loadCommonJsonnetLib(templates map[string]string) (err error) {
	if commonTemplateFiles == nil {
		if len(commonTemplateFilesPath) == 0 && setting.Config.V2Config != nil && setting.Config.V2Config.JsonnetConfig != nil {
			commonTemplateFilesPath = setting.Config.V2Config.JsonnetConfig.CommonTemplateFilesPath
		}
		if commonTemplateFilesPath == "" {
			return
		}
		commonTemplateFiles, err = loadFilesFromDisk(commonTemplateFilesPath)
		if err != nil {
			logrus.Errorf("failed to load common template files : %s", err.Error())
			return
		}
	}
	for key, value := range commonTemplateFiles {
		templates[path.Join(commonTemplateDir, filepath.Base(key))] = value
	}
	return nil
}

// LoadFilesFromDisk loads all files inside baseDir directory and its subdirectory recursively,
// mapping each file's path/content as a key/value into a map.
func loadFilesFromDisk(baseDir string) (map[string]string, error) {
	cacheFiles := make(map[string]string)
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			b, err := ioutil.ReadFile(path)
			if err != nil {
				logrus.Errorf("Read file \"%s\", err: %v", path, err)
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
