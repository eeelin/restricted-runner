package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunValidateSuccess(t *testing.T) {
	configPath := writeExampleConfig(t)
	payload := `{"version":"v1","request_id":"req-123","script":"homecloud/site/validate","argv":["sites/homes/ruyi/hass"],"env":{"TARGET":"server","ACTOR":"eeelin"},"stdin":"hello from stdin"}`

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"validate", "--config", configPath, "--caller", "github-actions-homecloud", "--target", "server", "--payload", payload}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%s", code, stderr.String())
	}
	var body map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &body); err != nil {
		t.Fatalf("decode stdout: %v", err)
	}
	if body["stage"] != "validated" {
		t.Fatalf("expected stage validated, got %v", body["stage"])
	}
}

func TestRunDispatchDryRunSuccess(t *testing.T) {
	configPath := writeExampleConfig(t)
	payload := `{"version":"v1","request_id":"req-124","script":"homecloud/site/apply","argv":["sites/homes/ruyi/hass"],"env":{"TARGET":"server","ACTOR":"eeelin"}}`

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"dispatch", "--dry-run", "--config", configPath, "--caller", "github-actions-homecloud", "--target", "server", "--payload", payload}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%s", code, stderr.String())
	}
	var body map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &body); err != nil {
		t.Fatalf("decode stdout: %v", err)
	}
	if body["stage"] != "dispatch_preflight" {
		t.Fatalf("expected stage dispatch_preflight, got %v", body["stage"])
	}
}

func TestRunDispatchExecuteSuccess(t *testing.T) {
	configPath := writeExampleConfig(t)
	payload := `{"version":"v1","request_id":"req-125","script":"homecloud/site/apply","argv":["sites/homes/ruyi/hass"],"env":{"TARGET":"server","ACTOR":"eeelin"}}`

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"dispatch", "--config", configPath, "--caller", "github-actions-homecloud", "--target", "server", "--payload", payload}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%s", code, stderr.String())
	}
	var body map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &body); err != nil {
		t.Fatalf("decode stdout: %v", err)
	}
	if body["stage"] != "executed" {
		t.Fatalf("expected stage executed, got %v", body["stage"])
	}
	if body["ok"] != true {
		t.Fatalf("expected ok=true, got %v", body["ok"])
	}
}

func TestRunDispatchRejectsMissingCaller(t *testing.T) {
	configPath := writeExampleConfig(t)
	payload := `{"version":"v1","request_id":"req-126","script":"homecloud/site/apply","argv":["sites/homes/ruyi/hass"],"env":{"TARGET":"server","ACTOR":"eeelin"}}`

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"dispatch", "--config", configPath, "--target", "server", "--payload", payload}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d, stderr=%s", code, stderr.String())
	}
	var body map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &body); err != nil {
		t.Fatalf("decode stdout: %v", err)
	}
	if body["stage"] != "input" {
		t.Fatalf("expected stage input, got %v", body["stage"])
	}
}

func writeExampleConfig(t *testing.T) string {
	t.Helper()
	repoRoot := findRepoRoot(t)
	rootPath := filepath.Join(repoRoot, "examples", "root")
	configContent := strings.ReplaceAll(`version: v1
root_path: ROOT_PATH_PLACEHOLDER

runtime:
  inject_runtime_env: true

audit:
  mode: stderr
  include_argv: true

callers:
  - id: github-actions-homecloud
    transport: ssh
    allowed_targets:
      - server
      - claw

scripts:
  - path: homecloud/site/validate
    allowed_callers:
      - github-actions-homecloud
    allowed_targets:
      - server
      - claw
    allow_argv: true
    allow_stdin: true
    allowed_env:
      - TARGET
      - WORKFLOW_RUN_ID
      - ACTOR
    required_env:
      - TARGET

  - path: homecloud/site/apply
    allowed_callers:
      - github-actions-homecloud
    allowed_targets:
      - server
    allow_argv: true
    allow_stdin: false
    allowed_env:
      - TARGET
      - WORKFLOW_RUN_ID
      - ACTOR
    required_env:
      - TARGET
`, "ROOT_PATH_PLACEHOLDER", rootPath)

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return configPath
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return filepath.Dir(filepath.Dir(cwd))
}
