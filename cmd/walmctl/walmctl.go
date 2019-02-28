package main

import (
	"os"

	walmctlEnv "walm/cmd/walmctl/environment"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"fmt"
)

var globalUsage = `walmctl controls the walm application lifecycle manager.
To begin working with walmctl,Find detail docs at:
http://172.16.1.168:8090/display/TOS/Walmctl+MileStone/

WARNING: 
Before use walmctl commands, you must refer walm host with --server/-s and namespace with --namespace/-n

Environment:
  $WALM_HOST		set a required walm host to substitute --server/-s in commands. The format is host:port (export $WALM_HOST=...)

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
		PersistentPreRun: func(*cobra.Command, []string) {
			if walmserver == "" {
				if viper.GetString("walm_host") == "" {
					fmt.Printf("Error: flag --server/-s required\n")
					os.Exit(1)
				} else {
					walmserver = viper.GetString("walm_host")
				}
			}

			if namespace == "" {
				fmt.Printf("Error: flag --namespace/-n required\n")
				os.Exit(1)
			}
		},

	}

	cmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	viper.AutomaticEnv()
	flags := cmd.PersistentFlags()

	flags.String("walm_host", "", "walm apiserver env. Overrides $WALM_HOST")
	flags.MarkHidden("walm_host")
	flags.StringVarP(&walmserver, "server", "s", "", "walm apiserver address (Required only once)")
	flags.StringVarP(&namespace, "namespace", "n", "", "kubernates namespace (Required only once)")

	viper.BindPFlag("walm_host", flags.Lookup("walm_host"))


	settings.AddFlags(flags)
	out := cmd.OutOrStdout()

	cmd.AddCommand(

		newCreateCmd(out),
		newUpdateCmd(out),
		newGetCmd(out),
		newListCmd(out),
		newDeleteCmd(out),

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
