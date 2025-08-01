/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	apidomain "github.com/r0x16/Raidark/shared/api/domain"
	drivermigration "github.com/r0x16/Raidark/shared/migration/driver"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
	"github.com/spf13/cobra"
)

// dbMigrationCmd represents the dbMigration command
var dbMigrationCmd = &cobra.Command{
	Use:   "dbmigrate",
	Short: "Migrate database schema",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		modules := cmd.Context().Value(modulesKey).([]apidomain.ApiModule)
		hub := cmd.Context().Value(hubKey).(*domprovider.ProviderHub)
		drivermigration.NewDbmigrate(hub, modules).Run()
	},
}

func init() {
	RootCmd.AddCommand(dbMigrationCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dbMigrationCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dbMigrationCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
