// Package scaffold composes selected modules into a justfile, wrapping each
// module's recipes in `# >>> justx:<id>` / `# <<< justx:<id>` provenance fences
// so v2's @upgrade can re-sync managed regions without clobbering user edits.
package scaffold

import (
	"github.com/amberpixels/just-x/internal/detect"
	"github.com/amberpixels/just-x/internal/modules"
)

// Compose renders the chosen modules into justfile text for the given project.
//
// M3 stub: returns empty output.
func Compose(p detect.Project, mods []modules.Module) (string, error) {
	_ = p
	_ = mods
	return "", nil
}
