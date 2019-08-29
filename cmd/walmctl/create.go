package main

import (
	"WarpCloud/walm/cmd/walmctl/util/walmctlclient"
	"WarpCloud/walm/pkg/models/release"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"k8s.io/klog"
	"path/filepath"
	projectModel "WarpCloud/walm/pkg/models/project"
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
	sourceType  string
	projectName string
	file        string
	name        string
	withchart   string
	timeoutSec  int64
	async       bool
	out         io.Writer
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
			if len(args) != 1 {
				return errors.New("arguments release/project required after command create")
			}
			if err := checkResourceType(args[0]); err != nil {
				return err
			}
			cc.sourceType = args[0]
			return cc.run()
		},
	}
	cmd.Flags().StringVarP(&cc.projectName, "project", "p", "", "operate resources of the project")
	cmd.Flags().StringVarP(&cc.file, "file", "f", "", "absolutely or relative path to source file")
	cmd.Flags().StringVar(&cc.name, "name", "", "name for release or project you create, overrides field name in file, required!!!")
	cmd.Flags().StringVar(&cc.withchart, "withchart", "", "update release with local chart, absolutely or relative path to source file")
	cmd.Flags().Int64Var(&cc.timeoutSec, "timeoutSec", 0, "timeout")
	cmd.Flags().BoolVar(&cc.async, "async", true, "whether asynchronous")

	cmd.MarkFlagRequired("file")
	return cmd
}

func (cc *createCmd) run() error {
	var (
		err          error
		filePath     string
		chartPath    string
		fileBytes	 []byte
		configValues map[string]interface{}
	)
	chartPath = cc.withchart
	if chartPath != "" {
		chartPath, err = filepath.Abs(cc.withchart)
		if err != nil {
			return err
		}
	}

	filePath, err = filepath.Abs(cc.file)
	if err != nil {
		return err
	}
	fileBytes, err = ioutil.ReadFile(filePath)
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
		if cc.projectName == "" {
			_, err = client.CreateRelease(namespace, chartPath, cc.name, cc.async, cc.timeoutSec, configValues)
		} else {
			_, err = client.AddReleaseInProject(namespace, cc.name, cc.projectName, cc.async, cc.timeoutSec, configValues)
		}
	} else {
		if cc.name == "" {
			return errProjectNameRequired
		}
		_, err = client.CreateProject(namespace, chartPath, cc.name, cc.async, cc.timeoutSec, configValues)
	}

	if err != nil {
		panic(err)
	}

	fmt.Printf("%s %s Create Succeed!\n", cc.sourceType, cc.name)
	return nil
}
