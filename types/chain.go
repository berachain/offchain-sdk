package types

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Chain interface {
	ChainReader
	ChainWriter
	ChainSubscriber
}

type ChainWriter interface {
	CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	SendTransaction(tx *types.Transaction) error
}

type ChainReader interface {
	CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error)
	CurrentBlock() (*types.Block, error)
	GetBlockByNumber(number uint64) (*types.Block, error)
	EstimateGas(ctx context.Context, call ethereum.CallMsg) (uint64, error)
	FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
}

type ChainSubscriber interface {
	SubscribeFilterLogs(q ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error)
}
