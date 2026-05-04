package cmd

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

var devicesCmd = &cobra.Command{
	Use:     "devices",
	Aliases: []string{"device"},
	Short:   "Manage devices",
}

var devicesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available devices",
	RunE: func(cmd *cobra.Command, args []string) error {
		return requestAndPrint(cmd, http.MethodGet, "/devices", nil, nil)
	},
}

var devicesGetCmd = &cobra.Command{
	Use:   "get DEVICE_ID",
	Short: "Get device details",
	Args:  requireExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return requestAndPrint(cmd, http.MethodGet, "/devices/"+url.PathEscape(args[0]), nil, nil)
	},
}

var devicesSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "List compact device summaries for agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := apiClient()
		if err != nil {
			return err
		}
		data, _, _, err := c.Do(cmd.Context(), http.MethodGet, "/devices", nil, nil)
		if err != nil {
			return err
		}
		var devices []deviceSummary
		if err := json.Unmarshal(data, &devices); err != nil {
			return err
		}
		for i := range devices {
			devices[i].BasisGates = cleanStrings(devices[i].BasisGates)
			devices[i].SupportedInstructions = cleanStrings(devices[i].SupportedInstructions)
			devices[i].Description = strings.TrimSpace(devices[i].Description)
		}
		return printValue(cmd, devices)
	},
}

type deviceSummary struct {
	DeviceID              string   `json:"device_id"`
	DeviceType            string   `json:"device_type"`
	Status                string   `json:"status"`
	NQubits               int      `json:"n_qubits,omitempty"`
	NPendingJobs          int      `json:"n_pending_jobs"`
	BasisGates            []string `json:"basis_gates,omitempty"`
	SupportedInstructions []string `json:"supported_instructions,omitempty"`
	AvailableAt           string   `json:"available_at,omitempty"`
	CalibratedAt          string   `json:"calibrated_at,omitempty"`
	Description           string   `json:"description,omitempty"`
}

func cleanStrings(values []string) []string {
	cleaned := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		cleaned = append(cleaned, value)
	}
	return cleaned
}

func init() {
	devicesCmd.AddCommand(devicesListCmd)
	devicesCmd.AddCommand(devicesGetCmd)
	devicesCmd.AddCommand(devicesSummaryCmd)
}
