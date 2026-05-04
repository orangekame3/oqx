package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is updated by tagpr and GoReleaser.
var Version = "0.0.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "oqx version "+Version)
	},
}
