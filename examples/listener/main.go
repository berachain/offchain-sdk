package main

import (
	"fmt"
	"os"

	"github.com/berachain/offchain-sdk/cmd"
	"github.com/berachain/offchain-sdk/config"
	"github.com/berachain/offchain-sdk/config/toml"
	"github.com/berachain/offchain-sdk/examples/listener/app"
	listenerconfig "github.com/berachain/offchain-sdk/examples/listener/config"
)

func main() {
	var target config.Config[listenerconfig.Config]
	toml.ReadIntoMap[config.Config[listenerconfig.Config]]("config.toml", &target)

	if err := cmd.BuildBasicRootCmd[listenerconfig.Config](&app.ListenerApp{}).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your app '%s'", err)
		os.Exit(1)
	}
}
