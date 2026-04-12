package executor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eeelin/restricted-runner/internal/config"
	"github.com/eeelin/restricted-runner/internal/policy"
	"github.com/eeelin/restricted-runner/internal/protocol"
)

var (
	ErrResolvedPathEscapesRoot = errors.New("resolved path escapes root")
	ErrExecutableNotFound      = errors.New("executable not found")
	ErrExecutableNotFile       = errors.New("resolved path is not a file")
	ErrExecutableNotExecutable = errors.New("resolved path is not executable")
)

type ResolveInput struct {
	Config config.Config
	Match  policy.MatchResult
	Request protocol.Request
}

type ResolveResult struct {
	ResolvedPath string
}

func Resolve(input ResolveInput) (ResolveResult, error) {
	root := filepath.Clean(input.Config.RootPath)
	joined := filepath.Join(root, input.Request.Script)
	resolved := filepath.Clean(joined)

	rootWithSep := root + string(os.PathSeparator)
	if resolved != root && !strings.HasPrefix(resolved, rootWithSep) {
		return ResolveResult{}, ErrResolvedPathEscapesRoot
	}

	return ResolveResult{ResolvedPath: resolved}, nil
}

func Preflight(input ResolveInput) (ResolveResult, error) {
	result, err := Resolve(input)
	if err != nil {
		return ResolveResult{}, err
	}

	info, err := os.Stat(result.ResolvedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ResolveResult{}, fmt.Errorf("%w: %s", ErrExecutableNotFound, result.ResolvedPath)
		}
		return ResolveResult{}, err
	}
	if !info.Mode().IsRegular() {
		return ResolveResult{}, fmt.Errorf("%w: %s", ErrExecutableNotFile, result.ResolvedPath)
	}
	if info.Mode()&0o111 == 0 {
		return ResolveResult{}, fmt.Errorf("%w: %s", ErrExecutableNotExecutable, result.ResolvedPath)
	}

	return result, nil
}
