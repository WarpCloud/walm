package main

import (
	walmctlEnv "WarpCloud/walm/cmd/walmctl/util/environment"
	"github.com/spf13/cobra"
	"os"
)

var globalUsage = `walmctl controls the walm application lifecycle manager.
To begin working with walmctl,Find detail docs at:
http://172.16.1.41:10080/zhiyangdai/walmdocs/blob/master/docs/features/walmctl%E8%AF%B4%E6%98%8E%E6%96%87%E6%A1%A3.md
Environment:
  $WALM_HOST		Set WALM_HOST env to substitute --server/-s in commands. The format is host:port (export $WALM_HOST=...)

`

var (
	settings walmctlEnv.EnvSettings
	walmserver string
	namespace string
)

func newRootCmd(args []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "walmctl",
		Short: "walmctl controls the walm application lifecycle manager",
		Long: globalUsage,
		SilenceUsage: true,

	}

	cmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	flags := cmd.PersistentFlags()

	flags.StringVarP(&walmserver, "server", "s", os.Getenv("WALM_HOST"), "walm apiserver address")
	flags.StringVarP(&namespace, "namespace", "n", "", "kubernates namespace")

	settings.AddFlags(flags)

	out := cmd.OutOrStdout()

	cmd.AddCommand(

		newCreateCmd(out),
		newUpdateCmd(out),
		newGetCmd(out),
		newListCmd(out),
		newDeleteCmd(out),

		newPackageCmd(out),
		newEditCmd(out),
		newLintCmd(),
		NewVersionCmd(),
	)

	flags.Parse(args)

	settings.Init(flags)

	return cmd
}

func main() {
	cmd := newRootCmd(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
