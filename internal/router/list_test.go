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
	// Recipe wrapped in color codes, comment preceded by an ANSI escape.
	in := "    \x1b[36mapp--build\x1b[0m   \x1b[34m# desc\x1b[0m"
	want := "    \x1b[36mapp:build\x1b[0m    \x1b[34m# desc\x1b[0m"
	if got := reverseLine(in, defaultCfg); got != want {
		t.Errorf("reverseLine(colored)\n  = %q\nwant %q", got, want)
	}
}

func TestReverseTranslateListPreservesStructure(t *testing.T) {
	in := "Available recipes:\n    app--build    # build\n    dev-x         # dev\n"
	want := "Available recipes:\n    app:build     # build\n    dev!          # dev\n"
	if got := reverseTranslateList(in, defaultCfg); got != want {
		t.Errorf("reverseTranslateList\n  = %q\nwant %q", got, want)
	}
}
