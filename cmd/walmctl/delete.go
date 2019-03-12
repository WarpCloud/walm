package main

import (
	"io"
	"github.com/spf13/cobra"
	"walm/cmd/walmctl/walmctlclient"
	"fmt"
	"github.com/pkg/errors"
)

const deleteDesc = `This command delete a walm release, project or a release of project.

support cmds format:
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
		Use:   "delete release/project releaseName/projectName",
		Short: "delete a release, project or a release of project",
		Long:  deleteDesc,
		RunE: func(cmd *cobra.Command, args []string) error {

			if namespace == "" {
				return errors.New("flag --namespace/-n required")
			}
			if walmserver == "" {
				return errors.New("flag --server/-s required")

			}
			if len(args) != 2 {
				return errors.New("arguments error, delete release [releaseName] or delete project [projectName]")
			}
			dc.sourceType = args[0]
			if dc.sourceType == "release" {
				dc.releaseName = args[1]
			} else if dc.sourceType == "project" {
				dc.projectName = args[1]
			} else {
				return errors.New("delete [args]: first arg must be release or project")
			}

			return dc.run()
		},
	}

	cmd.Flags().BoolVar(&dc.deletePvcs, "deletePvcs", true, "whether to delete pvcs related release")
	cmd.Flags().Int64Var(&dc.timeoutSec, "timeoutSec", 0, "timeout (default 0)")
	cmd.Flags().BoolVar(&dc.async, "async", true, "whether asynchronous")
	cmd.Flags().StringVarP(&dc.projectName, "project", "p", "", "operate resources of the project")
	return cmd
}

func (dc *deleteCmd) run() error {

	// Todo: [Bug] delete release which not exists, also return OK
	var err error

	client := walmctlclient.CreateNewClient(walmserver)

	if dc.sourceType == "project" {
		_, err = client.DeleteProject(namespace, dc.projectName, dc.async, dc.timeoutSec, dc.deletePvcs)

		if err != nil {
			return err
		}

		fmt.Printf("project %s deleted", dc.projectName)
	} else {
		if dc.projectName == "" {
			_, err = client.DeleteRelease(namespace, dc.releaseName, dc.async, dc.timeoutSec, dc.deletePvcs)
			if err != nil {
				return err
			}

			fmt.Printf("release %s deleted", dc.releaseName)

		} else {
			_, err = client.DeleteReleaseInProject(namespace, dc.projectName, dc.releaseName, dc.async, dc.timeoutSec, dc.deletePvcs)
			if err != nil {
				return err
			}
			fmt.Printf("release %s in project %s deleted", dc.releaseName, dc.projectName)
		}
	}

	return nil
}
