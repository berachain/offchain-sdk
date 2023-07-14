package main

import (
	"fmt"
	"os"

	"github.com/berachain/offchain-sdk/cmd"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := cmd.BuildRootCommand(
		"cron",
		"cron does crons",
		cobra.NoArgs,
	)

	rootCmd.AddCommand(
		cmd.BuildStartCommand("cron", cobra.ExactArgs(0)),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
