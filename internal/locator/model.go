package locator

type ScopeKind string

const (
	ScopeKindTest    ScopeKind = "test"
	ScopeKindSubtest ScopeKind = "subtest"
)

type Scope struct {
	Name           string
	Kind           ScopeKind
	StartLine      int
	EndLine        int
	NameResolvable bool
	Children       []*Scope
	Parent         *Scope
}

type Mode string

const (
	ModeLeaf    Mode = "leaf"
	ModeParent  Mode = "parent"
	ModeTest    Mode = "test"
	ModeFile    Mode = "file"
	ModePkg     Mode = "pkg"
	ModeProject Mode = "project"
	ModeAuto    Mode = "auto"
)

type ResolveOptions struct {
	ParentUp    int
	ProjectRoot string
}

type Resolution struct {
	Mode       Mode
	Effective  Mode
	PackageDir string
	ModuleRoot string
	FilePath   string
	RunPattern string
}
