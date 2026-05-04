package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

type announcementsListOptions struct {
	offset      string
	limit       string
	order       string
	currentTime string
}

var announcementsListOpts announcementsListOptions

var announcementsCmd = &cobra.Command{
	Use:     "announcements",
	Aliases: []string{"announcement"},
	Short:   "Read announcements",
}

var announcementsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List announcements",
	RunE: func(cmd *cobra.Command, args []string) error {
		return requestAndPrint(cmd, http.MethodGet, "/announcements", map[string]string{
			"offset":       announcementsListOpts.offset,
			"limit":        announcementsListOpts.limit,
			"order":        announcementsListOpts.order,
			"current_time": announcementsListOpts.currentTime,
		}, nil)
	},
}

var announcementsGetCmd = &cobra.Command{
	Use:   "get ANNOUNCEMENT_ID",
	Short: "Get selected announcement",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := requireExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if _, err := strconv.Atoi(args[0]); err != nil {
			return fmt.Errorf("announcement id must be an integer: %w", err)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return requestAndPrint(cmd, http.MethodGet, "/announcements/"+url.PathEscape(args[0]), nil, nil)
	},
}

func init() {
	announcementsListCmd.Flags().StringVar(&announcementsListOpts.offset, "offset", "", "Offset")
	announcementsListCmd.Flags().StringVar(&announcementsListOpts.limit, "limit", "", "Limit")
	announcementsListCmd.Flags().StringVar(&announcementsListOpts.order, "order", "", "ASC or DESC")
	announcementsListCmd.Flags().StringVar(&announcementsListOpts.currentTime, "current-time", "", "Filter active announcements at this RFC3339 time")

	announcementsCmd.AddCommand(announcementsListCmd)
	announcementsCmd.AddCommand(announcementsGetCmd)
}
