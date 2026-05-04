package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigStore(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	path, err := saveConfig(fileConfig{
		BaseURL:  "http://example.test",
		APIToken: "secret",
	})
	if err != nil {
		t.Fatalf("saveConfig returned error: %v", err)
	}
	if filepath.Base(path) != "config.json" {
		t.Fatalf("path = %s", path)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("mode = %o, want 600", got)
	}

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}
	if cfg.BaseURL != "http://example.test" || cfg.APIToken != "secret" {
		t.Fatalf("config = %#v", cfg)
	}

	if _, err := removeConfig(); err != nil {
		t.Fatalf("removeConfig returned error: %v", err)
	}
	cfg, err = loadConfig()
	if err != nil {
		t.Fatalf("loadConfig after remove returned error: %v", err)
	}
	if cfg != (fileConfig{}) {
		t.Fatalf("config after remove = %#v", cfg)
	}
}

func TestAuthStatusDoesNotPrintToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if _, err := saveConfig(fileConfig{BaseURL: "http://example.test", APIToken: "secret"}); err != nil {
		t.Fatal(err)
	}

	out, _, err := executeForTest("auth", "status")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if strings.Contains(out, "secret") {
		t.Fatalf("auth status leaked token: %s", out)
	}
	if !strings.Contains(out, "API token: saved") {
		t.Fatalf("output = %s", out)
	}
}

func TestLoadOqtopusConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	data := `[default]
url=https://api.example.test
api_token=secret
proxy=http://proxy.example.test
`
	if err := os.WriteFile(filepath.Join(home, ".oqtopus"), []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadOqtopusConfig()
	if err != nil {
		t.Fatalf("loadOqtopusConfig returned error: %v", err)
	}
	if cfg.BaseURL != "https://api.example.test" || cfg.APIToken != "secret" || cfg.Proxy != "http://proxy.example.test" {
		t.Fatalf("config = %#v", cfg)
	}
}
