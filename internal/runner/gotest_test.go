package runner

import (
	"reflect"
	"testing"

	"github.com/loheagn/gun/internal/locator"
)

func TestBuildInvocationWithRunPattern(t *testing.T) {
	res := locator.Resolution{
		Mode:       locator.ModeLeaf,
		PackageDir: "/tmp/pkg",
		RunPattern: "^TestA$/^sub$",
	}
	inv, err := BuildInvocation(res, []string{"-v"})
	if err != nil {
		t.Fatalf("BuildInvocation: %v", err)
	}
	if inv.Dir != "/tmp/pkg" {
		t.Fatalf("dir = %q", inv.Dir)
	}
	want := []string{"test", "-run", "^TestA$/^sub$", "-v", "."}
	if !reflect.DeepEqual(inv.Args, want) {
		t.Fatalf("args = %#v, want %#v", inv.Args, want)
	}
}

func TestBuildInvocationRejectsRunOverride(t *testing.T) {
	res := locator.Resolution{
		Mode:       locator.ModeLeaf,
		PackageDir: "/tmp/pkg",
		RunPattern: "^TestA$",
	}
	if _, err := BuildInvocation(res, []string{"-run", "TestB"}); err == nil {
		t.Fatalf("expected -run conflict error")
	}
}

func TestBuildInvocationProject(t *testing.T) {
	res := locator.Resolution{
		Mode:       locator.ModeProject,
		ModuleRoot: "/tmp/mod",
	}
	inv, err := BuildInvocation(res, []string{"-count=1"})
	if err != nil {
		t.Fatalf("BuildInvocation: %v", err)
	}
	if inv.Dir != "/tmp/mod" {
		t.Fatalf("dir = %q", inv.Dir)
	}
	want := []string{"test", "-count=1", "./..."}
	if !reflect.DeepEqual(inv.Args, want) {
		t.Fatalf("args = %#v, want %#v", inv.Args, want)
	}
}
