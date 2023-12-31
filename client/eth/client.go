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
	Dial() error
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
	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]ethcoretypes.Log, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*ethcoretypes.Header, error)
	PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SendTransaction(ctx context.Context, tx *ethcoretypes.Transaction) error
	SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery,
		ch chan<- ethcoretypes.Log) (ethereum.Subscription, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	TransactionByHash(ctx context.Context, hash common.Hash,
	) (tx *ethcoretypes.Transaction, isPending bool, err error)
}

type Writer interface {
	SendTransaction(ctx context.Context, tx *ethcoretypes.Transaction) error
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
	if c.Client != nil || c.wsclient != nil {
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
	if c == nil || c.wsclient == nil {
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
func (c *client) GetReceipts(
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
func (c *client) SubscribeNewHead(
	ctx context.Context) (chan *ethcoretypes.Header, ethereum.Subscription, error) {
	ch := make(chan *ethcoretypes.Header)
	sub, err := c.wsclient.SubscribeNewHead(ctx, ch)
	return ch, sub, err
}

func (c *client) SubscribeFilterLogs(
	ctx context.Context,
	q ethereum.FilterQuery, ch chan<- ethcoretypes.Log) (ethereum.Subscription, error) {
	return c.wsclient.SubscribeFilterLogs(ctx, q, ch)
}
