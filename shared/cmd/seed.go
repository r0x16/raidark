package cmd

import (
	apidomain "github.com/r0x16/Raidark/shared/api/domain"
	drivermigration "github.com/r0x16/Raidark/shared/migration/driver"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Load seed data into the database.",
	Run: func(cmd *cobra.Command, args []string) {
		modules := cmd.Context().Value(modulesKey).([]apidomain.ApiModule)
		hub := cmd.Context().Value(hubKey).(*domprovider.ProviderHub)
		drivermigration.NewSeeder(hub, modules).Run()
	},
}

func init() {
	dbMigrationCmd.AddCommand(seedCmd)
}
