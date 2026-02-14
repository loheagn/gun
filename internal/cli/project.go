package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/loheagn/gun/internal/input"
	"github.com/loheagn/gun/internal/locator"
	"github.com/loheagn/gun/internal/runner"
)

func newProjectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "project <file> <line> [project-root] | <file>:<line> [project-root]",
		Short: "Run tests for the whole project",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			positional, passthrough := splitArgs(cmd, args)
			target, root, inferredPassthrough, err := parseProjectTarget(positional, cmd.ArgsLenAtDash() >= 0)
			if err != nil {
				return err
			}
			passthrough = append(inferredPassthrough, passthrough...)
			res, err := locator.Resolve(locator.ModeProject, target.File, target.Line, locator.ResolveOptions{ProjectRoot: root})
			if err != nil {
				return err
			}
			inv, err := runner.BuildInvocation(res, passthrough)
			if err != nil {
				return err
			}
			return runner.Run(inv)
		},
	}
}

func parseProjectTarget(args []string, strict bool) (input.Target, string, []string, error) {
	if strict {
		target, root, err := input.ParseFileLineWithOptionalRoot(args)
		if err != nil {
			return input.Target{}, "", nil, err
		}
		return target, root, nil, nil
	}
	target, rest, err := consumeTargetPrefix(args)
	if err != nil {
		return input.Target{}, "", nil, err
	}
	root := ""
	if len(rest) > 0 && !strings.HasPrefix(rest[0], "-") {
		root = rest[0]
		rest = rest[1:]
	}
	return target, root, rest, nil
}
