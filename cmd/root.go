package cmd

import (
	"os"

	coreapp "github.com/berachain/offchain-sdk/core/app"
	"github.com/spf13/cobra"
)

// BuildBasicRootCmd builds a root command.
func BuildBasicRootCmd[C any](app coreapp.App[C]) *cobra.Command {
	rootCmd := BuildRootCommand(
		app.Name(),
		"Welcome to "+app.Name(),
		cobra.NoArgs,
	)

	rootCmd.AddCommand(
		StartCmd(app, os.Getenv("HOME")),
	)

	return rootCmd
}

// BuildRootCommand builds the root command.
func BuildRootCommand(name, short string, args cobra.PositionalArgs) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   name,
		Short: short,
		Args:  args,
		Run: func(cmd *cobra.Command, _ []string) {
			if err := cmd.Help(); err != nil {
				os.Exit(1)
			}
			os.Exit(0)
		},
	}
	return rootCmd
}
