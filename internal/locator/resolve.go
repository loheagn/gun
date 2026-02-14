package locator

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/loheagn/gun/internal/errs"
	"github.com/loheagn/gun/internal/project"
)

func Resolve(mode Mode, filePath string, line int, opts ResolveOptions) (Resolution, error) {
	res := Resolution{
		Mode:       mode,
		Effective:  mode,
		FilePath:   filePath,
		PackageDir: filepath.Dir(filePath),
	}

	switch mode {
	case ModePkg:
		return res, nil
	case ModeProject:
		root, err := project.ResolveRoot(filePath, opts.ProjectRoot)
		if err != nil {
			return Resolution{}, err
		}
		res.ModuleRoot = root
		return res, nil
	}

	fset, target, info, err := loadPackageTypes(filePath)
	if err != nil {
		return Resolution{}, err
	}
	fmtAliases, fmtDot := collectImportAliases(target, "fmt")
	strconvAliases, strconvDot := collectImportAliases(target, "strconv")
	ctx := &evalContext{
		info:           info,
		fmtAliases:     fmtAliases,
		fmtDot:         fmtDot,
		strconvAliases: strconvAliases,
		strconvDot:     strconvDot,
	}
	scan := scanTests(target, fset, ctx)
	if len(scan.Tests) == 0 {
		return Resolution{}, errs.New(errs.CodeUsage, "no top-level TestXxx found in file", nil)
	}

	switch mode {
	case ModeFile:
		res.RunPattern = buildFilePattern(scan.Tests)
		return res, nil
	case ModeLeaf, ModeParent, ModeTest, ModeAuto:
		top := findContainingTop(scan.Tests, line)
		if top == nil {
			return Resolution{}, errs.New(errs.CodeUsage, fmt.Sprintf("line %d is not inside any Test/t.Run block; try test/file/pkg/project", line), nil)
		}
		path := deepestPath(top, line, nil)
		return resolveFromPath(res, mode, path, opts.ParentUp)
	default:
		return Resolution{}, errs.New(errs.CodeUsage, fmt.Sprintf("unsupported mode %q", mode), nil)
	}
}

func resolveFromPath(res Resolution, mode Mode, path []*Scope, parentUp int) (Resolution, error) {
	if len(path) == 0 {
		return Resolution{}, errs.New(errs.CodeUsage, "internal error: empty test path", nil)
	}
	top := path[0]
	if top.Name == "" {
		return Resolution{}, errs.New(errs.CodeUsage, "internal error: top test has empty name", nil)
	}

	switch mode {
	case ModeTest:
		res.RunPattern = buildSegmentPattern([]string{top.Name})
		return res, nil
	case ModeLeaf:
		if len(path) == 1 {
			res.Effective = ModeTest
			res.RunPattern = buildSegmentPattern([]string{top.Name})
			return res, nil
		}
		if !allSubtestsResolvable(path) {
			return Resolution{}, errs.New(errs.CodeUsage, "unable to resolve subtest name for leaf; use test/file/pkg/project", nil)
		}
		res.RunPattern = buildSegmentPattern(pathNames(path))
		return res, nil
	case ModeParent:
		if parentUp <= 0 {
			parentUp = 1
		}
		selected := len(path) - 1 - parentUp
		if selected < 0 {
			selected = 0
		}
		targetPath := path[:selected+1]
		if selected == 0 {
			res.Effective = ModeTest
			res.RunPattern = buildSegmentPattern([]string{top.Name})
			return res, nil
		}
		if !allSubtestsResolvable(targetPath) {
			return Resolution{}, errs.New(errs.CodeUsage, "unable to resolve subtest name for parent; use test/file/pkg/project", nil)
		}
		res.RunPattern = buildSegmentPattern(pathNames(targetPath))
		return res, nil
	case ModeAuto:
		if len(path) == 1 {
			res.Effective = ModeTest
			res.RunPattern = buildSegmentPattern([]string{top.Name})
			return res, nil
		}
		if allSubtestsResolvable(path) {
			res.Effective = ModeLeaf
			res.RunPattern = buildSegmentPattern(pathNames(path))
			return res, nil
		}
		res.Effective = ModeTest
		res.RunPattern = buildSegmentPattern([]string{top.Name})
		return res, nil
	default:
		return Resolution{}, errs.New(errs.CodeUsage, fmt.Sprintf("unsupported path mode %q", mode), nil)
	}
}

func findContainingTop(tests []*Scope, line int) *Scope {
	for _, test := range tests {
		if containsLine(test, line) {
			return test
		}
	}
	return nil
}

func deepestPath(node *Scope, line int, path []*Scope) []*Scope {
	path = append(path, node)
	for _, child := range node.Children {
		if containsLine(child, line) {
			return deepestPath(child, line, path)
		}
	}
	return path
}

func containsLine(scope *Scope, line int) bool {
	return line >= scope.StartLine && line <= scope.EndLine
}

func allSubtestsResolvable(path []*Scope) bool {
	for i := 1; i < len(path); i++ {
		if !path[i].NameResolvable || path[i].Name == "" {
			return false
		}
	}
	return true
}

func pathNames(path []*Scope) []string {
	names := make([]string, 0, len(path))
	for _, scope := range path {
		names = append(names, scope.Name)
	}
	return names
}

func buildSegmentPattern(names []string) string {
	segments := make([]string, 0, len(names))
	for _, name := range names {
		segments = append(segments, "^"+regexp.QuoteMeta(name)+"$")
	}
	return strings.Join(segments, "/")
}

func buildFilePattern(tests []*Scope) string {
	names := make([]string, 0, len(tests))
	for _, test := range tests {
		names = append(names, regexp.QuoteMeta(test.Name))
	}
	sort.Strings(names)
	if len(names) == 1 {
		return "^" + names[0] + "$"
	}
	return "^(" + strings.Join(names, "|") + ")$"
}
