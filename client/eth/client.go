package eth

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	MaxRetries       = 3
	defaultRetryTime = 1 * time.Second
)

type Client interface {
	Dial() error
	Close() error
	Reader
	Writer
}

// Reader is the eth reader interface.
type Reader interface {
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	SubscribeNewHead(ctx context.Context) (chan *types.Header, ethereum.Subscription, error)
	BlockNumber(ctx context.Context) (uint64, error)
	ChainID(ctx context.Context) (*big.Int, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	BalanceAt(ctx context.Context, address common.Address, bn *big.Int) (*big.Int, error)
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)
	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery,
		ch chan<- types.Log) (ethereum.Subscription, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error)
}

type Writer interface {
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	CallContract(ctx context.Context, msg ethereum.CallMsg,
		blockNumber *big.Int) ([]byte, error)
}

// client is the indexer eth client.
type client struct {
	*ethclient.Client
	cfg      *Config
	wsclient *ethclient.Client
}

// NewClient returns a new client. It has both reader and writer privilege.
func NewClient(cfg *Config) Client {
	client := &client{
		cfg: cfg,
	}
	return client
}

// ==================================================================
// Client Lifecycle
// ==================================================================

// Dial dials the client.
func (c *client) Dial() error {
	if c.wsclient != nil {
		return ErrAlreadyDial
	}
	// TODO: manage context better
	ctx := context.Background()
	retries := 0
	var err error
	var httpclient, wsethclient *ethclient.Client
	for retries < MaxRetries {
		retries++
		httpclient, err = ethclient.DialContext(ctx, c.cfg.EthHTTPURL)
		if err == nil {
			c.Client = httpclient
			break
		}
		time.Sleep(defaultRetryTime)
	}
	if err != nil {
		return err
	}

	retries = 0
	for retries < MaxRetries {
		retries++
		wsethclient, err = ethclient.DialContext(ctx, c.cfg.EthWSURL)
		if err == nil {
			c.wsclient = wsethclient
			break
		}
		time.Sleep(defaultRetryTime)
	}
	if err != nil {
		return err
	}

	return nil
}

// Close closes the client.
func (c *client) Close() error {
	if c.wsclient == nil {
		return ErrClosed
	}
	c.Close()
	c.wsclient.Close()
	return nil
}

// ==================================================================
// Client Usage Methods
// ==================================================================

// GetReceipts returns the receipts for the given transactions.
func (c *client) GetReceipts(ctx context.Context, txs types.Transactions) (types.Receipts, error) {
	var receipts types.Receipts
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
func (c *client) SubscribeNewHead(ctx context.Context) (chan *types.Header, ethereum.Subscription, error) {
	ch := make(chan *types.Header)
	sub, err := c.wsclient.SubscribeNewHead(ctx, ch)
	return ch, sub, err
}
