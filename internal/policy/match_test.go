package policy

import (
	"testing"

	"github.com/eeelin/restricted-runner/internal/config"
	"github.com/eeelin/restricted-runner/internal/protocol"
)

func TestMatchAcceptsAllowedRequest(t *testing.T) {
	cfg := config.Config{
		Callers: []config.CallerConfig{{ID: "github-actions-homecloud", AllowedTargets: []string{"server"}}},
		Scripts: []config.ScriptConfig{{
			Path:           "homecloud/site/apply",
			AllowedCallers: []string{"github-actions-homecloud"},
			AllowedTargets: []string{"server"},
			AllowStdin:     false,
			AllowedEnv:     []string{"TARGET"},
			RequiredEnv:    []string{"TARGET"},
		}},
	}
	req := protocol.Request{
		Version:   protocol.VersionV1,
		RequestID: "req-123",
		Script:    "homecloud/site/apply",
		Env: map[string]string{
			"TARGET": "server",
		},
	}

	result, err := Match(MatchInput{Config: cfg, Request: req, CallerID: "github-actions-homecloud", Target: "server"})
	if err != nil {
		t.Fatalf("expected match success, got error: %v", err)
	}
	if result.Script.Path != "homecloud/site/apply" {
		t.Fatalf("unexpected script path: %s", result.Script.Path)
	}
}

func TestMatchRejectsUnknownScript(t *testing.T) {
	cfg := config.Config{
		Callers: []config.CallerConfig{{ID: "github-actions-homecloud"}},
		Scripts: []config.ScriptConfig{{Path: "homecloud/site/apply", AllowedCallers: []string{"github-actions-homecloud"}, AllowedTargets: []string{"server"}}},
	}
	req := protocol.Request{Version: protocol.VersionV1, RequestID: "req-123", Script: "homecloud/site/delete"}

	_, err := Match(MatchInput{Config: cfg, Request: req, CallerID: "github-actions-homecloud", Target: "server"})
	if err == nil || err.Error() != "script not allowed: homecloud/site/delete" {
		t.Fatalf("expected script not allowed error, got %v", err)
	}
}

func TestMatchRejectsCallerNotAllowed(t *testing.T) {
	cfg := config.Config{
		Callers: []config.CallerConfig{{ID: "github-actions-homecloud"}},
		Scripts: []config.ScriptConfig{{Path: "homecloud/site/apply", AllowedCallers: []string{"other-caller"}, AllowedTargets: []string{"server"}}},
	}
	req := protocol.Request{Version: protocol.VersionV1, RequestID: "req-123", Script: "homecloud/site/apply"}

	_, err := Match(MatchInput{Config: cfg, Request: req, CallerID: "github-actions-homecloud", Target: "server"})
	if err == nil || err.Error() != "caller not allowed: github-actions-homecloud" {
		t.Fatalf("expected caller not allowed error, got %v", err)
	}
}

func TestMatchRejectsTargetNotAllowed(t *testing.T) {
	cfg := config.Config{
		Callers: []config.CallerConfig{{ID: "github-actions-homecloud", AllowedTargets: []string{"server"}}},
		Scripts: []config.ScriptConfig{{Path: "homecloud/site/apply", AllowedCallers: []string{"github-actions-homecloud"}, AllowedTargets: []string{"server"}}},
	}
	req := protocol.Request{Version: protocol.VersionV1, RequestID: "req-123", Script: "homecloud/site/apply"}

	_, err := Match(MatchInput{Config: cfg, Request: req, CallerID: "github-actions-homecloud", Target: "claw"})
	if err == nil || err.Error() != "target not allowed: claw" {
		t.Fatalf("expected target not allowed error, got %v", err)
	}
}

func TestMatchRejectsStdinNotAllowed(t *testing.T) {
	stdin := "hello"
	cfg := config.Config{
		Callers: []config.CallerConfig{{ID: "github-actions-homecloud", AllowedTargets: []string{"server"}}},
		Scripts: []config.ScriptConfig{{Path: "homecloud/site/apply", AllowedCallers: []string{"github-actions-homecloud"}, AllowedTargets: []string{"server"}, AllowStdin: false}},
	}
	req := protocol.Request{Version: protocol.VersionV1, RequestID: "req-123", Script: "homecloud/site/apply", Stdin: &stdin}

	_, err := Match(MatchInput{Config: cfg, Request: req, CallerID: "github-actions-homecloud", Target: "server"})
	if err != ErrStdinNotAllowed {
		t.Fatalf("expected ErrStdinNotAllowed, got %v", err)
	}
}

func TestMatchRejectsUnexpectedEnvKey(t *testing.T) {
	cfg := config.Config{
		Callers: []config.CallerConfig{{ID: "github-actions-homecloud", AllowedTargets: []string{"server"}}},
		Scripts: []config.ScriptConfig{{Path: "homecloud/site/apply", AllowedCallers: []string{"github-actions-homecloud"}, AllowedTargets: []string{"server"}, AllowedEnv: []string{"TARGET"}}},
	}
	req := protocol.Request{Version: protocol.VersionV1, RequestID: "req-123", Script: "homecloud/site/apply", Env: map[string]string{"EXTRA": "1"}}

	_, err := Match(MatchInput{Config: cfg, Request: req, CallerID: "github-actions-homecloud", Target: "server"})
	if err == nil || err.Error() != "environment key not allowed: EXTRA" {
		t.Fatalf("expected env not allowed error, got %v", err)
	}
}

func TestMatchRejectsMissingRequiredEnv(t *testing.T) {
	cfg := config.Config{
		Callers: []config.CallerConfig{{ID: "github-actions-homecloud", AllowedTargets: []string{"server"}}},
		Scripts: []config.ScriptConfig{{
			Path:           "homecloud/site/apply",
			AllowedCallers: []string{"github-actions-homecloud"},
			AllowedTargets: []string{"server"},
			RequiredEnv:    []string{"TARGET"},
		}},
	}
	req := protocol.Request{Version: protocol.VersionV1, RequestID: "req-123", Script: "homecloud/site/apply"}

	_, err := Match(MatchInput{Config: cfg, Request: req, CallerID: "github-actions-homecloud", Target: "server"})
	if err == nil || err.Error() != "missing required environment key: TARGET" {
		t.Fatalf("expected missing required env error, got %v", err)
	}
}
