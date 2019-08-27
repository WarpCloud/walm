package main

import (
	"WarpCloud/walm/cmd/walmctl/util/walmctlclient"
	"flag"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"k8s.io/klog"
	"path/filepath"
)

const createDesc = `This command creates a walm release or project(collection of releases) 
along with the common yaml or json formatted file.
For example, 'helm create release -f txsql.yaml' will deploy a txsql
release in specific namespace.

'helm create' takes an argument values release or project to specify
the source type to create. Also, filepath and release/project name required.

Advanced:
There are options you can define, while usually not need.
--async(bool):     asynchronous or synchronous(default asynchronous, false)
--timeoutSec(int): timeout for work(default 0)
`

type createCmd struct {
	file       string
	chart      string
	sourceType string
	sourceName string
	timeoutSec int64
	async      bool
	out        io.Writer
}

func newCreateCmd(out io.Writer) *cobra.Command {
	cc := &createCmd{out: out}
	gofs := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(gofs)

	cmd := &cobra.Command{
		Use:                   "create release/project",
		Short:                 "create a release/project based on json/yaml",
		DisableFlagsInUseLine: true,
		Long:                  createDesc,

		RunE: func(cmd *cobra.Command, args []string) error {

			if walmserver == "" {
				return errServerRequired
			}
			if namespace == "" {
				return errNamespaceRequired
			}
			cc.sourceType = args[0]
			return cc.run()
		},
	}

	cmd.Flags().StringVarP(&cc.file, "file", "f", "", "absolutely or relative path to source file")
	cmd.Flags().StringVar(&cc.sourceName, "name", "", "releaseName or projectName, Overrides name value in file, optional for release, but required for project")
	cmd.Flags().Int64Var(&cc.timeoutSec, "timeoutSec", 0, "timeout")
	cmd.Flags().BoolVar(&cc.async, "async", false, "whether asynchronous")
	cmd.Flags().StringVarP(&cc.chart, "chart", "c", "", "absolutely or relative path to chart location")
	cmd.MarkFlagRequired("namespace")
	cmd.MarkFlagRequired("file")
	return cmd
}

func (cc *createCmd) run() error {
	var err error
	var configValues map[string]interface{}
	chartPath := cc.chart
	if chartPath != "" {
		chartPath, err = filepath.Abs(cc.chart)
		if err != nil {
			return err
		}
	}
	err = checkResourceType(cc.sourceType)
	if err != nil {
		return err
	}
	fileBytes, err := ioutil.ReadFile(cc.file)
	if err != nil {
		klog.Errorf("read file %s error %v", cc.file, err)
		return err
	}
	err = yaml.Unmarshal(fileBytes, &configValues)
	if err != nil {
		klog.Errorf("yaml Unmarshal file %s error %v", cc.file, err)
		return err
	}

	client := walmctlclient.CreateNewClient(walmserver)
	if err = client.ValidateHostConnect(); err != nil {
		return err
	}
	if cc.sourceType == "release" {
		_, err = client.CreateRelease(namespace, chartPath, cc.sourceName, cc.async, cc.timeoutSec, configValues)
	} else {
		if cc.sourceName == "" {
			return errProjectNameRequired
		}
		if chartPath != "" {
			return errors.New("project use chartfile currently not support")
		}
		_, err = client.CreateProject(namespace, chartPath, cc.sourceName, cc.async, cc.timeoutSec, configValues)
	}

	if err != nil {
		panic(err)
	}

	fmt.Printf("%s %s created", cc.sourceType, cc.sourceName)
	return nil
}
