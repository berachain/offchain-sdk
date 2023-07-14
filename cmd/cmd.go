package cmd

import (
	"os"
	"os/signal"

	baseapp "github.com/berachain/offchain-sdk/baseapp"
	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/log"
	"github.com/spf13/cobra"
)

// BuildRootCommand builds the root command.
func BuildRootCommand(name, short string, args cobra.PositionalArgs) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   name,
		Short: short,
		Args:  args,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}
	return rootCmd
}

// BuildStartCommand builds the start command.
func BuildStartCommand(appname string, args cobra.PositionalArgs) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Starts " + appname,
		Args:  args,
		Run: func(cmd *cobra.Command, args []string) {
			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, os.Interrupt)
			logger := log.NewBlankLogger(os.Stdout)
			ethConfig := eth.LoadConfig("")
			baseapp := baseapp.New(appname, logger, &ethConfig)
			baseapp.Start()

			// Wait for a signal to stop
			for range signalChan {
				baseapp.Stop()
				return
			}
		},
	}
}
