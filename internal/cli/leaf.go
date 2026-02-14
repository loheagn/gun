package cli

import (
	"github.com/spf13/cobra"

	"github.com/loheagn/gun/internal/locator"
)

func newLeafCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "leaf <file> <line> | <file>:<line>",
		Short: "Run the nearest leaf subtest; fallback to test when no t.Run exists",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMode(cmd, args, locator.ModeLeaf, locator.ResolveOptions{})
		},
	}
}
