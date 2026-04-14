package executor

import (
	"bytes"
	"os/exec"
	"sort"
	"time"

	"github.com/eeelin/restricted-runner/internal/config"
	"github.com/eeelin/restricted-runner/internal/policy"
	"github.com/eeelin/restricted-runner/internal/protocol"
)

type ExecuteInput struct {
	Config   config.Config
	Match    policy.MatchResult
	Request  protocol.Request
	CallerID string
	Target   string
}

func Execute(input ExecuteInput) (protocol.Result, error) {
	resolved, err := Preflight(ResolveInput{
		Config:  input.Config,
		Match:   input.Match,
		Request: input.Request,
	})
	if err != nil {
		return protocol.Result{}, err
	}

	cmd := exec.Command(resolved.ResolvedPath, input.Request.Argv...)
	cmd.Env = buildEnv(input, resolved)
	if input.Request.Stdin != nil {
		cmd.Stdin = bytes.NewBufferString(*input.Request.Stdin)
	}

	startedAt := time.Now().UTC()
	stdout, err := cmd.Output()
	finishedAt := time.Now().UTC()

	result := protocol.Result{
		OK:           err == nil,
		RequestID:    input.Request.RequestID,
		Script:       input.Request.Script,
		ResolvedPath: resolved.ResolvedPath,
		Stdout:       string(stdout),
		StartedAt:    startedAt,
		FinishedAt:   finishedAt,
		Metadata:     input.Request.Metadata,
	}

	if err == nil {
		result.ExitCode = 0
		return result, nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
		result.Stderr = string(exitErr.Stderr)
		return result, nil
	}

	return result, err
}

func buildEnv(input ExecuteInput, resolved ResolveResult) []string {
	values := map[string]string{}
	for key, value := range input.Request.Env {
		values[key] = value
	}
	if input.Config.Runtime.InjectRuntimeEnv {
		values["RR_REQUEST_ID"] = input.Request.RequestID
		values["RR_SCRIPT_PATH"] = input.Request.Script
		values["RR_ROOT_PATH"] = input.Config.RootPath
		values["RR_CALLER"] = input.CallerID
		values["RR_TARGET"] = input.Target
		values["RR_PROTOCOL_VERSION"] = input.Request.Version
		values["RR_RESOLVED_PATH"] = resolved.ResolvedPath
		if source, ok := input.Request.Metadata["source"]; ok {
			values["RR_SOURCE"] = source
		}
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	env := make([]string, 0, len(keys))
	for _, key := range keys {
		env = append(env, key+"="+values[key])
	}
	return env
}
