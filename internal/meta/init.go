package meta

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"golang.org/x/term"

	"github.com/amberpixels/just-x/internal/detect"
	"github.com/amberpixels/just-x/internal/modules"
	"github.com/amberpixels/just-x/internal/scaffold"
)

// justfilePerm is the mode for a generated justfile — a normal, world-readable
// project file meant to be committed.
const justfilePerm = 0o644

// runInit implements `@init`: detect the project, pick modules (interactively
// unless assumeYes or non-TTY), compose a fenced justfile, and write it.
// Refuses to overwrite an existing justfile (v1 has no merge — that's @upgrade).
func runInit(assumeYes bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	justfilePath := filepath.Join(cwd, "justfile")
	if exists(justfilePath) {
		return errors.New("justfile already exists — refusing to overwrite " +
			"(merging into existing files is `@upgrade`, landing in v2)")
	}

	p, err := detect.Detect(cwd)
	if err != nil {
		return err
	}
	applicable := modules.ForProject(p)
	if len(applicable) == 0 {
		return errors.New("no recognized project here — need a go.mod or package.json")
	}

	// Pre-tick from detection.
	selected := make(map[string]bool)
	for _, m := range applicable {
		if m.Detect(p) {
			selected[m.ID] = true
		}
	}

	if !assumeYes && isInteractive() {
		selected, err = runForm(applicable, selected)
		if err != nil {
			return err
		}
	}

	// Pull in modules required by the selection (e.g. ci → lint/test).
	selected = modules.ResolveRequires(applicable, selected)

	var chosen []modules.Module
	for _, m := range applicable {
		if selected[m.ID] {
			chosen = append(chosen, m)
		}
	}
	if len(chosen) == 0 {
		return errors.New("no modules selected — nothing to write")
	}

	out, err := scaffold.Compose(p, chosen)
	if err != nil {
		return err
	}
	if err := os.WriteFile(justfilePath, []byte(out), justfilePerm); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "✓ wrote justfile (%s project, %d module(s): %s)\n",
		p.Stack, len(chosen), joinIDs(chosen))
	return nil
}

// runForm shows the interactive module multiselect and returns the chosen IDs.
// Pre-ticking is driven by Option.Selected; chosen receives the result.
func runForm(applicable []modules.Module, preselected map[string]bool) (map[string]bool, error) {
	options := make([]huh.Option[string], len(applicable))
	for i, m := range applicable {
		options[i] = huh.NewOption(fmt.Sprintf("%-6s %s", m.ID, m.Title), m.ID).
			Selected(preselected[m.ID])
	}

	var chosen []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select justfile modules").
				Description("Space toggles · Enter confirms").
				Options(options...).
				Value(&chosen),
		),
	)
	if err := form.Run(); err != nil {
		return nil, err
	}

	out := make(map[string]bool, len(chosen))
	for _, id := range chosen {
		out[id] = true
	}
	return out, nil
}

func isInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func joinIDs(mods []modules.Module) string {
	ids := make([]string, len(mods))
	for i, m := range mods {
		ids[i] = m.ID
	}
	return strings.Join(ids, ", ")
}
