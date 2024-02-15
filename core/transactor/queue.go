package transactor

import (
	"github.com/berachain/offchain-sdk/core/transactor/types"
	queuetypes "github.com/berachain/offchain-sdk/types/queue/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// WrappedQueue is a wrapper around the queue with helper functions.
type WrappedQueue struct {
	queuetypes.Queue[*types.TxRequest]
}

// Push pushes a transaction request to the queue.
func (wq *WrappedQueue) Push(
	md *bind.MetaData, to common.Address, fn string, args ...interface{},
) (string, error) {
	abi, err := md.GetAbi()
	if err != nil {
		return "", err
	}

	bz, err := abi.Pack(fn, args...)
	if err != nil {
		return "", err
	}

	return wq.Queue.Push(types.NewTxRequest(to, 0, nil, nil, nil, bz))
}
