package jobs

import (
	"context"

	"github.com/berachain/offchain-sdk/job"
	sdk "github.com/berachain/offchain-sdk/types"
	"github.com/ethereum/go-ethereum"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// Compile time check to ensure that EthEventSub implements job.EthSubscribable.
var _ job.EthSubscribable = (*EthFilterSub)(nil)

// EthFilterSub allows you to subscribe a basic job to an ethereum event.
type EthFilterSub struct {
	job.Basic
	eventFilter ethereum.FilterQuery
	sub         ethereum.Subscription
}

// NewEthFilterSub creates a new EthFilterSub
// eventFilter is a ethereum.FilterQuery.
func NewEthFilterSub(job job.Basic, eventFilter ethereum.FilterQuery) *EthFilterSub {
	return &EthFilterSub{
		Basic:       job,
		eventFilter: eventFilter,
	}
}

// Subscribe subscribes to all events based on ethereum filter query.
func (j *EthFilterSub) Subscribe(ctx context.Context) (ethereum.Subscription, chan coretypes.Log) {
	sCtx := sdk.UnwrapSdkContext(ctx)
	ch := make(chan coretypes.Log)
	sub, err := sCtx.Chain().SubscribeFilterLogs(context.Background(), j.eventFilter, ch)
	j.sub = sub
	if err != nil {
		panic(err)
	}
	return sub, ch
}

// Unsubscribe unsubscribes from filter query.
func (j *EthFilterSub) Unsubscribe(_ context.Context) {
	j.sub.Unsubscribe()
}
