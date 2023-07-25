package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

type TxProps struct {
	to common.Address
	data []byte
}

type Wrapper interface {
	SendSignedTransaction(ctx context.Context, tx *types.Transaction) error
	SignTransaction()
}

type wrapper struct {
	ethclient Client
	nonceManager NonceManager
	cfg        *Config
}

func (s *wrapper) SendSignedTransaction(ctx context.Context, props TxProps) error {
	nonce, err := s.nonceManager.GetNonce(ctx, props.to)
	if err != nil {
		return err
	}
	// build a tx
	txdata := &types.DynamicFeeTx{
		ChainID:    &s.cfg.ChainID,
		Nonce:      nonce,
		To:        	&props.to,
		Gas:        300000, // todo: make it variable
		GasFeeCap:  newGwei(5),
		GasTipCap:  big.NewInt(2),
		// AccessList: ??,
		Data:       []byte{},
	}
	tx := types.NewTx(txdata)
	// sign it
	tx, _ = s.SignTransaction(tx, &s.cfg.ChainID)
	// send it
	return  s.ethclient.SendTransaction(ctx, tx)
}

func (s *wrapper) SignTransaction(tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	privateKey, err := crypto.HexToECDSA(s.cfg.PrivateKey)
	if err != nil {
		return nil, err
	}
	return types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
}

// func (s *wrapper) GetTransactionGasFees(ctx context.Context, gasMode GasMode, isPriority bool) error {
// 	// GAS_MODE_EIP1559
// 	// set maxPriorityFeePerGas
// 	// set maxFeePerGas
// }

func newGwei(n int64) *big.Int {
	return new(big.Int).Mul(big.NewInt(n), big.NewInt(params.GWei))
}
