package policy

import (
	"errors"
	"fmt"

	"github.com/eeelin/restricted-runner/internal/config"
	"github.com/eeelin/restricted-runner/internal/protocol"
)

var (
	ErrScriptNotAllowed   = errors.New("script not allowed")
	ErrCallerNotAllowed   = errors.New("caller not allowed")
	ErrTargetNotAllowed   = errors.New("target not allowed")
	ErrStdinNotAllowed    = errors.New("stdin not allowed")
	ErrEnvNotAllowed      = errors.New("environment key not allowed")
	ErrMissingRequiredEnv = errors.New("missing required environment key")
)

type MatchInput struct {
	Config   config.Config
	Request  protocol.Request
	CallerID string
	Target   string
}

type MatchResult struct {
	Caller config.CallerConfig
	Script config.ScriptConfig
}

func Match(input MatchInput) (MatchResult, error) {
	caller, ok := findCaller(input.Config.Callers, input.CallerID)
	if !ok {
		return MatchResult{}, fmt.Errorf("%w: %s", ErrCallerNotAllowed, input.CallerID)
	}
	if !containsOrEmpty(caller.AllowedTargets, input.Target) {
		return MatchResult{}, fmt.Errorf("%w: %s", ErrTargetNotAllowed, input.Target)
	}

	script, ok := findScript(input.Config.Scripts, input.Request.Script)
	if !ok {
		return MatchResult{}, fmt.Errorf("%w: %s", ErrScriptNotAllowed, input.Request.Script)
	}
	if !contains(script.AllowedCallers, input.CallerID) {
		return MatchResult{}, fmt.Errorf("%w: %s", ErrCallerNotAllowed, input.CallerID)
	}
	if !contains(script.AllowedTargets, input.Target) {
		return MatchResult{}, fmt.Errorf("%w: %s", ErrTargetNotAllowed, input.Target)
	}
	if input.Request.Stdin != nil && !script.AllowStdin {
		return MatchResult{}, ErrStdinNotAllowed
	}
	for key := range input.Request.Env {
		if !contains(script.AllowedEnv, key) {
			return MatchResult{}, fmt.Errorf("%w: %s", ErrEnvNotAllowed, key)
		}
	}
	for _, key := range script.RequiredEnv {
		if _, ok := input.Request.Env[key]; !ok {
			return MatchResult{}, fmt.Errorf("%w: %s", ErrMissingRequiredEnv, key)
		}
	}

	return MatchResult{Caller: caller, Script: script}, nil
}

func findCaller(callers []config.CallerConfig, id string) (config.CallerConfig, bool) {
	for _, caller := range callers {
		if caller.ID == id {
			return caller, true
		}
	}
	return config.CallerConfig{}, false
}

func findScript(scripts []config.ScriptConfig, path string) (config.ScriptConfig, bool) {
	for _, script := range scripts {
		if script.Path == path {
			return script, true
		}
	}
	return config.ScriptConfig{}, false
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsOrEmpty(values []string, want string) bool {
	if len(values) == 0 {
		return true
	}
	return contains(values, want)
}
