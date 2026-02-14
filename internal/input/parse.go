package input

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/loheagn/gun/internal/errs"
)

type Target struct {
	File string
	Line int
}

func ParseFileLine(args []string) (Target, error) {
	switch len(args) {
	case 1:
		file, line, ok := splitFileLine(args[0])
		if !ok {
			return Target{}, errs.New(errs.CodeUsage, "expected <file> <line> or <file>:<line>", nil)
		}
		return normalizeTarget(file, line)
	case 2:
		line, err := strconv.Atoi(args[1])
		if err != nil {
			return Target{}, errs.New(errs.CodeUsage, fmt.Sprintf("invalid line number %q", args[1]), err)
		}
		return normalizeTarget(args[0], line)
	default:
		return Target{}, errs.New(errs.CodeUsage, "expected <file> <line> or <file>:<line>", nil)
	}
}

func ParseFileLineWithOptionalRoot(args []string) (Target, string, error) {
	switch len(args) {
	case 1, 2:
		target, err := ParseFileLine(args[:1])
		if err == nil {
			root := ""
			if len(args) == 2 {
				root = args[1]
			}
			return target, root, nil
		}
		target, err2 := ParseFileLine(args[:2])
		if err2 == nil {
			return target, "", nil
		}
		return Target{}, "", err
	case 3:
		target, err := ParseFileLine(args[:2])
		if err != nil {
			return Target{}, "", err
		}
		return target, args[2], nil
	default:
		return Target{}, "", errs.New(errs.CodeUsage, "expected <file> <line> [project-root] or <file>:<line> [project-root]", nil)
	}
}

func splitFileLine(raw string) (string, int, bool) {
	idx := strings.LastIndex(raw, ":")
	if idx <= 0 || idx >= len(raw)-1 {
		return "", 0, false
	}
	line, err := strconv.Atoi(raw[idx+1:])
	if err != nil {
		return "", 0, false
	}
	return raw[:idx], line, true
}

func normalizeTarget(file string, line int) (Target, error) {
	if line <= 0 {
		return Target{}, errs.New(errs.CodeUsage, "line number must be > 0", nil)
	}
	if !strings.HasSuffix(file, "_test.go") {
		return Target{}, errs.New(errs.CodeUsage, "input file must end with _test.go", nil)
	}
	abs, err := filepath.Abs(file)
	if err != nil {
		return Target{}, errs.New(errs.CodeUsage, "failed to resolve file path", err)
	}
	abs = filepath.Clean(abs)
	st, err := os.Stat(abs)
	if err != nil {
		return Target{}, errs.New(errs.CodeUsage, fmt.Sprintf("test file %q not found", abs), err)
	}
	if st.IsDir() {
		return Target{}, errs.New(errs.CodeUsage, fmt.Sprintf("%q is a directory", abs), nil)
	}
	return Target{File: abs, Line: line}, nil
}
