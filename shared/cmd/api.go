/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"

	"github.com/r0x16/Raidark/shared/api"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
	"github.com/r0x16/Raidark/shared/providers/driver"
	"github.com/spf13/cobra"
)

// apiCmd represents the api command
var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Starts API Server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		// Configure providers in context
		providers := []domprovider.ProviderFactory{
			&driver.EnvProviderFactory{},
			&driver.LoggerProviderFactory{},
			&driver.DatastoreProviderFactory{},
			&driver.AuthProviderFactory{},
			&driver.ApiProviderFactory{},
		}

		ctxWithProviders := context.WithValue(ctx, "providers", providers)

		api := api.NewApi(ctxWithProviders)
		api.Run()
	},
}

func init() {
	RootCmd.AddCommand(apiCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// apiCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// apiCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
