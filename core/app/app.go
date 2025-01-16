package app

import (
	"context"

	"github.com/berachain/offchain-sdk/v2/log"
)

type App[C any] interface {
	Name() string
	Setup(ab Builder, config C, logger log.Logger) error
	Start(context.Context) error
	Stop()
}
