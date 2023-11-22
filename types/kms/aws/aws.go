package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	awskmstype "github.com/aws/aws-sdk-go-v2/service/kms/types"
	awsutils "github.com/berachain/offchain-sdk/types/aws"
	"github.com/berachain/offchain-sdk/types/kms/types"
)

// KeyManagementSystem is a wrapper around the AWS KMS client to provide
// additional functionality.
type KeyManagementSystem struct {
	// awsClient is the underlying AWS KMS client instance.
	client *awsClient
}

// NewKeyManagementSystem creates a new KeyManagementSystem instance.
func NewKeyManagementSystemFromConfig(cfg aws.Config) *KeyManagementSystem {
	return &KeyManagementSystem{
		client: newAwsClient(kms.NewFromConfig(cfg)),
	}
}

// NewKeyManagementSystem creates a new KeyManagementSystem instance with the
// provided AWS region and credentials.
func NewKeyManagementSystem(region, accessKeyID, secretKey string) *KeyManagementSystem {
	// Load the default AWS configuration.
	awsCfg, _ := config.LoadDefaultConfig(
		context.Background(), func(cfg *config.LoadOptions) error {
			// Set the AWS region.
			cfg.Region = region
			// Set the AWS credentials.
			cfg.Credentials = awsutils.NewCredentialsProvider(
				accessKeyID, secretKey,
			)
			// Return nil since no error occurred.
			return nil
		})

	// Return a new KeyManagementSystem instance created from the AWS KMS client.
	return NewKeyManagementSystemFromConfig(awsCfg)
}

// GenerateKey uses the AWS KMS to generate a new key.
func (awsKms *KeyManagementSystem) GenerateKey(ctx context.Context) (string, error) {
	input := &kms.CreateKeyInput{
		KeySpec:  awskmstype.KeySpecEccSecgP256k1,
		KeyUsage: awskmstype.KeyUsageTypeSignVerify,
	}

	result, err := awsKms.client.CreateKey(ctx, input)
	if err != nil {
		return "", err
	}
	return *result.KeyMetadata.KeyId, nil
}

// ListKeysByID retrieves all keys managed by the AWS KMS.
func (awsKms *KeyManagementSystem) ListKeysByID(ctx context.Context) ([]string, error) {
	input := &kms.ListKeysInput{}

	result, err := awsKms.client.ListKeys(ctx, input)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(result.Keys))
	for i, key := range result.Keys {
		keys[i] = *key.KeyId
	}
	return keys, nil
}

// GetSigner returns a TxSigner instance that uses the AWS KMS.
func (awsKms *KeyManagementSystem) GetSigner(id string) (types.TxSigner, error) {
	return newSigner(awsKms.client, id)
}
