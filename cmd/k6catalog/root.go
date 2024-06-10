package main

import (
	"github.com/spf13/cobra"
)

// newCmd returns a cobra.Command for k6build command
func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "k6catalg",
		Short: "k6 dependency catalog tool",
		// prevent the usage help to printed to stderr when an error is reported by a subcommand
		SilenceUsage: true,
		// this is needed to prevent cobra to print errors reported by subcommands in the stderr
		SilenceErrors: true,
	}

	return cmd
}
