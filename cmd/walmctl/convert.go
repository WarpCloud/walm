package main

import (
	"github.com/spf13/cobra"
)

const convertDesc = `
Convert a Docker Compose file
`

func newConvertCmd() *cobra.Command {
	lint := &lintOptions{chartPath: "."}

	cmd := &cobra.Command{
		Use:   "convert [file]",
		Short: "Convert a Docker Compose file",
		Long:  convertDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return lint.run()
		},
	}
	cmd.PersistentFlags().StringVar(&lint.chartPath, "chartPath", ".", "test transwarp chart path")

	return cmd
}
