package main

import (
	"os"

	"github.com/spf13/cobra"

	"walm/pkg/setting"
	. "walm/pkg/util/log"
)

var (
	conf setting.Config
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
		addFlagsConfig(newServCmd()),
		newVersionCmd(),
	)

	flags.Parse(args)

	conf.Init()

	return cmd
}

func main() {
	cmd := newRootCmd(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		Log.Errorln(err)
		os.Exit(1)
	}
}

/*
func setupConnection() error {

	if config, client, err := getKubeClient(settings.KubeContext); err != nil {
		Log.Errorf("get kubenetes config failed: %s", err)
	}
	return nil
}

// configForContext creates a Kubernetes REST client configuration for a given kubeconfig context.
func configForContext(context string) (*rest.Config, error) {
	config, err := kube.GetConfig(context).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("could not get Kubernetes config for context %q: %s", context, err)
	}
	return config, nil
}

// getKubeClient creates a Kubernetes config and client for a given kubeconfig context.
func getKubeClient(context string) (*rest.Config, *kubernetes.Clientset, error) {
	config, err := configForContext(context)
	if err != nil {
		return nil, nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get Kubernetes client: %s", err)
	}
	return config, client, nil
}
*/

// addFlagsTLS adds the flags for supporting client side TLS to the
// helm command (only those that invoke communicate to Tiller.)
func addFlagsConfig(cmd *cobra.Command) *cobra.Command {

	cmd.Flags().StringVar(&setting.ConfigPath, "conf", "/etc/walm/conf", "path of the config file")
	return cmd
}
