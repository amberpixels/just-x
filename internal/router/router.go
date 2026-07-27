// Package router translates expressive recipe names (`:`, `!`, `?`) into the
// mappings real `just` understands, then execs `just`. It is the thin,
// must-never-break core ported 1:1 from the original bash wrapper.
package router

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// exitCommandNotFound is the conventional shell exit code for a missing command.
const exitCommandNotFound = 127

// Recipe-listing flags intercepted for reverse-translation.
const (
	flagList     = "--list"
	flagListAbbr = "-l"
	flagSummary  = "--summary"
)

// config holds the character→replacement mappings, overridable via env vars
// (kept compatible with the original bash tool).
type config struct {
	bang     string // ! replacement   (default: -x)
	question string // ? replacement   (default: -q)
	colon    string // : replacement   (default: --)
}

func loadConfig() config {
	return config{
		bang:     envOr("JUST_X_BANG", "-x"),
		question: envOr("JUST_X_QUESTION", "-q"),
		colon:    envOr("JUST_X_COLON", "--"),
	}
}

func envOr(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

// Run dispatches a non-meta invocation to `just`: it intercepts
// --list/-l/--summary for reverse-translation, otherwise translates the recipe
// name and execs real `just`. Returns a process exit code (the exec path
// replaces the process and does not return on success).
func Run(args []string) int {
	cfg := loadConfig()

	if isListInvocation(args) {
		return runList(args, cfg)
	}
	return execJust(buildArgs(args, cfg))
}

// isListInvocation reports whether any arg is a recipe-listing flag, matching
// the original wrapper's interception of --list/-l/--summary.
func isListInvocation(args []string) bool {
	for _, a := range args {
		switch a {
		case flagList, flagListAbbr, flagSummary:
			return true
		}
	}
	return false
}

// buildArgs produces the argument list to hand to real `just`:
//   - leading VAR=VALUE overrides pass through untouched;
//   - a plain recipe (no special chars) takes the fast path unchanged;
//   - otherwise the recipe name is translated, its own args preserved.
func buildArgs(args []string, cfg config) []string {
	// Collect leading VAR=VALUE overrides (only before the recipe name).
	i := 0
	for i < len(args) && isVarAssignment(args[i]) {
		i++
	}
	pre := args[:i]
	rest := args[i:]

	out := make([]string, 0, len(args))
	out = append(out, pre...)

	if len(rest) == 0 {
		return out
	}

	cmd := rest[0]
	if !strings.ContainsAny(cmd, "!?:") {
		// Plain recipe fast path: recipe + its args, unchanged.
		return append(out, rest...)
	}

	out = append(out, translate(cmd, cfg))
	return append(out, rest[1:]...)
}

// translate rewrites a recipe name's special characters into their configured
// replacements. Order (bang, question, colon) matches the original tool.
func translate(name string, cfg config) string {
	name = replaceAll(name, "!", cfg.bang)
	name = replaceAll(name, "?", cfg.question)
	name = replaceAll(name, ":", cfg.colon)
	return name
}

// isVarAssignment matches the bash glob [A-Za-z_]*=* : starts with a letter or
// underscore and contains an '='.
func isVarAssignment(s string) bool {
	if s == "" {
		return false
	}
	c := s[0]
	letter := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '_'
	return letter && strings.Contains(s, "=")
}

// replaceAll is strings.ReplaceAll guarded against an empty pattern (which would
// otherwise splice the replacement between every character).
func replaceAll(s, pat, repl string) string {
	if pat == "" {
		return s
	}
	return strings.ReplaceAll(s, pat, repl)
}

// execJust replaces the current process with real `just` so exit codes,
// signals and the TTY pass through unchanged.
func execJust(args []string) int {
	path, err := exec.LookPath("just")
	if err != nil {
		fmt.Fprintln(os.Stderr, "justx: `just` not found on PATH")
		return exitCommandNotFound
	}
	argv := append([]string{"just"}, args...)
	// On success this does not return.
	err = syscall.Exec(path, argv, os.Environ())
	fmt.Fprintf(os.Stderr, "justx: exec just: %v\n", err)
	return 1
}

// runList runs `just --color always <args>` and reverse-translates the recipe
// names in the output (mapped → expressive), re-padding comment alignment.
func runList(args []string, cfg config) int {
	path, err := exec.LookPath("just")
	if err != nil {
		fmt.Fprintln(os.Stderr, "justx: `just` not found on PATH")
		return exitCommandNotFound
	}

	cmdArgs := append([]string{"--color", "always"}, args...)
	cmd := exec.CommandContext(context.Background(), path, cmdArgs...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	runErr := cmd.Run()

	fmt.Fprint(os.Stdout, reverseTranslateList(stdout.String(), cfg))

	if exitErr, ok := errors.AsType[*exec.ExitError](runErr); ok {
		return exitErr.ExitCode()
	}
	if runErr != nil {
		return 1
	}
	return 0
}
