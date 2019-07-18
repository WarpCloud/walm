/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	fakedisc "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/helm/cmd/helm/require"
	"k8s.io/helm/pkg/action"
	"k8s.io/helm/pkg/chart/loader"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/hapi/release"
	"k8s.io/helm/pkg/storage"
	"k8s.io/helm/pkg/storage/driver"
	"k8s.io/helm/pkg/tiller/environment"
)

var defaultKubeVersion = fmt.Sprintf("%s.%s", chartutil.DefaultKubeVersion.Major, chartutil.DefaultKubeVersion.Minor)

const templateDesc = `
Render chart templates locally and display the output.

This does not require a Kubernetes connection. However, any values that would normally
be retrieved in-cluster will be faked locally. Additionally, no validation is
performed on the resulting manifest files. As a result, there is no assurance that a
file generated from this command will be valid to Kubernetes.
`

type templateOptions struct {
	nameTemplate string // --name-template
	showNotes    bool   // --notes
	releaseName  string // --name
	kubeVersion  string // --kube-version
	outputDir    string // --output-dir

	valuesOptions

	chartPath string
}

func newTemplateCmd(out io.Writer) *cobra.Command {
	o := &templateOptions{}

	cmd := &cobra.Command{
		Use:   "template CHART",
		Short: fmt.Sprintf("locally render templates"),
		Long:  templateDesc,
		Args:  require.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// verify chart path exists
			if _, err := os.Stat(args[0]); err == nil {
				if o.chartPath, err = filepath.Abs(args[0]); err != nil {
					return err
				}
			} else {
				return err
			}
			return o.run(out)
		},
	}

	f := cmd.Flags()
	f.BoolVar(&o.showNotes, "notes", false, "show the computed NOTES.txt file as well")
	f.StringVarP(&o.releaseName, "name", "", "RELEASE-NAME", "release name")
	f.StringVar(&o.nameTemplate, "name-template", "", "specify template used to name the release")
	f.StringVar(&o.kubeVersion, "kube-version", defaultKubeVersion, "kubernetes version used as Capabilities.KubeVersion.Major/Minor")
	f.StringVar(&o.outputDir, "output-dir", "", "writes the executed templates to files in output-dir instead of stdout")
	o.valuesOptions.addFlags(f)

	return cmd
}

func (o *templateOptions) run(out io.Writer) error {
	// get combined values and create config
	config, err := o.mergedValues()
	if err != nil {
		return err
	}

	// If template is specified, try to run the template.
	if o.nameTemplate != "" {
		o.releaseName, err = templateName(o.nameTemplate)
		if err != nil {
			return err
		}
	}

	// Check chart dependencies to make sure all are present in /charts
	c, err := loader.Load(o.chartPath)
	if err != nil {
		return err
	}

	if req := c.Metadata.Dependencies; req != nil {
		if err := checkDependencies(c, req); err != nil {
			return err
		}
	}

	if err := chartutil.ProcessDependencies(c, config); err != nil {
		return err
	}

	// dry-run using the Kubernetes mock
	disc, err := createFakeDiscovery(o.kubeVersion)
	if err != nil {
		return errors.Wrap(err, "could not parse a kubernetes version")
	}

	customConfig := &action.Configuration{
		// Add mock objects in here so it doesn't use Kube API server
		Releases:   storage.Init(driver.NewMemory()),
		KubeClient: &environment.PrintingKubeClient{Out: ioutil.Discard},
		Discovery:  disc,
		Log: func(format string, v ...interface{}) {
			fmt.Fprintf(out, format, v...)
		},
	}
	inst := action.NewInstall(customConfig)
	inst.DryRun = true
	inst.Replace = true // Skip running the name check
	inst.ReleaseName = o.releaseName
	rel, err := inst.Run(c, config)
	if err != nil {
		return err
	}

	if o.outputDir != "" {
		return o.writeAsFiles(rel)
	}
	fmt.Fprintln(out, rel.Manifest)
	if o.showNotes {
		fmt.Fprintf(out, "---\n# Source: %s/templates/NOTES.txt\n", c.Name())
		fmt.Fprintln(out, rel.Info.Notes)
	}
	return nil
}

func (o *templateOptions) writeAsFiles(rel *release.Release) error {
	if _, err := os.Stat(o.outputDir); os.IsNotExist(err) {
		return errors.Errorf("output-dir '%s' does not exist", o.outputDir)
	}
	// At one point we parsed out the returned manifest and created multiple files.
	// I'm not totally sure what the use case was for that.
	filename := filepath.Join(o.outputDir, rel.Name+".yaml")
	return ioutil.WriteFile(filename, []byte(rel.Manifest), 0644)
}

// createFakeDiscovery creates a discovery client and seeds it with mock data.
func createFakeDiscovery(verStr string) (discovery.DiscoveryInterface, error) {
	disc := fake.NewSimpleClientset().Discovery()
	if verStr != "" {
		kv, err := semver.NewVersion(verStr)
		if err != nil {
			return disc, errors.Wrap(err, "could not parse a kubernetes version")
		}
		disc.(*fakedisc.FakeDiscovery).FakedServerVersion = &version.Info{
			Major:      fmt.Sprintf("%d", kv.Major()),
			Minor:      fmt.Sprintf("%d", kv.Minor()),
			GitVersion: fmt.Sprintf("v%d.%d.0", kv.Major(), kv.Minor()),
		}
	}
	return disc, nil
}
