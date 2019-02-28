package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var globalUsage = ``

func newWalmClientCmd(args []string) *cobra.Command {
	walmClientCmd := &cobra.Command{
		Use:          "walmcli",
		Short:        "Walm Client For Walm Server.",
		Long:         globalUsage,
		SilenceUsage: true,
	}
	flags := walmClientCmd.PersistentFlags()

	walmClientCmd.AddCommand(
		newLintCmd(),
	)

	flags.Parse(args)

	return walmClientCmd
}

func main() {
	walmClientCmd := newWalmClientCmd(os.Args[1:])
	if err := walmClientCmd.Execute(); err != nil {
		logrus.Errorln(err)
		os.Exit(1)
	}
}
