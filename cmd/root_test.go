package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/pflag"
)

func executeForTest(args ...string) (string, string, error) {
	var out, errOut bytes.Buffer
	opts.baseURL = getenv(envBaseURL, defaultBaseURL)
	opts.token = os.Getenv(envToken)
	opts.apiToken = os.Getenv(envAPIToken)
	opts.timeout = 30 * time.Second
	opts.output = "pretty"
	opts.quiet = false
	rootCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		flag.Changed = false
	})
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errOut)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return out.String(), errOut.String(), err
}

func TestDevicesList(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/devices" {
			t.Fatalf("path = %s, want /devices", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"device_id":"SVSim"}]`))
	}))
	defer server.Close()

	out, _, err := executeForTest("--base-url", server.URL, "--token", "tok", "devices", "list")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if gotAuth != "Bearer tok" {
		t.Fatalf("Authorization = %q, want Bearer tok", gotAuth)
	}
	if !strings.Contains(out, `"device_id": "SVSim"`) {
		t.Fatalf("output = %s", out)
	}
}

func TestDevicesListWithAPIToken(t *testing.T) {
	var gotAPIToken string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIToken = r.Header.Get("q-api-token")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	_, _, err := executeForTest("--base-url", server.URL, "--api-token", "api-secret", "devices", "list")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if gotAPIToken != "api-secret" {
		t.Fatalf("q-api-token = %q, want api-secret", gotAPIToken)
	}
}

func TestDevicesSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"device_id":"qulacs","device_type":"simulator","status":"available","n_qubits":16,"n_pending_jobs":1,"basis_gates":["x"],"supported_instructions":["measure"," barrier"],"description":"Qulacs Simulator\n","device_info":"large"}]`))
	}))
	defer server.Close()

	out, _, err := executeForTest("--base-url", server.URL, "--output", "json", "devices", "summary")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if strings.Contains(out, "device_info") {
		t.Fatalf("summary leaked device_info: %s", out)
	}
	if !strings.Contains(out, `"device_id":"qulacs"`) {
		t.Fatalf("output = %s", out)
	}
	if strings.Contains(out, `" barrier"`) || strings.Contains(out, `\n`) {
		t.Fatalf("summary was not cleaned: %s", out)
	}
}

func TestJobsListQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("status"); got != "submitted" {
			t.Fatalf("status = %q, want submitted", got)
		}
		if got := q.Get("page"); got != "2" {
			t.Fatalf("page = %q, want 2", got)
		}
		if got := q.Get("size"); got != "20" {
			t.Fatalf("size = %q, want 20", got)
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	_, _, err := executeForTest("--base-url", server.URL, "jobs", "list", "--status", "submitted", "--page", "2", "--size", "20")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
}

func TestUserUpdateBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/users/me" {
			t.Fatalf("path = %s, want /users/me", r.URL.Path)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["name"] != "Alice" || body["organization"] != "Example" {
			t.Fatalf("body = %#v", body)
		}
		_, _ = w.Write([]byte(`{"name":"Alice","organization":"Example"}`))
	}))
	defer server.Close()

	_, _, err := executeForTest("--base-url", server.URL, "user", "update", "--name", "Alice", "--organization", "Example")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
}

func TestOutputJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"b":2,"a":1}`))
	}))
	defer server.Close()

	out, _, err := executeForTest("--base-url", server.URL, "--output", "json", "devices", "list")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if out != `{"a":1,"b":2}`+"\n" {
		t.Fatalf("output = %q", out)
	}
}

func TestRawCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/custom/path" {
			t.Fatalf("path = %s, want /custom/path", r.URL.Path)
		}
		if got := r.URL.Query().Get("x"); got != "1" {
			t.Fatalf("query x = %q, want 1", got)
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	out, _, err := executeForTest("--base-url", server.URL, "--output", "json", "raw", "GET", "/custom/path", "--query", "x=1")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if strings.TrimSpace(out) != `{"ok":true}` {
		t.Fatalf("output = %q", out)
	}
}

func TestJobsWait(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		status := "running"
		if calls > 1 {
			status = "succeeded"
		}
		_, _ = w.Write([]byte(`{"job_id":"job-1","status":"` + status + `"}`))
	}))
	defer server.Close()

	out, _, err := executeForTest("--base-url", server.URL, "--output", "json", "jobs", "wait", "job-1", "--interval", "1ms", "--timeout", "1s")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !strings.Contains(out, `"status":"succeeded"`) {
		t.Fatalf("output = %q", out)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestJobsResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"job_id":"job-1","status":"succeeded","job_info":{"message":"done","result":{"sampling":{"counts":{"00":10}}},"program":["large"]}}`))
	}))
	defer server.Close()

	out, _, err := executeForTest("--base-url", server.URL, "--output", "json", "jobs", "result", "job-1")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if strings.Contains(out, "program") {
		t.Fatalf("result leaked program: %s", out)
	}
	if !strings.Contains(out, `"counts":{"00":10}`) {
		t.Fatalf("output = %s", out)
	}
}

func TestJobsSubmitSampling(t *testing.T) {
	program := filepath.Join(t.TempDir(), "bell.qasm")
	if err := os.WriteFile(program, []byte("OPENQASM 3;"), 0o600); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["device_id"] != "qulacs" || body["job_type"] != "sampling" {
			t.Fatalf("body = %#v", body)
		}
		if body["shots"].(float64) != 123 {
			t.Fatalf("shots = %#v", body["shots"])
		}
		_, _ = w.Write([]byte(`{"job_id":"job-1"}`))
	}))
	defer server.Close()

	out, _, err := executeForTest("--base-url", server.URL, "--output", "json", "jobs", "submit-sampling", "--device", "qulacs", "--program", program, "--shots", "123")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if strings.TrimSpace(out) != `{"job_id":"job-1"}` {
		t.Fatalf("output = %s", out)
	}
}

func TestExamplesSubmitJobUsesRunnableBellProgram(t *testing.T) {
	out, _, err := executeForTest("--output", "json", "examples", "submit-job", "--device", "qulacs", "--shots", "123")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	for _, want := range []string{`include \"stdgates.inc\"`, `cx q[0], q[1]`, `"shots":123`} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q: %s", want, out)
		}
	}
	if strings.Contains(out, "cnot") {
		t.Fatalf("output still uses cnot: %s", out)
	}
}

func TestJobsSubmitSamplingHelpIncludesBellGuidance(t *testing.T) {
	out, _, err := executeForTest("jobs", "submit-sampling", "--help")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	for _, want := range []string{`include "stdgates.inc"`, `cx q[0], q[1]`, `--program bell.qasm`} {
		if !strings.Contains(out, want) {
			t.Fatalf("help missing %q: %s", want, out)
		}
	}
}

func TestContextCommand(t *testing.T) {
	out, _, err := executeForTest("--output", "json", "context")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !strings.Contains(out, `"safe_default_device":"qulacs"`) {
		t.Fatalf("output = %s", out)
	}
}

func TestExitCodeForHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	_, _, err := executeForTest("--base-url", server.URL, "devices", "list")
	if err == nil {
		t.Fatal("expected error")
	}
	if got := ExitCode(err); got != 2 {
		t.Fatalf("ExitCode = %d, want 2", got)
	}
}

func TestVersion(t *testing.T) {
	out, _, err := executeForTest("version")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if strings.TrimSpace(out) != "oqx version "+Version {
		t.Fatalf("output = %q", out)
	}
}

func TestGlobalDefaults(t *testing.T) {
	if opts.baseURL == "" || opts.timeout <= 0 {
		t.Fatalf("invalid defaults: %#v", opts)
	}
	if opts.timeout > time.Minute {
		t.Fatalf("timeout default = %s, want <= 1m", opts.timeout)
	}
}
