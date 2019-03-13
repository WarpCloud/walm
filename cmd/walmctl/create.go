package main

import (
	"io"
	"github.com/pkg/errors"
	"path/filepath"
	"walm/cmd/walmctl/walmctlclient"
	"github.com/spf13/cobra"
	"fmt"
)

const createDesc = `This command creates a walm release or project along with existing yaml/json files.
'walmctl create' takes a parameter which decide the source type, release or project.
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
		Long:  createDesc,
		RunE: func(cmd *cobra.Command, args []string) error {

			if namespace == "" {
				return errors.New("flag --namespace/-n required")
			}
			if walmserver == "" {
				return errors.New("flag --server/-s required")
			}
			if len(args) != 1 {
				return errors.New("arguments release/project required after command create")
			}

			cc.sourceType = args[0]

			if cc.sourceType != "release" && cc.sourceType != "project" {
				return errors.New("arguments error, release/project accept only")
			}
			return cc.run()
		},
	}

	cmd.Flags().StringVarP(&cc.file, "file", "f", "", "absolutely or relative path to source file")
	cmd.Flags().StringVar(&cc.sourceName, "name", "", "releaseName or projectName, Overrides name value in file, optional for release, but required for project")
	cmd.Flags().Int64Var(&cc.timeoutSec, "timeoutSec", 0, "timeout, (default 0)")
	cmd.Flags().BoolVar(&cc.async, "async", true, "whether asynchronous")

	cmd.MarkFlagRequired("file")
	cmd.MarkFlagRequired("name")

	return cmd
}

func (c *createCmd) run() error {

	// Todo:// read url files
	var err error

	filePath, err := filepath.Abs(c.file)
	if err != nil {
		return err
	}

	client := walmctlclient.CreateNewClient(walmserver)

	if c.sourceType == "release" {
		_, err = client.CreateRelease(namespace, c.sourceName, c.async, c.timeoutSec, filePath)

	} else {
		_, err = client.CreateProject(namespace, c.sourceName, c.async, c.timeoutSec, filePath)
	}

	if err != nil {
		return err
	}

	fmt.Printf("%s %s created", c.sourceType, c.sourceName)
	return nil
}
