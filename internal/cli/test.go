package cli

import (
	"github.com/spf13/cobra"

	"github.com/loheagn/gun/internal/locator"
)

func newTestCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "test <file> <line> | <file>:<line>",
		Short: "Run the containing top-level TestXxx",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMode(cmd, args, locator.ModeTest, locator.ResolveOptions{})
		},
	}
}
