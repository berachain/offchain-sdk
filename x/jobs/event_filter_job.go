package jobs

import (
	"context"

	"github.com/berachain/offchain-sdk/v2/job"
	sdk "github.com/berachain/offchain-sdk/v2/types"

	"github.com/ethereum/go-ethereum"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// Compile time check to ensure that EthFilterSub implements job.EthSubscribable, and optionally
// the basic job's Setup and Teardown methods.
var (
	_ job.EthSubscribable = (*EthFilterSub)(nil)
	_ job.HasSetup        = (*EthFilterSub)(nil)
	_ job.HasTeardown     = (*EthFilterSub)(nil)
)

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
func (j *EthFilterSub) Subscribe(
	ctx context.Context,
) (ethereum.Subscription, chan coretypes.Log, error) {
	sCtx := sdk.UnwrapContext(ctx)
	ch := make(chan coretypes.Log)
	sub, err := sCtx.Chain().SubscribeFilterLogs(ctx, j.eventFilter, ch)
	if err != nil {
		return nil, nil, err
	}
	j.sub = sub
	return sub, ch, nil
}

// Unsubscribe unsubscribes from filter query.
func (j *EthFilterSub) Unsubscribe(_ context.Context) {
	if j.sub != nil {
		j.sub.Unsubscribe()
	}
}

func (j *EthFilterSub) Setup(ctx context.Context) error {
	if setupJob, ok := j.Basic.(job.HasSetup); ok {
		return setupJob.Setup(ctx)
	}
	return nil
}

func (j *EthFilterSub) Teardown() error {
	if setupJob, ok := j.Basic.(job.HasTeardown); ok {
		return setupJob.Teardown()
	}
	return nil
}
