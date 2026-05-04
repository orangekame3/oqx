package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/orangekame3/oqx/internal/cli"
	"github.com/spf13/cobra"
)

const (
	defaultBaseURL = "http://localhost:8080"
	envBaseURL     = "OQX_BASE_URL"
	envToken       = "OQX_TOKEN"
	envAPIToken    = "OQX_API_TOKEN"
	envOqtopusURL  = "OQTOPUS_URL"
	envOqtopusAPI  = "OQTOPUS_API_TOKEN"
)

type globalOptions struct {
	baseURL  string
	token    string
	apiToken string
	timeout  time.Duration
	output   string
	quiet    bool
}

var opts = globalOptions{
	baseURL: getenv(envBaseURL, defaultBaseURL),
	token:   os.Getenv(envToken),
	timeout: 30 * time.Second,
}

var rootCmd = &cobra.Command{
	Use:   "oqx",
	Short: "CLI for the OQTOPUS Cloud User API",
	Long: `oqx is a CLI for the OQTOPUS Cloud User API.

Use oqx auth login, OQX_BASE_URL/OQX_TOKEN, or global flags to configure requests.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func SetOutput(stdout, stderr io.Writer) {
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
}

func init() {
	rootCmd.PersistentFlags().StringVar(&opts.baseURL, "base-url", opts.baseURL, "User API base URL")
	rootCmd.PersistentFlags().StringVar(&opts.token, "token", opts.token, "Bearer token")
	rootCmd.PersistentFlags().StringVar(&opts.apiToken, "api-token", opts.apiToken, "OQTOPUS API token for q-api-token header")
	rootCmd.PersistentFlags().DurationVar(&opts.timeout, "timeout", opts.timeout, "HTTP timeout")
	rootCmd.PersistentFlags().StringVar(&opts.output, "output", "pretty", "Output format: pretty, json, raw")
	rootCmd.PersistentFlags().BoolVarP(&opts.quiet, "quiet", "q", false, "Suppress non-data messages")

	rootCmd.AddCommand(devicesCmd)
	rootCmd.AddCommand(jobsCmd)
	rootCmd.AddCommand(apiTokenCmd)
	rootCmd.AddCommand(announcementsCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(userCmd)
	rootCmd.AddCommand(settingsCmd)
	rootCmd.AddCommand(rawCmd)
	rootCmd.AddCommand(examplesCmd)
	rootCmd.AddCommand(contextCmd)
	rootCmd.AddCommand(versionCmd)
}

func apiClient() (*cli.Client, error) {
	resolved, err := resolveOptions()
	if err != nil {
		return nil, err
	}
	if resolved.apiToken != "" {
		return cli.NewClientWithAuth(resolved.baseURL, "q-api-token", resolved.apiToken, resolved.timeout)
	}
	return cli.NewClient(resolved.baseURL, resolved.token, resolved.timeout)
}

func resolveOptions() (globalOptions, error) {
	resolved := globalOptions{
		baseURL: defaultBaseURL,
		timeout: opts.timeout,
		output:  opts.output,
		quiet:   opts.quiet,
	}

	cfg, err := loadConfig()
	if err != nil {
		return resolved, err
	}
	oqtopusCfg, err := loadOqtopusConfig()
	if err != nil {
		return resolved, err
	}
	if oqtopusCfg.BaseURL != "" {
		resolved.baseURL = oqtopusCfg.BaseURL
	}
	if oqtopusCfg.APIToken != "" {
		resolved.apiToken = oqtopusCfg.APIToken
	}
	if cfg.BaseURL != "" {
		resolved.baseURL = cfg.BaseURL
	}
	if cfg.APIToken != "" {
		resolved.apiToken = cfg.APIToken
	}
	if cfg.Token != "" {
		resolved.token = cfg.Token
	}
	if v := os.Getenv(envOqtopusURL); v != "" {
		resolved.baseURL = v
	}
	if v := os.Getenv(envOqtopusAPI); v != "" {
		resolved.apiToken = v
	}
	if v := os.Getenv(envBaseURL); v != "" {
		resolved.baseURL = v
	}
	if v := os.Getenv(envAPIToken); v != "" {
		resolved.apiToken = v
	}
	if v := os.Getenv(envToken); v != "" {
		resolved.token = v
		resolved.apiToken = ""
	}
	if rootCmd.PersistentFlags().Changed("base-url") {
		resolved.baseURL = opts.baseURL
	}
	if rootCmd.PersistentFlags().Changed("api-token") {
		resolved.apiToken = opts.apiToken
	}
	if rootCmd.PersistentFlags().Changed("token") {
		resolved.token = opts.token
		resolved.apiToken = ""
	}
	return resolved, nil
}

func printData(cmd *cobra.Command, data []byte) error {
	var out []byte
	switch opts.output {
	case "pretty":
		out = cli.PrettyJSON(data)
	case "json":
		out = cli.JSON(data)
	case "raw":
		out = data
	default:
		return fmt.Errorf("unsupported output format %q", opts.output)
	}
	if len(out) == 0 {
		return nil
	}
	_, err := cmd.OutOrStdout().Write(out)
	return err
}

func requestAndPrint(cmd *cobra.Command, method, path string, query map[string]string, body []byte) error {
	c, err := apiClient()
	if err != nil {
		return err
	}
	data, _, _, err := c.Do(cmd.Context(), method, path, values(query), body)
	if err != nil {
		return err
	}
	return printData(cmd, data)
}

func printValue(cmd *cobra.Command, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return printData(cmd, data)
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var httpErr *cli.HTTPError
	if errors.As(err, &httpErr) {
		switch {
		case httpErr.StatusCode == http.StatusUnauthorized || httpErr.StatusCode == http.StatusForbidden:
			return 2
		case httpErr.StatusCode == http.StatusNotFound:
			return 3
		case httpErr.StatusCode >= 400 && httpErr.StatusCode < 500:
			return 4
		case httpErr.StatusCode >= 500:
			return 5
		default:
			return 1
		}
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return 6
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return 6
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return 6
	}
	return 1
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func requireExactArgs(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			return fmt.Errorf("accepts %d arg(s), received %d", n, len(args))
		}
		return nil
	}
}
