package cli

import (
	"github.com/spf13/cobra"

	"github.com/loheagn/gun/internal/locator"
)

func newPkgCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "pkg <file> <line> | <file>:<line>",
		Short: "Run all tests in the package containing the file",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMode(cmd, args, locator.ModePkg, locator.ResolveOptions{})
		},
	}
}
