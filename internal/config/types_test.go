package config

import (
	"testing"
	"time"
)

func TestConfigValidateAcceptsMinimalValidConfig(t *testing.T) {
	cfg := Config{
		Version:  VersionV1,
		RootPath: "/opt/restricted-runner/root",
		Scripts: []ScriptConfig{
			{
				Path:           "homecloud/site/apply",
				AllowedCallers: []string{"github-actions-homecloud"},
				AllowedTargets: []string{"server"},
				AllowArgv:      true,
				Timeout:        5 * time.Minute,
			},
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestConfigValidateRejectsMissingVersion(t *testing.T) {
	cfg := Config{RootPath: "/opt/restricted-runner/root", Scripts: []ScriptConfig{{Path: "homecloud/site/apply"}}}

	if err := cfg.Validate(); err != ErrMissingVersion {
		t.Fatalf("expected ErrMissingVersion, got %v", err)
	}
}

func TestConfigValidateRejectsRelativeRootPath(t *testing.T) {
	cfg := Config{Version: VersionV1, RootPath: "relative/root", Scripts: []ScriptConfig{{Path: "homecloud/site/apply"}}}

	if err := cfg.Validate(); err != ErrRootPathNotAbsolute {
		t.Fatalf("expected ErrRootPathNotAbsolute, got %v", err)
	}
}

func TestConfigValidateRejectsMissingScripts(t *testing.T) {
	cfg := Config{Version: VersionV1, RootPath: "/opt/restricted-runner/root"}

	if err := cfg.Validate(); err != ErrMissingScripts {
		t.Fatalf("expected ErrMissingScripts, got %v", err)
	}
}

func TestConfigValidateRejectsDuplicateCallerID(t *testing.T) {
	cfg := Config{
		Version:  VersionV1,
		RootPath: "/opt/restricted-runner/root",
		Callers: []CallerConfig{
			{ID: "github-actions-homecloud"},
			{ID: "github-actions-homecloud"},
		},
		Scripts: []ScriptConfig{{Path: "homecloud/site/apply"}},
	}

	if err := cfg.Validate(); err == nil || err.Error() != "duplicate caller id: github-actions-homecloud" {
		t.Fatalf("expected duplicate caller id error, got %v", err)
	}
}

func TestConfigValidateRejectsDuplicateScriptPath(t *testing.T) {
	cfg := Config{
		Version:  VersionV1,
		RootPath: "/opt/restricted-runner/root",
		Scripts: []ScriptConfig{
			{Path: "homecloud/site/apply"},
			{Path: "homecloud/site/apply"},
		},
	}

	if err := cfg.Validate(); err == nil || err.Error() != "duplicate script path: homecloud/site/apply" {
		t.Fatalf("expected duplicate script path error, got %v", err)
	}
}

func TestScriptConfigValidateRejectsInvalidPath(t *testing.T) {
	script := ScriptConfig{Path: "../escape"}

	if err := script.Validate(); err == nil || err.Error() != "invalid script path" {
		t.Fatalf("expected invalid script path error, got %v", err)
	}
}

func TestScriptConfigValidateRejectsReservedEnvKeys(t *testing.T) {
	script := ScriptConfig{
		Path:       "homecloud/site/apply",
		AllowedEnv: []string{"RR_TARGET"},
	}

	if err := script.Validate(); err == nil || err.Error() != "reserved environment key conflict: RR_TARGET" {
		t.Fatalf("expected reserved env conflict, got %v", err)
	}
}

func TestCallerConfigValidateRejectsMissingID(t *testing.T) {
	caller := CallerConfig{}

	if err := caller.Validate(); err != ErrMissingCallerID {
		t.Fatalf("expected ErrMissingCallerID, got %v", err)
	}
}
