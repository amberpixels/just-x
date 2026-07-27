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
		{"shell in bin", []string{"bin/resolve.sh"}, StackShell, ""},
		{"shell in scripts", []string{"scripts/release.sh"}, StackShell, ""},
		{"shell at root", []string{"install.sh"}, StackShell, ""},
		{"shell .bash extension", []string{"bin/lib.bash"}, StackShell, ""},
		{"go wins over shell", []string{"go.mod", "scripts/build.sh"}, StackGo, ""},
		{"node wins over shell", []string{"package.json", "scripts/build.sh"}, StackNode, "npm"},
		{"unknown", []string{"README.md"}, StackUnknown, ""},
		{"nested scripts do not count", []string{"vendor/lib/deep.sh"}, StackUnknown, ""},
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

// TestDetectShellShebang covers the extensionless case, which is how scripts
// meant to be run as commands are conventionally named. Content matters here,
// so these cannot use the empty-file write helper.
func TestDetectShellShebang(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    Stack
	}{
		{"env bash", "#!/usr/bin/env bash\necho hi\n", StackShell},
		{"absolute sh", "#!/bin/sh\necho hi\n", StackShell},
		{"env -S with flags", "#!/usr/bin/env -S bash -e\necho hi\n", StackShell},
		{"zsh", "#!/bin/zsh\necho hi\n", StackShell},
		{"python is not shell", "#!/usr/bin/env python3\nprint('hi')\n", StackUnknown},
		{"no shebang", "just some text\n", StackUnknown},
		{"shebang not on first line", "# a comment\n#!/bin/bash\n", StackUnknown},
		{"empty file", "", StackUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			bin := filepath.Join(dir, "bin")
			if err := os.MkdirAll(bin, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(bin, "deploy"), []byte(tt.content), 0o755); err != nil {
				t.Fatal(err)
			}

			p, err := Detect(dir)
			if err != nil {
				t.Fatal(err)
			}
			if p.Stack != tt.want {
				t.Errorf("Stack = %q, want %q", p.Stack, tt.want)
			}
		})
	}
}

// TestIsShellFileIgnoresOtherExtensions guards the ext switch: a non-shell
// extension must never fall through to the shebang check, or a .py file with a
// `#!/bin/sh`-lookalike line would register as shell.
func TestIsShellFileIgnoresOtherExtensions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "script.py")
	if err := os.WriteFile(path, []byte("#!/bin/bash\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if isShellFile(path) {
		t.Errorf("%s should not be treated as a shell file", path)
	}
}

func TestDetectStandardgo(t *testing.T) {
	tests := []struct {
		name  string
		goMod string
		want  bool
	}{
		{"no tool directive", "module example.com/m\n\ngo 1.26\n", false},
		{
			"single-line tool directive",
			"module example.com/m\n\ngo 1.26\n\ntool github.com/amberpixels/standardgo/cmd/standardgo\n",
			true,
		},
		{
			"tool block",
			"module example.com/m\n\ngo 1.26\n\ntool (\n\tgolang.org/x/tools/cmd/stringer\n\tgithub.com/amberpixels/standardgo/cmd/standardgo\n)\n",
			true,
		},
		{
			"tool block without standardgo",
			"module example.com/m\n\ngo 1.26\n\ntool (\n\tgolang.org/x/tools/cmd/stringer\n)\n",
			false,
		},
		{
			"directive with trailing comment",
			"module example.com/m\n\ntool github.com/amberpixels/standardgo/cmd/standardgo // pinned\n",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(tt.goMod), 0o644); err != nil {
				t.Fatal(err)
			}

			p, err := Detect(dir)
			if err != nil {
				t.Fatal(err)
			}
			if p.Standardgo != tt.want {
				t.Errorf("Standardgo = %v, want %v", p.Standardgo, tt.want)
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
