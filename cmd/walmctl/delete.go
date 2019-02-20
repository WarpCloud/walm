package main

import (
	"io"
	"github.com/spf13/cobra"
	"walm/cmd/walmctl/walmctlclient"
	"fmt"
	"errors"
)

const deleteDesc = `
This command delete a walm release, project or a release of project.
Usage:
walmctl delete release releaseName
walmctl delete project projectName
walmctl delete release releaseName -p projectName

`

type deleteCmd struct {
	sourceType  string
	releaseName string
	projectName	string
	deletePvcs  bool
	timeoutSec  int64
	async       bool
	out         io.Writer
}

func newDeleteCmd(out io.Writer) *cobra.Command {
	dc := &deleteCmd{out: out}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete a release, project or a release of project",
		Long:  deleteDesc,
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 2 {
				return errors.New("arguments error, delete release [releaseName] or delete project [projectName]")
			}
			dc.sourceType = args[0]
			if dc.sourceType == "release" {
				dc.releaseName = args[1]
			} else if dc.sourceType == "project" {
				dc.projectName = args[1]
			} else {
				return errors.New("delete [args]: first arg must be one of: release|project")
			}

			return dc.run()
		},
	}

	cmd.Flags().BoolVar(&dc.deletePvcs, "deletePvcs", true, "whether to delete pvcs related release")
	cmd.Flags().Int64Var(&dc.timeoutSec, "timeoutSec", 0, "timeout for task complete")
	cmd.Flags().BoolVar(&dc.async, "async", false, "asynchronous: true, synchronous: false")
	cmd.Flags().StringVarP(&dc.projectName, "project", "p", "", "指定一个 project 进行操作")
	return cmd
}

func (c *deleteCmd) run() error {

	//var resp *resty.Response
	var err error

	client := walmctlclient.CreateNewClient(walmserver)

	if c.sourceType == "project" {
		_, err = client.DeleteProject(namespace, c.projectName, c.async, c.timeoutSec, c.deletePvcs)
	} else {
		if c.projectName == "" {
			_, err = client.DeleteRelease(namespace, c.releaseName, c.async, c.timeoutSec, c.deletePvcs)
		} else {
			_, err = client.DeleteReleaseInProject(namespace, c.projectName, c.releaseName, c.async, c.timeoutSec, c.deletePvcs)
		}
	}

	if err != nil {
		return err
	}

	fmt.Sprintf("release %s in namespace %s deleted.", c.releaseName, namespace)
	return nil
}
