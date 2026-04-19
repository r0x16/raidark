package cmd

import (
	apidomain "github.com/r0x16/Raidark/shared/api/domain"
	drivermigration "github.com/r0x16/Raidark/shared/migration/driver"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
	"github.com/spf13/cobra"
)

var dbMigrationCmd = &cobra.Command{
	Use:   "dbmigrate",
	Short: "Run database schema migrations.",
	Run: func(cmd *cobra.Command, args []string) {
		modules := cmd.Context().Value(modulesKey).([]apidomain.ApiModule)
		hub := cmd.Context().Value(hubKey).(*domprovider.ProviderHub)
		drivermigration.NewDbmigrate(hub, modules).Run()
	},
}

func init() {
	RootCmd.AddCommand(dbMigrationCmd)
}
