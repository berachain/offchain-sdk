package main

import (
	"fmt"
	"os"

	"github.com/berachain/offchain-sdk/cmd"
	"github.com/berachain/offchain-sdk/examples/listener/app"
	listenerconfig "github.com/berachain/offchain-sdk/examples/listener/config"
)

func main() {
	if err := cmd.BuildBasicRootCmd[listenerconfig.Config](&app.ListenerApp{}).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your app '%s'", err)
		os.Exit(1)
	}
}
