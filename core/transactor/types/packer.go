package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// Packer struct for packing metadata.
type Packer struct {
	*bind.MetaData
}

// CreateTxRequest function for creating transaction request.
func (p *Packer) CreateTxRequest(
	to common.Address, // address to send transaction to
	value *big.Int, // value to be sent in the transaction (optional)
	gasTipCap *big.Int, // gas tip cap for the transaction (optional)
	gasFeeCap *big.Int, // gas fee cap for the transaction (optional)
	gasLimit uint64, // gas limit for the transaction (optional)
	method string, // method to be called in the transaction
	args ...any, // arguments for the method (optional)
) (*TxRequest, error) { // returns a transaction request or an error
	abi, err := p.GetAbi() // get the ABI from the metadata
	if err != nil {
		return nil, err
	}

	bz, err := abi.Pack(method, args...) // pack the method and arguments into the ABI
	if err != nil {
		return nil, err
	}

	return &TxRequest{
		To:        &to,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		Value:     value,
		Data:      bz,
	}, nil // return a new transaction request
}

// GetCallResponse function for unpacking the return data from a call response.
func (p *Packer) GetCallResponse(method string, ret []byte) ([]any, error) {
	abi, err := p.GetAbi() // get the ABI from the metadata
	if err != nil {
		return nil, err
	}

	return abi.Unpack(method, ret) // unpack the response
}

// MustGetEventSig returns the event signature for the given event name in the packer's ABI.
func (p *Packer) MustGetEventSig(eventName string) common.Hash {
	abi, err := p.GetAbi() // get the ABI from the metadata
	if err != nil {
		panic(err)
	}
	return abi.Events[eventName].ID
}
