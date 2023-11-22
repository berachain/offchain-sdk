package factory

import (
	"context"
	"math/big"

	"github.com/berachain/offchain-sdk/contracts/bindings"
	"github.com/berachain/offchain-sdk/core/transactor/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// forge create Multicall3 --rpc-url=http://devnet.beraswillmakeit.com:8545
// --private-key=0xfffdbb37105441e14b0ee6330d855d8504ff39e705c3afa8f859ac9865f99306.
type Multicall3Batcher struct {
	contractAddress common.Address
}

// NewMulticall3Batcher creates a new Multicall3Batcher instance.
func NewMulticall3Batcher(address common.Address) *Multicall3Batcher {
	return &Multicall3Batcher{
		contractAddress: address,
	}
}

// BatchTxRequests creates a batched transaction request for the given transaction requests.
func (mc *Multicall3Batcher) BatchTxRequests(
	_ context.Context,
	txReqs []*types.TxRequest,
) *types.TxRequest {
	calls := make([]bindings.Multicall3Call, len(txReqs))
	for i, txReq := range txReqs {
		call := bindings.Multicall3Call{
			Target:   txReq.To,
			CallData: txReq.Data,
		}
		calls[i] = call
	}

	txRequest, err := (&types.Packer[*bind.MetaData]{Metadata: bindings.Multicall3MetaData}).
		CreateTxRequest(
			mc.contractAddress,
			big.NewInt(0),
			"aggregate",
			calls,
		)
	if err != nil {
		return nil
	}

	return txRequest
}
