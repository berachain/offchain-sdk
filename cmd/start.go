package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/berachain/offchain-sdk/baseapp"
	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/cmd/flags"
	"github.com/berachain/offchain-sdk/config"
	"github.com/berachain/offchain-sdk/config/toml"
	coreapp "github.com/berachain/offchain-sdk/core/app"
	"github.com/berachain/offchain-sdk/log"
	"github.com/berachain/offchain-sdk/server"
	"github.com/spf13/cobra"
)

// StartCmdOptions defines options that can be customized in `StartCmdWithOptions`,.
type StartCmdOptions struct{}

// StartCmd runs the application passed in.
func StartCmd[C any](app coreapp.App[C], defaultAppHome string) *cobra.Command {
	return StartCmdWithOptions[C](app, defaultAppHome, StartCmdOptions{})
}

// StartCmdWithOptions runs the service passed in.
func StartCmdWithOptions[C any](
	app coreapp.App[C], defaultAppHome string, _ StartCmdOptions,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the service",
		Long:  `Run the service`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create a context that will be cancelled when the user presses Ctrl+C
			// (process receives termination signal).
			logger := log.NewBlankLogger(cmd.OutOrStdout())
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)

			configPath, err := cmd.Flags().GetString(flags.ConfigPath)
			if err != nil {
				return err
			}

			if configPath == "" {
				configPath = defaultAppHome
			}

			var cfg config.Config[C]
			// Check if we should override the config with environment variables.
			envOverride, err := cmd.Flags().GetBool(flags.EnvOverride)
			if err != nil {
				return err
			}

			envOverridePrefix, err := cmd.Flags().GetString(flags.EnvOverridePrefix)
			if err != nil {
				return err
			}
			if err = toml.LoadConfig[config.Config[C]](
				configPath, &cfg, envOverride, envOverridePrefix); err != nil {
				return err
			}

			ab := baseapp.NewAppBuilder(app.Name())

			// // Maybe move this to BuildApp?
			// ethClient := eth.NewHealthCheckedClient(&cfg.Eth)
			// ab.RegisterEthClient(ethClient)

			// if err = ethClient.Dial(); err != nil {
			// 	logger.Error("failed to dial chain node", err)
			// }

			cp, err := eth.NewConnectionPoolImpl(cfg.ConnectionPool, logger)
			if err != nil {
				return err
			}

			cpi, err := eth.NewChainProviderImpl(cp)
			if err != nil {
				return err
			}

			if err = cpi.DialContext(ctx, ""); err != nil {
				return err
			}

			ab.RegisterEthClient(cpi)

			// Maybe move this to BuildApp?
			svr := server.New(&cfg.Server)
			ab.RegisterHTTPServer(svr)

			// Build the application, then start it.
			app.Setup(ab, cfg.App, logger)
			if err = app.Start(ctx); err != nil {
				return err
			}

			// Wait for the context to be done.
			<-ctx.Done()
			err = ctx.Err()

			defer func() {
				logger.Info("received interrupt signal, exiting gracefully...", ctx.Done())
				app.Stop()
				stop()
			}()

			// TODO: should we return error here based on ctx.Err()?
			if errors.Is(err, context.Canceled) {
				return nil
			}

			return err
		},
	}

	cmd.Flags().String(flags.ConfigPath, flags.DefaultConfigPath, "The config directory")
	cmd.Flags().Bool(flags.EnvOverride, false, "Override config with environment variables")
	cmd.Flags().String(flags.EnvOverridePrefix, "", "Prefix for environment variables")
	return cmd
}
