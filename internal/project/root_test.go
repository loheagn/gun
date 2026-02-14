package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindModuleRoot(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "mod")
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	got, err := FindModuleRoot(nested)
	if err != nil {
		t.Fatalf("FindModuleRoot: %v", err)
	}
	if got != root {
		t.Fatalf("root = %q, want %q", got, root)
	}
}

func TestResolveRootOverride(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "mod")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	got, err := ResolveRoot("/unused/path/file_test.go", root)
	if err != nil {
		t.Fatalf("ResolveRoot override: %v", err)
	}
	if got != root {
		t.Fatalf("root = %q, want %q", got, root)
	}
}
