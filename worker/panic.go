package worker

import (
	"runtime/debug"

	"github.com/berachain/offchain-sdk/log"
)

// PanicHandler builds a panic handler for the worker pool that logs the panic.
func PanicHandler(logger log.Logger) func(interface{}) {
	return func(panic interface{}) {
		logger.Error("Worker exits from a panic", "reason", panic, "stack-trace", string(debug.Stack()))
	}
}
