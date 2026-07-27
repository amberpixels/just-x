// Package detect inspects a directory tree for project signals (go.mod,
// package.json, …) and reports the stack so @init can pre-tick the right
// modules.
package detect

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// Stack identifies a detected project type.
type Stack string

const (
	StackUnknown Stack = ""
	StackGo      Stack = "go"
	StackNode    Stack = "node"
	StackShell   Stack = "shell"
)

// Project is the result of detection: the stack plus useful context for
// module composition (project root, package manager, …).
type Project struct {
	Root           string // directory where the project signal was found
	Stack          Stack
	PackageManager string // for Node: pnpm | yarn | bun | npm
	Standardgo     bool   // for Go: standardgo is pinned as a tool dependency
}

// Detect walks up from dir looking for project signals and returns the first
// match. If nothing is recognized it returns a Project with StackUnknown rooted
// at dir. When multiple stacks are present in one directory, Go wins (a
// deliberate v1 simplification; revisit if it bites), and shell loses to
// everything — see inspect.
func Detect(dir string) (Project, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return Project{}, err
	}

	cur := abs
	for {
		if p, ok := inspect(cur); ok {
			return p, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break // reached filesystem root
		}
		cur = parent
	}

	return Project{Root: abs, Stack: StackUnknown}, nil
}

// inspect checks a single directory for a known project signal.
//
// Shell is checked last, and deliberately so: it is the one stack with no
// manifest to key off, so its signal is "there are shell scripts here", which a
// Go or Node repo shipping a couple of helper scripts also satisfies. Testing it
// after the manifest stacks keeps those repos on their real stack.
func inspect(dir string) (Project, bool) {
	if goMod := filepath.Join(dir, "go.mod"); fileExists(goMod) {
		return Project{Root: dir, Stack: StackGo, Standardgo: usesStandardgo(goMod)}, true
	}
	if fileExists(filepath.Join(dir, "package.json")) {
		return Project{Root: dir, Stack: StackNode, PackageManager: detectNodePM(dir)}, true
	}
	if hasShellScripts(dir) {
		return Project{Root: dir, Stack: StackShell}, true
	}
	return Project{}, false
}

// shellScriptDirs are the conventional homes for a shell project's scripts,
// relative to the project root. The root itself is included so a repo that keeps
// its scripts at the top level is still recognized.
var shellScriptDirs = []string{".", "bin", "scripts"}

// hasShellScripts reports whether dir holds at least one shell script, in the
// directory itself or in a conventional script subdirectory. The search is
// deliberately shallow: it answers "is this a shell project", not "find every
// script" — that is `shfmt -f`'s job at recipe run time.
func hasShellScripts(dir string) bool {
	for _, sub := range shellScriptDirs {
		entries, err := os.ReadDir(filepath.Join(dir, sub))
		if err != nil {
			continue // absent or unreadable is simply "no scripts here"
		}
		for _, e := range entries {
			if !e.IsDir() && isShellFile(filepath.Join(dir, sub, e.Name())) {
				return true
			}
		}
	}
	return false
}

// isShellFile reports whether a path is a shell script, by extension or, for
// extensionless files, by shebang. The shebang case is what catches scripts
// meant to be run as commands, which are conventionally named `bin/deploy`
// rather than `bin/deploy.sh`.
func isShellFile(path string) bool {
	switch filepath.Ext(path) {
	case ".sh", ".bash":
		return true
	case "":
		return hasShellShebang(path)
	default:
		return false
	}
}

// shellShebangs are the interpreters worth linting with shellcheck.
var shellShebangs = []string{"sh", "bash", "dash", "ksh", "zsh"}

// shebangPeek is how many bytes to read looking for the shebang line. A shebang
// is the first line or it is not a shebang, and no plausible one is this long.
const shebangPeek = 128

// hasShellShebang reports whether a file starts with a shell shebang. Every
// field of the line is tested, not just the last, so `#!/bin/sh`,
// `#!/usr/bin/env bash` and `#!/usr/bin/env -S bash -e` all match.
func hasShellShebang(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, shebangPeek)
	n, _ := f.Read(buf)
	line, _, _ := strings.Cut(string(buf[:n]), "\n")
	if !strings.HasPrefix(line, "#!") {
		return false
	}
	for field := range strings.FieldsSeq(line) {
		if slices.Contains(shellShebangs, filepath.Base(field)) {
			return true
		}
	}
	return false
}

// Node package managers.
const (
	pmPnpm = "pnpm"
	pmYarn = "yarn"
	pmBun  = "bun"
	pmNpm  = "npm"
)

// detectNodePM infers the Node package manager from the lockfile present,
// defaulting to npm.
func detectNodePM(dir string) string {
	switch {
	case fileExists(filepath.Join(dir, "pnpm-lock.yaml")):
		return pmPnpm
	case fileExists(filepath.Join(dir, "yarn.lock")):
		return pmYarn
	case fileExists(filepath.Join(dir, "bun.lockb")), fileExists(filepath.Join(dir, "bun.lock")):
		return pmBun
	default:
		return pmNpm
	}
}

// standardgoTool is the tool directive a project gains from
// `go get -tool github.com/amberpixels/standardgo/cmd/standardgo`.
const standardgoTool = "github.com/amberpixels/standardgo/cmd/standardgo"

// usesStandardgo reports whether go.mod pins standardgo as a tool dependency.
//
// The tool directive is the signal rather than a binary on PATH: standardgo is
// deliberately not installed globally, and its whole point is that the version
// is pinned per project.
//
// Handles both directive shapes go.mod uses: the single-line `tool <path>` and
// the block `tool ( ... )` that appears once a module has several tools.
func usesStandardgo(goModPath string) bool {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return false
	}

	inBlock := false
	for line := range strings.Lines(string(data)) {
		line, _, _ = strings.Cut(line, "//")
		fields := strings.Fields(line)
		switch {
		case inBlock:
			if len(fields) == 1 && fields[0] == ")" {
				inBlock = false
			} else if len(fields) == 1 && fields[0] == standardgoTool {
				return true
			}
		case len(fields) == 2 && fields[0] == "tool":
			switch fields[1] {
			case "(":
				inBlock = true
			case standardgoTool:
				return true
			}
		}
	}

	return false
}

// OnPath reports whether an executable is available on PATH. Used by modules to
// tell which referenced tools are already installed.
func OnPath(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
