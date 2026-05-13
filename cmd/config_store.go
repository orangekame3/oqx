package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const configDirName = "oqx"
const envOqtopusConfigPath = "OQTOPUS_CONFIG_PATH"

type fileConfig struct {
	BaseURL  string `json:"base_url,omitempty"`
	Token    string `json:"token,omitempty"`
	APIToken string `json:"api_token,omitempty"`
	Proxy    string `json:"proxy,omitempty"`
}

func configPath() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, configDirName, "config.json"), nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configDirName, "config.json"), nil
}

func loadOqtopusConfig() (fileConfig, error) {
	path, err := oqtopusConfigPath()
	if err != nil {
		return fileConfig{}, err
	}
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fileConfig{}, nil
		}
		return fileConfig{}, err
	}
	defer func() {
		_ = file.Close()
	}()

	var cfg fileConfig
	inDefault := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inDefault = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "["), "]")) == "default"
			continue
		}
		if !inDefault {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "url":
			cfg.BaseURL = strings.TrimSpace(value)
		case "api_token":
			cfg.APIToken = strings.TrimSpace(value)
		case "proxy":
			cfg.Proxy = strings.TrimSpace(value)
		}
	}
	if err := scanner.Err(); err != nil {
		return fileConfig{}, err
	}
	return cfg, nil
}

func oqtopusConfigPath() (string, error) {
	if path := os.Getenv(envOqtopusConfigPath); path != "" {
		return path, nil
	}
	path := os.ExpandEnv("~/.oqtopus")
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
	}
	return path, nil
}

func loadConfig() (fileConfig, error) {
	path, err := configPath()
	if err != nil {
		return fileConfig{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fileConfig{}, nil
		}
		return fileConfig{}, err
	}
	var cfg fileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fileConfig{}, fmt.Errorf("read config %s: %w", path, err)
	}
	return cfg, nil
}

func saveConfig(cfg fileConfig) (string, error) {
	path, err := configPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func removeConfig() (string, error) {
	path, err := configPath()
	if err != nil {
		return "", err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	return path, nil
}
