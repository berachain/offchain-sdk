package jobs

import (
	"context"

	"github.com/berachain/offchain-sdk/job"
	sdk "github.com/berachain/offchain-sdk/types"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// Compile time check to ensure that EthEventSub implements job.EthSubscribable, and optionally the
// basic job's Setup and Teardown methods.
var (
	_ job.EthSubscribable = (*EthEventSub)(nil)
	_ job.HasSetup        = (*EthEventSub)(nil)
	_ job.HasTeardown     = (*EthEventSub)(nil)
)

// EthEventSub allows you to subscribe a basic job to an ethereum event.
type EthEventSub struct {
	job.Basic
	contractAddress common.Address
	event           string
	sub             ethereum.Subscription
}

// NewEthSub creates a new EthEventSub.
func NewEthSub(job job.Basic, contractAddr string, event string) *EthEventSub {
	return &EthEventSub{
		Basic:           job,
		contractAddress: common.HexToAddress(contractAddr),
		event:           event,
	}
}

// Subscribe subscribes to an ethereum event.
func (j *EthEventSub) Subscribe(
	ctx context.Context,
) (ethereum.Subscription, chan any, error) {
	sCtx := sdk.UnwrapContext(ctx)
	logCh := make(chan coretypes.Log)
	sub, err := sCtx.Chain().SubscribeFilterLogs(ctx, ethereum.FilterQuery{
		Addresses: []common.Address{j.contractAddress},
		Topics:    [][]common.Hash{{crypto.Keccak256Hash([]byte(j.event))}},
	}, logCh)
	if err != nil {
		return nil, nil, err
	}
	j.sub = sub

	ch := make(chan any)
	go func() {
		defer close(ch)
		for {
			select {
			case val, ok := <-logCh:
				if !ok {
					return
				}
				ch <- val
			case <-ctx.Done():
				return
			}
		}
	}()

	return sub, ch, nil
}

// Unsubscribe unsubscribes from an ethereum event.
func (j *EthEventSub) Unsubscribe(_ context.Context) {
	if j.sub != nil {
		j.sub.Unsubscribe()
	}
}

func (j *EthEventSub) Setup(ctx context.Context) error {
	if setupJob, ok := j.Basic.(job.HasSetup); ok {
		return setupJob.Setup(ctx)
	}
	return nil
}

func (j *EthEventSub) Teardown() error {
	if setupJob, ok := j.Basic.(job.HasTeardown); ok {
		return setupJob.Teardown()
	}
	return nil
}
