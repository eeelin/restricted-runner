package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/eeelin/restricted-runner/internal/protocol"
)

const VersionV1 = "v1"

var (
	ErrMissingVersion      = errors.New("missing version")
	ErrUnsupportedVersion  = errors.New("unsupported version")
	ErrMissingRootPath     = errors.New("missing root_path")
	ErrRootPathNotAbsolute = errors.New("root_path must be absolute")
	ErrMissingScripts      = errors.New("missing scripts")
	ErrMissingScriptPath   = errors.New("missing script path")
	ErrDuplicateScriptPath = errors.New("duplicate script path")
	ErrMissingCallerID     = errors.New("missing caller id")
	ErrDuplicateCallerID   = errors.New("duplicate caller id")
)

type Config struct {
	Version  string         `json:"version" yaml:"version"`
	RootPath string         `json:"root_path" yaml:"root_path"`
	Runtime  RuntimeConfig  `json:"runtime,omitempty" yaml:"runtime,omitempty"`
	Audit    AuditConfig    `json:"audit,omitempty" yaml:"audit,omitempty"`
	Callers  []CallerConfig `json:"callers,omitempty" yaml:"callers,omitempty"`
	Scripts  []ScriptConfig `json:"scripts" yaml:"scripts"`
}

type RuntimeConfig struct {
	DefaultTimeout      time.Duration `json:"default_timeout,omitempty" yaml:"default_timeout,omitempty"`
	MaxStdoutBytes      int64         `json:"max_stdout_bytes,omitempty" yaml:"max_stdout_bytes,omitempty"`
	MaxStderrBytes      int64         `json:"max_stderr_bytes,omitempty" yaml:"max_stderr_bytes,omitempty"`
	AllowStdinByDefault bool          `json:"allow_stdin_by_default,omitempty" yaml:"allow_stdin_by_default,omitempty"`
	InjectRuntimeEnv    bool          `json:"inject_runtime_env,omitempty" yaml:"inject_runtime_env,omitempty"`
}

type AuditConfig struct {
	Mode           string `json:"mode,omitempty" yaml:"mode,omitempty"`
	IncludeEnvKeys bool   `json:"include_env_keys,omitempty" yaml:"include_env_keys,omitempty"`
	IncludeArgv    bool   `json:"include_argv,omitempty" yaml:"include_argv,omitempty"`
}

type CallerConfig struct {
	ID             string   `json:"id" yaml:"id"`
	Transport      string   `json:"transport,omitempty" yaml:"transport,omitempty"`
	AllowedTargets []string `json:"allowed_targets,omitempty" yaml:"allowed_targets,omitempty"`
}

type ScriptConfig struct {
	Path           string        `json:"path" yaml:"path"`
	AllowedCallers []string      `json:"allowed_callers,omitempty" yaml:"allowed_callers,omitempty"`
	AllowedTargets []string      `json:"allowed_targets,omitempty" yaml:"allowed_targets,omitempty"`
	AllowArgv      bool          `json:"allow_argv,omitempty" yaml:"allow_argv,omitempty"`
	AllowStdin     bool          `json:"allow_stdin,omitempty" yaml:"allow_stdin,omitempty"`
	AllowedEnv     []string      `json:"allowed_env,omitempty" yaml:"allowed_env,omitempty"`
	RequiredEnv    []string      `json:"required_env,omitempty" yaml:"required_env,omitempty"`
	Timeout        time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Version) == "" {
		return ErrMissingVersion
	}
	if c.Version != VersionV1 {
		return fmt.Errorf("%w: %s", ErrUnsupportedVersion, c.Version)
	}
	if strings.TrimSpace(c.RootPath) == "" {
		return ErrMissingRootPath
	}
	if !filepath.IsAbs(c.RootPath) {
		return ErrRootPathNotAbsolute
	}
	if len(c.Scripts) == 0 {
		return ErrMissingScripts
	}

	callerIDs := map[string]struct{}{}
	for _, caller := range c.Callers {
		if err := caller.Validate(); err != nil {
			return err
		}
		if _, exists := callerIDs[caller.ID]; exists {
			return fmt.Errorf("%w: %s", ErrDuplicateCallerID, caller.ID)
		}
		callerIDs[caller.ID] = struct{}{}
	}

	scriptPaths := map[string]struct{}{}
	for _, script := range c.Scripts {
		if err := script.Validate(); err != nil {
			return err
		}
		if _, exists := scriptPaths[script.Path]; exists {
			return fmt.Errorf("%w: %s", ErrDuplicateScriptPath, script.Path)
		}
		scriptPaths[script.Path] = struct{}{}
	}

	return nil
}

func (c CallerConfig) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return ErrMissingCallerID
	}
	return nil
}

func (s ScriptConfig) Validate() error {
	if strings.TrimSpace(s.Path) == "" {
		return ErrMissingScriptPath
	}
	if err := protocol.ValidateScriptPath(s.Path); err != nil {
		if errors.Is(err, protocol.ErrMissingScript) {
			return ErrMissingScriptPath
		}
		return err
	}
	for _, key := range s.AllowedEnv {
		if protocol.IsReservedEnvKey(key) {
			return fmt.Errorf("reserved environment key conflict: %s", key)
		}
	}
	for _, key := range s.RequiredEnv {
		if protocol.IsReservedEnvKey(key) {
			return fmt.Errorf("reserved environment key conflict: %s", key)
		}
	}
	return nil
}
