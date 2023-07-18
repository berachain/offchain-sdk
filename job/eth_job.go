package job

import (
	"context"

	sdk "github.com/berachain/offchain-sdk/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var _ EthSubscribable = (*EthEventSub)(nil)

type EthEventSub struct {
	exec            func(context.Context, any) (any, error)
	contractAddress common.Address
	event           string
	sub             ethereum.Subscription
}

func NewEthSub(contractAddr common.Address, event string, exec func(context.Context, any) (any, error)) *EthEventSub {
	return &EthEventSub{
		exec:            exec,
		contractAddress: contractAddr,
		event:           event,
	}
}

func (j *EthEventSub) Execute(ctx context.Context, args any) (any, error) {
	Sctx := sdk.UnwrapSdkContext(ctx)
	Sctx.Logger().Info("executing eth sub", "args", args)
	return j.exec(ctx, args)
}

func (j *EthEventSub) Subscribe(ctx context.Context) (ethereum.Subscription, chan coretypes.Log) {
	sCtx := sdk.UnwrapSdkContext(ctx)
	ch := make(chan coretypes.Log)
	sub, err := sCtx.Chain().SubscribeFilterLogs(ethereum.FilterQuery{
		Addresses: []common.Address{j.contractAddress},
		Topics:    [][]common.Hash{{crypto.Keccak256Hash([]byte(j.event))}},
	}, ch)
	if err != nil {
		panic(err)
	}
	return sub, ch
}

func (j *EthEventSub) Unsubscribe(ctx context.Context) {
	j.sub.Unsubscribe()
}
