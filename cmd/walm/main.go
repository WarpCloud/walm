package main

import (
	"os"

	"github.com/spf13/cobra"

	. "walm/pkg/util/log"
)

var globalUsage = `The Warp application lifecycle manager

To begin working with walm, run the 'walm serv' command:

	$ walm serv

Environment:
  $KUBECONFIG         set an alternative Kubernetes configuration file (default "~/.kube/config")
`

func newRootCmd(args []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "walm",
		Short:        "The Warp application lifecycle manager.",
		Long:         globalUsage,
		SilenceUsage: true,
	}
	flags := cmd.PersistentFlags()

	cmd.AddCommand(
		newServCmd(),
		newVersionCmd(),
	)

	flags.Parse(args)

	return cmd
}

func main() {
	cmd := newRootCmd(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		Log.Errorln(err)
		os.Exit(1)
	}
}
