package router

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Reverse-translation of `just --list` output: recipe names are shown in their
// mapped form (e.g. app--build), so we rewrite them back to the expressive form
// (app:build) for display and re-pad comment alignment, since the expressive
// names are shorter than the mapped ones. Only the recipe name is rewritten:
// parameters, defaults and doc comments print exactly as `just` prints them.

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

// reverseTranslateSummary processes `just --summary` output, which holds
// nothing but space-separated recipe names, so every word is a name.
func reverseTranslateSummary(input string, cfg config) string {
	return reverseName(input, cfg)
}

// reverseLine rewrites the recipe name on one output line (mapped →
// expressive), then adds compensating spaces before the comment marker so
// columns stay aligned. The name is the first word after the indent; lines
// without an indent (the header) and `[group]` headings are left alone.
func reverseLine(line string, cfg config) string {
	body := strings.TrimLeft(line, " ")
	indent := line[:len(line)-len(body)]
	if indent == "" || strings.HasPrefix(ansiRe.ReplaceAllString(body, ""), "[") {
		return line
	}

	end := strings.IndexByte(body, ' ')
	if end < 0 {
		end = len(body)
	}
	name, rest := body[:end], body[end:]

	mapped := reverseName(name, cfg)
	extra := visibleWidth(name) - visibleWidth(mapped)
	if extra > 0 {
		rest = padComment(rest, extra)
	}

	return indent + mapped + rest
}

// reverseName undoes translate: configured replacements back to ! ? and :.
func reverseName(name string, cfg config) string {
	name = replaceAll(name, cfg.bang, "!")
	name = replaceAll(name, cfg.question, "?")
	name = replaceAll(name, cfg.colon, ":")
	return name
}

// padComment inserts n spaces in front of the comment marker in s, if any.
func padComment(s string, n int) string {
	pad := strings.Repeat(" ", n)
	if loc := ansiHashRe.FindStringIndex(s); loc != nil {
		return s[:loc[0]] + pad + s[loc[0]:]
	}
	if loc := commentRe.FindStringIndex(s); loc != nil {
		return s[:loc[0]] + pad + s[loc[0]:]
	}
	return s
}

// visibleWidth counts the characters of s that reach the terminal.
func visibleWidth(s string) int {
	return utf8.RuneCountInString(ansiRe.ReplaceAllString(s, ""))
}
