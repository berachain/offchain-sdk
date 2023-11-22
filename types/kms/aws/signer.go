package aws

import (
	"bytes"
	"context"
	"math/big"

	kmstypes "github.com/berachain/offchain-sdk/types/kms/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

// Ensure `signer` implements the TxSigner interface.
var _ kmstypes.TxSigner = (*signer)(nil)

// signer implements the bind.SignerFn interface for signing transactions
// using AWS KMS.
type signer struct {
	// The AWS KMS key ID to use for signing.
	keyID string
	// The AWS KMS client.
	client *awsClient
	// The publickey bytes associated with the KMS key ID.
	pubKeyBytes []byte
	// CacheAddress to prevent re-deriving the address every time.
	addressCache common.Address
}

// newSigner creates a new signer instance.
//
// Parameters:
// - client: The AWS KMS client
// - keyID: The KMS key ID to use for signing
//
// Returns:
// - *signer: The new signer instance.
func newSigner(client *awsClient, keyID string) (*signer, error) {
	pubkey, err := client.GetPubkeyByID(context.Background(), keyID)
	if err != nil {
		return nil, err
	}

	return &signer{
		client:       client,
		keyID:        keyID,
		pubKeyBytes:  secp256k1.S256().Marshal(pubkey.X, pubkey.Y),
		addressCache: crypto.PubkeyToAddress(*pubkey),
	}, nil
}

// Address returns the Ethereum address for the signer's public key.
func (s *signer) Address() common.Address {
	return s.addressCache
}

// SignerFunc returns a SignerFn that can sign Ethereum transactions
// using the AWS KMS.
//
// Parameters:
// - ctx: Context for the operation
// - chainID: Ethereum chain ID
//
// Returns:
// - bind.SignerFn: Signer function for use with contract bindings
// - error: Any error that occurred.
func (s *signer) SignerFunc(ctx context.Context, chainID *big.Int) (bind.SignerFn, error) {
	signer := types.LatestSignerForChainID(chainID)
	return func(addr common.Address, tx *types.Transaction) (*types.Transaction, error) {
		if addr != s.Address() {
			return nil, bind.ErrNotAuthorized
		}

		txHashBytes := signer.Hash(tx).Bytes()

		rBytes, sBytes, err := s.client.SignWithKms(ctx, s.keyID, txHashBytes)
		if err != nil {
			return nil, err
		}

		// Adjust S value from signature according to Ethereum standard
		sBigInt := new(big.Int).SetBytes(sBytes)
		if sBigInt.Cmp(secp256k1HalfN) > 0 {
			sBytes = new(big.Int).Sub(secp256k1N, sBigInt).Bytes()
		}

		signature, err := assembleEthereumSignature(s.pubKeyBytes, txHashBytes, rBytes, sBytes)
		if err != nil {
			return nil, err
		}

		return tx.WithSignature(signer, signature)
	}, nil
}

// assembleEthereumSignature assembles a signature from R, S and public key.
//
// Parameters:
// - expectedPublicKeyBytes: Expected public key
// - txHash: Transaction hash
// - r: Signature R value
// - s: Signature S value
//
// Returns:
// - []byte: The assembled Ethereum signature
// - error: Any error that occurred.
func assembleEthereumSignature(
	expectedPublicKeyBytes []byte, txHash []byte, r []byte, s []byte,
) ([]byte, error) {
	rsSignature := append(padBufferTo32Bytes(r), padBufferTo32Bytes(s)...)
	signature := appendByteToSignature(rsSignature, 0)

	recoveredPublicKeyBytes, err := crypto.Ecrecover(txHash, signature)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(recoveredPublicKeyBytes, expectedPublicKeyBytes) {
		signature = appendByteToSignature(rsSignature, 1)
		recoveredPublicKeyBytes, err = crypto.Ecrecover(txHash, signature)
		if err != nil {
			return nil, err
		}

		if !bytes.Equal(recoveredPublicKeyBytes, expectedPublicKeyBytes) {
			return nil, ErrPublicKeyReconstruction
		}
	}

	return signature, nil
}

// appendByteToSignature appends a byte to a signature.
func appendByteToSignature(signature []byte, b byte) []byte {
	return append(signature, b)
}

// padBufferTo32Bytes pads a byte buffer to 32 bytes.
func padBufferTo32Bytes(buffer []byte) []byte {
	// Trim leading zeros from the buffer
	buffer = bytes.TrimLeft(buffer, "\x00")
	// Prepend zero bytes to the buffer until its length is 32 bytes
	for len(buffer) < 32 {
		zeroBuf := []byte{0}
		buffer = append(zeroBuf, buffer...)
	}
	return buffer
}
