package cmd

import (
	"context"
	"os"

	"github.com/berachain/offchain-sdk/log"
	"github.com/spf13/cobra"
)

type App interface {
	Name() string
	Setup(ab AppBuilder, logger log.Logger)
	Start(context.Context) error
}

// BuildBasicRootCmd builds a root command.
func BuildBasicRootCmd(app App) *cobra.Command {
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
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				os.Exit(1)
			}
			os.Exit(0)
		},
	}
	return rootCmd
}
