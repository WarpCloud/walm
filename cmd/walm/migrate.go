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

		RunE: func(cmd *cobra.Command, args []string) error {
			return mc.run()
		},

		PostRun: func(_ *cobra.Command, _ []string) {
			defer models.CloseDB()
		},
	}
	return cmd
}

func (mc *migrateCmd) run() error {
	return models.AutoMigrate(&conf)
}
