package main

import (
	"fmt"
	"os"

	"github.com/berachain/offchain-sdk/v2/cmd"
	"github.com/berachain/offchain-sdk/v2/examples/listener/app"
	listenerconfig "github.com/berachain/offchain-sdk/v2/examples/listener/config"
)

func main() {
	if err := cmd.BuildBasicRootCmd[listenerconfig.Config](&app.ListenerApp{}).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your app '%s'", err)
		os.Exit(1)
	}
}
