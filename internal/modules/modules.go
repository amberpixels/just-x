// Package modules holds the built-in, data-driven justfile modules. Each
// module contributes a few recipes; @init toggles modules and the scaffold
// package assembles the selected ones into a justfile.
//
// Recipe bodies are text/template strings rendered with the detected
// detect.Project. They use [[ ]] delimiters (set by the scaffold package) so
// they never collide with just's own {{ }} interpolation syntax.
package modules

import "github.com/amberpixels/just-x/internal/detect"

// Module is a composable unit of a justfile (fmt, lint, test, build, ci, …).
type Module struct {
	ID       string                      // "lint", "test", "build", "ci", "fmt"
	Title    string                      // shown in the huh form
	Tools    []string                    // binaries the recipes call (v2 @doctor)
	Detect   func(p detect.Project) bool // whether to pre-tick this module
	Recipes  string                      // [[ ]]-delimited template body
	Requires []string                    // other module IDs this depends on
}

// Module IDs.
const (
	modFmt   = "fmt"
	modLint  = "lint"
	modTest  = "test"
	modBuild = "build"
	modCI    = "ci"
)

// ForProject returns the modules applicable to a detected project, in display
// order. Returns nil for an unrecognized stack.
func ForProject(p detect.Project) []Module {
	switch p.Stack {
	case detect.StackGo:
		return goModules()
	case detect.StackNode:
		return nodeModules()
	case detect.StackUnknown:
		return nil
	default:
		return nil
	}
}

// ResolveRequires expands a selection to include every module required
// (transitively) by a selected one. Cycle-safe.
func ResolveRequires(applicable []Module, selected map[string]bool) map[string]bool {
	byID := make(map[string]Module, len(applicable))
	for _, m := range applicable {
		byID[m.ID] = m
	}

	out := make(map[string]bool)
	var add func(id string)
	add = func(id string) {
		if out[id] {
			return
		}
		m, ok := byID[id]
		if !ok {
			return
		}
		out[id] = true
		for _, req := range m.Requires {
			add(req)
		}
	}
	for id, ok := range selected {
		if ok {
			add(id)
		}
	}
	return out
}

func always(detect.Project) bool { return true }

func goModules() []Module {
	return []Module{
		{
			ID:      modFmt,
			Title:   "format code (go fmt)",
			Tools:   []string{"go"},
			Detect:  always,
			Recipes: "# format Go code\nfmt:\n    go fmt ./...\n",
		},
		{
			ID:      modLint,
			Title:   "lint code (golangci-lint)",
			Tools:   []string{"golangci-lint"},
			Detect:  func(detect.Project) bool { return detect.OnPath("golangci-lint") },
			Recipes: "# lint Go code\nlint:\n    golangci-lint run\n",
		},
		{
			ID:      modTest,
			Title:   "run tests (go test)",
			Tools:   []string{"go"},
			Detect:  always,
			Recipes: "# run tests\ntest:\n    go test ./...\n",
		},
		{
			ID:      modBuild,
			Title:   "build (go build)",
			Tools:   []string{"go"},
			Detect:  always,
			Recipes: "# build\nbuild:\n    go build ./...\n",
		},
		{
			ID:       modCI,
			Title:    "aggregate checks (fmt + lint + test)",
			Detect:   always,
			Requires: []string{modFmt, modLint, modTest},
			Recipes:  "# run all checks\nci: fmt lint test\n",
		},
	}
}

func nodeModules() []Module {
	// Recipes run scripts via the detected package manager. `<pm> run <script>`
	// works for npm, pnpm, yarn and bun alike.
	return []Module{
		{
			ID:      modFmt,
			Title:   "format code",
			Detect:  always,
			Recipes: "# format code\nfmt:\n    [[.PackageManager]] run format\n",
		},
		{
			ID:      modLint,
			Title:   "lint code",
			Detect:  always,
			Recipes: "# lint code\nlint:\n    [[.PackageManager]] run lint\n",
		},
		{
			ID:      modTest,
			Title:   "run tests",
			Detect:  always,
			Recipes: "# run tests\ntest:\n    [[.PackageManager]] run test\n",
		},
		{
			ID:      modBuild,
			Title:   "build project",
			Detect:  always,
			Recipes: "# build\nbuild:\n    [[.PackageManager]] run build\n",
		},
		{
			ID:       modCI,
			Title:    "aggregate checks (fmt + lint + test)",
			Detect:   always,
			Requires: []string{modFmt, modLint, modTest},
			Recipes:  "# run all checks\nci: fmt lint test\n",
		},
	}
}
