package utils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type CredentialsProvider struct {
	accessKeyID string
	secretKey   string
}

func NewCredentialsProvider(accessKeyID, secretKey string) *CredentialsProvider {
	return &CredentialsProvider{
		accessKeyID: accessKeyID,
		secretKey:   secretKey,
	}
}

func (cp *CredentialsProvider) Retrieve(_ context.Context) (aws.Credentials, error) {
	return aws.Credentials{
		AccessKeyID:     cp.accessKeyID,
		SecretAccessKey: cp.secretKey,
	}, nil
}
