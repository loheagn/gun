package input

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFileLine(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a_test.go")
	if err := os.WriteFile(file, []byte("package x\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	target, err := ParseFileLine([]string{file, "12"})
	if err != nil {
		t.Fatalf("ParseFileLine split args: %v", err)
	}
	if target.Line != 12 {
		t.Fatalf("line = %d, want 12", target.Line)
	}
	if target.File != file {
		t.Fatalf("file = %q, want %q", target.File, file)
	}

	target2, err := ParseFileLine([]string{file + ":9"})
	if err != nil {
		t.Fatalf("ParseFileLine combined arg: %v", err)
	}
	if target2.Line != 9 {
		t.Fatalf("line = %d, want 9", target2.Line)
	}
	if target2.File != file {
		t.Fatalf("file = %q, want %q", target2.File, file)
	}
}

func TestParseFileLineRejectsInvalid(t *testing.T) {
	dir := t.TempDir()
	nonTestFile := filepath.Join(dir, "a.go")
	if err := os.WriteFile(nonTestFile, []byte("package x\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if _, err := ParseFileLine([]string{nonTestFile, "1"}); err == nil {
		t.Fatalf("expected _test.go suffix error")
	}

	testFile := filepath.Join(dir, "b_test.go")
	if err := os.WriteFile(testFile, []byte("package x\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := ParseFileLine([]string{testFile, "0"}); err == nil {
		t.Fatalf("expected invalid line error")
	}
}

func TestParseFileLineWithOptionalRoot(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a_test.go")
	if err := os.WriteFile(file, []byte("package x\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	target, root, err := ParseFileLineWithOptionalRoot([]string{file + ":3", "/tmp/root"})
	if err != nil {
		t.Fatalf("ParseFileLineWithOptionalRoot combined: %v", err)
	}
	if target.Line != 3 || target.File != file || root != "/tmp/root" {
		t.Fatalf("unexpected parse result: %+v root=%q", target, root)
	}

	target2, root2, err := ParseFileLineWithOptionalRoot([]string{file, "7", "/tmp/root2"})
	if err != nil {
		t.Fatalf("ParseFileLineWithOptionalRoot split: %v", err)
	}
	if target2.Line != 7 || target2.File != file || root2 != "/tmp/root2" {
		t.Fatalf("unexpected parse result: %+v root=%q", target2, root2)
	}
}
