package locator

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/loheagn/gun/internal/errs"
)

type evalContext struct {
	info           *types.Info
	fmtAliases     map[string]bool
	fmtDot         bool
	strconvAliases map[string]bool
	strconvDot     bool
}

func loadPackageTypes(filePath string) (*token.FileSet, *ast.File, *types.Info, error) {
	absFile, err := filepath.Abs(filePath)
	if err != nil {
		return nil, nil, nil, errs.New(errs.CodeUsage, "failed to resolve target file", err)
	}
	absFile = filepath.Clean(absFile)

	probeSet := token.NewFileSet()
	probeFile, err := parser.ParseFile(probeSet, absFile, nil, 0)
	if err != nil {
		return nil, nil, nil, errs.New(errs.CodeUsage, "failed to parse target file", err)
	}

	dir := filepath.Dir(absFile)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, nil, errs.New(errs.CodeUsage, "failed to read package directory", err)
	}

	fset := token.NewFileSet()
	var files []*ast.File
	var target *ast.File
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		f, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			continue
		}
		if f.Name == nil || f.Name.Name != probeFile.Name.Name {
			continue
		}
		files = append(files, f)
		if filepath.Clean(path) == absFile {
			target = f
		}
	}

	if target == nil {
		return nil, nil, nil, errs.New(errs.CodeUsage, fmt.Sprintf("target file %q not in package set", absFile), nil)
	}

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	cfg := types.Config{
		Importer: importer.Default(),
		Error:    func(error) {},
	}
	_, _ = cfg.Check(probeFile.Name.Name, fset, files, info)

	return fset, target, info, nil
}

func collectImportAliases(file *ast.File, importPath string) (map[string]bool, bool) {
	aliases := make(map[string]bool)
	dot := false
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil || path != importPath {
			continue
		}
		if spec.Name == nil {
			base := importPath
			if idx := strings.LastIndex(importPath, "/"); idx >= 0 {
				base = importPath[idx+1:]
			}
			aliases[base] = true
			continue
		}
		switch spec.Name.Name {
		case ".":
			dot = true
		case "_":
			continue
		default:
			aliases[spec.Name.Name] = true
		}
	}
	return aliases, dot
}

func evalString(expr ast.Expr, ctx *evalContext) (string, bool) {
	if v, ok := constValue(expr, ctx); ok && v.Kind() == constant.String {
		return constant.StringVal(v), true
	}

	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind != token.STRING {
			return "", false
		}
		v, err := strconv.Unquote(e.Value)
		if err != nil {
			return "", false
		}
		return v, true
	case *ast.ParenExpr:
		return evalString(e.X, ctx)
	case *ast.BinaryExpr:
		if e.Op != token.ADD {
			return "", false
		}
		lhs, ok := evalString(e.X, ctx)
		if !ok {
			return "", false
		}
		rhs, ok := evalString(e.Y, ctx)
		if !ok {
			return "", false
		}
		return lhs + rhs, true
	case *ast.CallExpr:
		if isFmtSprintfCall(e.Fun, ctx) {
			if len(e.Args) < 1 {
				return "", false
			}
			format, ok := evalString(e.Args[0], ctx)
			if !ok {
				return "", false
			}
			vals := make([]any, 0, len(e.Args)-1)
			for _, arg := range e.Args[1:] {
				v, ok := evalAny(arg, ctx)
				if !ok {
					return "", false
				}
				vals = append(vals, v)
			}
			return fmt.Sprintf(format, vals...), true
		}
		if isStrconvItoaCall(e.Fun, ctx) {
			if len(e.Args) != 1 {
				return "", false
			}
			i, ok := evalInt(e.Args[0], ctx)
			if !ok {
				return "", false
			}
			return strconv.Itoa(int(i)), true
		}
	}

	return "", false
}

func evalAny(expr ast.Expr, ctx *evalContext) (any, bool) {
	if s, ok := evalString(expr, ctx); ok {
		return s, true
	}
	if i, ok := evalInt(expr, ctx); ok {
		return int(i), true
	}
	if f, ok := evalFloat(expr, ctx); ok {
		return f, true
	}
	if b, ok := evalBool(expr, ctx); ok {
		return b, true
	}
	return nil, false
}

func evalInt(expr ast.Expr, ctx *evalContext) (int64, bool) {
	if v, ok := constValue(expr, ctx); ok {
		if i, ok := constant.Int64Val(v); ok {
			return i, true
		}
	}
	switch e := expr.(type) {
	case *ast.ParenExpr:
		return evalInt(e.X, ctx)
	case *ast.UnaryExpr:
		v, ok := evalInt(e.X, ctx)
		if !ok {
			return 0, false
		}
		switch e.Op {
		case token.SUB:
			return -v, true
		case token.ADD:
			return v, true
		}
	}
	return 0, false
}

func evalFloat(expr ast.Expr, ctx *evalContext) (float64, bool) {
	if v, ok := constValue(expr, ctx); ok {
		if f, ok := constant.Float64Val(v); ok {
			return f, true
		}
	}
	return 0, false
}

func evalBool(expr ast.Expr, ctx *evalContext) (bool, bool) {
	if v, ok := constValue(expr, ctx); ok && v.Kind() == constant.Bool {
		return constant.BoolVal(v), true
	}
	return false, false
}

func constValue(expr ast.Expr, ctx *evalContext) (constant.Value, bool) {
	if ctx == nil || ctx.info == nil {
		return nil, false
	}
	if tv, ok := ctx.info.Types[expr]; ok && tv.Value != nil {
		return tv.Value, true
	}
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return nil, false
	}
	if obj, ok := ctx.info.Uses[ident].(*types.Const); ok {
		return obj.Val(), true
	}
	if obj, ok := ctx.info.Defs[ident].(*types.Const); ok {
		return obj.Val(), true
	}
	return nil, false
}

func isFmtSprintfCall(fun ast.Expr, ctx *evalContext) bool {
	switch f := fun.(type) {
	case *ast.SelectorExpr:
		if f.Sel.Name != "Sprintf" {
			return false
		}
		ident, ok := f.X.(*ast.Ident)
		if !ok {
			return false
		}
		return ctx != nil && ctx.fmtAliases[ident.Name]
	case *ast.Ident:
		return ctx != nil && ctx.fmtDot && f.Name == "Sprintf"
	default:
		return false
	}
}

func isStrconvItoaCall(fun ast.Expr, ctx *evalContext) bool {
	switch f := fun.(type) {
	case *ast.SelectorExpr:
		if f.Sel.Name != "Itoa" {
			return false
		}
		ident, ok := f.X.(*ast.Ident)
		if !ok {
			return false
		}
		return ctx != nil && ctx.strconvAliases[ident.Name]
	case *ast.Ident:
		return ctx != nil && ctx.strconvDot && f.Name == "Itoa"
	default:
		return false
	}
}
