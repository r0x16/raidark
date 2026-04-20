package cmd

import (
	"github.com/r0x16/Raidark/shared/api"
	domapi "github.com/r0x16/Raidark/shared/api/domain"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
	"github.com/spf13/cobra"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the HTTP API server.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		hub := ctx.Value(hubKey).(*domprovider.ProviderHub)
		modules := ctx.Value(modulesKey).([]domapi.ApiModule)

		api := api.NewApi(hub, modules)
		api.Run()
	},
}

func init() {
	RootCmd.AddCommand(apiCmd)
}
