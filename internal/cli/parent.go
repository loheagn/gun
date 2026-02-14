package cli

import (
	"github.com/spf13/cobra"

	"github.com/loheagn/gun/internal/errs"
	"github.com/loheagn/gun/internal/locator"
)

func newParentCommand() *cobra.Command {
	var up int
	cmd := &cobra.Command{
		Use:   "parent <file> <line> | <file>:<line>",
		Short: "Run parent scope of the nearest node",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if up < 1 {
				return errs.New(errs.CodeUsage, "--up must be >= 1", nil)
			}
			return runMode(cmd, args, locator.ModeParent, locator.ResolveOptions{ParentUp: up})
		},
	}
	cmd.Flags().IntVar(&up, "up", 1, "ancestor levels to move up")
	return cmd
}
