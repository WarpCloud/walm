package main

import (
	"io"
	"github.com/pkg/errors"
	"path/filepath"
	"walm/cmd/walmctl/walmctlclient"
	"github.com/spf13/cobra"
	"fmt"
)

const createDesc = `
This command creates a walm release along with the yaml/json files.
walmctl[create] takes a flag means the filepath of sourcefile 
example: walmctl create xxx.json/xxx.yaml
`

type createCmd struct {
	file 		string
	sourceType	string
	name 	 	string
	timeoutSec  int64
	async       bool
	out      	io.Writer
}


func newCreateCmd(out io.Writer) *cobra.Command {
	cc := &createCmd{out:out}

	cmd := &cobra.Command{
		Use: "create",
		Short: "create a release/project based on json/yaml",
		Long: createDesc,
		RunE: func(cmd *cobra.Command, args []string) error {

			// Todo:// Judge file type suites[json/yaml]

			//
			if len(args) != 1 {
				return errors.New("arguments release|project required after command create")
			}

			cc.sourceType = args[0]
			if cc.sourceType != "release" && cc.sourceType != "project" {
				return errors.New("arguments error, release/project accept only")
			}
			if cc.sourceType == "project" && cc.name == "" {
				return errors.New("flag --name required when create project")
			}

			if cc.file == "" {
				return errors.New("flag -f/--file required after command create")
			}

			return cc.run()
		},
	}

	cmd.Flags().StringVarP(&cc.file, "file", "f", "", "指定创建的资源的路径")
	cmd.Flags().StringVar(&cc.name, "name", "", "指定创建的 project/release 名称")
	cmd.Flags().Int64Var(&cc.timeoutSec, "timeoutSec", 0, "timeout 超时")
	cmd.Flags().BoolVar(&cc.async, "async", false, "是否异步")
	return cmd
}


func (c *createCmd) run() error {

	// Todo:// read url files
	//var resp *resty.Response
	var err error

	filePath, err := filepath.Abs(c.file)
	if err != nil {
		return err
	}



	client := walmctlclient.CreateNewClient(walmserver)

	if c.sourceType == "release" {
		_, err = client.CreateRelease(namespace, c.name, filePath)

	} else {
		_, err = client.CreateProject(namespace, c.name, c.async, c.timeoutSec, filePath)
	}

	if err != nil {
		return err
	}

	fmt.Sprintf("%s %s created", c.sourceType, c.name)
	return nil
}