package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/eeelin/restricted-runner/internal/config"
	"github.com/eeelin/restricted-runner/internal/policy"
	"github.com/eeelin/restricted-runner/internal/protocol"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: restricted-runner <validate|version>")
		return 2
	}

	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, version)
		return 0
	case "validate":
		return runValidate(args[1:], stdin, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", args[0])
		return 2
	}
}

func runValidate(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(stderr)

	callerID := fs.String("caller", "", "caller identity")
	target := fs.String("target", "", "logical target")
	payload := fs.String("payload", "", "request payload JSON")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	requestBytes := []byte(*payload)
	if len(requestBytes) == 0 {
		data, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "failed to read stdin: %v\n", err)
			return 1
		}
		requestBytes = data
	}

	var req protocol.Request
	if err := json.Unmarshal(requestBytes, &req); err != nil {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "decode", "error": err.Error()})
		return 1
	}
	if err := req.Validate(); err != nil {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "request_validate", "error": err.Error()})
		return 1
	}
	if *callerID == "" {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "input", "error": "missing caller"})
		return 1
	}
	if *target == "" {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "input", "error": "missing target"})
		return 1
	}

	cfg := placeholderConfig()
	if err := cfg.Validate(); err != nil {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "config_validate", "error": err.Error()})
		return 1
	}

	match, err := policy.Match(policy.MatchInput{
		Config:   cfg,
		Request:  req,
		CallerID: *callerID,
		Target:   *target,
	})
	if err != nil {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "policy_match", "error": err.Error()})
		return 1
	}

	writeJSON(stdout, map[string]any{
		"ok":      true,
		"stage":   "validated",
		"request": req,
		"match": map[string]any{
			"caller": match.Caller.ID,
			"script": match.Script.Path,
		},
	})
	return 0
}

func placeholderConfig() config.Config {
	return config.Config{
		Version:  config.VersionV1,
		RootPath: "/opt/restricted-runner/root",
		Callers: []config.CallerConfig{
			{ID: "github-actions-homecloud", AllowedTargets: []string{"server", "claw"}},
		},
		Scripts: []config.ScriptConfig{
			{
				Path:           "homecloud/site/validate",
				AllowedCallers: []string{"github-actions-homecloud"},
				AllowedTargets: []string{"server", "claw"},
				AllowArgv:      true,
				AllowStdin:     true,
				AllowedEnv:     []string{"TARGET", "WORKFLOW_RUN_ID", "ACTOR"},
				RequiredEnv:    []string{"TARGET"},
			},
			{
				Path:           "homecloud/site/apply",
				AllowedCallers: []string{"github-actions-homecloud"},
				AllowedTargets: []string{"server"},
				AllowArgv:      true,
				AllowStdin:     false,
				AllowedEnv:     []string{"TARGET", "WORKFLOW_RUN_ID", "ACTOR"},
				RequiredEnv:    []string{"TARGET"},
			},
		},
	}
}

func writeJSON(w io.Writer, value any) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(value)
}
