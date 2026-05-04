package cmd

import (
	"net/http"

	"github.com/spf13/cobra"
)

var apiTokenCmd = &cobra.Command{
	Use:     "api-token",
	Aliases: []string{"token"},
	Short:   "Manage API tokens",
}

var apiTokenCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an API token",
	RunE: func(cmd *cobra.Command, args []string) error {
		return requestAndPrint(cmd, http.MethodPost, "/api-token", nil, nil)
	},
}

var apiTokenStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get API token status",
	RunE: func(cmd *cobra.Command, args []string) error {
		return requestAndPrint(cmd, http.MethodGet, "/api-token/status", nil, nil)
	},
}

var apiTokenDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an API token",
	RunE: func(cmd *cobra.Command, args []string) error {
		return requestAndPrint(cmd, http.MethodDelete, "/api-token", nil, nil)
	},
}

func init() {
	apiTokenCmd.AddCommand(apiTokenCreateCmd)
	apiTokenCmd.AddCommand(apiTokenStatusCmd)
	apiTokenCmd.AddCommand(apiTokenDeleteCmd)
}
