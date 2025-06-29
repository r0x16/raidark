/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// seedCmd represents the init command
var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seeding initial data to database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Seed logic here
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
