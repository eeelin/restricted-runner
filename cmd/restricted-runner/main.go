package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/eeelin/restricted-runner/internal/config"
	"github.com/eeelin/restricted-runner/internal/executor"
	"github.com/eeelin/restricted-runner/internal/policy"
	"github.com/eeelin/restricted-runner/internal/protocol"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: restricted-runner <validate|dispatch|version>")
		return 2
	}

	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, version)
		return 0
	case "validate":
		return runValidate(args[1:], stdin, stdout, stderr)
	case "dispatch":
		return runDispatch(args[1:], stdin, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", args[0])
		return 2
	}
}

func runValidate(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	request, cfg, callerID, target, match, ok := prepareRequest(args, stdin, stdout, stderr, "validate")
	if !ok {
		return 1
	}

	writeJSON(stdout, map[string]any{
		"ok":      true,
		"stage":   "validated",
		"request": request,
		"config": map[string]any{
			"root_path": cfg.RootPath,
		},
		"match": map[string]any{
			"caller": match.Caller.ID,
			"target": target,
			"script": match.Script.Path,
		},
	})
	_ = callerID
	return 0
}

func runDispatch(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("dispatch", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dryRun := fs.Bool("dry-run", false, "perform preflight only")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	request, cfg, callerID, target, match, ok := prepareRequest(fs.Args(), stdin, stdout, stderr, "dispatch")
	if !ok {
		return 1
	}

	resolved, err := executor.Preflight(executor.ResolveInput{
		Config:  cfg,
		Match:   match,
		Request: request,
	})
	if err != nil {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "preflight", "error": err.Error()})
		return 1
	}

	writeJSON(stdout, map[string]any{
		"ok":            true,
		"stage":         "dispatch_preflight",
		"dry_run":       *dryRun,
		"request":       request,
		"resolved_path": resolved.ResolvedPath,
		"match": map[string]any{
			"caller": match.Caller.ID,
			"target": target,
			"script": match.Script.Path,
		},
	})
	_ = callerID
	return 0
}

func prepareRequest(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, name string) (protocol.Request, config.Config, string, string, policy.MatchResult, bool) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)

	callerID := fs.String("caller", "", "caller identity")
	target := fs.String("target", "", "logical target")
	configPath := fs.String("config", "", "path to YAML config")
	payload := fs.String("payload", "", "request payload JSON")

	if err := fs.Parse(args); err != nil {
		return protocol.Request{}, config.Config{}, "", "", policy.MatchResult{}, false
	}

	requestBytes := []byte(*payload)
	if len(requestBytes) == 0 {
		data, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "failed to read stdin: %v\n", err)
			return protocol.Request{}, config.Config{}, "", "", policy.MatchResult{}, false
		}
		requestBytes = data
	}

	var req protocol.Request
	if err := json.Unmarshal(requestBytes, &req); err != nil {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "decode", "error": err.Error()})
		return protocol.Request{}, config.Config{}, "", "", policy.MatchResult{}, false
	}
	if err := req.Validate(); err != nil {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "request_validate", "error": err.Error()})
		return protocol.Request{}, config.Config{}, "", "", policy.MatchResult{}, false
	}
	if *callerID == "" {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "input", "error": "missing caller"})
		return protocol.Request{}, config.Config{}, "", "", policy.MatchResult{}, false
	}
	if *target == "" {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "input", "error": "missing target"})
		return protocol.Request{}, config.Config{}, "", "", policy.MatchResult{}, false
	}
	if *configPath == "" {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "input", "error": "missing config"})
		return protocol.Request{}, config.Config{}, "", "", policy.MatchResult{}, false
	}

	cfg, err := config.LoadFile(*configPath)
	if err != nil {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "config_load", "error": err.Error()})
		return protocol.Request{}, config.Config{}, "", "", policy.MatchResult{}, false
	}

	match, err := policy.Match(policy.MatchInput{
		Config:   cfg,
		Request:  req,
		CallerID: *callerID,
		Target:   *target,
	})
	if err != nil {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "policy_match", "error": err.Error()})
		return protocol.Request{}, config.Config{}, "", "", policy.MatchResult{}, false
	}

	return req, cfg, *callerID, *target, match, true
}

func writeJSON(w io.Writer, value any) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(value)
}
