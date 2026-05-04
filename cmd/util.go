package cmd

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
)

func values(items map[string]string) url.Values {
	v := url.Values{}
	for key, value := range items {
		if value != "" {
			v.Set(key, value)
		}
	}
	return v
}

func readBody(path string) ([]byte, error) {
	var data []byte
	var err error
	if path == "-" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return nil, err
	}
	if !json.Valid(data) {
		return nil, fmt.Errorf("%s is not valid JSON", path)
	}
	return data, nil
}

func updateUserBody(file, name, organization string) ([]byte, error) {
	if file != "" {
		return readBody(file)
	}
	body := map[string]string{}
	if name != "" {
		body["name"] = name
	}
	if organization != "" {
		body["organization"] = organization
	}
	if len(body) == 0 {
		return nil, errors.New("set --name, --organization, or --file")
	}
	return json.Marshal(body)
}

func writeSSELog(data []byte, output string) error {
	var resp struct {
		File string `json:"file"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}
	if resp.File == "" {
		return errors.New("response does not contain file")
	}
	decoded, err := base64.StdEncoding.DecodeString(resp.File)
	if err != nil {
		return fmt.Errorf("decode sselog file: %w", err)
	}
	return os.WriteFile(output, decoded, 0o600)
}
