package eth

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethcoretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	MaxRetries       = 3
	defaultRetryTime = 1 * time.Second
)

type Client interface {
	DialContext(ctx context.Context, rawurl string) error
	Close() error
	Reader
	Writer
}

// Reader is the eth reader interface.
type Reader interface {
	BlockByNumber(ctx context.Context, number *big.Int) (*ethcoretypes.Block, error)
	BlockReceipts(
		ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash,
	) ([]*ethcoretypes.Receipt, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethcoretypes.Receipt, error)
	SubscribeNewHead(ctx context.Context) (chan *ethcoretypes.Header, ethereum.Subscription, error)
	BlockNumber(ctx context.Context) (uint64, error)
	ChainID(ctx context.Context) (*big.Int, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]ethcoretypes.Log, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*ethcoretypes.Header, error)
	PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	SendTransaction(ctx context.Context, tx *ethcoretypes.Transaction) error
	SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery,
		ch chan<- ethcoretypes.Log) (ethereum.Subscription, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	TransactionByHash(ctx context.Context, hash common.Hash,
	) (tx *ethcoretypes.Transaction, isPending bool, err error)
	TxPoolContent(
		ctx context.Context) (
		map[string]map[string]map[string]*ethcoretypes.Transaction, error)
}

type Writer interface {
	SendTransaction(ctx context.Context, tx *ethcoretypes.Transaction) error
}

// client is the indexer eth client.
type ExtendedEthClient struct {
	*ethclient.Client
}

// ==================================================================
// Client Lifecycle
// ==================================================================

func (c *ExtendedEthClient) DialContext(ctx context.Context, rawurl string) error {
	if c.Client != nil {
		return nil
	}

	var err error
	c.Client, err = ethclient.DialContext(ctx, rawurl)
	return err
}

// Close closes the client.
func (c *ExtendedEthClient) Close() error {
	if c == nil {
		return ErrClosed
	}
	c.Close()
	return nil
}

// ==================================================================
// Client Usage Methods
// ==================================================================

// GetReceipts returns the receipts for the given transactions.
func (c *ExtendedEthClient) GetReceipts(
	ctx context.Context, txs ethcoretypes.Transactions) (ethcoretypes.Receipts, error) {
	var receipts ethcoretypes.Receipts
	for _, tx := range txs {
		receipt, err := c.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			return nil, err
		}
		receipts = append(receipts, receipt)
	}
	return receipts, nil
}

// SubscribeNewHead subscribes to new block headers.
func (c *ExtendedEthClient) SubscribeNewHead(
	ctx context.Context) (chan *ethcoretypes.Header, ethereum.Subscription, error) {
	ch := make(chan *ethcoretypes.Header)
	sub, err := c.Client.SubscribeNewHead(ctx, ch)
	return ch, sub, err
}

func (c *ExtendedEthClient) SubscribeFilterLogs(
	ctx context.Context,
	q ethereum.FilterQuery, ch chan<- ethcoretypes.Log) (ethereum.Subscription, error) {
	return c.Client.SubscribeFilterLogs(ctx, q, ch)
}

func (c *ExtendedEthClient) TxPoolContent(
	ctx context.Context,
) (map[string]map[string]map[string]*ethcoretypes.Transaction, error) {
	// var result map[string]map[string]map[string]*ethcoretypes.Transaction
	var result map[string]map[string]map[string]*ethcoretypes.Transaction
	if err := c.Client.Client().CallContext(ctx, &result, "txpool_content"); err != nil {
		return nil, err
	}
	return result, nil
}
