package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/berachain/offchain-sdk/baseapp"
	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/cmd/flags"
	"github.com/berachain/offchain-sdk/log"
	"github.com/spf13/cobra"
)

// StartCmdOptions defines options that can be customized in `StartCmdWithOptions`,.
type StartCmdOptions struct{}

// StartCmd runs the application passed in.
func StartCmd(app App, defaultAppHome string) *cobra.Command {
	return StartCmdWithOptions(app, defaultAppHome, StartCmdOptions{})
}

// StartCmdWithOptions runs the service passed in.
func StartCmdWithOptions(app App, defaultAppHome string, _ StartCmdOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the service",
		Long:  `Run the service`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create a context that will be cancelled when the user presses Ctrl+C
			// (process receives termination signal).
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			configPath, err := cmd.Flags().GetString(flags.ConfigPath)
			if err != nil {
				return err
			}

			if configPath == "" {
				configPath = defaultAppHome
			}

			ab := &baseapp.AppBuilder{}

			// TODO MOVE / handle all generic configs.
			ethConfig := eth.LoadConfig(configPath)
			ethClient := eth.NewClient(&ethConfig)
			ab.RegisterEthClient(ethClient)

			// Build the application, then start it.
			app.Setup(ab, log.NewBlankLogger(cmd.OutOrStdout()))
			if err = app.Start(ctx); err != nil {
				return err
			}

			// Wait for the context to be done.
			<-ctx.Done()
			// TODO: should we return error here based on ctx.Err()?
			return nil
		},
	}

	cmd.Flags().String(flags.ConfigPath, flags.DefaultConfigPath, "The config directory")
	return cmd
}
