package locator

import (
	"strings"
	"testing"

	"github.com/loheagn/gun/internal/testutil"
)

func TestResolveLeafDeepest(t *testing.T) {
	file := testutil.FixtureFile(t)
	line := testutil.MarkerLine(t, file, "inner")

	res, err := Resolve(ModeLeaf, file, line, ResolveOptions{})
	if err != nil {
		t.Fatalf("Resolve leaf: %v", err)
	}
	if res.Effective != ModeLeaf {
		t.Fatalf("effective = %q, want %q", res.Effective, ModeLeaf)
	}
	if res.RunPattern != "^TestAlpha$/^outer$/^inner$" {
		t.Fatalf("run pattern = %q", res.RunPattern)
	}
}

func TestResolveLeafFallsBackToTestWhenNoSubtest(t *testing.T) {
	file := testutil.FixtureFile(t)
	line := testutil.MarkerLine(t, file, "nosub")

	res, err := Resolve(ModeLeaf, file, line, ResolveOptions{})
	if err != nil {
		t.Fatalf("Resolve leaf fallback: %v", err)
	}
	if res.Effective != ModeTest {
		t.Fatalf("effective = %q, want %q", res.Effective, ModeTest)
	}
	if res.RunPattern != "^TestNoSub$" {
		t.Fatalf("run pattern = %q", res.RunPattern)
	}
}

func TestResolveLeafRejectsDynamicSubtest(t *testing.T) {
	file := testutil.FixtureFile(t)
	line := testutil.MarkerLine(t, file, "dynamic")

	_, err := Resolve(ModeLeaf, file, line, ResolveOptions{})
	if err == nil {
		t.Fatalf("expected dynamic name error")
	}
	if !strings.Contains(err.Error(), "unable to resolve subtest name") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveParentUp(t *testing.T) {
	file := testutil.FixtureFile(t)
	line := testutil.MarkerLine(t, file, "inner")

	res, err := Resolve(ModeParent, file, line, ResolveOptions{ParentUp: 1})
	if err != nil {
		t.Fatalf("Resolve parent up=1: %v", err)
	}
	if res.RunPattern != "^TestAlpha$/^outer$" {
		t.Fatalf("run pattern = %q", res.RunPattern)
	}

	res2, err := Resolve(ModeParent, file, line, ResolveOptions{ParentUp: 99})
	if err != nil {
		t.Fatalf("Resolve parent up=99: %v", err)
	}
	if res2.Effective != ModeTest {
		t.Fatalf("effective = %q, want %q", res2.Effective, ModeTest)
	}
	if res2.RunPattern != "^TestAlpha$" {
		t.Fatalf("run pattern = %q", res2.RunPattern)
	}
}

func TestResolveAutoFallsBackOnDynamicSubtest(t *testing.T) {
	file := testutil.FixtureFile(t)
	line := testutil.MarkerLine(t, file, "dynamic")

	res, err := Resolve(ModeAuto, file, line, ResolveOptions{})
	if err != nil {
		t.Fatalf("Resolve auto: %v", err)
	}
	if res.Effective != ModeTest {
		t.Fatalf("effective = %q, want %q", res.Effective, ModeTest)
	}
	if res.RunPattern != "^TestAlpha$" {
		t.Fatalf("run pattern = %q", res.RunPattern)
	}
}

func TestResolveConstAndFormatNames(t *testing.T) {
	file := testutil.FixtureFile(t)

	constLine := testutil.MarkerLine(t, file, "const_concat")
	resConst, err := Resolve(ModeLeaf, file, constLine, ResolveOptions{})
	if err != nil {
		t.Fatalf("Resolve const concat: %v", err)
	}
	if resConst.RunPattern != "^TestConst$/^prefix-const-sub$" {
		t.Fatalf("run pattern = %q", resConst.RunPattern)
	}

	fmtLine := testutil.MarkerLine(t, file, "fmt_sprintf")
	resFmt, err := Resolve(ModeLeaf, file, fmtLine, ResolveOptions{})
	if err != nil {
		t.Fatalf("Resolve fmt sprintf: %v", err)
	}
	if resFmt.RunPattern != "^TestConst$/^fmt-7$" {
		t.Fatalf("run pattern = %q", resFmt.RunPattern)
	}

	itoaLine := testutil.MarkerLine(t, file, "itoa")
	resItoa, err := Resolve(ModeLeaf, file, itoaLine, ResolveOptions{})
	if err != nil {
		t.Fatalf("Resolve strconv.Itoa: %v", err)
	}
	if resItoa.RunPattern != "^TestConst$/^11$" {
		t.Fatalf("run pattern = %q", resItoa.RunPattern)
	}
}

func TestResolveFilePatternAndOutsideError(t *testing.T) {
	file := testutil.FixtureFile(t)
	res, err := Resolve(ModeFile, file, 1, ResolveOptions{})
	if err != nil {
		t.Fatalf("Resolve file: %v", err)
	}
	want := "^(TestAlpha|TestBeta|TestConst|TestNoSub)$"
	if res.RunPattern != want {
		t.Fatalf("run pattern = %q, want %q", res.RunPattern, want)
	}

	outside := testutil.MarkerLine(t, file, "outside")
	if _, err := Resolve(ModeTest, file, outside, ResolveOptions{}); err == nil {
		t.Fatalf("expected outside line error")
	}
}
