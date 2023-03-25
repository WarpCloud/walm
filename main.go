package main

import (
	"flag"
	"os"

	"WarpCloud/walm/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog"
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

func initKubeLogs() {
	gofs := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(gofs)
	pflag.CommandLine.AddGoFlagSet(gofs)
	pflag.CommandLine.Set("logtostderr", "true")
	pflag.CommandLine.Set("v", "1")
}

func main() {
	initKubeLogs()

	klog.Infof("Walm Start...")
	rootCmd := newRootCmd(os.Args[1:])
	if err := rootCmd.Execute(); err != nil {
		klog.Errorln(err)
		os.Exit(1)
	}
}
