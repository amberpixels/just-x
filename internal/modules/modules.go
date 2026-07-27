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
//
// fmt, lint and fix are three distinct verbs, and the distinction is the whole
// reason they are separate modules:
//
//	lint  reads code and reports findings. Never mutates.
//	fmt   rewrites code to canonical form. Mutates, but decides nothing:
//	      every input has exactly one formatted output.
//	fix   applies the mechanical subset of lint findings. Mutates, and does
//	      make judgement calls, so it is the one to run on a clean tree.
//
// Only lint is safe in CI.
const (
	modFmt   = "fmt"
	modLint  = "lint"
	modFix   = "fix"
	modTest  = "test"
	modBuild = "build"
	modCI    = "ci"
)

// Titles for the three verbs above, shown in the @init form. They are shared
// across stacks because the promise each verb makes is the same whatever the
// language; only the command behind it changes.
const (
	titleFmt  = "format code (rewrites, decides nothing)"
	titleLint = "lint code (reports, never mutates)"
	titleFix  = "auto-fix findings (rewrites, makes judgement calls)"
)

// ForProject returns the modules applicable to a detected project, in display
// order. Returns nil for an unrecognized stack.
func ForProject(p detect.Project) []Module {
	switch p.Stack {
	case detect.StackGo:
		return goModules()
	case detect.StackNode:
		return nodeModules()
	case detect.StackShell:
		return shellModules()
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
			ID:     modFmt,
			Title:  titleFmt,
			Tools:  []string{"go"},
			Detect: always,
			Recipes: "# format Go code - rewrites to canonical form\nfmt:\n" +
				"[[ if .Standardgo ]]    go tool standardgo fmt ./...\n" +
				"[[ else ]]    go fmt ./...\n[[ end ]]",
		},
		{
			ID:     modLint,
			Title:  titleLint,
			Tools:  []string{"go"},
			Detect: func(p detect.Project) bool { return p.Standardgo || detect.OnPath("golangci-lint") },
			Recipes: "# lint Go code - reports findings, changes nothing\nlint:\n" +
				"[[ if .Standardgo ]]    go tool standardgo ./...\n" +
				"[[ else ]]    golangci-lint run\n[[ end ]]",
		},
		{
			ID:     modFix,
			Title:  titleFix,
			Tools:  []string{"go"},
			Detect: func(p detect.Project) bool { return p.Standardgo || detect.OnPath("golangci-lint") },
			Recipes: "# auto-fix what can be fixed - run on a clean tree and read the diff\nfix:\n" +
				"[[ if .Standardgo ]]    go tool standardgo ./... --fix\n" +
				"[[ else ]]    golangci-lint run --fix\n[[ end ]]",
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
			Title:    "aggregate checks (lint + test)",
			Detect:   always,
			Requires: []string{modLint, modTest},
			// Deliberately no fmt: CI verifies, it does not rewrite the tree it
			// was handed. Formatting problems still fail this recipe, because
			// linting reports them as ordinary findings.
			Recipes: "# run all checks - read-only, safe for CI\nci: lint test\n",
		},
	}
}

// Shell recipe fragments.
//
// discoverShell lists the scripts to check. shfmt's own -f finds shell files by
// extension and by shebang, matching what detection looked for, so a script
// named `bin/deploy` is covered without being listed anywhere.
//
// `xargs -r` keeps a tree whose scripts have all been removed from invoking
// shellcheck with no arguments, which it treats as an error. GNU xargs runs the
// command on empty input; BSD does not; -r makes both skip it.
const (
	toolShfmt      = "shfmt"
	toolShellcheck = "shellcheck"

	discoverShell = toolShfmt + " -f ."
	// shfmtStyle is the house format: two-space indent, indented switch cases.
	shfmtStyle = "-i 2 -ci"
)

func shellModules() []Module {
	return []Module{
		{
			ID:     modFmt,
			Title:  titleFmt,
			Tools:  []string{toolShfmt},
			Detect: always,
			Recipes: "# format shell scripts - rewrites to canonical form\nfmt:\n" +
				"    shfmt -w " + shfmtStyle + " .\n",
		},
		{
			ID:     modLint,
			Title:  titleLint,
			Tools:  []string{toolShellcheck, toolShfmt},
			Detect: func(detect.Project) bool { return detect.OnPath(toolShellcheck) },
			Recipes: "# lint shell scripts - reports findings, changes nothing\nlint:\n" +
				"    " + discoverShell + " | xargs -r shellcheck\n" +
				"    shfmt -d " + shfmtStyle + " .\n",
		},
		{
			ID:     modFix,
			Title:  titleFix,
			Tools:  []string{toolShellcheck, toolShfmt},
			Detect: func(detect.Project) bool { return detect.OnPath(toolShellcheck) },
			// Both halves make judgement calls, which is why they are here and not
			// in fmt: shellcheck's diff format emits a patch of the findings it can
			// mechanically fix, and shfmt -s rewrites code it considers redundant.
			// --allow-empty is required because a clean tree yields no patch at all.
			Recipes: "# auto-fix what can be fixed - run on a clean tree and read the diff\nfix:\n" +
				"    " + discoverShell + " | xargs -r shellcheck -f diff | git apply --allow-empty\n" +
				"    shfmt -w -s " + shfmtStyle + " .\n",
		},
		{
			ID:       modCI,
			Title:    "aggregate checks (lint)",
			Detect:   always,
			Requires: []string{modLint},
			// No test module for shell: there is no runner conventional enough to
			// assume, so ci aggregates lint alone rather than naming a recipe that
			// would not exist.
			Recipes: "# run all checks - read-only, safe for CI\nci: lint\n",
		},
	}
}

func nodeModules() []Module {
	// Recipes run scripts via the detected package manager. `<pm> run <script>`
	// works for npm, pnpm, yarn and bun alike.
	return []Module{
		{
			ID:      modFmt,
			Title:   titleFmt,
			Detect:  always,
			Recipes: "# format code - rewrites to canonical form\nfmt:\n    [[.PackageManager]] run format\n",
		},
		{
			ID:      modLint,
			Title:   titleLint,
			Detect:  always,
			Recipes: "# lint code - reports findings, changes nothing\nlint:\n    [[.PackageManager]] run lint\n",
		},
		{
			ID:      modFix,
			Title:   titleFix,
			Detect:  always,
			Recipes: "# auto-fix what can be fixed\nfix:\n    [[.PackageManager]] run fix\n",
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
			Title:    "aggregate checks (lint + test)",
			Detect:   always,
			Requires: []string{modLint, modTest},
			// No fmt here either: `<pm> run format` conventionally maps to
			// prettier --write, which rewrites the tree.
			Recipes: "# run all checks - read-only, safe for CI\nci: lint test\n",
		},
	}
}
