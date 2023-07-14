package eth

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethcoretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const MaxRetries = 3

// Client is the indexer eth client.
type Reader interface {
	Dial() error
	Close() error
	GetBlockByNumber(ctx context.Context, number uint64) (*ethcoretypes.Block, error)
	GetReceipts(ctx context.Context, txs ethcoretypes.Transactions) (ethcoretypes.Receipts, error)
	GetReceipt(ctx context.Context, txHash common.Hash) (*ethcoretypes.Receipt, error)
	SubscribeNewHead(ctx context.Context) (chan *ethcoretypes.Header, ethereum.Subscription, error)
	BlockNumber(ctx context.Context) (uint64, error)
	ChainID(ctx context.Context) (*big.Int, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethcoretypes.Receipt, error)
	GetBalance(ctx context.Context, address common.Address) (*big.Int, error)
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	// EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
	// FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]ethcoretypes.Log, error)
	// HeaderByNumber(ctx context.Context, number *big.Int) (*ethcoretypes.Header, error)
	// PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error)
	// PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	// SendTransaction(ctx context.Context, tx *ethcoretypes.Transaction) error
	// SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- ethcoretypes.Log) (ethereum.Subscription, error)
	// SuggestGasPrice(ctx context.Context) (*big.Int, error)
}

// client is the indexer eth client.
type client struct {
	cfg        *Config
	httpclient *ethclient.Client
	wsclient   *ethclient.Client
}

// NewClient returns a new client.
func NewReader(cfg *Config) Reader {
	return &client{
		cfg: cfg,
	}
}

// ==================================================================
// Client Lifecycle
// ==================================================================

// Dial dials the client.
func (c *client) Dial() error {
	if c.httpclient != nil || c.wsclient != nil {
		return ErrAlreadyDial
	}
	// TODO: manage context better
	ctx := context.Background()
	retries := 0
	var err error
	var httpclient, wsethclient *ethclient.Client
	for retries < MaxRetries {
		retries++
		httpclient, err = ethclient.DialContext(ctx, c.cfg.EthHttpURL)
		if err == nil {
			c.httpclient = httpclient
			break
		}
		fmt.Println("failed to dial http client", err, "retrying in 5, retries - ", retries)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		panic(err)
	}

	retries = 0
	for retries < MaxRetries {
		retries++
		wsethclient, err = ethclient.DialContext(ctx, c.cfg.EthWSURL)
		if err == nil {
			c.wsclient = wsethclient
			break
		}
		fmt.Println("failed to dial ws client", err, "retrying in 5, retries - ", retries)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		panic(err)
	}

	return nil
}

// Close closes the client.
func (c *client) Close() error {
	if c.httpclient == nil || c.wsclient == nil {
		return ErrClosed
	}
	c.httpclient.Close()
	c.wsclient.Close()
	return nil
}

// ==================================================================
// Client Usage Methods
// ==================================================================

// GetBlockByNumber returns the block for the given block number.
func (c *client) GetBlockByNumber(ctx context.Context, number uint64) (*ethcoretypes.Block, error) {
	return c.httpclient.BlockByNumber(ctx, big.NewInt(int64(number)))
}

// GetBalance returns the balance for the given address.
func (c *client) GetBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	return c.httpclient.BalanceAt(ctx, address, nil)
}

// GetReceipts returns the receipts for the given transactions.
func (c *client) GetReceipts(ctx context.Context, txs ethcoretypes.Transactions) (ethcoretypes.Receipts, error) {
	var receipts ethcoretypes.Receipts
	for _, tx := range txs {
		receipt, err := c.httpclient.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			return nil, err
		}
		receipts = append(receipts, receipt)
	}
	return receipts, nil
}

// GetReceipt returns the receipt for the given transaction hash.
func (c *client) GetReceipt(ctx context.Context, txHash common.Hash) (*ethcoretypes.Receipt, error) {
	return c.httpclient.TransactionReceipt(ctx, txHash)
}

// SubscribeNewHead subscribes to new block headers.
func (c *client) SubscribeNewHead(ctx context.Context) (chan *ethcoretypes.Header, ethereum.Subscription, error) {
	ch := make(chan *ethcoretypes.Header)
	sub, err := c.wsclient.SubscribeNewHead(ctx, ch)
	return ch, sub, err
}

// BlockNumber returns the current block number.
func (c *client) BlockNumber(ctx context.Context) (uint64, error) {
	return c.httpclient.BlockNumber(ctx)
}

// ChainID returns the current chain ID.
func (c *client) ChainID(ctx context.Context) (*big.Int, error) {
	return c.httpclient.ChainID(ctx)
}

func (c *client) TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethcoretypes.Receipt, error) {
	return c.httpclient.TransactionReceipt(ctx, txHash)
}

func (c *client) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	return c.httpclient.CodeAt(ctx, contract, blockNumber)
}

func (c *client) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return c.httpclient.CallContract(ctx, msg, blockNumber)
}

// func (c *client) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
// 	return c.httpclient.EstimateGas(ctx, msg)
// }

// func (c *client) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]ethcoretypes.Log, error) {
// 	return c.httpclient.FilterLogs(ctx, q)
// }

// func (c *client) HeaderByNumber(ctx context.Context, number *big.Int) (*ethcoretypes.Header, error) {
// 	return c.httpclient.HeaderByNumber(ctx, number)
// }

// func (c *client) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
// 	return c.httpclient.PendingCodeAt(ctx, account)
// }

// func (c *client) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
// 	return c.httpclient.PendingNonceAt(ctx, account)
// }

// func (c *client) SendTransaction(ctx context.Context, tx *ethcoretypes.Transaction) error {
// 	return c.httpclient.SendTransaction(ctx, tx)
// }

// func (c *client) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- ethcoretypes.Log) (ethereum.Subscription, error) {
// 	return c.httpclient.SubscribeFilterLogs(ctx, q, ch)
// }

// func (c *client) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
// 	return c.httpclient.SuggestGasPrice(ctx)
// }
