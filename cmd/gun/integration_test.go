package main_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/loheagn/gun/internal/testutil"
)

var (
	repoRoot  string
	gunBinary string
)

func TestMain(m *testing.M) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "resolve cwd:", err)
		os.Exit(1)
	}
	repoRoot = filepath.Clean(filepath.Join(wd, "..", ".."))
	tmpDir, err := os.MkdirTemp("", "gun-bin-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, "create temp dir:", err)
		os.Exit(1)
	}
	gunBinary = filepath.Join(tmpDir, "gun")
	buildCmd := exec.Command("go", "build", "-o", gunBinary, "./cmd/gun")
	buildCmd.Dir = repoRoot
	if out, err := buildCmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build gun failed: %v\n%s", err, string(out))
		_ = os.RemoveAll(tmpDir)
		os.Exit(1)
	}

	code := m.Run()
	_ = os.RemoveAll(tmpDir)
	os.Exit(code)
}

func TestLeafRunsDeepestSubtest(t *testing.T) {
	file := testutil.FixtureFile(t)
	line := testutil.MarkerLine(t, file, "inner")
	out, err := runGun(t, "leaf", file, strconv.Itoa(line), "--", "-v")
	if err != nil {
		t.Fatalf("gun leaf failed: %v\n%s", err, out)
	}
	mustContain(t, out, "RUN:Alpha/outer/inner")
}

func TestLeafFallsBackToTestWhenNoSubtest(t *testing.T) {
	file := testutil.FixtureFile(t)
	line := testutil.MarkerLine(t, file, "nosub")
	out, err := runGun(t, "leaf", file+":"+strconv.Itoa(line), "--", "-v")
	if err != nil {
		t.Fatalf("gun leaf fallback failed: %v\n%s", err, out)
	}
	mustContain(t, out, "RUN:NoSub")
}

func TestLeafDynamicNameFails(t *testing.T) {
	file := testutil.FixtureFile(t)
	line := testutil.MarkerLine(t, file, "dynamic")
	out, err := runGun(t, "leaf", file, strconv.Itoa(line), "--", "-v")
	if err == nil {
		t.Fatalf("expected leaf dynamic failure")
	}
	if code := exitCode(err); code != 2 {
		t.Fatalf("exit code = %d, want 2\n%s", code, out)
	}
	mustContain(t, out, "unable to resolve subtest name")
}

func TestAutoModeFallbackOnDynamicName(t *testing.T) {
	file := testutil.FixtureFile(t)
	line := testutil.MarkerLine(t, file, "dynamic")
	out, err := runGun(t, file, strconv.Itoa(line), "--", "-v")
	if err != nil {
		t.Fatalf("gun auto failed: %v\n%s", err, out)
	}
	mustContain(t, out, "RUN:Alpha/outer/inner")
	mustContain(t, out, "RUN:Alpha/dyn")
}

func TestParentAndOverflowUp(t *testing.T) {
	file := testutil.FixtureFile(t)
	line := testutil.MarkerLine(t, file, "inner")

	out, err := runGun(t, "parent", file, strconv.Itoa(line), "--up", "1", "--", "-v")
	if err != nil {
		t.Fatalf("gun parent --up 1 failed: %v\n%s", err, out)
	}
	mustContain(t, out, "RUN:Alpha/outer/inner")
	mustContain(t, out, "RUN:Alpha/dyn")

	out2, err := runGun(t, "parent", file, strconv.Itoa(line), "--up", "99", "--", "-v")
	if err != nil {
		t.Fatalf("gun parent --up 99 failed: %v\n%s", err, out2)
	}
	mustContain(t, out2, "RUN:Alpha/outer/inner")
	mustContain(t, out2, "RUN:Alpha/dyn")
}

func TestProjectDefaultAndOverrideRoot(t *testing.T) {
	file := testutil.FixtureFile(t)
	line := testutil.MarkerLine(t, file, "nosub")
	root := testutil.FixtureRoot(t)

	out, err := runGun(t, "project", file, strconv.Itoa(line), "--", "-run", "TestNoSub", "-v")
	if err != nil {
		t.Fatalf("gun project default failed: %v\n%s", err, out)
	}
	mustContain(t, out, "RUN:NoSub")

	out2, err := runGun(t, "project", file, strconv.Itoa(line), root, "--", "-run", "TestNoSub", "-v")
	if err != nil {
		t.Fatalf("gun project override failed: %v\n%s", err, out2)
	}
	mustContain(t, out2, "RUN:NoSub")
}

func TestLeafRejectsRunOverride(t *testing.T) {
	file := testutil.FixtureFile(t)
	line := testutil.MarkerLine(t, file, "inner")
	out, err := runGun(t, "leaf", file, strconv.Itoa(line), "--", "-run", "TestNoSub")
	if err == nil {
		t.Fatalf("expected run override conflict")
	}
	if code := exitCode(err); code != 2 {
		t.Fatalf("exit code = %d, want 2\n%s", code, out)
	}
	mustContain(t, out, "do not pass -run")
}

func runGun(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(gunBinary, args...)
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func mustContain(t *testing.T, out, want string) {
	t.Helper()
	if !strings.Contains(out, want) {
		t.Fatalf("output missing %q\n%s", want, out)
	}
}

func exitCode(err error) int {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return 0
}
