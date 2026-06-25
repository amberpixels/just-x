package router

import (
	"regexp"
	"strings"
)

// Reverse-translation of `just --list` output: recipe names are shown in their
// mapped form (e.g. app--build), so we rewrite them back to the expressive form
// (app:build) for display and re-pad comment alignment, since the expressive
// names are shorter than the mapped ones. Ported 1:1 from the bash awk script.

var (
	ansiRe     = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	commentRe  = regexp.MustCompile(` +# `)
	ansiHashRe = regexp.MustCompile(`\x1b\[[0-9;]*m#`)
)

// reverseTranslateList processes each line of `just --list` output.
func reverseTranslateList(input string, cfg config) string {
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		lines[i] = reverseLine(line, cfg)
	}
	return strings.Join(lines, "\n")
}

// reverseLine rewrites one output line: mapped → expressive names, then adds
// compensating spaces before the comment marker so columns stay aligned.
func reverseLine(line string, cfg config) string {
	// Strip ANSI to locate the comment and count replacements.
	clean := ansiRe.ReplaceAllString(line, "")

	extra := 0
	if loc := commentRe.FindStringIndex(clean); loc != nil {
		leftClean := clean[:loc[0]]
		extra += countOf(leftClean, cfg.colon) * (len(cfg.colon) - 1)
		extra += countOf(leftClean, cfg.bang) * (len(cfg.bang) - 1)
		extra += countOf(leftClean, cfg.question) * (len(cfg.question) - 1)
	}

	// Reverse the mappings on the full (still-colored) line.
	line = replaceAll(line, cfg.bang, "!")
	line = replaceAll(line, cfg.question, "?")
	line = replaceAll(line, cfg.colon, ":")

	// Re-pad: insert the lost width back in front of the comment marker.
	if extra > 0 {
		pad := strings.Repeat(" ", extra)
		if loc := ansiHashRe.FindStringIndex(line); loc != nil {
			line = line[:loc[0]] + pad + line[loc[0]:]
		} else if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i] + pad + line[i:]
		}
	}

	return line
}

// countOf counts non-overlapping occurrences of pat in s (0 for an empty pat).
func countOf(s, pat string) int {
	if pat == "" {
		return 0
	}
	return strings.Count(s, pat)
}
