package main

import (
	"walm/models"

	. "walm/pkg/util/log"

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
	if err := models.AutoMigrate(&conf); err != nil {
		Log.Fatal(err)
	}
}
