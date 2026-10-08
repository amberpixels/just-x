package router

import "testing"

func TestReverseLineNoColor(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			// colon: app--build → app:build; one char shorter, so +1 space before #.
			"colon recipe realigns comment",
			"    app--build target    # builds the app",
			"    app:build target     # builds the app",
		},
		{
			"bang recipe realigns comment",
			"    greet-x target    # greet",
			"    greet! target     # greet",
		},
		{
			"question recipe realigns comment",
			"    ready-q    # check",
			"    ready?     # check",
		},
		{
			"no comment, no padding",
			"    app--build target",
			"    app:build target",
		},
		{
			"plain recipe untouched",
			"    build    # regular build",
			"    build    # regular build",
		},
		{
			// two colons → two chars shorter → +2 spaces.
			"multiple replacements accumulate padding",
			"    a--b--c    # x",
			"    a:b:c      # x",
		},
		{
			"colon in comment untouched",
			"    run *args    # just run --json",
			"    run *args    # just run --json",
		},
		{
			"bang and question in comment untouched",
			"    search    # grep -q and xargs -x",
			"    search    # grep -q and xargs -x",
		},
		{
			// the -- sits left of the comment but outside the name: no rewrite, no padding.
			"colon in default value untouched",
			"    fetch flags=\"--depth 1\" # x",
			"    fetch flags=\"--depth 1\" # x",
		},
		{
			"mapped name, comment with mapping",
			"    app--build    # just run --json",
			"    app:build     # just run --json",
		},
		{
			"nested module recipe",
			"        sub--r # sub recipe --x",
			"        sub:r  # sub recipe --x",
		},
		{
			"group heading untouched",
			"    [my--grp]",
			"    [my--grp]",
		},
		{
			"header untouched",
			"Available recipes:",
			"Available recipes:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reverseLine(tt.in, defaultCfg); got != tt.want {
				t.Errorf("reverseLine(%q)\n  = %q\nwant %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestReverseLineColored(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			// Recipe wrapped in color codes, comment preceded by an ANSI escape.
			"colored name realigns comment",
			"    \x1b[36mapp--build\x1b[0m   \x1b[34m# desc\x1b[0m",
			"    \x1b[36mapp:build\x1b[0m    \x1b[34m# desc\x1b[0m",
		},
		{
			"colored comment untouched",
			"    run \x1b[35m*\x1b[0m\x1b[36margs\x1b[0m   \x1b[34m#\x1b[0m \x1b[34mrun --json -q -x\x1b[0m",
			"    run \x1b[35m*\x1b[0m\x1b[36margs\x1b[0m   \x1b[34m#\x1b[0m \x1b[34mrun --json -q -x\x1b[0m",
		},
		{
			"colored default value untouched",
			"    fetch \x1b[36mflags\x1b[0m=\x1b[32m\"--depth 1\"\x1b[0m \x1b[34m#\x1b[0m \x1b[34mx\x1b[0m",
			"    fetch \x1b[36mflags\x1b[0m=\x1b[32m\"--depth 1\"\x1b[0m \x1b[34m#\x1b[0m \x1b[34mx\x1b[0m",
		},
		{
			"colored mapped name, comment with mapping",
			"    app--build   \x1b[34m#\x1b[0m \x1b[34mrun --json\x1b[0m",
			"    app:build    \x1b[34m#\x1b[0m \x1b[34mrun --json\x1b[0m",
		},
		{
			"colored group heading untouched",
			"    \x1b[1;33m[my--grp]\x1b[0m",
			"    \x1b[1;33m[my--grp]\x1b[0m",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reverseLine(tt.in, defaultCfg); got != tt.want {
				t.Errorf("reverseLine(%q)\n  = %q\nwant %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestReverseTranslateSummary(t *testing.T) {
	in := "app--build fetch g--one ready-q sub::sub--r\n"
	want := "app:build fetch g:one ready? sub::sub:r\n"
	if got := reverseTranslateSummary(in, defaultCfg); got != want {
		t.Errorf("reverseTranslateSummary\n  = %q\nwant %q", got, want)
	}
}

func TestReverseTranslateListPreservesStructure(t *testing.T) {
	in := "Available recipes:\n    app--build    # build\n    dev-x         # dev\n"
	want := "Available recipes:\n    app:build     # build\n    dev!          # dev\n"
	if got := reverseTranslateList(in, defaultCfg); got != want {
		t.Errorf("reverseTranslateList\n  = %q\nwant %q", got, want)
	}
}
