package detect

import (
	"os"
	"path/filepath"
	"testing"
)

// write creates files (relative to dir) with empty content.
func write(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, n := range names {
		p := filepath.Join(dir, n)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDetectStack(t *testing.T) {
	tests := []struct {
		name   string
		files  []string
		wantSt Stack
		wantPM string
	}{
		{"go", []string{"go.mod"}, StackGo, ""},
		{"node pnpm", []string{"package.json", "pnpm-lock.yaml"}, StackNode, "pnpm"},
		{"node yarn", []string{"package.json", "yarn.lock"}, StackNode, "yarn"},
		{"node bun", []string{"package.json", "bun.lockb"}, StackNode, "bun"},
		{"node npm lock", []string{"package.json", "package-lock.json"}, StackNode, "npm"},
		{"node no lock defaults npm", []string{"package.json"}, StackNode, "npm"},
		{"go wins over node", []string{"go.mod", "package.json"}, StackGo, ""},
		{"unknown", []string{"README.md"}, StackUnknown, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			write(t, dir, tt.files...)

			p, err := Detect(dir)
			if err != nil {
				t.Fatal(err)
			}
			if p.Stack != tt.wantSt {
				t.Errorf("Stack = %q, want %q", p.Stack, tt.wantSt)
			}
			if p.PackageManager != tt.wantPM {
				t.Errorf("PackageManager = %q, want %q", p.PackageManager, tt.wantPM)
			}
			if p.Root != dir {
				t.Errorf("Root = %q, want %q", p.Root, dir)
			}
		})
	}
}

func TestDetectWalksUp(t *testing.T) {
	root := t.TempDir()
	write(t, root, "go.mod")
	deep := filepath.Join(root, "internal", "pkg")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}

	p, err := Detect(deep)
	if err != nil {
		t.Fatal(err)
	}
	if p.Stack != StackGo {
		t.Errorf("Stack = %q, want %q", p.Stack, StackGo)
	}
	if p.Root != root {
		t.Errorf("Root = %q, want %q (should be where go.mod lives)", p.Root, root)
	}
}

func TestDetectUnknownRootsAtDir(t *testing.T) {
	dir := t.TempDir()
	p, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if p.Stack != StackUnknown {
		t.Errorf("Stack = %q, want unknown", p.Stack)
	}
	if p.Root != dir {
		t.Errorf("Root = %q, want %q", p.Root, dir)
	}
}
