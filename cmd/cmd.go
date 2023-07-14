package cmd

import (
	"os"
	"os/signal"

	baseapp "github.com/berachain/offchain-sdk/baseapp"
	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/log"
	"github.com/spf13/cobra"
)

// AppBuilder is a builder for an app. It follows a basic factory pattern.
type AppBuilder interface {
	AppName() string
	BuildApp(log.Logger, *eth.Config) *baseapp.BaseApp
}

// BuildBasicRootCmd builds a root command.
func BuildBasicRootCmd(ab AppBuilder) *cobra.Command {
	rootCmd := BuildRootCommand(
		"cron",
		"cron does crons",
		cobra.NoArgs,
	)

	rootCmd.AddCommand(
		BuildStartCommand(ab, cobra.ExactArgs(0)),
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

// BuildStartCommand builds the start command.
func BuildStartCommand(appBuilder AppBuilder, args cobra.PositionalArgs) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Starts " + appBuilder.AppName(),
		Args:  args,
		Run: func(cmd *cobra.Command, args []string) {

			// Setup channel to manage shutdown signal from the os.
			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, os.Interrupt)

			// Create a logger
			logger := log.NewBlankLogger(os.Stdout)

			// Load the eth config
			ethConfig := eth.LoadConfig("")

			// Build the baseapp
			app := appBuilder.BuildApp(logger, &ethConfig)

			// Start the app
			app.Start()

			// Wait for a signal to shutdown the app
			for range signalChan {
				app.Stop()
				logger.Info(appBuilder.AppName() + " stopped gracefully")
				return
			}
		},
	}
}
