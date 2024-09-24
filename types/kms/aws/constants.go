package aws

import (
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
)

// Constants used for defining the size of the public key cache,
// the message type for an AWS KMS sign operation,
// and the signing algorithm for the AWS KMS sign operation.
const (
	pubkeyCacheSize                     = 128
	awsKmsSignOperationMessageType      = "DIGEST"
	awsKmsSignOperationSigningAlgorithm = "ECDSA_SHA_256"
)

var (
	// secp256k1N is a global variable that represents the N parameter of the
	// crypto.S256() elliptic curve.
	secp256k1N = crypto.S256().Params().N //nolint:gochecknoglobals // This is a constant.

	// secp256k1HalfN is a global variable that represents half of the secp256k1N value.
	secp256k1HalfN = new(big.Int).
			Div(secp256k1N, big.NewInt(2)) //nolint:gochecknoglobals,mnd // constant.
)
