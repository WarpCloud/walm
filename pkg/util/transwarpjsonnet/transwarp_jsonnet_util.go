package transwarpjsonnet

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"encoding/json"
	"io"
	"strings"
	"fmt"
	"github.com/google/go-jsonnet"
	"path/filepath"
	"path"
	"gopkg.in/yaml.v2"
	jsonnetAst "github.com/google/go-jsonnet/ast"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

func RegisterNativeFuncs(vm *jsonnet.VM) {
	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "parseYaml",
		Params: []jsonnetAst.Identifier{"yaml"},
		Func: func(args []interface{}) (res interface{}, err error) {
			ret := []interface{}{}
			data := []byte(args[0].(string))
			d := k8syaml.NewYAMLToJSONDecoder(bytes.NewReader(data))
			for {
				var doc interface{}
				if err := d.Decode(&doc); err != nil {
					if err == io.EOF {
						break
					}
					return nil, err
				}
				ret = append(ret, doc)
			}
			return ret, nil
		},
	})
}

func renderMainJsonnetFile(templateFiles map[string]string, configValues map[string]interface{}) (jsonStr string, err error) {
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

func BuildNotRenderedFileName(fileName string) (notRenderFileName string) {
	notRenderFileName = path.Join(path.Dir(fileName),  path.Base(fileName) + TranswarpJsonetFileSuffix)
	return
}

func buildKubeResourcesByJsonStr(jsonStr string) (resources map[string][]byte, err error) {
	// key: resource.json, value: resource template(map)
	resourcesMap := make(map[string]map[string]interface{})
	err = json.Unmarshal([]byte(jsonStr), &resourcesMap)
	if err != nil {
		logrus.Errorf("failed to unmarshal json string : %s", err.Error())
		return nil, err
	}

	resources = map[string][]byte{}
	for fileName, resource := range resourcesMap {
		resourceBytes, err := yaml.Marshal(resource)
		if err != nil {
			logrus.Errorf("failed to marshal resource to yaml bytes : %s", err.Error())
			return nil, err
		}
		resources[fileName] = resourceBytes
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
	Data map[string]jsonnet.Contents
}

func (importer *MemoryImporter) Import(importedFrom, importedPath string) (content jsonnet.Contents, foundAt string, err  error) {
	path := filepath.Join(filepath.Dir(importedFrom), importedPath)
	// Separator would be \ in windows
	path = filepath.ToSlash(path)
	if c, ok := importer.Data[path]; ok {
		content = c
		foundAt = path
	} else {
		err = fmt.Errorf("Import not available %v", path)
	}

	return
}

func MakeMemoryVM(data map[string]jsonnet.Contents) *jsonnet.VM {
	vm := jsonnet.MakeVM()
	RegisterNativeFuncs(vm)
	vm.Importer(&MemoryImporter{Data: data})
	return vm
}

// parseTemplateWithTLAString parse the templates by specifying values of Top-Level Arguments (TLA)
// The TLAs comes from external json string.
func parseTemplateWithTLAString(templatePath string, tlaVar string, tlaValue string, templateData map[string]string) (string, error) {
	templateContents := map[string]jsonnet.Contents{}
	for k, v := range templateData {
		templateContents[k] = jsonnet.MakeContents(v)
	}
	vm := MakeMemoryVM(templateContents)
	if tlaVar != "" {
		vm.TLACode(tlaVar, tlaValue)
	}

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
