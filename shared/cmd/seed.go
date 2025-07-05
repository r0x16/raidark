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

// seedCmd represents the init command
var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seeding initial data to database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		modules := cmd.Context().Value(modulesKey).([]apidomain.ApiModule)
		hub := cmd.Context().Value(hubKey).(*domprovider.ProviderHub)
		drivermigration.NewSeeder(hub, modules).Run()
	},
}

func init() {
	dbMigrationCmd.AddCommand(seedCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
