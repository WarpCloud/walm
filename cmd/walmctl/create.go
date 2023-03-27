package main

import (
	"WarpCloud/walm/cmd/walmctl/util"
	"WarpCloud/walm/cmd/walmctl/util/walmctlclient"
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
	sourceType  string
	projectName string
	file        string
	name        string
	withchart   string
	timeoutSec  int64
	async       bool
	dryrun      bool
	out         io.Writer
}

func newCreateCmd(out io.Writer) *cobra.Command {
	cc := &createCmd{out: out}

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
	cmd.PersistentFlags().StringVarP(&cc.projectName, "project", "p", "", "operate resources of the project")
	cmd.PersistentFlags().StringVarP(&cc.file, "file", "f", "", "absolutely or relative path to source file")
	cmd.PersistentFlags().StringVar(&cc.name, "name", "", "name for release or project you create, overrides field name in file, required!!!")
	cmd.PersistentFlags().StringVar(&cc.withchart, "withchart", "", "update release with local chart, absolutely or relative path to source file")
	cmd.PersistentFlags().Int64Var(&cc.timeoutSec, "timeoutSec", 0, "timeout")
	cmd.PersistentFlags().BoolVar(&cc.async, "async", false, "whether asynchronous")
	cmd.PersistentFlags().BoolVar(&cc.dryrun, "dryrun", false, "dry run")

	cmd.MarkPersistentFlagRequired("name")
	return cmd
}

func (cc *createCmd) run() error {
	var (
		err          error
		filePath     string
		chartPath    string
		fileBytes    []byte
		configValues map[string]interface{}
	)
	chartPath = cc.withchart
	if chartPath != "" {
		chartPath, err = filepath.Abs(cc.withchart)
		if err != nil {
			return err
		}
	}

	if cc.file != "" {
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
	}

	destConfigValues, _, _, err := util.SmartConfigValues(configValues)
	if err != nil {
		klog.Errorf("smart yaml Unmarshal file %s error %v", cc.file, err)
		return err
	}

	client, err := walmctlclient.CreateNewClient(walmserver, enableTLS, rootCA)
	if err != nil {
		klog.Errorf("failed to create walmctl client: %s", err.Error())
		return err
	}
	if err = client.ValidateHostConnect(walmserver); err != nil {
		return err
	}
	if cc.sourceType == "release" {
		if cc.dryrun {
			klog.Infof("Dry Run %s %s", namespace, cc.name)
			response, err := client.DryRunCreateRelease(namespace, chartPath, cc.name, destConfigValues)
			klog.Infof("%v", response)
			if err != nil {
				klog.Errorf("error %v", err)
			}
			return nil
		}
		if cc.projectName == "" {
			_, err = client.CreateRelease(namespace, chartPath, cc.name, cc.async, cc.timeoutSec, destConfigValues)
		} else {
			_, err = client.AddReleaseInProject(namespace, cc.name, cc.projectName, cc.async, cc.timeoutSec, destConfigValues)
		}
	} else {
		if cc.name == "" {
			return errProjectNameRequired
		}
		_, err = client.CreateProject(namespace, chartPath, cc.name, cc.async, cc.timeoutSec, destConfigValues)
	}

	if err != nil {
		panic(err)
	}

	fmt.Printf("%s %s Create Succeed!\n", cc.sourceType, cc.name)
	return nil
}
