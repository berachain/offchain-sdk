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

// Compile time check to ensure that EthEventSub implements job.EthSubscribable.
var _ job.EthSubscribable = (*EthEventSub)(nil)

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
func (j *EthEventSub) Subscribe(ctx context.Context) (ethereum.Subscription, chan coretypes.Log) {
	sCtx := sdk.UnwrapContext(ctx)
	ch := make(chan coretypes.Log)
	sub, err := sCtx.Chain().SubscribeFilterLogs(context.Background(), ethereum.FilterQuery{
		Addresses: []common.Address{j.contractAddress},
		Topics:    [][]common.Hash{{crypto.Keccak256Hash([]byte(j.event))}},
	}, ch)
	j.sub = sub
	if err != nil {
		panic(err)
	}
	return sub, ch
}

// Unsubscribe unsubscribes from an ethereum event.
func (j *EthEventSub) Unsubscribe(_ context.Context) {
	j.sub.Unsubscribe()
}
