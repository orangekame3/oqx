package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

var rawFile string
var rawQuery []string

var rawCmd = &cobra.Command{
	Use:   "raw METHOD PATH",
	Short: "Call an arbitrary User API endpoint",
	Args:  requireExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		method := strings.ToUpper(args[0])
		if method == "" {
			return fmt.Errorf("method is required")
		}
		path := args[1]
		if !strings.HasPrefix(path, "/") {
			return fmt.Errorf("path must start with /")
		}

		var body []byte
		if rawFile != "" {
			var err error
			body, err = readBody(rawFile)
			if err != nil {
				return err
			}
		}

		c, err := apiClient()
		if err != nil {
			return err
		}
		data, _, _, err := c.Do(cmd.Context(), method, path, rawQueryValues(rawQuery), body)
		if err != nil {
			return err
		}
		return printData(cmd, data)
	},
}

func init() {
	rawCmd.Flags().StringVarP(&rawFile, "file", "f", "", "JSON request body file, or - for stdin")
	rawCmd.Flags().StringArrayVar(&rawQuery, "query", nil, "Query parameter as key=value; repeatable")
}

func rawQueryValues(items []string) url.Values {
	v := url.Values{}
	for _, item := range items {
		key, value, ok := strings.Cut(item, "=")
		if !ok {
			v.Add(item, "")
			continue
		}
		v.Add(key, value)
	}
	return v
}
