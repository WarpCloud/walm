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
)

var longLintHelp = `
This command takes a path to a chart and runs a series of tests to verify that
the chart is well-formed.

If the linter encounters things that will cause the chart to fail installation,
it will emit [ERROR] messages. If it encounters issues that break with convention
or recommendation, it will emit [WARNING] messages.
`

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

	return cmd
}

func (lint *lintOptions) run() error {
	if lint.ciPath == "" {
		lint.ciPath = path.Join(lint.chartPath, "ci")
	}

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

		err = transwarpjsonnet.ProcessJsonnetChart(rawChart, testCase.caseNamespace, testCase.caseName,
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
		caseName: "dummycase",
		caseNamespace: "ci-test",
		userConfigs: map[string]interface{}{},
		dependencyConfigs: map[string]interface{}{},
		dependencies: map[string]string{},
		releaseLabels: map[string]string{},
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

func mockInst() *action.Install{
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
