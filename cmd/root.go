package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// BuildBasicRootCmd builds a root command.
func BuildBasicRootCmd(ab AppBuilder) *cobra.Command {
	rootCmd := BuildRootCommand(
		ab.AppName(),
		"Welcome to "+ab.AppName(),
		cobra.NoArgs,
	)

	rootCmd.AddCommand(
		StartCmd(ab, os.Getenv("HOME")),
	)

	return rootCmd
}

// BuildRootCommand builds the root command.
func BuildRootCommand(name, short string, args cobra.PositionalArgs) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   name,
		Short: short,
		Args:  args,
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				os.Exit(1)
			}
			os.Exit(0)
		},
	}
	return rootCmd
}
