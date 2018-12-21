package helm

import (
	"github.com/sirupsen/logrus"
	"encoding/json"
	"strings"
	"fmt"
	"github.com/google/go-jsonnet"
	"path/filepath"
	"sort"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"path"
)

func renderJsonnetFiles(templateFiles map[string]string, configValues map[string]interface{}) (jsonStr string, err error) {
	mainJsonFileName, err := getMainJsonnetFile(templateFiles)
	if err != nil {
		logrus.Errorf("failed to get main jsonnet file : %s", err.Error())
		return "", err
	}

	tlaValue, err := json.Marshal(configValues)
	if err != nil {
		logrus.Errorf("failed to marshal config values : %s", err.Error())
		return "", err
	}

	jsonStr, err = parseTemplateWithTLAString(mainJsonFileName, "config", string(tlaValue), templateFiles)
	if err != nil {
		logrus.Errorf("failed to parse main jsonnet template file : %s", err.Error())
		return "", err
	}
	return
}

func buildNotRenderedFileName(fileName string) (notRenderFileName string) {
	notRenderFileName = path.Join(path.Dir(fileName),  "NOTRENDER-" + path.Base(fileName))
	return
}

func buildK8sResourcesByJsonStr(jsonStr string) (resources map[string]runtime.Object, err error) {
	// key: resource.json, value: resource template(map)
	resourcesMap := make(map[string]map[string]interface{})
	err = json.Unmarshal([]byte(jsonStr), &resourcesMap)
	if err != nil {
		logrus.Errorf("failed to unmarshal json string : %s", err.Error())
		return nil, err
	}

	//TODO use hook to fix the issue that configMap, secret load first
	keys := make([]string, 0, len(resourcesMap))
	for k := range resourcesMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// WARP-28357 reorder keys to confirm configmap will be synced before other resources, e.g. statefulset
	startAt := 0
	for keyIdx, key := range keys {
		switch resourcesMap[key]["kind"] {
		case "ConfigMap", "Secret":
			if startAt < keyIdx {
				keys[startAt], keys[keyIdx] = keys[keyIdx], keys[startAt]
			}
			startAt = startAt + 1
		default:
		}
	}

	resources = make(map[string]runtime.Object, len(resourcesMap))
	for _, key := range keys {
		resource := resourcesMap[key]
		apiVersion, exists := resource["apiVersion"]
		if !exists {
			err = fmt.Errorf("%s does not have apiVersion", key)
			logrus.Errorf(err.Error())
			return
		}
		kind, exists := resource["kind"]
		if !exists {
			err = fmt.Errorf("%s does not have kind", key)
			logrus.Errorf(err.Error())
			return
		}
		var group, version string
		gvs := strings.Split(apiVersion.(string), "/")
		if len(gvs) == 2 {
			group, version = gvs[0], gvs[1]
		} else {
			group, version = "", apiVersion.(string)
		}
		defaultGVK := schema.GroupVersionKind{Group: group, Version: version, Kind: kind.(string)}
		resourceBytes, err := json.Marshal(resource)
		if err != nil {
			logrus.Errorf("failed to marshal resource : %s", key)
			return nil, err
		}

		decoder := scheme.Codecs.UniversalDecoder(defaultGVK.GroupVersion())
		obj, gvk, err := decoder.Decode(resourceBytes, &defaultGVK, nil)
		if err != nil {
			logrus.Errorf("failed to decode resource : %s", key)
			return nil, err
		}
		if gvk.GroupVersion() != defaultGVK.GroupVersion() {
			err = fmt.Errorf("API version in the data (%s) does not match expected API version (%s)", gvk.GroupVersion().String(), defaultGVK.GroupVersion().String())
			logrus.Errorf(err.Error())
			return nil, err
		}

		resources[key] = obj
	}
	return
}

func getMainJsonnetFile(templateFiles map[string]string) (string, error) {
	for fileName := range templateFiles {
		if strings.HasSuffix(fileName, "main.jsonnet") {
			return fileName, nil
		}
	}
	return "", fmt.Errorf("failed to find main jsonnet file")
}

type MemoryImporter struct {
	Data map[string]string
}

func (importer *MemoryImporter) Import(dir, importedPath string) (*jsonnet.ImportedData, error) {
	path := filepath.Join(dir, importedPath)
	// Separator would be \ in windows
	path = filepath.ToSlash(path)
	if content, ok := importer.Data[path]; ok {
		return &jsonnet.ImportedData{Content: content, FoundHere: path}, nil
	}
	return nil, fmt.Errorf("Import not available %v", path)
}

func MakeMemoryVM(data map[string]string) *jsonnet.VM {
	vm := jsonnet.MakeVM()
	vm.Importer(&MemoryImporter{Data: data})
	return vm
}

// parseTemplateWithTLAString parse the templates by specifying values of Top-Level Arguments (TLA)
// The TLAs comes from external json string.
func parseTemplateWithTLAString(templatePath string, tlaVar string, tlaValue string, templateData map[string]string) (string, error) {
	vm := MakeMemoryVM(templateData)
	vm.TLACode(tlaVar, tlaValue)
	if _, ok := templateData[templatePath]; !ok {
		logrus.Errorf("failed to find entrance of template : %s", templatePath)
		return "", fmt.Errorf("failed to find entrance of template : %s", templatePath)
	}
	output, err := vm.EvaluateSnippet(templatePath, templateData[templatePath])
	if err != nil {
		logrus.Errorf("failed to parse template %s, %s=%s, error: %+v", templatePath, tlaVar, tlaValue, err)
		return "", err
	}
	return string(output), nil
}
