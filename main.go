package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/sirupsen/logrus"

	"walm/cmd"
)

var globalUsage = `The Warp application lifecycle manager

To begin working with walm, run the 'walm serv' command:

    $ walm serv

Environment:
  $KUBECONFIG         set an alternative Kubernetes configuration file (default "~/.kube/config")
`

func newRootCmd(args []string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "walm",
		Short:        "The Warp application lifecycle manager.",
		Long:         globalUsage,
		SilenceUsage: true,
	}
	flags := rootCmd.PersistentFlags()

	rootCmd.AddCommand(
		cmd.NewServCmd(),
		cmd.NewVersionCmd(),
	)

	flags.Parse(args)

	return rootCmd
}

func main() {
	logrus.Infof("Walm Start...")
	rootCmd := newRootCmd(os.Args[1:])
	if err := rootCmd.Execute(); err != nil {
		logrus.Errorln(err)
		os.Exit(1)
	}
}