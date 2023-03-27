package main

import (
	"WarpCloud/walm/cmd/walmctl/util/editor"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	"k8s.io/kubernetes/pkg/kubectl/util/templates"
)

var (
	editLong = templates.LongDesc(i18n.T(`
		Edit a resource from the default editor.

		The edit command allows you to directly edit any API resource you can retrieve via the
		command line tools. It will open the editor defined by your KUBE_EDITOR, or EDITOR
		environment variables, or fall back to 'vi' for Linux or 'notepad' for Windows.
		You can edit multiple objects, although changes are applied one at a time. The command
		accepts filenames as well as command line arguments, although the files you point to must
		be previously saved versions of resources.

		Editing is done with the API version used to fetch the resource.
		To edit using a specific API version, fully-qualify the resource, version, and group.

		The default format is YAML. To edit in JSON, specify "-o json".

		The flag --windows-line-endings can be used to force Windows line endings,
		otherwise the default for your operating system will be used.

		In the event an error occurs while updating, a temporary file will be created on disk
		that contains your unapplied changes. The most common error when updating a resource
		is another editor changing the resource on the server. When this occurs, you will have
		to apply your changes to the newer version of the resource, or update your temporary
		saved copy to include the latest resource version.`))

	editExample = templates.Examples(i18n.T(`
		# Edit the service named 'docker-registry':
		walmctl edit release/docker-registry

		# Use an alternative editor
		KUBE_EDITOR="nano" kubectl edit svc/docker-registry

		# Edit the job 'myjob' in JSON using the v1 API format:
		kubectl edit job.v1.batch/myjob -o json

		# Edit the deployment 'mydeployment' in YAML and save the modified config in its annotation:
		kubectl edit deployment/mydeployment -o yaml --save-config`))
)

func newEditCmd(out io.Writer) *cobra.Command {
	ioStreams := genericclioptions.IOStreams{Out: out}
	o := editor.NewEditOptions(editor.NormalEditMode, ioStreams)

	cmd := &cobra.Command{
		Use:                   "edit RESOURCE NAME",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Edit a resource on the server"),
		Long:                  editLong,
		Example:               fmt.Sprintf(editExample),
		RunE: func(cmd *cobra.Command, args []string) error {
			if walmserver == "" {
				return errServerRequired
			}
			if namespace == "" {
				return errNamespaceRequired
			}
			if len(args) != 2 {
				return errors.New("arguments error, get release/project releaseName/projectName")
			}
			if checkResourceType(args[0]) != nil {
				return checkResourceType(args[0])
			}
			o.Namespace = namespace
			o.WalmServer = walmserver
			o.EnableTLS = enableTLS
			o.RootCA = rootCA
			o.SourceType = args[0]
			o.SourceName = args[1]
			return o.Run()
		},
	}

	// bind flag structs
	cmd.PersistentFlags().StringVarP(o.PrintFlags.OutputFormat, "output", "o", "yaml", "-o, --output='': Output format for detail description. Support: json, yaml")

	return cmd
}
