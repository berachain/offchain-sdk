package aws

import (
	"context"
	"crypto/ecdsa"
	"encoding/asn1"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/berachain/offchain-sdk/types/kms/types"
	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/ethereum/go-ethereum/crypto"
)

// awsClient is a wrapper around the AWS KMS client, that extends it's
// functionality.
type awsClient struct {
	// The underlying AWS KMS client
	*kms.Client
	// Cache for public keys mapped to their KMS key ID
	pubkeyCache *lru.Cache[string, *ecdsa.PublicKey]
}

// newAwsClient creates a new awsClient instance.
func newAwsClient(client *kms.Client) *awsClient {
	// Create new LRU cache for public keys
	pubkeyCache, err := lru.New[string, *ecdsa.PublicKey](pubkeyCacheSize)
	if err != nil {
		panic(err)
	}

	return &awsClient{
		// Set the underlying KMS client
		Client: client,
		// Set the public key cache
		pubkeyCache: pubkeyCache,
	}
}

// GetPubkeyByID retrieves the public key associated with the given KMS key ID.
//
// Parameters:
// - ctx: Context for the operation
// - keyID: The KMS key ID to retrieve the public key for
//
// Returns:
// - *ecdsa.PublicKey: The public key associated with the key ID
// - error: Any error that occurred.
func (c *awsClient) GetPubkeyByID(ctx context.Context, keyID string) (*ecdsa.PublicKey, error) {
	// Check cache for existing public key
	pubkey, found := c.pubkeyCache.Get(keyID)
	if !found {
		// Public key not in cache, retrieve from KMS
		pubKeyBytes, err := c.retrieveDerEncodedPublicKeyFromKms(ctx, keyID)
		if err != nil {
			return nil, err
		}

		// Construct public key from DER-encoded bytes
		pubkey, err = crypto.UnmarshalPubkey(pubKeyBytes)
		if err != nil {
			return nil, err
		}

		// Add public key to cache
		c.pubkeyCache.Add(keyID, pubkey)
	}

	return pubkey, nil
}

// signWithKms signs a transaction hash using AWS KMS.
//
// Parameters:
// - ctx: Context for the operation
// - keyID: The KMS key ID to use for signing
// - txHashBytes: The transaction hash to sign
//
// Returns:
// - rBytes: The R component of the signature
// - sBytes: The S component of the signature
// - error: Any error that occurred.
func (c *awsClient) SignWithKms(
	ctx context.Context, keyID string, txHashBytes []byte,
) ([]byte, []byte, error) {
	// Create KMS Sign input
	signInput := &kms.SignInput{
		KeyId:            aws.String(keyID),
		SigningAlgorithm: awsKmsSignOperationSigningAlgorithm,
		MessageType:      awsKmsSignOperationMessageType,
		Message:          txHashBytes,
	}

	// Call KMS Sign API
	signOutput, err := c.Sign(ctx, signInput)
	if err != nil {
		return nil, nil, err
	}

	// Unmarshal signature into R and S components
	var sigAsn1 types.Asn1EcSig
	_, err = asn1.Unmarshal(signOutput.Signature, &sigAsn1)
	if err != nil {
		return nil, nil, err
	}

	return sigAsn1.R.Bytes, sigAsn1.S.Bytes, nil
}

// retrieveDerEncodedPublicKeyFromKms retrieves the DER-encoded public key from AWS KMS.
//
// Parameters:
// - ctx: Context for the operation
// - keyID: The KMS key ID to retrieve the public key for
//
// Returns:
// - []byte: The DER-encoded public key
// - error: Any error that occurred.
func (c *awsClient) retrieveDerEncodedPublicKeyFromKms(
	ctx context.Context, keyID string,
) ([]byte, error) {
	// Call the KMS GetPublicKey method to retrieve the public key.
	getPubKeyOutput, err := c.GetPublicKey(ctx, &kms.GetPublicKeyInput{
		KeyId: aws.String(keyID),
	})
	if err != nil {
		return nil, fmt.Errorf("[%w] can not get public key from KMS for KeyId=%s", err, keyID)
	}

	// Unmarshal the ASN.1-encoded public key into an asn1EcPublicKey struct.
	var asn1pubk types.Asn1EcPublicKey
	_, err = asn1.Unmarshal(getPubKeyOutput.PublicKey, &asn1pubk)
	if err != nil {
		return nil, fmt.Errorf("[%w] can not parse asn1 public key for KeyId=%s", err, keyID)
	}

	// Return the public key as a byte slice.
	return asn1pubk.PublicKey.Bytes, nil
}
