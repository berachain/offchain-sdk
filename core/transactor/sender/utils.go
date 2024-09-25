package sender

import (
	"fmt"
	"math/big"

	"github.com/holiman/uint256"

	coretypes "github.com/ethereum/go-ethereum/core/types"
)

var (
	multiplier = big.NewInt(11500) //nolint:mnd // its okay.
	quotient   = big.NewInt(10000) //nolint:mnd // its okay.
)

// BumpGas bumps the gas on a tx by a 15% increase.
func BumpGas(tx *coretypes.Transaction) *coretypes.Transaction {
	var innerTx coretypes.TxData
	switch tx.Type() {
	case coretypes.DynamicFeeTxType, coretypes.BlobTxType:
		// Bump the existing gas tip cap 15% (10% is required but add a buffer to be safe).
		bumpedGasTipCap := new(big.Int).Mul(tx.GasTipCap(), multiplier)
		bumpedGasTipCap = new(big.Int).Quo(bumpedGasTipCap, quotient)

		// Bump the existing gas fee cap 15% (only 10% required but add a buffer to be safe).
		bumpedGasFeeCap := new(big.Int).Mul(tx.GasFeeCap(), multiplier)
		bumpedGasFeeCap = new(big.Int).Quo(bumpedGasFeeCap, quotient)

		if tx.Type() == coretypes.BlobTxType {
			// Bump the existing blob gas fee cap 15%. // TODO: verify that this is correct.
			bumpedBlobGasFeeCap := new(big.Int).Mul(tx.BlobGasFeeCap(), multiplier)
			bumpedBlobGasFeeCap = new(big.Int).Quo(bumpedBlobGasFeeCap, quotient)

			innerTx = &coretypes.BlobTx{
				Nonce:      tx.Nonce(),
				To:         *tx.To(),
				Gas:        tx.Gas(),
				Value:      uint256.MustFromBig(tx.Value()),
				Data:       tx.Data(),
				GasTipCap:  uint256.MustFromBig(bumpedGasTipCap),
				GasFeeCap:  uint256.MustFromBig(bumpedGasFeeCap),
				BlobFeeCap: uint256.MustFromBig(bumpedBlobGasFeeCap),
				BlobHashes: tx.BlobHashes(),
				Sidecar:    tx.BlobTxSidecar(),
			}
		} else {
			innerTx = &coretypes.DynamicFeeTx{
				ChainID:   tx.ChainId(),
				Nonce:     tx.Nonce(),
				GasTipCap: bumpedGasTipCap,
				GasFeeCap: bumpedGasFeeCap,
				Gas:       tx.Gas(),
				To:        tx.To(),
				Value:     tx.Value(),
				Data:      tx.Data(),
			}
		}
	case coretypes.LegacyTxType, coretypes.AccessListTxType:
		// Bump the gas price by 15% (10% is required but we add a buffer to be safe).
		bumpedGasPrice := new(big.Int).Mul(tx.GasPrice(), multiplier)
		bumpedGasPrice = new(big.Int).Quo(bumpedGasPrice, quotient)

		if tx.Type() == coretypes.AccessListTxType {
			innerTx = &coretypes.AccessListTx{
				ChainID:    tx.ChainId(),
				Nonce:      tx.Nonce(),
				GasPrice:   bumpedGasPrice,
				Gas:        tx.Gas(),
				To:         tx.To(),
				Value:      tx.Value(),
				Data:       tx.Data(),
				AccessList: tx.AccessList(),
			}
		} else {
			innerTx = &coretypes.LegacyTx{
				Nonce:    tx.Nonce(),
				To:       tx.To(),
				Gas:      tx.Gas(),
				GasPrice: bumpedGasPrice,
				Value:    tx.Value(),
				Data:     tx.Data(),
			}
		}
	default:
		panic(fmt.Sprintf("trying to bump gas on unknown tx type (%d)", tx.Type()))
	}

	return coretypes.NewTx(innerTx)
}

// SetNonce sets the given nonce on a tx.
func SetNonce(tx *coretypes.Transaction, nonce uint64) *coretypes.Transaction {
	var innerTx coretypes.TxData
	switch tx.Type() {
	case coretypes.DynamicFeeTxType:
		innerTx = &coretypes.DynamicFeeTx{
			ChainID:   tx.ChainId(),
			Nonce:     nonce,
			GasTipCap: tx.GasTipCap(),
			GasFeeCap: tx.GasFeeCap(),
			Gas:       tx.Gas(),
			To:        tx.To(),
			Value:     tx.Value(),
			Data:      tx.Data(),
		}
	case coretypes.LegacyTxType:
		innerTx = &coretypes.LegacyTx{
			Nonce:    nonce,
			To:       tx.To(),
			Gas:      tx.Gas(),
			GasPrice: tx.GasPrice(),
			Value:    tx.Value(),
			Data:     tx.Data(),
		}
	case coretypes.AccessListTxType:
		innerTx = &coretypes.AccessListTx{
			ChainID:    tx.ChainId(),
			Nonce:      nonce,
			GasPrice:   tx.GasPrice(),
			Gas:        tx.Gas(),
			To:         tx.To(),
			Value:      tx.Value(),
			Data:       tx.Data(),
			AccessList: tx.AccessList(),
		}
	case coretypes.BlobTxType:
		innerTx = &coretypes.BlobTx{
			Nonce:      nonce,
			To:         *tx.To(),
			Gas:        tx.Gas(),
			Value:      uint256.MustFromBig(tx.Value()),
			Data:       tx.Data(),
			GasTipCap:  uint256.MustFromBig(tx.GasTipCap()),
			GasFeeCap:  uint256.MustFromBig(tx.GasFeeCap()),
			BlobFeeCap: uint256.MustFromBig(tx.BlobGasFeeCap()),
			BlobHashes: tx.BlobHashes(),
			Sidecar:    tx.BlobTxSidecar(),
		}
	default:
		panic(fmt.Sprintf("trying to set nonce on unknown tx type (%d)", tx.Type()))
	}

	return coretypes.NewTx(innerTx)
}
