package main

import (
	"flag"
	"os"

	walmctlEnv "WarpCloud/walm/cmd/walmctl/util/environment"
	walmSetting "WarpCloud/walm/pkg/setting"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog"
)

var globalUsage = `walmctl controls the walm application lifecycle manager.
To begin working with walmctl,Find detail docs at:
https://github.com/WarpCloud/walm/tree/master/docs/walmcli.md

Environment:
  $WALMSERVER		Set WALMSERVER env to substitute --server/-s in commands. The format is host:port (export $WALMSERVER=...)
  $ROOTCA	        Set ROOTCA env to substitute --rootCA in commands. which stores CA root certificate(export $WALMSERVER=...)
[WARNING] If the walm server use https, --tls=false required !!!
`

var (
	settings   walmctlEnv.EnvSettings
	walmserver string
	namespace  string
	rootCA	   string
	enableTLS  bool
)

func newRootCmd(args []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "walmctl",
		Short:        "walmctl controls the walm application lifecycle manager",
		Long:         globalUsage,
		SilenceUsage: true,
	}

	cmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	flags := cmd.PersistentFlags()

	flags.StringVarP(&walmserver, "server", "s", os.Getenv("WALMSERVER"), "walm apiserver address")
	flags.StringVarP(&namespace, "namespace", "n", "default", "kubernetes namespace")
	flags.BoolVar(&enableTLS, "tls", false, "enable send request use https")
	flags.StringVar(&rootCA, "rootCA", os.Getenv("ROOTCA"), "CA root certificate (public key)")

	settings.AddFlags(flags)

	out := cmd.OutOrStdout()

	cmd.AddCommand(
		newCreateCmd(out),
		newUpdateCmd(out),
		newGetCmd(out),
		newListCmd(out),
		newDeleteCmd(out),
		newDiffCmd(out),
		newMigrationCmd(out),
		newPackageCmd(out),
		newEditCmd(out),
		newSyncCmd(out),
		newLintCmd(),
		newComposeCmd(),
		newVersionCmd(),
	)

	flags.Parse(args)

	settings.Init(flags)
	walmSetting.InitDummyConfig()

	return cmd
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
	cmd := newRootCmd(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
