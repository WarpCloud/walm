package helm

import (
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/chart"
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
	"mime/multipart"
	"k8s.io/helm/pkg/chart/loader"
	"strings"
	"github.com/pkg/errors"
	"k8s.io/helm/pkg/ignore"
	"k8s.io/helm/pkg/sympath"
	"compress/gzip"
	"archive/tar"
	"bytes"
	"io"
	"fmt"
)

const (
	commonTemplateDir = "templates/applib/ksonnet-lib"
)

var commonTemplateFilesPath string
var commonTemplateFiles map[string]string

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

	nativeChart.Values = defaultValues

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

	nativeChart.Templates = []*chart.File{}
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
		nativeChart.Templates = append(nativeChart.Templates, &chart.File{
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
	MergeValues(configValues, jsonnetChart.Values)
	//TODO merge system values

	MergeValues(configValues, dependencyConfigs)

	configValues["Transwarp_Install_ID"] = name
	configValues["Transwarp_Install_Namespace"] = namespace
	configValues["TosVersion"] = "1.9"
	configValues["Customized_Namespace"] = namespace
	MergeValues(configValues, userConfigs)

	MergeValues(jsonDefaultValues, configValues)
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

func LoadChartByPath(chartPath string) (isJsonnetChart bool, nativeChart, jsonnetChart *chart.Chart,err error) {
	nativeChart, err = loader.Load(chartPath)
	if err != nil {
		logrus.Errorf("failed to load chart : %s", err.Error())
		return
	}

	isJsonnetChart, jsonnetChart, err = parseJsonnetChart(chartPath, nativeChart)
	return
}

func LoadChartByArchive(chartArchive multipart.File) (isJsonnetChart bool, nativeChart, jsonnetChart *chart.Chart,err error) {
	defer chartArchive.Close()
	nativeChart, err = loader.LoadArchive(chartArchive)
	if err != nil {
		logrus.Errorf("failed to load chart : %s", err.Error())
		return
	}

	files, err := loadArchive(chartArchive)
	if err != nil {
		return false, nil,nil, err
	}
	isJsonnetChart, jsonnetChart, err = loadFiles(files, nativeChart)
	return
}

func parseJsonnetChart(chartPath string, nativeChart *chart.Chart) (isJsonnetChart bool, chart *chart.Chart,err error) {
	fi, err := os.Stat(chartPath)
	if err != nil {
		return false, nil, err
	}

	files := []*BufferedFile{}
	if fi.IsDir() {
		files, err = loadDir(chartPath)
		if err != nil {
			return
		}
	} else {
		raw, err := os.Open(chartPath)
		if err != nil {
			return false, nil, err
		}
		defer raw.Close()
		files, err = loadArchive(raw)
		if err != nil {
			return false, nil, err
		}
	}

	isJsonnetChart, chart, err = loadFiles(files, nativeChart)
	return
}

type BufferedFile struct {
	Name string
	Data []byte
}

func loadDir(dir string) (files []*BufferedFile, err error) {
	topdir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	files = []*BufferedFile{}

	rules := ignore.Empty()
	ifile := filepath.Join(topdir, ignore.HelmIgnore)
	if _, err := os.Stat(ifile); err == nil {
		r, err := ignore.ParseFile(ifile)
		if err != nil {
			return nil, err
		}
		rules = r
	}
	rules.AddDefaults()


	topdir += string(filepath.Separator)

	walk := func(name string, fi os.FileInfo, err error) error {
		n := strings.TrimPrefix(name, topdir)
		if n == "" {
			// No need to process top level. Avoid bug with helmignore .* matching
			// empty names. See issue 1779.
			return nil
		}

		// Normalize to / since it will also work on Windows
		n = filepath.ToSlash(n)

		if err != nil {
			return err
		}
		if fi.IsDir() {
			// Directory-based ignore rules should involve skipping the entire
			// contents of that directory.
			if rules.Ignore(n, fi) {
				return filepath.SkipDir
			}
			return nil
		}

		// If a .helmignore file matches, skip this file.
		if rules.Ignore(n, fi) {
			return nil
		}

		data, err := ioutil.ReadFile(name)
		if err != nil {
			return errors.Wrapf(err, "error reading %s", n)
		}

		files = append(files, &BufferedFile{Name: n, Data: data})
		return nil
	}
	if err = sympath.Walk(topdir, walk); err != nil {
		return
	}
	return
}

func loadArchive(in io.Reader) ([]*BufferedFile, error) {
	unzipped, err := gzip.NewReader(in)
	if err != nil {
		return nil, err
	}
	defer unzipped.Close()

	files := []*BufferedFile{}
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

		files = append(files, &BufferedFile{Name: n, Data: b.Bytes()})
		b.Reset()
	}

	if len(files) == 0 {
		return nil, errors.New("no files in chart archive")
	}

	return files, nil
}

func loadFiles(files []*BufferedFile, nativeChat *chart.Chart) (bool, *chart.Chart, error) {
	jsonnetChart := new(chart.Chart)
	isJsonnetChart := false

	for _, f := range files {
		if strings.HasPrefix(f.Name, "template-jsonnet/") {
			isJsonnetChart = true
			cname := strings.TrimPrefix(f.Name, "template-jsonnet/")
			if strings.IndexAny(cname, "._") == 0 {
				// Ignore charts/ that start with . or _.
				continue
			}
			jsonnetChart.Templates = append(jsonnetChart.Templates, &chart.File{Name: fmt.Sprintf("templates/%s", cname), Data: f.Data})
		}
	}

	if isJsonnetChart {
		jsonnetChart.Values = nativeChat.Values
		jsonnetChart.Metadata = nativeChat.Metadata
		jsonnetChart.Files = nativeChat.Files
		return true, jsonnetChart, nil
	} else {
		return false, nativeChat, nil
	}


}