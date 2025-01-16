package worker

import (
	"runtime/debug"

	"github.com/berachain/offchain-sdk/v2/log"
)

// PanicHandler builds a panic handler for the worker pool that logs the panic.
func PanicHandler(logger log.Logger) func(interface{}) {
	return func(panicData interface{}) {
		logger.Error(
			"Worker exits from a panic",
			"reason", panicData, "stack-trace", string(debug.Stack()),
		)
	}
}
