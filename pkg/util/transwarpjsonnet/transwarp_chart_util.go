package transwarpjsonnet

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"

	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/chart"
	"k8s.io/helm/pkg/chart/loader"
	"k8s.io/helm/pkg/walm/plugins"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"transwarp/release-config/pkg/apis/transwarp/v1beta1"


	"walm/pkg/setting"
	"walm/pkg/util"
)

const (
	CommonTemplateDir           = "applib/ksonnet-lib"
	TranswarpJsonetFileSuffix   = ".transwarp-jsonnet.yaml"
	TranswarpJsonnetTemplateDir = "template-jsonnet/"
	TranswarpJsonetAppLibDir    = "applib/"
	TranswarpMetadataDir        = "transwarp-meta/"
	TranswarpCiDir              = "ci/"
)

var commonTemplateFilesPath string
var commonTemplateFiles map[string]string

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
			cacheFiles[strings.TrimPrefix(filepath.ToSlash(path), baseDir)] = string(b)
		}
		return nil
	})
	if err != nil {
		return cacheFiles, err
	}
	return cacheFiles, nil
}

func loadCommonJsonnetLib(templates map[string]string) (err error) {
	if commonTemplateFiles == nil {
		if len(commonTemplateFilesPath) == 0 && setting.Config.JsonnetConfig != nil {
			commonTemplateFilesPath = setting.Config.JsonnetConfig.CommonTemplateFilesPath
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
		templates[path.Join(CommonTemplateDir, key)] = value
	}
	return nil
}

func buildConfigValuesToRender(rawChart *chart.Chart, namespace, name string, userConfigs, dependencyConfigs map[string]interface{}) (configValues map[string]interface{}, err error) {
	configValues = map[string]interface{}{}
	util.MergeValues(configValues, rawChart.Values)
	//TODO merge system values

	util.MergeValues(configValues, dependencyConfigs)

	configValues["helmReleaseName"] = name
	configValues["helmReleaseNamespace"] = namespace
	configValues["chartVersion"] = rawChart.Metadata.Version
	configValues["chartName"] = rawChart.Metadata.Name
	configValues["chartAppVersion"] = rawChart.Metadata.AppVersion
	configValues["Transwarp_Install_Namespace"] = namespace

	util.MergeValues(configValues, userConfigs)

	return configValues, nil
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
func ProcessJsonnetChart(rawChart *chart.Chart, releaseNamespace, releaseName string,
	userConfigs, dependencyConfigs map[string]interface{}, dependencies, releaseLabels map[string]string,
) error {
	jsonnetTemplateFiles := make(map[string]string, 0)
	rawChartFiles := []*chart.File{}
	for _, f := range rawChart.Files {
		if strings.HasPrefix(f.Name, TranswarpJsonnetTemplateDir) {
			cname := strings.TrimPrefix(f.Name, TranswarpJsonnetTemplateDir)
			if strings.IndexAny(cname, "._") == 0 {
				// Ignore charts/ that start with . or _.
				continue
			}
			appcname := path.Join(releaseName, rawChart.Metadata.AppVersion, TranswarpJsonnetTemplateDir, cname)
			jsonnetTemplateFiles[appcname] = string(f.Data)
		} else if strings.HasPrefix(f.Name, TranswarpJsonetAppLibDir) {
			jsonnetTemplateFiles[f.Name] = string(f.Data)
		} else if !strings.HasPrefix(f.Name, TranswarpMetadataDir) && !strings.HasPrefix(f.Name, TranswarpCiDir) {
			rawChartFiles = append(rawChartFiles, f)
		}
	}

	autoGenReleaseConfig, err := buildAutoGenReleaseConfig(releaseNamespace, releaseName,
		rawChart.Metadata.Name, rawChart.Metadata.Version, rawChart.Metadata.AppVersion,
		releaseLabels, dependencies, dependencyConfigs, userConfigs)
	if err != nil {
		logrus.Errorf("failed to auto gen release config : %s", err.Error())
		return err
	}
	rawChart.Templates = append(rawChart.Templates, &chart.File{
		Name: BuildNotRenderedFileName("autogen-releaseconfig.json"),
		Data: autoGenReleaseConfig,
	})
	rawChart.Files = rawChartFiles

	if len(jsonnetTemplateFiles) == 0 {
		// native chart
		logrus.Infof("chart %s is native chart", rawChart.Metadata.Name)
		return nil
	}
	// load values.yaml
	valueYamlContent, err := json.Marshal(rawChart.Values)
	jsonnetTemplateFiles[path.Join(releaseName, rawChart.Metadata.AppVersion, "values.yaml")] = string(valueYamlContent)

	loadCommonJsonnetLib(jsonnetTemplateFiles)

	configValues, err := buildConfigValuesToRender(rawChart, releaseNamespace, releaseName, userConfigs, dependencyConfigs)
	if err != nil {
		logrus.Errorf("failed to build config values to render jsonnet template files : %s", err.Error())
		return err
	}

	jsonStr, err := renderMainJsonnetFile(jsonnetTemplateFiles, configValues)
	if err != nil {
		logrus.Errorf("failed to render jsonnet files : %s", err.Error())
		return err
	}

	kubeResources, err := buildKubeResourcesByJsonStr(jsonStr)
	if err != nil {
		logrus.Errorf("failed to build native chart templates : %s", err.Error())
		return err
	}

	for fileName, kubeResourceBytes := range kubeResources {
		rawChart.Templates = append(rawChart.Templates, &chart.File{
			Name: BuildNotRenderedFileName(fileName),
			Data: kubeResourceBytes,
		})
	}
	return nil
}


func buildAutoGenReleaseConfig(releaseNamespace, releaseName, chartName, chartVersion, chartAppVersion string,
	labels, dependencies map[string]string, dependencyConfigs, userConfigs map[string]interface{}) ([]byte, error) {
	if labels == nil {
		labels = map[string]string{}
	}
	labels[plugins.AutoGenLabelKey] = "true"

	releaseConfig := &v1beta1.ReleaseConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ReleaseConfig",
			APIVersion: "apiextensions.transwarp.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: releaseNamespace,
			Name:      releaseName,
			Labels:    labels,
		},
		Spec: v1beta1.ReleaseConfigSpec{
			DependenciesConfigValues: dependencyConfigs,
			ChartVersion:             chartVersion,
			ChartName:                chartName,
			ChartAppVersion:          chartAppVersion,
			ConfigValues:             userConfigs,
			Dependencies:             dependencies,
			OutputConfig:             map[string]interface{}{},
		},
	}

	releaseConfigBytes, err := yaml.Marshal(releaseConfig)
	if err != nil {
		logrus.Errorf("failed to marshal release config : %s", err.Error())
		return nil, err
	}
	return releaseConfigBytes, nil
}

func LoadArchive(in io.Reader) ([]*loader.BufferedFile, error) {
	unzipped, err := gzip.NewReader(in)
	if err != nil {
		return nil, err
	}
	defer unzipped.Close()

	files := []*loader.BufferedFile{}
	tr := tar.NewReader(unzipped)
	for {
		b := bytes.NewBuffer(nil)
		hd, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if hd.FileInfo().IsDir() {
			// Use this instead of hd.Typeflag because we don't have to do any
			// inference chasing.
			continue
		}

		// Archive could contain \ if generated on Windows
		delimiter := "/"
		if strings.ContainsRune(hd.Name, '\\') {
			delimiter = "\\"
		}

		parts := strings.Split(hd.Name, delimiter)
		n := strings.Join(parts[1:], delimiter)

		// Normalize the path to the / delimiter
		n = strings.Replace(n, delimiter, "/", -1)

		if parts[0] == "Chart.yaml" {
			return nil, errors.New("chart yaml not in base directory")
		}

		if _, err := io.Copy(b, tr); err != nil {
			return nil, err
		}

		files = append(files, &loader.BufferedFile{Name: n, Data: b.Bytes()})
		b.Reset()
	}

	if len(files) == 0 {
		return nil, errors.New("no files in chart archive")
	}

	return files, nil
}
