package main

import (
	"WarpCloud/walm/cmd/walmctl/util/walmctlclient"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io"
	"k8s.io/klog"
)

const deleteDesc = `Delete walm resources by source name.
delete release, project or release in project.
support format:
walmctl delete release releaseName
walmctl delete project projectName
walmctl delete release releaseName -p projectName

Note that the delete command does NOT do resource version checks, so if someone submits an update to
a resource right when you submit a delete, their update will be lost along with the rest of the
resource.
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
		DisableFlagsInUseLine: true,
		Short: "delete a release, project or a release of project",
		Long:  deleteDesc,
		RunE: func(cmd *cobra.Command, args []string) error {

			if walmserver == "" {
				return errServerRequired
			}
			if args[0] != "namespace" && namespace == "" {
				return errNamespaceRequired
			}

			if len(args) != 2 {
				return errors.New("arguments error, two args required")
			}

			err := checkResourceType(args[0])
			if err != nil {
				return err
			}
			dc.sourceType = args[0]

			if dc.sourceType == "release" {
				dc.releaseName = args[1]
			} else if dc.sourceType == "project" {
				dc.projectName = args[1]
			} else {
				namespace = args[1]
			}
			return dc.run()
		},
	}

	cmd.PersistentFlags().BoolVar(&dc.deletePvcs, "deletePvcs", true, "whether to delete pvcs related release")
	cmd.PersistentFlags().Int64Var(&dc.timeoutSec, "timeoutSec", 0, "timeout (default 0)")
	cmd.PersistentFlags().BoolVar(&dc.async, "async", true, "whether asynchronous")
	cmd.PersistentFlags().StringVarP(&dc.projectName, "project", "p", "", "operate resource of project")

	return cmd
}

func (dc *deleteCmd) run() error {

	var err error
	client, err := walmctlclient.CreateNewClient(walmserver, enableTLS, rootCA)
	if err != nil {
		klog.Errorf("failed to create walmctl client: %s", err.Error())
		return err
	}
	if err = client.ValidateHostConnect(walmserver); err != nil {
		return err
	}
	if dc.sourceType == "namespace" {
		_, err = client.DeleteTenant(namespace)
		if err != nil {
			klog.Errorf("failed to delete namespace %s : %s", namespace, err.Error())
			return err
		}
		fmt.Printf("namespace %s deleted\n", namespace)
	} else if dc.sourceType == "project" {
		_, err = client.DeleteProject(namespace, dc.projectName, dc.async, dc.timeoutSec, dc.deletePvcs)
		if err != nil {
			return err
		}
		fmt.Printf("project %s deleted\n", dc.projectName)

	} else {
		if dc.projectName == "" {
			_, err = client.DeleteRelease(namespace, dc.releaseName, dc.async, dc.timeoutSec, dc.deletePvcs)
			if err != nil {
				return err
			}

			fmt.Printf("release %s deleted\n", dc.releaseName)

		} else {
			_, err = client.DeleteReleaseInProject(namespace, dc.projectName, dc.releaseName, dc.async, dc.timeoutSec, dc.deletePvcs)
			if err != nil {
				return err
			}
			fmt.Printf("release %s in project %s deleted\n", dc.releaseName, dc.projectName)
		}
	}

	return nil
}
