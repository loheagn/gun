package cli

import (
	"github.com/spf13/cobra"

	"github.com/loheagn/gun/internal/errs"
	"github.com/loheagn/gun/internal/input"
	"github.com/loheagn/gun/internal/locator"
	"github.com/loheagn/gun/internal/runner"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "gun",
		Short:         "Run nearby Go tests by file and line",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, passthrough, err := targetAndPassthrough(cmd, args)
			if err != nil {
				return err
			}
			res, err := locator.Resolve(locator.ModeAuto, target.File, target.Line, locator.ResolveOptions{})
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

	cmd.AddCommand(
		newLeafCommand(),
		newParentCommand(),
		newTestCommand(),
		newFileCommand(),
		newPkgCommand(),
		newProjectCommand(),
	)
	return cmd
}

func ExitCode(err error) int {
	return errs.ExitCode(err)
}

func splitArgs(cmd *cobra.Command, args []string) ([]string, []string) {
	idx := cmd.ArgsLenAtDash()
	if idx < 0 {
		return args, nil
	}
	return args[:idx], args[idx:]
}

func consumeTargetPrefix(args []string) (input.Target, []string, error) {
	var firstErr error
	if len(args) >= 2 {
		target, err := input.ParseFileLine(args[:2])
		if err == nil {
			return target, args[2:], nil
		}
		firstErr = err
	}
	if len(args) >= 1 {
		target, err := input.ParseFileLine(args[:1])
		if err == nil {
			return target, args[1:], nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	if firstErr != nil {
		return input.Target{}, nil, firstErr
	}
	return input.Target{}, nil, errs.New(errs.CodeUsage, "expected <file> <line> or <file>:<line>", nil)
}

func targetAndPassthrough(cmd *cobra.Command, args []string) (input.Target, []string, error) {
	positional, passthrough := splitArgs(cmd, args)
	if cmd.ArgsLenAtDash() >= 0 {
		target, err := input.ParseFileLine(positional)
		if err != nil {
			return input.Target{}, nil, err
		}
		return target, passthrough, nil
	}
	target, rest, err := consumeTargetPrefix(positional)
	if err != nil {
		return input.Target{}, nil, err
	}
	return target, append(rest, passthrough...), nil
}

func runMode(cmd *cobra.Command, args []string, mode locator.Mode, opts locator.ResolveOptions) error {
	target, passthrough, err := targetAndPassthrough(cmd, args)
	if err != nil {
		return err
	}
	res, err := locator.Resolve(mode, target.File, target.Line, opts)
	if err != nil {
		return err
	}
	inv, err := runner.BuildInvocation(res, passthrough)
	if err != nil {
		return err
	}
	return runner.Run(inv)
}
