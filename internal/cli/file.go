package cli

import (
	"github.com/spf13/cobra"

	"github.com/loheagn/gun/internal/locator"
)

func newFileCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "file <file> <line> | <file>:<line>",
		Short: "Run all top-level tests in the given file",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMode(cmd, args, locator.ModeFile, locator.ResolveOptions{})
		},
	}
}
