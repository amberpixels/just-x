package router

import (
	"reflect"
	"testing"
)

var defaultCfg = config{bang: "-x", question: "-q", colon: "--"}

func TestTranslate(t *testing.T) {
	tests := []struct {
		name string
		in   string
		cfg  config
		want string
	}{
		{"bang", "dev!", defaultCfg, "dev-x"},
		{"question", "ready?", defaultCfg, "ready-q"},
		{"colon", "app:build", defaultCfg, "app--build"},
		{"plain unchanged", "test", defaultCfg, "test"},
		{"multiple colons", "a:b:c", defaultCfg, "a--b--c"},
		{"mixed", "app:dev!", defaultCfg, "app--dev-x"},
		{"custom bang", "dev!", config{bang: "-bang", question: "-q", colon: "--"}, "dev-bang"},
		{"custom colon", "app:build", config{bang: "-x", question: "-q", colon: "__"}, "app__build"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := translate(tt.in, tt.cfg); got != tt.want {
				t.Errorf("translate(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsVarAssignment(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"my_var=hello", true},
		{"FOO=bar", true},
		{"_x=1", true},
		{"a=", true},
		{"plain", false},
		{"foo=bar=baz", true},
		{"=leading", false},
		{"1abc=x", false}, // must start with letter or underscore
		{"", false},
		{"app:build", false},
	}
	for _, tt := range tests {
		if got := isVarAssignment(tt.in); got != tt.want {
			t.Errorf("isVarAssignment(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{"empty", nil, []string{}},
		{"plain recipe fast path", []string{"test"}, []string{"test"}},
		{"plain recipe with args", []string{"plain", "mytarget"}, []string{"plain", "mytarget"}},
		{"translate bang recipe", []string{"greet!", "mytarget"}, []string{"greet-x", "mytarget"}},
		{"translate colon recipe", []string{"app:build"}, []string{"app--build"}},
		{
			"var override + translated recipe + args",
			[]string{"my_var=hello", "greet!", "mytarget"},
			[]string{"my_var=hello", "greet-x", "mytarget"},
		},
		{
			"var override + plain recipe + args",
			[]string{"my_var=hello", "plain", "mytarget"},
			[]string{"my_var=hello", "plain", "mytarget"},
		},
		{
			// A recipe arg that looks like VAR=VALUE must not be consumed as an override.
			"trailing VAR=VALUE-looking arg stays an arg",
			[]string{"plain", "foo=bar"},
			[]string{"plain", "foo=bar"},
		},
		{
			"multiple leading overrides",
			[]string{"A=1", "B=2", "app:build", "x"},
			[]string{"A=1", "B=2", "app--build", "x"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildArgs(tt.in, defaultCfg)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildArgs(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsListInvocation(t *testing.T) {
	yes := [][]string{{"--list"}, {"-l"}, {"--summary"}, {"app:build", "--list"}}
	no := [][]string{{}, {"build"}, {"app:test"}, {"--listfoo"}}
	for _, a := range yes {
		if !isListInvocation(a) {
			t.Errorf("isListInvocation(%q) = false, want true", a)
		}
	}
	for _, a := range no {
		if isListInvocation(a) {
			t.Errorf("isListInvocation(%q) = true, want false", a)
		}
	}
}
