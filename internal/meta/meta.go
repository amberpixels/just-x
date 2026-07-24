// Package meta dispatches `@`-sigil meta-commands — handled by justx, never
// passed to `just`. v1 ships `@init`; `@help` and `@version` are cheap built-ins.
// The meta surface is built on urfave/cli/v3 so each command (and future ones
// like @add/@upgrade/@doctor) gets its own flags and help for free.
package meta

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

// Version is the justx version.
const Version = "0.1.0"

// Run handles an `@`-sigil meta invocation. args[0] is the meta-command
// including its leading `@` (the sigil split happens in main). Returns a
// process exit code.
func Run(args []string) int {
	cmd := newRootCommand()
	// urfave expects argv[0] to be the program name.
	if err := cmd.Run(context.Background(), append([]string{"justx"}, args...)); err != nil {
		fmt.Fprintf(os.Stderr, "justx: %v\n", err)
		return 1
	}
	return 0
}

func newRootCommand() *cli.Command {
	root := &cli.Command{
		Name:    "justx",
		Usage:   "the justfile companion",
		Version: Version,
		// We run our own `@`-sigil dispatch, so suppress urfave's auto help/version
		// commands and flags — the surface is the `@`-prefixed commands below.
		HideHelp:    true,
		HideVersion: true,
		Commands: []*cli.Command{
			{
				Name:  "@init",
				Usage: "detect project + scaffold a justfile",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "skip the interactive form; use the pre-ticked modules",
					},
				},
				Action: func(_ context.Context, c *cli.Command) error {
					return runInit(c.Bool("yes"))
				},
			},
			{
				Name:  "@help",
				Usage: "show this help",
				Action: func(_ context.Context, c *cli.Command) error {
					return cli.ShowRootCommandHelp(c)
				},
			},
			{
				Name:  "@version",
				Usage: "print version",
				Action: func(_ context.Context, _ *cli.Command) error {
					fmt.Fprintf(os.Stdout, "justx %s\n", Version)
					return nil
				},
			},
		},
		// urfave routes an unmatched `@command` here as a stray arg (this fires,
		// not CommandNotFound, which only triggers via the help path). A stray
		// arg is an unknown meta-command; otherwise just show help.
		Action: func(_ context.Context, c *cli.Command) error {
			if name := c.Args().First(); name != "" {
				fmt.Fprintf(os.Stderr, "justx: unknown meta-command %q\n\n", name)
				_ = cli.ShowRootCommandHelp(c)
				return cli.Exit("", 2)
			}
			return cli.ShowRootCommandHelp(c)
		},
	}
	return root
}
