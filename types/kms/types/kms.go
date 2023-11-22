package types

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// KeyManagementSystem is an interface that combines the KeyGenerator and KeyViewer interfaces.
type KeyManagementSystem interface {
	KeyViewer
}

// KeyViewer is an interface that defines a method for retrieving a
// TxSigner instance associated with a key ID.
type KeyViewer interface {
	// ListKeysByID returns a list of all keys managed by the KMS.
	ListKeysByID() ([]string, error)
	// GetSigner returns a TxSigner instance associated with the given key ID.
	GetSigner(id string) TxSigner
}

// TxSigner is an interface that defines a method for signing a transaction generated from a bind.
// The Address method returns the Ethereum address associated with the signer.
// The SignerFunc method returns a function that can be used to sign a transaction generated
// from a bind.
type TxSigner interface {
	// Address returns the Ethereum address associated with the signer.
	Address() common.Address
	// SignerFunc returns a function that can be used to sign a transaction generated from a bind.
	SignerFunc(context.Context, *big.Int) (bind.SignerFn, error)
}
