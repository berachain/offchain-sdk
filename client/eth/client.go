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

type Client interface {
	DialContext(ctx context.Context, rawurl string) error
	Close() error
	Health() bool
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
	SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery,
		ch chan<- ethcoretypes.Log) (ethereum.Subscription, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	TransactionByHash(ctx context.Context, hash common.Hash,
	) (tx *ethcoretypes.Transaction, isPending bool, err error)

	/*
		TxPoolContentFrom returns the pending and queued transactions of this address.
		Example response:

		{
			"pending": {
				"1": {
					// transaction details...
				},
				...
			},
			"queued": {
				"3": {
					// transaction details...
				},
				...
			}
		}
	*/
	TxPoolContentFrom(ctx context.Context, address common.Address) (
		map[string]map[string]*ethcoretypes.Transaction, error,
	)

	/*
		TxPoolInspect returns the textual summary of all pending and queued transactions.
		Example response:

		{
			"pending": {
				"0x12345": {
					"1": "0x12345789: 1 wei + 2 gas x 3 wei"
				},
				...
			},
			"queued": {
				"0x67890": {
					"2": "0x12345789: 1 wei + 2 gas x 3 wei"
				},
				...
			}
		}
	*/
	TxPoolInspect(ctx context.Context) (map[string]map[common.Address]map[string]string, error)
}

type Writer interface {
	SendTransaction(ctx context.Context, tx *ethcoretypes.Transaction) error
}

// client is the indexer eth client.
type ExtendedEthClient struct {
	*ethclient.Client
	rpcTimeout time.Duration
}

func NewExtendedEthClient(c *ethclient.Client, rpcTimeout time.Duration) *ExtendedEthClient {
	return &ExtendedEthClient{
		Client:     c,
		rpcTimeout: rpcTimeout,
	}
}

// ==================================================================
// Client Lifecycle
// ==================================================================

func (c *ExtendedEthClient) DialContext(ctx context.Context, rawurl string) error {
	if c.Client != nil {
		return nil
	}

	var err error
	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
	c.Client, err = ethclient.DialContext(ctxWithTimeout, rawurl)
	cancel()
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

func (c *ExtendedEthClient) Health() bool {
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), c.rpcTimeout)
	_, err := c.ChainID(ctxWithTimeout)
	cancel()
	return err == nil
}

// ==================================================================
// Client Usage Methods
// ==================================================================

// GetReceipts returns the receipts for the given transactions.
func (c *ExtendedEthClient) GetReceipts(
	ctx context.Context, txs ethcoretypes.Transactions) (ethcoretypes.Receipts, error) {
	var receipts ethcoretypes.Receipts
	for _, tx := range txs {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
		receipt, err := c.TransactionReceipt(ctxWithTimeout, tx.Hash())
		cancel()
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
	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
	sub, err := c.Client.SubscribeNewHead(ctxWithTimeout, ch)
	cancel()
	return ch, sub, err
}

func (c *ExtendedEthClient) SubscribeFilterLogs(
	ctx context.Context,
	q ethereum.FilterQuery, ch chan<- ethcoretypes.Log) (ethereum.Subscription, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
	defer cancel()
	return c.Client.SubscribeFilterLogs(ctxWithTimeout, q, ch)
}

func (c *ExtendedEthClient) TxPoolContentFrom(
	ctx context.Context, address common.Address,
) (map[string]map[string]*ethcoretypes.Transaction, error) {
	var result map[string]map[string]*ethcoretypes.Transaction
	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
	defer cancel()
	if err := c.Client.Client().CallContext(
		ctxWithTimeout, &result, "txpool_contentFrom", address,
	); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *ExtendedEthClient) TxPoolInspect(
	ctx context.Context,
) (map[string]map[common.Address]map[string]string, error) {
	var result map[string]map[common.Address]map[string]string
	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.rpcTimeout)
	defer cancel()
	if err := c.Client.Client().CallContext(
		ctxWithTimeout, &result, "txpool_inspect",
	); err != nil {
		return nil, err
	}
	return result, nil
}
