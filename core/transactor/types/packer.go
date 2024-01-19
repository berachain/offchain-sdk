package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// Metadata interface for getting ABI.
type Metadata interface {
	GetAbi() (*abi.ABI, error)
}

// Packer struct for packing metadata.
type Packer struct {
	Metadata
}

// CreateTxRequest function for creating transaction request.
func (p *Packer) CreateTxRequest(
	to common.Address, // address to send transaction to
	value *big.Int, // value to be sent in the transaction
	method string, // method to be called in the transaction
	args ...interface{}, // arguments for the method
) (*TxRequest, error) { // returns a transaction request or an error
	abi, err := p.Metadata.GetAbi() // get the ABI from the metadata
	if err != nil {
		return nil, err
	}

	bz, err := abi.Pack(method, args...) // pack the method and arguments into the ABI
	if err != nil {
		return nil, err
	}

	return &TxRequest{
		To:    to,
		Data:  bz,
		Value: value,
	}, nil // return a new transaction request
}
