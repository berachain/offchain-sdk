package app

import (
	"context"

	"github.com/berachain/offchain-sdk/log"
)

type App[C any] interface {
	Name() string
	Setup(ab Builder, config C, logger log.Logger)
	Start(context.Context) error
	Stop()
}
