package cmd

import (
	"net/http"

	"github.com/spf13/cobra"
)

var settingsCmd = &cobra.Command{
	Use:     "settings",
	Aliases: []string{"setting"},
	Short:   "Read system settings",
}

var settingsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get current settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		return requestAndPrint(cmd, http.MethodGet, "/system/settings", nil, nil)
	},
}

func init() {
	settingsCmd.AddCommand(settingsGetCmd)
}
