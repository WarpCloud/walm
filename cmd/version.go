package main

import (
	"walm/pkg/version"

	"github.com/spf13/cobra"
)

const versionDesc = `
This command print version of walm.`

type vCmd struct {
}

func newVersionCmd() *cobra.Command {
	vc := &vCmd{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "print version",
		Long:  versionDesc,

		Run: func(cmd *cobra.Command, args []string) {
			defer vc.run()
		},
	}
	return cmd
}

func (vc *vCmd) run() {
	version.PrintVersionInfo()
}
