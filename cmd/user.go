package cmd

import (
	"net/http"

	"github.com/spf13/cobra"
)

var userUpdateFile string
var userName string
var userOrganization string

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage the current user",
}

var userGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get current user",
	RunE: func(cmd *cobra.Command, args []string) error {
		return requestAndPrint(cmd, http.MethodGet, "/users/me", nil, nil)
	},
}

var userUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update current user",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := updateUserBody(userUpdateFile, userName, userOrganization)
		if err != nil {
			return err
		}
		return requestAndPrint(cmd, http.MethodPatch, "/users/me", nil, body)
	},
}

var userDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete current user",
	RunE: func(cmd *cobra.Command, args []string) error {
		return requestAndPrint(cmd, http.MethodDelete, "/users/me", nil, nil)
	},
}

func init() {
	userUpdateCmd.Flags().StringVarP(&userUpdateFile, "file", "f", "", "JSON request body file, or - for stdin")
	userUpdateCmd.Flags().StringVar(&userName, "name", "", "User name")
	userUpdateCmd.Flags().StringVar(&userOrganization, "organization", "", "Organization")

	userCmd.AddCommand(userGetCmd)
	userCmd.AddCommand(userUpdateCmd)
	userCmd.AddCommand(userDeleteCmd)
}
