package executor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eeelin/restricted-runner/internal/config"
	"github.com/eeelin/restricted-runner/internal/policy"
	"github.com/eeelin/restricted-runner/internal/protocol"
)

func TestResolveReturnsPathUnderRoot(t *testing.T) {
	root := t.TempDir()
	result, err := Resolve(ResolveInput{
		Config: config.Config{RootPath: root},
		Match:  policy.MatchResult{},
		Request: protocol.Request{Script: "homecloud/site/apply"},
	})
	if err != nil {
		t.Fatalf("expected resolve success, got error: %v", err)
	}
	want := filepath.Join(root, "homecloud/site/apply")
	if result.ResolvedPath != want {
		t.Fatalf("expected %s, got %s", want, result.ResolvedPath)
	}
}

func TestResolveRejectsEscapeFromRoot(t *testing.T) {
	root := t.TempDir()
	_, err := Resolve(ResolveInput{
		Config: config.Config{RootPath: root},
		Request: protocol.Request{Script: "../escape"},
	})
	if err != ErrResolvedPathEscapesRoot {
		t.Fatalf("expected ErrResolvedPathEscapesRoot, got %v", err)
	}
}

func TestPreflightRejectsMissingExecutable(t *testing.T) {
	root := t.TempDir()
	_, err := Preflight(ResolveInput{
		Config: config.Config{RootPath: root},
		Request: protocol.Request{Script: "homecloud/site/apply"},
	})
	if err == nil || err.Error() == "" {
		t.Fatalf("expected executable not found error, got %v", err)
	}
}

func TestPreflightRejectsNonExecutableFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "homecloud/site/apply")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := Preflight(ResolveInput{
		Config: config.Config{RootPath: root},
		Request: protocol.Request{Script: "homecloud/site/apply"},
	})
	if err == nil || err.Error() == "" {
		t.Fatalf("expected non-executable error, got %v", err)
	}
}

func TestPreflightAcceptsExecutableFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "homecloud/site/apply")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write file: %v", err)
	}

	result, err := Preflight(ResolveInput{
		Config: config.Config{RootPath: root},
		Request: protocol.Request{Script: "homecloud/site/apply"},
	})
	if err != nil {
		t.Fatalf("expected preflight success, got error: %v", err)
	}
	if result.ResolvedPath != path {
		t.Fatalf("expected %s, got %s", path, result.ResolvedPath)
	}
}
