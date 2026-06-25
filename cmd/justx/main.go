// Command justx is the justfile companion: an expressive-name router for `just`
// plus meta-commands (the `@`-sigil namespace) for scaffolding and upgrading justfiles.
package main

import (
	"os"

	"github.com/amberpixels/just-x/internal/meta"
	"github.com/amberpixels/just-x/internal/router"
)

func main() {
	args := os.Args[1:]

	// Meta-commands live under the `@` sigil and are never passed to `just`.
	// If argv[1] starts with `@`, justx handles it itself.
	if len(args) > 0 && len(args[0]) > 0 && args[0][0] == '@' {
		os.Exit(meta.Run(args))
	}

	// Everything else is the router: translate expressive recipe names and exec `just`.
	os.Exit(router.Run(args))
}
