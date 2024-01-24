package eth

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

// ErrClientNotFound is an error that is returned when a client is not found.
var (
	ErrClientNotFound = errors.New("client not found")
)

// ChainProvider is an interface that groups the Reader, Writer, and ConnectionPool interfaces.
type ChainProvider interface {
	Reader
	Writer
	ConnectionPool
}

// ChainProviderImpl is an implementation of the ChainProvider interface.
type ChainProviderImpl struct {
	ConnectionPool
}

// NewChainProviderImpl creates a new ChainProviderImpl with the given ConnectionPool.
func NewChainProviderImpl(pool ConnectionPool) (Client, error) {
	return &ChainProviderImpl{pool}, nil
}

// ==================================================================
// Implementations of Reader and Writer
// ==================================================================

// BlockByNumber returns the block for the given number.
func (c *ChainProviderImpl) BlockByNumber(
	ctx context.Context, num *big.Int) (*types.Block, error) {
	if client, ok := c.GetHTTP(); ok {
		return client.BlockByNumber(ctx, num)
	}
	return nil, ErrClientNotFound
}

// BlockReceipts returns the receipts for the given block number or hash.
func (c *ChainProviderImpl) BlockReceipts(
	ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) ([]*types.Receipt, error) {
	if client, ok := c.GetHTTP(); ok {
		return client.BlockReceipts(ctx, blockNrOrHash)
	}
	return nil, ErrClientNotFound
}

// TransactionReceipt returns the receipt for the given transaction hash.
func (c *ChainProviderImpl) TransactionReceipt(
	ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	if client, ok := c.GetHTTP(); ok {
		return client.TransactionReceipt(ctx, txHash)
	}
	return nil, ErrClientNotFound
}

// SubscribeNewHead subscribes to new head events.
func (c *ChainProviderImpl) SubscribeNewHead(
	ctx context.Context) (chan *types.Header, ethereum.Subscription, error) {
	if client, ok := c.GetWS(); ok {
		return client.SubscribeNewHead(ctx)
	}
	return nil, nil, ErrClientNotFound
}

// BlockNumber returns the current block number.
func (c *ChainProviderImpl) BlockNumber(ctx context.Context) (uint64, error) {
	if client, ok := c.GetHTTP(); ok {
		return client.BlockNumber(ctx)
	}
	return 0, ErrClientNotFound
}

// ChainID returns the current chain ID.
func (c *ChainProviderImpl) ChainID(ctx context.Context) (*big.Int, error) {
	if client, ok := c.GetHTTP(); ok {
		return client.ChainID(ctx)
	}
	return nil, ErrClientNotFound
}

// BalanceAt returns the balance of the given address at the given block number.
func (c *ChainProviderImpl) BalanceAt(
	ctx context.Context, address common.Address, blockNumber *big.Int) (*big.Int, error) {
	if client, ok := c.GetHTTP(); ok {
		return client.BalanceAt(ctx, address, blockNumber)
	}
	return nil, ErrClientNotFound
}

// CodeAt returns the code of the given account at the given block number.
func (c *ChainProviderImpl) CodeAt(
	ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	if client, ok := c.GetHTTP(); ok {
		return client.CodeAt(ctx, account, blockNumber)
	}
	return nil, ErrClientNotFound
}

// EstimateGas estimates the gas needed to execute a specific transaction.
func (c *ChainProviderImpl) EstimateGas(
	ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	if client, ok := c.GetHTTP(); ok {
		return client.EstimateGas(ctx, msg)
	}
	return 0, ErrClientNotFound
}

// FilterLogs returns the logs that satisfy the given filter query.
func (c *ChainProviderImpl) FilterLogs(
	ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	if client, ok := c.GetHTTP(); ok {
		return client.FilterLogs(ctx, q)
	}
	return nil, ErrClientNotFound
}

// HeaderByNumber returns the header of the block with the given number.
func (c *ChainProviderImpl) HeaderByNumber(
	ctx context.Context, number *big.Int) (*types.Header, error) {
	if client, ok := c.GetHTTP(); ok {
		return client.HeaderByNumber(ctx, number)
	}
	return nil, ErrClientNotFound
}

// PendingCodeAt returns the code of the given account in the pending state.
func (c *ChainProviderImpl) PendingCodeAt(
	ctx context.Context, account common.Address) ([]byte, error) {
	if client, ok := c.GetHTTP(); ok {
		return client.PendingCodeAt(ctx, account)
	}
	return nil, ErrClientNotFound
}

// PendingNonceAt returns the nonce of the given account in the pending state.
func (c *ChainProviderImpl) PendingNonceAt(
	ctx context.Context, account common.Address) (uint64, error) {
	if client, ok := c.GetHTTP(); ok {
		return client.PendingNonceAt(ctx, account)
	}
	return 0, ErrClientNotFound
}

// PendingNonceAt returns the nonce of the given account in the pending state.
func (c *ChainProviderImpl) NonceAt(
	ctx context.Context, account common.Address, bn *big.Int) (uint64, error) {
	if client, ok := c.GetHTTP(); ok {
		return client.NonceAt(ctx, account, bn)
	}
	return 0, ErrClientNotFound
}

// SendTransaction sends the given transaction.
func (c *ChainProviderImpl) SendTransaction(
	ctx context.Context, tx *types.Transaction) error {
	if client, ok := c.GetHTTP(); ok {
		return client.SendTransaction(ctx, tx)
	}
	return ErrClientNotFound
}

// SubscribeFilterLogs subscribes to new log events that satisfy the given filter query.
func (c *ChainProviderImpl) SubscribeFilterLogs(
	ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log,
) (ethereum.Subscription, error) {
	if client, ok := c.GetWS(); ok {
		return client.SubscribeFilterLogs(ctx, q, ch)
	}
	return nil, ErrClientNotFound
}

// SuggestGasPrice suggests a gas price.
func (c *ChainProviderImpl) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	if client, ok := c.GetHTTP(); ok {
		return client.SuggestGasPrice(ctx)
	}
	return nil, ErrClientNotFound
}

// CallContract calls a contract with the given message at the given block number.
func (c *ChainProviderImpl) CallContract(
	ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int,
) ([]byte, error) {
	if client, ok := c.GetHTTP(); ok {
		return client.CallContract(ctx, msg, blockNumber)
	}
	return nil, ErrClientNotFound
}

// SuggestGasTipCap suggests a gas tip cap.
func (c *ChainProviderImpl) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	if client, ok := c.GetHTTP(); ok {
		return client.SuggestGasTipCap(ctx)
	}
	return nil, ErrClientNotFound
}

// TransactionByHash returns the transaction with the given hash.
func (c *ChainProviderImpl) TransactionByHash(
	ctx context.Context, hash common.Hash,
) (*types.Transaction, bool, error) {
	if client, ok := c.GetHTTP(); ok {
		return client.TransactionByHash(ctx, hash)
	}
	return nil, false, ErrClientNotFound
}

// "id": 1,
// "result": {
// 	"pending": {
// 		"0xe74aA377Dbc22450349774d1C427337995120DCB": {
// 			"3698316": {
// 				"blockHash": null,
// 				"blockNumber": null,
// 				"from": "0xe74aa377dbc22450349774d1c427337995120dcb",
// 				"gas": "0x715b",

func (c *ChainProviderImpl) TxPoolContent(ctx context.Context) (
	map[string]map[string]map[string]*types.Transaction, error,
) {
	if client, ok := c.GetHTTP(); ok {
		return client.TxPoolContent(ctx)
	}
	return nil, ErrClientNotFound
}
