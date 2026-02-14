package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/loheagn/gun/internal/errs"
)

func ResolveRoot(filePath string, override string) (string, error) {
	if override != "" {
		abs, err := filepath.Abs(override)
		if err != nil {
			return "", errs.New(errs.CodeUsage, "failed to resolve project root path", err)
		}
		abs = filepath.Clean(abs)
		st, err := os.Stat(abs)
		if err != nil {
			return "", errs.New(errs.CodeUsage, fmt.Sprintf("project root %q not found", abs), err)
		}
		if !st.IsDir() {
			return "", errs.New(errs.CodeUsage, fmt.Sprintf("project root %q is not a directory", abs), nil)
		}
		if _, err := os.Stat(filepath.Join(abs, "go.mod")); err != nil {
			return "", errs.New(errs.CodeUsage, fmt.Sprintf("project root %q does not contain go.mod", abs), err)
		}
		return abs, nil
	}
	return FindModuleRoot(filepath.Dir(filePath))
}

func FindModuleRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", errs.New(errs.CodeUsage, "failed to resolve directory", err)
	}
	dir = filepath.Clean(dir)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errs.New(errs.CodeUsage, fmt.Sprintf("no go.mod found from %q upward", startDir), nil)
		}
		dir = parent
	}
}
