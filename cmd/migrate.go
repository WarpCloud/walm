package main

import (
	"walm/models"

	"github.com/spf13/cobra"
)

const migrateDesc = `
This command enalbe auto migrate walm DB.`

type migrateCmd struct {
}

func newMigrateCmd() *cobra.Command {
	mc := &migrateCmd{}

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "auto migrate db of walm",
		Long:  migrateDesc,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return models.Init(&settings)
		},

		Run: func(cmd *cobra.Command, args []string) {
			defer mc.run()
		},

		PostRun: func(_ *cobra.Command, _ []string) {
			defer models.CloseDB()
		},
	}
	return cmd
}

func (mc *migrateCmd) run() {
	models.AutoMigrate()
}
