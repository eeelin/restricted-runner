package protocol

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"time"
)

const VersionV1 = "v1"

var (
	ErrMissingVersion         = errors.New("missing version")
	ErrUnsupportedVersion     = errors.New("unsupported version")
	ErrMissingRequestID       = errors.New("missing request_id")
	ErrMissingScript          = errors.New("missing script")
	ErrInvalidScriptPath      = errors.New("invalid script path")
	ErrReservedEnvKeyConflict = errors.New("reserved environment key conflict")
)

type Request struct {
	Version   string            `json:"version"`
	RequestID string            `json:"request_id"`
	Script    string            `json:"script"`
	Argv      []string          `json:"argv,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	Stdin     *string           `json:"stdin,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type Result struct {
	OK           bool              `json:"ok"`
	RequestID    string            `json:"request_id,omitempty"`
	Script       string            `json:"script,omitempty"`
	ResolvedPath string            `json:"resolved_path,omitempty"`
	ExitCode     int               `json:"exit_code"`
	Stdout       string            `json:"stdout,omitempty"`
	Stderr       string            `json:"stderr,omitempty"`
	StartedAt    time.Time         `json:"started_at,omitempty"`
	FinishedAt   time.Time         `json:"finished_at,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

func (r Request) Validate() error {
	if strings.TrimSpace(r.Version) == "" {
		return ErrMissingVersion
	}
	if r.Version != VersionV1 {
		return fmt.Errorf("%w: %s", ErrUnsupportedVersion, r.Version)
	}
	if strings.TrimSpace(r.RequestID) == "" {
		return ErrMissingRequestID
	}
	if err := ValidateScriptPath(r.Script); err != nil {
		return err
	}
	for key := range r.Env {
		if IsReservedEnvKey(key) {
			return fmt.Errorf("%w: %s", ErrReservedEnvKeyConflict, key)
		}
	}
	return nil
}

func ValidateScriptPath(script string) error {
	raw := strings.TrimSpace(script)
	if raw == "" {
		return ErrMissingScript
	}
	if strings.HasPrefix(raw, "/") {
		return ErrInvalidScriptPath
	}
	cleaned := path.Clean(raw)
	if cleaned == "." || cleaned == "" {
		return ErrInvalidScriptPath
	}
	parts := strings.Split(cleaned, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return ErrInvalidScriptPath
		}
	}
	if cleaned != raw {
		return ErrInvalidScriptPath
	}
	return nil
}

func IsReservedEnvKey(key string) bool {
	return strings.HasPrefix(key, "RR_")
}
