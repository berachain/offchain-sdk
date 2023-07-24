package eth

import (
	"sync"
	"github.com/ethereum/go-ethereum/common"

	"context"
)

type NonceManager interface {
	GetNonce(context.Context, common.Address) (uint64, error)
	IncrementNonce(common.Address)
	SetNonce(common.Address, uint64)
	Reset()
}

type nonceManager struct {
	// stores last nonce used for each address
	nonce     map[common.Address]uint64
	ethclient Client
	mutex     sync.Mutex
}

func NewNonceManager() NonceManager {
	return &nonceManager{
		nonce: make(map[common.Address]uint64),
	}
}

func (nm *nonceManager) GetNonce(ctx context.Context, address common.Address) (uint64, error) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	if nonce, ok := nm.nonce[address]; ok {
		return nonce, nil
	}
	return nm.getNonceFromChain(ctx, address)
}

func (nm *nonceManager) IncrementNonce(address common.Address) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	nm.nonce[address] = nm.nonce[address] + 1
	return
}

// Use this function when you want to skip a nonce and put transactions pending in mempool 
func (nm *nonceManager) SetNonce(address common.Address, nonce uint64) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	nm.nonce[address] = nonce
	return
}

func (nm *nonceManager) getNonceFromChain(ctx context.Context, address common.Address) (uint64, error) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	nonce, err := nm.ethclient.PendingNonceAt(ctx, address)
	if err != nil {
		return 0, err
	}
	nm.nonce[address] = nonce
	return nonce, err
}

// Reset is expected to be used if state of nonce manager is ever out of date.
func (nm *nonceManager) Reset() {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	nm.nonce = make(map[common.Address]uint64)
}
