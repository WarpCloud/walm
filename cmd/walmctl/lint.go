package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"helm.sh/helm/pkg/action"
	"helm.sh/helm/pkg/chart"
	"helm.sh/helm/pkg/chart/loader"
	"helm.sh/helm/pkg/chartutil"
	"helm.sh/helm/pkg/kube"
	"helm.sh/helm/pkg/release"
	"helm.sh/helm/pkg/storage"
	"helm.sh/helm/pkg/storage/driver"
	"io/ioutil"
	"k8s.io/klog"
	"os"
	"path"
	"path/filepath"
	"strings"

	metainfo "WarpCloud/walm/pkg/models/release"
	"WarpCloud/walm/pkg/util"
	"WarpCloud/walm/pkg/util/transwarpjsonnet"
	"bytes"
	"encoding/json"
	"github.com/ghodss/yaml"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var longLintHelp = `
This command takes a path to a chart and runs a series of tests to verify that
the chart is well-formed.

If the linter encounters things that will cause the chart to fail installation,
it will emit [ERROR] messages. If it encounters issues that break with convention
or recommendation, it will emit [WARNING] messages.
`

// Todo: marshall metainfo.yaml to defined structure and add validate method in class

type lintOptions struct {
	chartPath  string
	ciPath     string
	kubeconfig string
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
	cmd.MarkPersistentFlagRequired("chartPath")
	cmd.MarkPersistentFlagRequired("kubeconfig")
	return cmd
}

func (lint *lintOptions) run() error {
	isOpenSource := false
	/* whether chart is openSource */
	var metaData map[string]interface{}
	metaDataByte, err := ioutil.ReadFile(path.Join(lint.chartPath, "Chart.yaml"))
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(metaDataByte, &metaData)
	if err != nil {
		return err
	}

	if fmt.Sprint(metaData["engine"]) != "jsonnet" {
		isOpenSource = true
	}
	fmt.Println(isOpenSource)

	/* check charts */
	if lint.ciPath == "" {
		lint.ciPath = path.Join(lint.chartPath, "ci")
	}

	valuesPath := path.Join(lint.chartPath, "values.yaml")
	metainfoPath := path.Join(lint.chartPath, "transwarp-meta/metainfo.yaml")

	valuesByte, err := ioutil.ReadFile(valuesPath)
	if err != nil {
		return err
	}

	metainfoByte, err := ioutil.ReadFile(metainfoPath)
	if err != nil {
		return err
	}

	/* validate yaml format */
	valuesByte, err = yaml.YAMLToJSON(valuesByte)
	if err != nil {
		return errors.Errorf("values.yaml \n%s", err.Error())
	}

	metainfoByte, err = yaml.YAMLToJSON(metainfoByte)
	if err != nil {
		return errors.Errorf("metainfo.yaml \n%s", err.Error())
	}

	/* reject unknown fields */
	var chartMetaInfo metainfo.ChartMetaInfo
	dec := json.NewDecoder(bytes.NewReader(metainfoByte))
	dec.DisallowUnknownFields()

	if err = dec.Decode(&chartMetaInfo); err != nil {
		klog.Errorf("check json decode error %v", err)
		return err
	}

	err = json.Unmarshal(metainfoByte, &chartMetaInfo)
	if err != nil {
		klog.Errorf("json unmarshal metainfo error %v", err)
		return err
	}

	/* check metainfo valid */
	configMaps, err := chartMetaInfo.CheckMetainfoValidate()
	if err != nil {
		return errors.Errorf("metainfo error: %s", err.Error())
	}
	klog.Infof("metainfo.yaml is valid, start check params in values.yaml...")

	/* check params in values */
	err = chartMetaInfo.CheckParamsInValues(string(valuesByte), configMaps)
	if err != nil {
		klog.Warningf("[Warning] check params in values err: %v", err)
	}

	klog.Info("values.yaml is valid...")

	chartLoader, err := loader.Loader(lint.chartPath)
	if err != nil {
		klog.Errorf("read chart error %v...", err)
		return err
	}

	rawChart, err := chartLoader.Load()
	if err != nil {
		klog.Errorf("load chart error %v...", err)
		return err
	}

	if !isOpenSource {
		err = lint.loadJsonnetAppLib(rawChart)
		if err != nil {
			klog.Errorf("load common lib error %v...", err)
			return err
		}
	}

	if req := rawChart.Metadata.Dependencies; req != nil {
		if err := checkDependencies(rawChart, req); err != nil {
			klog.Errorf("check dependencies error %v...", err)
			return err
		}
	}

	testCases, err := lint.loadCICases()
	if err != nil {
		klog.Errorf("find ci cases error %v...", err)
		return err
	}
	for _, testCase := range testCases {
		valueOverride := map[string]interface{}{}
		util.MergeValues(valueOverride, testCase.userConfigs, false)
		util.MergeValues(valueOverride, testCase.dependencyConfigs, false)

		if err := chartutil.ProcessDependencies(rawChart, valueOverride); err != nil {
			return err
		}
		repo := ""
		err = transwarpjsonnet.ProcessJsonnetChart(
			repo, rawChart, testCase.caseNamespace,
			testCase.caseName, testCase.userConfigs, testCase.dependencyConfigs,
			testCase.dependencies, testCase.releaseLabels, "", nil,
			nil, "",
		)
		if err != nil {
			klog.Errorf("processJsonnetChart error %v", err)
			return err
		}

		inst := lint.mockInst()
		inst.Namespace = testCase.caseNamespace
		inst.ReleaseName = testCase.caseName
		rel, err := inst.Run(rawChart, valueOverride)
		if err != nil {
			return err
		}
		klog.Infof("dry run release %s %s success", inst.Namespace, inst.ReleaseName)
		expectCasePath := path.Join(lint.ciPath, "_expect-cases", testCase.caseName)
		fileByte, err := ioutil.ReadFile(expectCasePath)
		if err != nil {
			//return err
		}

		expectChart := string(fileByte)
		err = checkGenReleaseConfig(expectChart, rel.Manifest)
		if err != nil {
			//return err
		}
		lint.writeAsFiles(rel)
	}

	return nil
}

func checkGenReleaseConfig(expectChart string, outputChart string) error {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(expectChart, outputChart, true)
	if len(diffs) > 2 {
		return errors.Errorf("rendered template result is not expected. There are %d diff places.\n%s", len(diffs)-1, dmp.DiffPrettyText(diffs[1:]))
	}
	klog.Infof("rendered template result is expected.")
	return nil
}

func (lint *lintOptions) loadCICases() ([]lintTestCase, error) {
	testCases := make([]lintTestCase, 0)
	cifiles, err := ioutil.ReadDir(lint.ciPath)
	if err != nil {
		return nil, nil
	}

	for _, cifile := range cifiles {
		if !cifile.IsDir() {
			userConfigByte, err := ioutil.ReadFile(path.Join(lint.ciPath, cifile.Name()))
			if err != nil {
				return nil, err
			}
			userConfigByte, err = yaml.YAMLToJSON(userConfigByte)
			if err != nil {
				return nil, err
			}
			userConfig := map[string]interface{}{}
			err = json.Unmarshal(userConfigByte, &userConfig)

			if err != nil {
				err = errors.Errorf("%s in\n %s", err.Error(), path.Join(lint.ciPath, cifile.Name()))
				return nil, err
			}

			dummyCase := lintTestCase{
				caseName:          cifile.Name(),
				caseNamespace:     "ci-test",
				userConfigs:       userConfig,
				dependencyConfigs: map[string]interface{}{},
				dependencies:      map[string]string{},
				releaseLabels:     map[string]string{},
			}

			testCases = append(testCases, dummyCase)
		}
	}

	return testCases, nil
}

func (lint *lintOptions) writeAsFiles(rel *release.Release) error {
	outputDir := path.Join(lint.ciPath, "_output-cases")
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.MkdirAll(outputDir, 0755)
	}
	// At one point we parsed out the returned manifest and created multiple files.
	// I'm not totally sure what the use case was for that.
	filename := filepath.Join(outputDir, rel.Name)
	klog.Infof("start write result to %s", filename)
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
				klog.Errorf("Read file \"%s\", err: %v", path, err)
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

func (lint *lintOptions) mockInst() *action.Install {
	// dry-run using the Kubernetes mock
	kc := kube.New(kube.GetConfig(lint.kubeconfig, "", ""), nil)
	customConfig := &action.Configuration{
		// Add mock objects in here so it doesn't use Kube API server
		Releases:         storage.Init(driver.NewMemory()),
		KubeClient:       kc,
		RESTClientGetter: kube.GetConfig(lint.kubeconfig, "", ""),
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
