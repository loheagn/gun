package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func RepoRoot(tb testing.TB) string {
	tb.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		tb.Fatalf("failed to resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
}

func FixtureFile(tb testing.TB) string {
	tb.Helper()
	return filepath.Join(RepoRoot(tb), "testdata", "fixturemod", "sample", "sample_test.go")
}

func FixtureRoot(tb testing.TB) string {
	tb.Helper()
	return filepath.Join(RepoRoot(tb), "testdata", "fixturemod")
}

func MarkerLine(tb testing.TB, filePath string, marker string) int {
	tb.Helper()
	data, err := os.ReadFile(filePath)
	if err != nil {
		tb.Fatalf("read fixture file: %v", err)
	}
	needle := "marker:" + marker
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if strings.Contains(line, needle) {
			return i + 1
		}
	}
	tb.Fatalf("marker %q not found", marker)
	return 0
}
