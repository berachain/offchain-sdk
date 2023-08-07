package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/berachain/offchain-sdk/log"
	"github.com/spf13/cobra"
)

// StartCmdOptions defines options that can be customized in `StartCmdWithOptions`,.
type StartCmdOptions struct{}

// StartCmd runs the application passed in.
func StartCmd(ab AppBuilder, defaultAppHome string) *cobra.Command {
	return StartCmdWithOptions(ab, defaultAppHome, StartCmdOptions{})
}

// StartCmdWithOptions runs the service passed in.
func StartCmdWithOptions(ab AppBuilder, _ string, _ StartCmdOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the service",
		Long:  `Run the service`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Create a context that will be cancelled when the user presses Ctrl+C
			// (process receives termination signal).
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			app := ab.BuildApp(log.NewBlankLogger(cmd.OutOrStdout()))
			if err := app.Start(ctx); err != nil {
				return err
			}

			// Wait on ctx.Done
			<-ctx.Done()
			// TODO: should we return error here based on ctx.Err()?
			return nil
		},
	}

	// cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	return cmd
}
