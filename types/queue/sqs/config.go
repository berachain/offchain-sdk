package sqs

// Refer to aws.Config for more details.
type Config struct {
	Region      string
	AccessKeyID string
	SecretKey   string
	QueueURL    string
}
