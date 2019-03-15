package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/helm/pkg/action"
	"k8s.io/helm/pkg/chart"
	"k8s.io/helm/pkg/chart/loader"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/hapi/release"
	"k8s.io/helm/pkg/storage"
	"k8s.io/helm/pkg/storage/driver"
	"k8s.io/helm/pkg/tiller/environment"
	"os"
	"path"
	"path/filepath"
	"strings"
	"walm/pkg/util"
	"walm/pkg/util/transwarpjsonnet"
	"strconv"
	"github.com/tidwall/gjson"
	"github.com/ghodss/yaml"
)

var longLintHelp = `
This command takes a path to a chart and runs a series of tests to verify that
the chart is well-formed.

If the linter encounters things that will cause the chart to fail installation,
it will emit [ERROR] messages. If it encounters issues that break with convention
or recommendation, it will emit [WARNING] messages.
`
// Todo: check and raise warning signal when unrecognized fields in metainfo.yaml

type lintOptions struct {
	chartPath  string
	ciPath     string
	kubeconfig string
}

type lintTypeCheck struct {
	mapKey   string
	Type     string
	required bool
	path     string
}

type lintTestCase struct {
	caseName          string
	caseNamespace     string
	userConfigs       map[string]interface{}
	dependencyConfigs map[string]interface{}
	dependencies      map[string]string
	releaseLabels     map[string]string
}

func newLintCmd() *cobra.Command {
	lint := &lintOptions{chartPath: "."}

	cmd := &cobra.Command{
		Use:   "lint PATH",
		Short: "examines a chart for possible issues",
		Long:  longLintHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			return lint.run()
		},
	}
	cmd.PersistentFlags().StringVar(&lint.chartPath, "chartPath", ".", "test transwarp chart path")
	cmd.PersistentFlags().StringVar(&lint.ciPath, "ciPath", "", "test chart ci path")
	cmd.PersistentFlags().StringVar(&lint.kubeconfig, "kubeconfig", "kubeconfig", "kubeconfig path")

	return cmd
}

func (lint *lintOptions) run() error {
	if lint.ciPath == "" {
		lint.ciPath = path.Join(lint.chartPath, "ci")
	}

	metainfoPath := path.Join(lint.chartPath, "transwarp-meta/metainfo.yaml")
	valuesPath := path.Join(lint.chartPath, "values.yaml")
	err := checkMapKeys(metainfoPath, valuesPath)
	if err != nil {
		return err
	}

	logrus.Println("map keys check correct")

	// 2. generate cases and dry run
	chartLoader, err := loader.Loader(lint.chartPath)
	if err != nil {
		return err
	}

	rawChart, err := chartLoader.Load()
	if err != nil {
		return err
	}
	err = lint.loadJsonnetAppLib(rawChart)
	if err != nil {
		return err
	}

	if req := rawChart.Metadata.Dependencies; req != nil {
		if err := checkDependencies(rawChart, req); err != nil {
			return err
		}
	}

	testCases, err := lint.loadCICases()
	for _, testCase := range testCases {
		valueOverride := map[string]interface{}{}
		util.MergeValues(valueOverride, testCase.userConfigs)
		util.MergeValues(valueOverride, testCase.dependencyConfigs)

		if err := chartutil.ProcessDependencies(rawChart, valueOverride); err != nil {
			return err
		}
		repo := ""
		err = transwarpjsonnet.ProcessJsonnetChart(repo, rawChart, testCase.caseNamespace, testCase.caseName,
			testCase.userConfigs, testCase.dependencyConfigs, testCase.dependencies, testCase.releaseLabels)

		inst := mockInst()
		inst.Namespace = testCase.caseNamespace
		inst.ReleaseName = testCase.caseName
		rel, err := inst.Run(rawChart, valueOverride)
		if err != nil {
			return err
		}

		lint.writeAsFiles(rel)
	}

	return nil
}

func (lint *lintOptions) loadCICases() ([]lintTestCase, error) {
	testCases := make([]lintTestCase, 0)
	dummyCase := lintTestCase{
		caseName:          "dummycase",
		caseNamespace:     "ci-test",
		userConfigs:       map[string]interface{}{},
		dependencyConfigs: map[string]interface{}{},
		dependencies:      map[string]string{},
		releaseLabels:     map[string]string{},
	}

	testCases = append(testCases, dummyCase)
	return testCases, nil
}

func (lint *lintOptions) writeAsFiles(rel *release.Release) error {
	outputDir := path.Join(lint.ciPath, "_output-cases")
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.MkdirAll(outputDir, 0755)
	}
	// At one point we parsed out the returned manifest and created multiple files.
	// I'm not totally sure what the use case was for that.
	filename := filepath.Join(outputDir, rel.Name+".yaml")
	return ioutil.WriteFile(filename, []byte(rel.Manifest), 0644)
}

func (lint *lintOptions) loadJsonnetAppLib(ch *chart.Chart) error {
	appLibDir := path.Join(lint.chartPath, "../../applib")
	err := filepath.Walk(appLibDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			b, err := ioutil.ReadFile(path)
			if err != nil {
				logrus.Errorf("Read file \"%s\", err: %v", path, err)
				return err
			}

			appSubPaths := strings.Split(path, "applib")
			chartAppLibName := "applib" + appSubPaths[1]
			file := chart.File{
				Name: chartAppLibName,
				Data: b,
			}
			ch.Files = append(ch.Files, &file)
		}
		return nil
	})

	return err
}

func checkMapKeys(metainfoPath string, valuesPath string) error {

	// check if metainfo and valuesInfo exists
	metainfoData, err := ioutil.ReadFile(metainfoPath)
	if err != nil {
		return err
	}

	// check suit yaml format
	newMetainfoData, err := yaml.YAMLToJSON(metainfoData)

	if  err != nil {
		return errors.Errorf("%s in metainfo.yaml", err.Error())

	}

	formatMetainfoData, err := yaml.JSONToYAML(newMetainfoData)
	if err != nil {
		return errors.Errorf("format %s to yaml style error", metainfoPath)
	}
	ioutil.WriteFile(metainfoPath, formatMetainfoData, 0666)

	if err != nil {
		return err
	}

	// get all map keys
	metainfoStr := string(newMetainfoData)
	typeChecks, err := getTypeCheck(metainfoStr)
	if err != nil {
		return err
	}
	// check all map keys
	valuesData, err := ioutil.ReadFile(valuesPath)
	if err != nil {
		return err
	}
	newValuesData, err := yaml.YAMLToJSON(valuesData)
	if err != nil {
		return errors.Errorf("%s in values.yaml", err.Error())
	}
	valuesStr := string(newValuesData)

	typeMap := make(map[string]string)
	typeMap["String"] = "string"
	typeMap["True"] = "boolean"
	typeMap["False"] = "boolean"
	typeMap["Number"] = "number"
	// Null, JSON -->  no type, env_list

	for _, typeCheck := range typeChecks {

		// check map key exist
		result := gjson.Get(valuesStr, typeCheck.mapKey)

		if !result.Exists() {
			return errors.Errorf("%s not exist in values.yaml", typeCheck.mapKey)
		}

		if strings.Contains(typeCheck.mapKey, "priority") {
			if result.Type.String() == "Number" {
				continue
			} else {
				return errors.Errorf("%s value type error, number type required", typeCheck.mapKey)
			}
		}

		if typeCheck.Type == "envType" || typeCheck.Type == "" {
			continue
		}

		// check map key type
		metainfoType := typeCheck.Type
		valuesType := typeMap[result.Type.String()]
		if valuesType == "" && typeCheck.required == false {
			continue
		}
		if metainfoType != valuesType {
			return errors.Errorf("%s value type error, %s in metainfo.yaml while %s in values.yaml",
				typeCheck.mapKey, metainfoType, valuesType)
		}
	}
	return nil
}

func getTypeCheck(metainfoStr string) ([]lintTypeCheck, error) {

	var typeChecks []lintTypeCheck
	var err error

	roleSize := int(gjson.Get(metainfoStr, "roles.#").Int())
	for i := 0; i < roleSize; i++ {
		prePath := "roles." + strconv.Itoa(i)

		// required fields, name, mapKey, type
		baseConfigSize := int(gjson.Get(metainfoStr, prePath+".baseConfig.#").Int())
		rolesName := gjson.Get(metainfoStr, prePath+".baseConfig.#.name").Array()
		rolesMapkey := gjson.Get(metainfoStr, prePath+".baseConfig.#.mapKey").Array()
		rolesType := gjson.Get(metainfoStr, prePath+".baseConfig.#.type").Array()

		if baseConfigSize != len(rolesName) || len(rolesName) != len(rolesMapkey) || len(rolesMapkey) != len(rolesType){
			err = errors.Errorf("name, mapKey, type are required inner each element of %s", prePath+".baseConfig")
			return nil, err
		}


		var rolesRequire []gjson.Result
		for configIndex := 0; configIndex < baseConfigSize; configIndex++ {

			require := gjson.Get(metainfoStr, prePath+".baseConfig." + strconv.Itoa(configIndex) + ".required")
			rolesRequire = append(rolesRequire, require)
		}

		for j := 0; j < len(rolesMapkey); j++ {

			var typeCheck lintTypeCheck
			if !rolesMapkey[j].Exists() {
				err = errors.New(prePath + ".baseConfig." + strconv.Itoa(j) + ".mapKey" + "not exist.")
			}

			if rolesMapkey[j].String() == "" {
				err = errors.New(prePath + ".baseConfig." + strconv.Itoa(j) + ".mapKey can't be empty")
			}

			if !rolesType[j].Exists() {
				err = errors.New(prePath + ".baseConfig." + strconv.Itoa(j) + ".type" + "not exist.")
			}

			if err != nil {
				return nil, err
			}

			typeCheck.mapKey = rolesMapkey[j].String()
			typeCheck.required = rolesRequire[j].Bool()
			typeCheck.Type = rolesType[j].String()
			typeChecks = append(typeChecks, typeCheck)
		}

		resources := gjson.GetMany(metainfoStr,
			prePath+".resources.limitsMemoryKey.mapKey",
			prePath+".resources.limitsCpuKey.mapKey",
			prePath+".resources.limitsGpuKey.mapKey",
			prePath+".resources.requestsMemoryKey.mapKey",
			prePath+".resources.requestsCpuKey.mapKey",
			prePath+".resources.requestsGpuKey.mapKey", )

		if gjson.Get(metainfoStr, prePath+".resources").Exists() {
			for _, resource := range resources {

				var typeCheck lintTypeCheck

				if !resource.Exists() {
					err = errors.Errorf("not enough fields in %s", prePath + ".resources")
					return nil, err
				} else if resource.String() == "" {
					err = errors.Errorf("mapKey can not be null in %s.resources", prePath + "。resources")
					return nil, err
				}

				typeCheck.mapKey = resource.String()
				typeCheck.required = true
				typeChecks = append(typeChecks, typeCheck)
			}
		}

		paramsSize := int(gjson.Get(metainfoStr, "params.#").Int())
		paramsMapKey := gjson.Get(metainfoStr, "params.#.mapKey").Array()
		paramsType := gjson.Get(metainfoStr, "params.#.type").Array()
		paramsRequire := gjson.Get(metainfoStr, "params.#.required").Array()

		for k := 0; k < paramsSize; k++ {
			var typeCheck lintTypeCheck
			if !paramsMapKey[k].Exists() {
				err = errors.New("params." + strconv.Itoa(k) + ".mapKey" + "not exist.")
			}

			if !paramsType[k].Exists() {
				err = errors.New("params." + strconv.Itoa(k) + ".type" + "not exist.")
			}

			if err != nil {
				return nil, err
			}
			typeCheck.mapKey = paramsMapKey[k].String()
			typeCheck.required = paramsRequire[k].Bool()
			typeCheck.Type = paramsType[k].String()
			typeChecks = append(typeChecks, typeCheck)
		}

	}
	return typeChecks, err
}

func mockInst() *action.Install {
	// dry-run using the Kubernetes mock
	disc := fake.NewSimpleClientset().Discovery()

	customConfig := &action.Configuration{
		// Add mock objects in here so it doesn't use Kube API server
		Releases:   storage.Init(driver.NewMemory()),
		KubeClient: &environment.PrintingKubeClient{Out: ioutil.Discard},
		Discovery:  disc,
		Log: func(format string, v ...interface{}) {
			fmt.Fprintf(os.Stdout, format, v...)
		},
	}
	inst := action.NewInstall(customConfig)
	inst.DryRun = true
	inst.Replace = true // Skip running the name check

	return inst
}

func checkDependencies(ch *chart.Chart, reqs []*chart.Dependency) error {
	var missing []string

OUTER:
	for _, r := range reqs {
		for _, d := range ch.Dependencies() {
			if d.Name() == r.Name {
				continue OUTER
			}
		}
		missing = append(missing, r.Name)
	}

	if len(missing) > 0 {
		return errors.Errorf("found in Chart.yaml, but missing in charts/ directory: %s", strings.Join(missing, ", "))
	}
	return nil
}
