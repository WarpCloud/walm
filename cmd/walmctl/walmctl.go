package main

import (
	"os"
	walmctlEnv "WarpCloud/walm/cmd/walmctl/environment"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var globalUsage = `walmctl controls the walm application lifecycle manager.
To begin working with walmctl,Find detail docs at:
http://172.16.1.41:10080/zhiyangdai/WalmctlDocs
Environment:
  $WALM_HOST		set walm host to substitute --server/-s in commands. The format is host:port (export $WALM_HOST=...)

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

	viper.AutomaticEnv()
	flags := cmd.PersistentFlags()

	flags.String("walm_host", "", "walm apiserver env. Overrides $WALM_HOST")
	flags.MarkHidden("walm_host")
	flags.StringVarP(&walmserver, "server", "s", "", "walm apiserver address")
	flags.StringVarP(&namespace, "namespace", "n", "", "kubernates namespace")

	viper.BindPFlag("walm_host", flags.Lookup("walm_host"))


	settings.AddFlags(flags)
	out := cmd.OutOrStdout()

	cmd.AddCommand(

		newCreateCmd(out),
		newUpdateCmd(out),
		newGetCmd(out),
		newListCmd(out),
		newDeleteCmd(out),

		newPackageCmd(out),
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
