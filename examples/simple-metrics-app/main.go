package main

import (
	"fmt"
	"os"

	"github.com/berachain/offchain-sdk/cmd"
	"github.com/berachain/offchain-sdk/examples/simple-metrics-app/app"
)

func main() {
	if err := cmd.BuildBasicRootCmd(&app.SimpleMetricsApp{}).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your app '%s'", err)
		os.Exit(1)
	}
}
