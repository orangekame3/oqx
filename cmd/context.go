package cmd

import (
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Print agent-oriented usage context",
	RunE: func(cmd *cobra.Command, args []string) error {
		resolved, err := resolveOptions()
		if err != nil {
			return err
		}
		auth := "none"
		if resolved.apiToken != "" {
			auth = "q-api-token"
		} else if resolved.token != "" {
			auth = "bearer"
		}
		ctx := map[string]any{
			"purpose": "Control OQTOPUS Cloud User API from coding agents and scripts.",
			"configuration": map[string]any{
				"effective_base_url": resolved.baseURL,
				"effective_auth":     auth,
				"preferred_env": map[string]string{
					"base_url":  "OQTOPUS_URL or OQX_BASE_URL",
					"api_token": "OQTOPUS_API_TOKEN or OQX_API_TOKEN",
				},
				"sdk_config": "~/.oqtopus [default] url=... api_token=...",
			},
			"safe_default_device": "qulacs",
			"job_statuses": map[string]any{
				"non_final": []string{"submitted", "ready", "running"},
				"final":     []string{"succeeded", "failed", "cancelled"},
			},
			"recommended_workflow": []string{
				"oqx --output json context",
				"oqx --output json devices summary",
				"oqx examples submit-job --device qulacs --shots 1000 > job.json",
				"oqx --output json jobs submit --file job.json",
				"oqx --output json jobs wait JOB_ID --timeout 10m",
				"oqx --output json jobs result JOB_ID",
			},
			"commands": map[string]string{
				"devices_summary":      "Compact device list without large calibration payloads.",
				"jobs_submit":          "Submit an explicit JSON body.",
				"jobs_submit_sampling": "Build and submit a simple sampling job from an OPENQASM 3 file.",
				"jobs_wait":            "Poll until final status.",
				"jobs_result":          "Return only job_info.result with job_id/status/message.",
				"raw":                  "Escape hatch for arbitrary User API endpoints.",
			},
			"notes": []string{
				"Prefer --output json for machine consumption.",
				"Prefer simulator device qulacs unless the user explicitly asks for QPU hardware.",
				"Do not call delete or submit jobs with large shots unless the user clearly requests it.",
			},
		}
		return printValue(cmd, ctx)
	},
}
