package main

import (
	"os"

	"github.com/spf13/cobra"

	"walm/pkg/setting"
	. "walm/pkg/util/log"

	"k8s.io/client-go/kubernetes"
)

var (
	tlsCaCertFile string // path to TLS CA certificate file
	tlsCertFile   string // path to TLS certificate file
	tlsKeyFile    string // path to TLS key file
	tlsEnable     bool   // enable TLS

	tlsCaCertDefault = "$WALM_HOME/ca.pem"
	tlsCertDefault   = "$WALM_HOME/cert.pem"
	tlsKeyDefault    = "$WALM_HOME/key.pem"

	client   *kubernetes.Clientset
	settings setting.Config
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
		PersistentPreRunE: func(*cobra.Command, []string) error {
			tlsCaCertFile = os.ExpandEnv(tlsCaCertFile)
			tlsCertFile = os.ExpandEnv(tlsCertFile)
			tlsKeyFile = os.ExpandEnv(tlsKeyFile)

			return nil //setupConnection()
		},
	}
	flags := cmd.PersistentFlags()

	settings.AddFlags(flags)

	cmd.AddCommand(
		// chart commands
		addFlagsTLS(newServCmd()),
		addFlagsTLS(newVersionCmd()),
		newMigrateCmd(),
	)

	flags.Parse(args)
	// set defaults from environment
	settings.Init(flags)

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
func addFlagsTLS(cmd *cobra.Command) *cobra.Command {

	// add flags
	cmd.Flags().StringVar(&tlsCaCertFile, "tls-ca-cert", tlsCaCertDefault, "path to TLS CA certificate file")
	cmd.Flags().StringVar(&tlsCertFile, "tls-cert", tlsCertDefault, "path to TLS certificate file")
	cmd.Flags().StringVar(&tlsKeyFile, "tls-key", tlsKeyDefault, "path to TLS key file")
	//cmd.Flags().BoolVar(&tlsVerify, "tls-verify", false, "enable TLS for request and verify remote")
	cmd.Flags().BoolVar(&tlsEnable, "tls", false, "enable TLS for request")
	return cmd
}
