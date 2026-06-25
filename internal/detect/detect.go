// Package detect inspects a directory tree for project signals (go.mod,
// package.json, pyproject.toml, …) and reports the stack so @init can
// pre-tick the right modules.
package detect

// Stack identifies a detected project type.
type Stack string

const (
	StackUnknown Stack = ""
	StackGo      Stack = "go"
	StackNode    Stack = "node"
)

// Project is the result of detection: the stack plus useful context for
// module composition (package manager, tools already on PATH, …).
type Project struct {
	Root           string
	Stack          Stack
	PackageManager string   // e.g. "pnpm", "yarn", "npm" for Node
	ToolsOnPath    []string // referenced tools found on PATH
}

// Detect walks up from dir looking for project signals.
//
// M2 stub: returns an empty Project.
func Detect(dir string) (Project, error) {
	_ = dir
	return Project{}, nil
}
