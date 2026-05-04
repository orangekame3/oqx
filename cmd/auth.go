package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var authToken string
var authAPIToken string
var authTokenStdin bool
var authBearer bool

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage saved authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Save API URL and API token",
	Long: `Save API URL and API token to the oqx config file.

By default oqx uses the q-api-token header, matching quri-parts-oqtopus.
Use --bearer to save a Bearer token for Authorization instead.
Use --token-stdin to avoid storing the token in shell history.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		token := authAPIToken
		if token == "" {
			token = authToken
		}
		if authTokenStdin {
			read, err := readTokenFromStdin(cmd)
			if err != nil {
				return err
			}
			token = read
		}
		if token == "" {
			return errors.New("set --api-token, --token, or --token-stdin")
		}

		resolved, err := resolveOptions()
		if err != nil {
			return err
		}
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		cfg.BaseURL = resolved.baseURL
		if authBearer {
			cfg.Token = token
			cfg.APIToken = ""
		} else {
			cfg.APIToken = token
			cfg.Token = ""
		}
		path, err := saveConfig(cfg)
		if err != nil {
			return err
		}
		if opts.quiet {
			return nil
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Saved credentials to %s\n", path)
		return err
	},
}

func readTokenFromStdin(cmd *cobra.Command) (string, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeCharDevice != 0 {
		if !opts.quiet {
			_, _ = fmt.Fprint(cmd.ErrOrStderr(), "Token: ")
		}
		token, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}
		return strings.TrimSpace(token), nil
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication configuration status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		path, err := configPath()
		if err != nil {
			return err
		}
		oqtopusCfg, err := loadOqtopusConfig()
		if err != nil {
			return err
		}
		oqtopusPath, err := oqtopusConfigPath()
		if err != nil {
			return err
		}
		resolved, err := resolveOptions()
		if err != nil {
			return err
		}

		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = defaultBaseURL
		}
		apiTokenStatus := "not saved"
		if cfg.APIToken != "" {
			apiTokenStatus = "saved"
		}
		bearerStatus := "not saved"
		if cfg.Token != "" {
			bearerStatus = "saved"
		}
		sdkStatus := "not found"
		if oqtopusCfg.BaseURL != "" || oqtopusCfg.APIToken != "" {
			sdkStatus = "found"
		}
		effectiveAuth := "none"
		if resolved.apiToken != "" {
			effectiveAuth = "q-api-token"
		} else if resolved.token != "" {
			effectiveAuth = "bearer"
		}
		_, err = fmt.Fprintf(
			cmd.OutOrStdout(),
			"Config: %s\nSDK config: %s (%s)\nBase URL: %s\nAPI token: %s\nBearer token: %s\nEffective base URL: %s\nEffective auth: %s\n",
			path,
			oqtopusPath,
			sdkStatus,
			baseURL,
			apiTokenStatus,
			bearerStatus,
			resolved.baseURL,
			effectiveAuth,
		)
		return err
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove saved authentication",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := removeConfig()
		if err != nil {
			return err
		}
		if opts.quiet {
			return nil
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed credentials from %s\n", path)
		return err
	},
}

func init() {
	authLoginCmd.Flags().StringVar(&authAPIToken, "api-token", "", "OQTOPUS API token to save")
	authLoginCmd.Flags().StringVar(&authToken, "token", "", "Token to save; saved as API token unless --bearer is set")
	authLoginCmd.Flags().BoolVar(&authTokenStdin, "token-stdin", false, "Read token from stdin")
	authLoginCmd.Flags().BoolVar(&authBearer, "bearer", false, "Save token as Authorization bearer token instead of q-api-token")

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
}
