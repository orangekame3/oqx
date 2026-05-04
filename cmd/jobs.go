package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
)

type jobsListOptions struct {
	fields    string
	startTime string
	endTime   string
	status    string
	q         string
	page      string
	size      string
	order     string
}

var jobsListOpts jobsListOptions
var submitFile string
var submitSamplingDevice string
var submitSamplingProgram string
var submitSamplingName string
var submitSamplingDescription string
var submitSamplingShots int
var sselogOutput string
var waitInterval time.Duration
var waitTimeout time.Duration

var jobsCmd = &cobra.Command{
	Use:     "jobs",
	Aliases: []string{"job"},
	Short:   "Manage quantum jobs",
}

var jobsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List quantum jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		return requestAndPrint(cmd, http.MethodGet, "/jobs", map[string]string{
			"fields":     jobsListOpts.fields,
			"start_time": jobsListOpts.startTime,
			"end_time":   jobsListOpts.endTime,
			"status":     jobsListOpts.status,
			"q":          jobsListOpts.q,
			"page":       jobsListOpts.page,
			"size":       jobsListOpts.size,
			"order":      jobsListOpts.order,
		}, nil)
	},
}

var jobsSubmitCmd = &cobra.Command{
	Use:   "submit --file BODY.json",
	Short: "Submit a quantum job",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readBody(submitFile)
		if err != nil {
			return err
		}
		return requestAndPrint(cmd, http.MethodPost, "/jobs", nil, body)
	},
}

var jobsSubmitSamplingCmd = &cobra.Command{
	Use:   "submit-sampling --device DEVICE_ID --program PROGRAM.qasm",
	Short: "Submit a sampling job from an OPENQASM 3 program",
	RunE: func(cmd *cobra.Command, args []string) error {
		program, err := os.ReadFile(submitSamplingProgram)
		if err != nil {
			return err
		}
		body, err := samplingJobBody(submitSamplingDevice, string(program), submitSamplingName, submitSamplingDescription, submitSamplingShots)
		if err != nil {
			return err
		}
		return requestAndPrint(cmd, http.MethodPost, "/jobs", nil, body)
	},
}

var jobsGetCmd = jobByIDCommand("get", "Get selected job", http.MethodGet, "")
var jobsStatusCmd = jobByIDCommand("status", "Get selected job status", http.MethodGet, "/status")
var jobsCancelCmd = jobByIDCommand("cancel", "Cancel selected job", http.MethodPost, "/cancel")
var jobsDeleteCmd = jobByIDCommand("delete", "Delete selected job", http.MethodDelete, "")

var jobsSSELogCmd = &cobra.Command{
	Use:   "sselog JOB_ID",
	Short: "Get selected job SSE log",
	Args:  requireExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := apiClient()
		if err != nil {
			return err
		}
		data, _, _, err := c.Do(cmd.Context(), http.MethodGet, "/jobs/"+url.PathEscape(args[0])+"/sselog", nil, nil)
		if err != nil {
			return err
		}
		if sselogOutput != "" {
			return writeSSELog(data, sselogOutput)
		}
		return printData(cmd, data)
	},
}

var jobsWaitCmd = &cobra.Command{
	Use:   "wait JOB_ID",
	Short: "Wait until a job reaches a final status",
	Args:  requireExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := apiClient()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		if waitTimeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, waitTimeout)
			defer cancel()
		}

		path := "/jobs/" + url.PathEscape(args[0]) + "/status"
		ticker := time.NewTicker(waitInterval)
		defer ticker.Stop()

		for {
			data, _, _, err := c.Do(ctx, http.MethodGet, path, nil, nil)
			if err != nil {
				return err
			}
			status, err := jobStatus(data)
			if err != nil {
				return err
			}
			if isFinalJobStatus(status) {
				return printData(cmd, data)
			}

			select {
			case <-ctx.Done():
				return fmt.Errorf("wait for job %s timed out: %w", args[0], ctx.Err())
			case <-ticker.C:
			}
		}
	},
}

var jobsResultCmd = &cobra.Command{
	Use:   "result JOB_ID",
	Short: "Print only the result portion of a job",
	Args:  requireExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := apiClient()
		if err != nil {
			return err
		}
		data, _, _, err := c.Do(cmd.Context(), http.MethodGet, "/jobs/"+url.PathEscape(args[0]), nil, nil)
		if err != nil {
			return err
		}
		result, err := jobResult(data)
		if err != nil {
			return err
		}
		return printValue(cmd, result)
	},
}

func jobByIDCommand(use, short, method, suffix string) *cobra.Command {
	return &cobra.Command{
		Use:   use + " JOB_ID",
		Short: short,
		Args:  requireExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return requestAndPrint(cmd, method, "/jobs/"+url.PathEscape(args[0])+suffix, nil, nil)
		},
	}
}

func init() {
	jobsListCmd.Flags().StringVar(&jobsListOpts.fields, "fields", "", "Comma-separated response fields")
	jobsListCmd.Flags().StringVar(&jobsListOpts.startTime, "start-time", "", "Filter jobs submitted at or after this RFC3339 time")
	jobsListCmd.Flags().StringVar(&jobsListOpts.endTime, "end-time", "", "Filter jobs submitted at or before this RFC3339 time")
	jobsListCmd.Flags().StringVar(&jobsListOpts.status, "status", "", "Filter by status")
	jobsListCmd.Flags().StringVar(&jobsListOpts.q, "q", "", "Search text")
	jobsListCmd.Flags().StringVar(&jobsListOpts.page, "page", "", "Page number")
	jobsListCmd.Flags().StringVar(&jobsListOpts.size, "size", "", "Page size")
	jobsListCmd.Flags().StringVar(&jobsListOpts.order, "order", "", "ASC or DESC")

	jobsSubmitCmd.Flags().StringVarP(&submitFile, "file", "f", "", "JSON request body file, or - for stdin")
	_ = jobsSubmitCmd.MarkFlagRequired("file")
	jobsSubmitSamplingCmd.Flags().StringVar(&submitSamplingDevice, "device", "qulacs", "Device ID")
	jobsSubmitSamplingCmd.Flags().StringVar(&submitSamplingProgram, "program", "", "OPENQASM 3 program file")
	jobsSubmitSamplingCmd.Flags().StringVar(&submitSamplingName, "name", "Bell State Sampling", "Job name")
	jobsSubmitSamplingCmd.Flags().StringVar(&submitSamplingDescription, "description", "", "Job description")
	jobsSubmitSamplingCmd.Flags().IntVar(&submitSamplingShots, "shots", 1000, "Number of shots")
	_ = jobsSubmitSamplingCmd.MarkFlagRequired("program")

	jobsSSELogCmd.Flags().StringVarP(&sselogOutput, "output", "o", "", "Output file for decoded SSE log")
	jobsWaitCmd.Flags().DurationVar(&waitInterval, "interval", 5*time.Second, "Polling interval")
	jobsWaitCmd.Flags().DurationVar(&waitTimeout, "timeout", 10*time.Minute, "Maximum wait duration")

	jobsCmd.AddCommand(jobsListCmd)
	jobsCmd.AddCommand(jobsSubmitCmd)
	jobsCmd.AddCommand(jobsSubmitSamplingCmd)
	jobsCmd.AddCommand(jobsGetCmd)
	jobsCmd.AddCommand(jobsStatusCmd)
	jobsCmd.AddCommand(jobsResultCmd)
	jobsCmd.AddCommand(jobsCancelCmd)
	jobsCmd.AddCommand(jobsDeleteCmd)
	jobsCmd.AddCommand(jobsSSELogCmd)
	jobsCmd.AddCommand(jobsWaitCmd)
}

func samplingJobBody(deviceID, program, name, description string, shots int) ([]byte, error) {
	if deviceID == "" {
		return nil, errors.New("device is required")
	}
	if program == "" {
		return nil, errors.New("program is required")
	}
	if shots < 1 {
		return nil, errors.New("shots must be >= 1")
	}
	body := map[string]any{
		"name":        name,
		"description": description,
		"device_id":   deviceID,
		"job_type":    "sampling",
		"job_info": map[string]any{
			"program": []string{program},
		},
		"transpiler_info": map[string]any{},
		"simulator_info":  map[string]any{},
		"mitigation_info": map[string]any{},
		"shots":           shots,
	}
	return json.Marshal(body)
}

type resultView struct {
	JobID   string `json:"job_id,omitempty"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Result  any    `json:"result,omitempty"`
}

func jobResult(data []byte) (resultView, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return resultView{}, err
	}
	view := resultView{}
	if v, ok := raw["job_id"].(string); ok {
		view.JobID = v
	}
	if v, ok := raw["status"].(string); ok {
		view.Status = v
	}
	jobInfo, _ := raw["job_info"].(map[string]any)
	if jobInfo != nil {
		if v, ok := jobInfo["message"].(string); ok {
			view.Message = v
		}
		view.Result = jobInfo["result"]
	}
	return view, nil
}

func jobStatus(data []byte) (string, error) {
	var resp struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	if resp.Status == "" {
		return "", errors.New("job status response does not contain status")
	}
	return resp.Status, nil
}

func isFinalJobStatus(status string) bool {
	switch status {
	case "succeeded", "failed", "cancelled":
		return true
	default:
		return false
	}
}
