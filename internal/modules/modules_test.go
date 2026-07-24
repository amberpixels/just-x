package modules

import (
	"testing"

	"github.com/amberpixels/just-x/internal/detect"
)

func ids(mods []Module) []string {
	out := make([]string, len(mods))
	for i, m := range mods {
		out[i] = m.ID
	}
	return out
}

func TestForProject(t *testing.T) {
	tests := []struct {
		stack detect.Stack
		want  []string
	}{
		{detect.StackGo, []string{"fmt", "lint", "fix", "test", "build", "ci"}},
		{detect.StackNode, []string{"fmt", "lint", "fix", "test", "build", "ci"}},
		{detect.StackUnknown, nil},
	}
	for _, tt := range tests {
		got := ids(ForProject(detect.Project{Stack: tt.stack}))
		if len(got) != len(tt.want) {
			t.Fatalf("stack %q: got %v, want %v", tt.stack, got, tt.want)
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("stack %q: got %v, want %v", tt.stack, got, tt.want)
				break
			}
		}
	}
}

func TestResolveRequiresPullsInDeps(t *testing.T) {
	applicable := goModules()
	// Select only ci; its Requires should pull in lint and test.
	got := ResolveRequires(applicable, map[string]bool{"ci": true})
	for _, want := range []string{"ci", "lint", "test"} {
		if !got[want] {
			t.Errorf("expected %q to be resolved, got %v", want, got)
		}
	}
	if got["build"] {
		t.Errorf("build should not be pulled in by ci, got %v", got)
	}
}

// TestCIPullsInNothingThatMutates is the point of splitting fmt, lint and fix:
// ci verifies the tree it was handed and must never rewrite it. fmt and fix
// both mutate, so neither may be reachable from ci.
func TestCIPullsInNothingThatMutates(t *testing.T) {
	for _, mods := range [][]Module{goModules(), nodeModules()} {
		got := ResolveRequires(mods, map[string]bool{"ci": true})
		for _, mutating := range []string{"fmt", "fix"} {
			if got[mutating] {
				t.Errorf("ci must not pull in %q, which rewrites code; got %v", mutating, got)
			}
		}
	}
}

func TestResolveRequiresIgnoresUnselected(t *testing.T) {
	applicable := goModules()
	got := ResolveRequires(applicable, map[string]bool{"fmt": true, "lint": false})
	if !got["fmt"] {
		t.Errorf("fmt should be selected")
	}
	if got["lint"] {
		t.Errorf("lint was false, should not be selected: %v", got)
	}
}
