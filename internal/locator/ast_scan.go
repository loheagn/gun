package locator

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode"
	"unicode/utf8"
)

type scanResult struct {
	Tests []*Scope
}

func scanTests(file *ast.File, fset *token.FileSet, eval *evalContext) *scanResult {
	testingAliases, testingDot := collectImportAliases(file, "testing")
	res := &scanResult{}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Body == nil {
			continue
		}
		if !isTopLevelTestName(fn.Name.Name) {
			continue
		}
		tVar, ok := extractTestingParamName(fn.Type, testingAliases, testingDot)
		if !ok {
			continue
		}
		testScope := &Scope{
			Name:           fn.Name.Name,
			Kind:           ScopeKindTest,
			StartLine:      fset.Position(fn.Body.Pos()).Line,
			EndLine:        fset.Position(fn.Body.End()).Line,
			NameResolvable: true,
		}
		if tVar != "" {
			scanStmtList(fn.Body.List, tVar, testScope, fset, testingAliases, testingDot, eval)
		}
		res.Tests = append(res.Tests, testScope)
	}
	return res
}

func scanStmtList(stmts []ast.Stmt, currentT string, parent *Scope, fset *token.FileSet, testingAliases map[string]bool, testingDot bool, eval *evalContext) {
	for _, stmt := range stmts {
		scanStmt(stmt, currentT, parent, fset, testingAliases, testingDot, eval)
	}
}

func scanStmt(stmt ast.Stmt, currentT string, parent *Scope, fset *token.FileSet, testingAliases map[string]bool, testingDot bool, eval *evalContext) {
	ast.Inspect(stmt, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncLit:
			return false
		case *ast.CallExpr:
			child, nextT, callbackBody := matchSubtestCall(node, currentT, fset, testingAliases, testingDot, eval)
			if child == nil {
				return true
			}
			child.Parent = parent
			parent.Children = append(parent.Children, child)
			if callbackBody != nil && nextT != "" {
				scanStmtList(callbackBody.List, nextT, child, fset, testingAliases, testingDot, eval)
			}
		}
		return true
	})
}

func matchSubtestCall(call *ast.CallExpr, currentT string, fset *token.FileSet, testingAliases map[string]bool, testingDot bool, eval *evalContext) (*Scope, string, *ast.BlockStmt) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Run" {
		return nil, "", nil
	}
	recv, ok := sel.X.(*ast.Ident)
	if !ok || recv.Name != currentT {
		return nil, "", nil
	}
	if len(call.Args) < 2 {
		return nil, "", nil
	}

	name, resolvable := evalString(call.Args[0], eval)
	callback, _ := call.Args[1].(*ast.FuncLit)
	startLine := fset.Position(call.Pos()).Line
	endLine := fset.Position(call.End()).Line
	nextT := ""
	if callback != nil && callback.Body != nil {
		startLine = fset.Position(callback.Body.Pos()).Line
		endLine = fset.Position(callback.Body.End()).Line
		nextT, _ = extractTestingParamName(callback.Type, testingAliases, testingDot)
	}

	child := &Scope{
		Name:           name,
		Kind:           ScopeKindSubtest,
		StartLine:      startLine,
		EndLine:        endLine,
		NameResolvable: resolvable,
	}
	if !resolvable {
		child.Name = ""
	}
	if callback == nil {
		return child, "", nil
	}
	return child, nextT, callback.Body
}

func isTopLevelTestName(name string) bool {
	if !strings.HasPrefix(name, "Test") {
		return false
	}
	if len(name) == len("Test") {
		return true
	}
	r, _ := utf8.DecodeRuneInString(name[len("Test"):])
	return !unicode.IsLower(r)
}

func extractTestingParamName(ft *ast.FuncType, aliases map[string]bool, dot bool) (string, bool) {
	if ft == nil || ft.Params == nil || len(ft.Params.List) != 1 {
		return "", false
	}
	param := ft.Params.List[0]
	if !isTestingTType(param.Type, aliases, dot) {
		return "", false
	}
	if len(param.Names) == 0 {
		return "", true
	}
	name := param.Names[0].Name
	if name == "_" {
		return "", false
	}
	return name, true
}

func isTestingTType(expr ast.Expr, aliases map[string]bool, dot bool) bool {
	star, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}
	switch x := star.X.(type) {
	case *ast.SelectorExpr:
		ident, ok := x.X.(*ast.Ident)
		if !ok {
			return false
		}
		return aliases[ident.Name] && x.Sel.Name == "T"
	case *ast.Ident:
		return dot && x.Name == "T"
	default:
		return false
	}
}
