// Package modules holds the built-in, data-driven justfile modules. Each
// module contributes a few recipes; @init toggles modules and the scaffold
// package assembles the selected ones into a justfile.
package modules

import "github.com/amberpixels/just-x/internal/detect"

// Module is a composable unit of a justfile (fmt, lint, test, build, ci, …).
type Module struct {
	ID       string                      // "lint", "test", "build", "ci", "fmt"
	Title    string                      // shown in the huh form
	Tools    []string                    // binaries the recipes call (v2 @doctor)
	Detect   func(p detect.Project) bool // pre-tick heuristic from detection
	Recipes  string                      // text/template body (fence-wrapped on render)
	Requires []string                    // other module IDs this depends on
}

// Registry is the ordered set of built-in modules.
//
// M3 stub: populated with Go/Node modules.
var Registry = []Module{}
