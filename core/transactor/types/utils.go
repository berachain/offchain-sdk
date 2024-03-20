package types

import (
	"github.com/ethereum/go-ethereum"
	coretypes "github.com/ethereum/go-ethereum/core/types"
)

// CallMsgFromTx creates a new ethereum.CallMsg from a coretypes.Transaction.
func CallMsgFromTx(tx *coretypes.Transaction) *ethereum.CallMsg {
	return &ethereum.CallMsg{
		To:        tx.To(),
		Gas:       tx.Gas(),
		GasFeeCap: tx.GasFeeCap(),
		GasTipCap: tx.GasTipCap(),
		Value:     tx.Value(),
		Data:      tx.Data(),
	}
}
