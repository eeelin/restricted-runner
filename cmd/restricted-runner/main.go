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

type commonInputs struct {
	Request    protocol.Request
	Config     config.Config
	CallerID   string
	Target     string
	Match      policy.MatchResult
	RequestRaw []byte
}

func runValidate(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(stderr)

	callerID := fs.String("caller", "", "caller identity")
	target := fs.String("target", "", "logical target")
	configPath := fs.String("config", "", "path to YAML config")
	payload := fs.String("payload", "", "request payload JSON")

	inputs, ok := prepareCommonInputs(fs, args, stdin, stdout, *callerID, *target, *configPath, *payload)
	if !ok {
		return 1
	}

	writeJSON(stdout, map[string]any{
		"ok":      true,
		"stage":   "validated",
		"request": inputs.Request,
		"config": map[string]any{
			"root_path": inputs.Config.RootPath,
		},
		"match": map[string]any{
			"caller": inputs.Match.Caller.ID,
			"target": inputs.Target,
			"script": inputs.Match.Script.Path,
		},
	})
	return 0
}

func runDispatch(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("dispatch", flag.ContinueOnError)
	fs.SetOutput(stderr)

	callerID := fs.String("caller", "", "caller identity")
	target := fs.String("target", "", "logical target")
	configPath := fs.String("config", "", "path to YAML config")
	payload := fs.String("payload", "", "request payload JSON")
	dryRun := fs.Bool("dry-run", false, "perform preflight only")

	inputs, ok := prepareCommonInputs(fs, args, stdin, stdout, *callerID, *target, *configPath, *payload)
	if !ok {
		return 1
	}

	resolved, err := executor.Preflight(executor.ResolveInput{
		Config:  inputs.Config,
		Match:   inputs.Match,
		Request: inputs.Request,
	})
	if err != nil {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "preflight", "error": err.Error()})
		return 1
	}

	if *dryRun {
		writeJSON(stdout, map[string]any{
			"ok":            true,
			"stage":         "dispatch_preflight",
			"dry_run":       true,
			"request":       inputs.Request,
			"resolved_path": resolved.ResolvedPath,
			"match": map[string]any{
				"caller": inputs.Match.Caller.ID,
				"target": inputs.Target,
				"script": inputs.Match.Script.Path,
			},
		})
		return 0
	}

	result, err := executor.Execute(executor.ExecuteInput{
		Config:   inputs.Config,
		Match:    inputs.Match,
		Request:  inputs.Request,
		CallerID: inputs.CallerID,
		Target:   inputs.Target,
	})
	if err != nil {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "execute", "error": err.Error()})
		return 1
	}
	writeJSON(stdout, map[string]any{
		"ok":      result.OK,
		"stage":   "executed",
		"dry_run": false,
		"result":  result,
		"match": map[string]any{
			"caller": inputs.Match.Caller.ID,
			"target": inputs.Target,
			"script": inputs.Match.Script.Path,
		},
	})
	if result.OK {
		return 0
	}
	return 1
}

func prepareCommonInputs(fs *flag.FlagSet, args []string, stdin io.Reader, stdout io.Writer, callerID, target, configPath, payload string) (commonInputs, bool) {
	if err := fs.Parse(args); err != nil {
		return commonInputs{}, false
	}

	callerID = fs.Lookup("caller").Value.String()
	target = fs.Lookup("target").Value.String()
	configPath = fs.Lookup("config").Value.String()
	payload = fs.Lookup("payload").Value.String()

	requestBytes := []byte(payload)
	if len(requestBytes) == 0 {
		data, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read stdin: %v\n", err)
			return commonInputs{}, false
		}
		requestBytes = data
	}

	var req protocol.Request
	if err := json.Unmarshal(requestBytes, &req); err != nil {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "decode", "error": err.Error()})
		return commonInputs{}, false
	}
	if err := req.Validate(); err != nil {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "request_validate", "error": err.Error()})
		return commonInputs{}, false
	}
	if callerID == "" {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "input", "error": "missing caller"})
		return commonInputs{}, false
	}
	if target == "" {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "input", "error": "missing target"})
		return commonInputs{}, false
	}
	if configPath == "" {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "input", "error": "missing config"})
		return commonInputs{}, false
	}

	cfg, err := config.LoadFile(configPath)
	if err != nil {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "config_load", "error": err.Error()})
		return commonInputs{}, false
	}

	match, err := policy.Match(policy.MatchInput{
		Config:   cfg,
		Request:  req,
		CallerID: callerID,
		Target:   target,
	})
	if err != nil {
		writeJSON(stdout, map[string]any{"ok": false, "stage": "policy_match", "error": err.Error()})
		return commonInputs{}, false
	}

	return commonInputs{
		Request:    req,
		Config:     cfg,
		CallerID:   callerID,
		Target:     target,
		Match:      match,
		RequestRaw: requestBytes,
	}, true
}

func writeJSON(w io.Writer, value any) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(value)
}
