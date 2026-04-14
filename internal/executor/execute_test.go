package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eeelin/restricted-runner/internal/config"
	"github.com/eeelin/restricted-runner/internal/policy"
	"github.com/eeelin/restricted-runner/internal/protocol"
)

func TestBuildEnvInjectsRuntimeKeys(t *testing.T) {
	input := ExecuteInput{
		Config: config.Config{RootPath: "/opt/restricted-runner/root", Runtime: config.RuntimeConfig{InjectRuntimeEnv: true}},
		Request: protocol.Request{
			Version:   protocol.VersionV1,
			RequestID: "req-123",
			Script:    "homecloud/site/apply",
			Env:       map[string]string{"TARGET": "server"},
			Metadata:  map[string]string{"source": "github-actions"},
		},
		CallerID: "github-actions-homecloud",
		Target:   "server",
	}

	env := buildEnv(input, ResolveResult{ResolvedPath: "/opt/restricted-runner/root/homecloud/site/apply"})
	joined := strings.Join(env, "\n")
	for _, expected := range []string{
		"TARGET=server",
		"RR_REQUEST_ID=req-123",
		"RR_SCRIPT_PATH=homecloud/site/apply",
		"RR_ROOT_PATH=/opt/restricted-runner/root",
		"RR_CALLER=github-actions-homecloud",
		"RR_TARGET=server",
		"RR_PROTOCOL_VERSION=v1",
		"RR_RESOLVED_PATH=/opt/restricted-runner/root/homecloud/site/apply",
		"RR_SOURCE=github-actions",
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected env to contain %q, got %s", expected, joined)
		}
	}
}

func TestExecuteRunsScript(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "homecloud/site/apply")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	script := "#!/bin/sh\necho hello-$1\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write file: %v", err)
	}

	result, err := Execute(ExecuteInput{
		Config: config.Config{RootPath: root, Runtime: config.RuntimeConfig{InjectRuntimeEnv: true}},
		Match:  policy.MatchResult{},
		Request: protocol.Request{
			Version:   protocol.VersionV1,
			RequestID: "req-123",
			Script:    "homecloud/site/apply",
			Argv:      []string{"world"},
		},
		CallerID: "github-actions-homecloud",
		Target:   "server",
	})
	if err != nil {
		t.Fatalf("expected execute success, got error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected result ok, got false")
	}
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", result.ExitCode)
	}
	if strings.TrimSpace(result.Stdout) != "hello-world" {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
}
