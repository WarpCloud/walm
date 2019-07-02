package main

import (
	"io"
	"path/filepath"
	"WarpCloud/walm/cmd/walmctl/util/walmctlclient"
	"github.com/spf13/cobra"
	"fmt"
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
	sourceType string
	sourceName string
	timeoutSec int64
	async      bool
	out        io.Writer
}

func newCreateCmd(out io.Writer) *cobra.Command {
	cc := &createCmd{out: out}

	cmd := &cobra.Command{
		Use:   "create release/project",
		Short: "create a release/project based on json/yaml",
		DisableFlagsInUseLine: true,
		Long:  createDesc,

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
	cmd.MarkFlagRequired("namespace")
	cmd.MarkFlagRequired("file")
	return cmd
}

func (cc *createCmd) run() error {

	err := checkResourceType(cc.sourceType)
	if err != nil {
		return err
	}

	filePath, err := filepath.Abs(cc.file)
	if err != nil {
		return err
	}

	client := walmctlclient.CreateNewClient(walmserver)

	if cc.sourceType == "release" {
		_, err = client.CreateRelease(namespace, cc.sourceName, cc.async, cc.timeoutSec, filePath)

	} else {
		if cc.sourceName == "" {
			return errProjectNameRequired
		}
		_, err = client.CreateProject(namespace, cc.sourceName, cc.async, cc.timeoutSec, filePath)
	}

	if err != nil {
		return err
	}

	fmt.Printf("%s %s created", cc.sourceType, cc.sourceName)
	return nil
}
