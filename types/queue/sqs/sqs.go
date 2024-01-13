package sqs

import (
	"context"
	"reflect"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	awsutils "github.com/berachain/offchain-sdk/types/aws"
	"github.com/berachain/offchain-sdk/types/queue/types"
)

// SQSClient is an interface that defines the necessary methods for interacting
// with the SQS service.
type Client interface {
	SendMessage(ctx context.Context,
		params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	ReceiveMessage(ctx context.Context,
		params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context,
		params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
	GetQueueAttributes(ctx context.Context, params *sqs.GetQueueAttributesInput,
		optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error)
}

// Queue is a wrapper struct around the SQS API.
type Queue[T types.Marshallable] struct {
	svc         Client
	queueURL    string
	inProcessMu *sync.RWMutex
	inProcess   map[string]string
}

// NewQueue creates a new SQS object with the specified queue URL.
func NewQueueFromConfig[T types.Marshallable](cfg aws.Config, queueURL string) (*Queue[T], error) {
	return &Queue[T]{
		svc:         sqs.NewFromConfig(cfg),
		queueURL:    queueURL,
		inProcessMu: new(sync.RWMutex),
		inProcess:   make(map[string]string),
	}, nil
}

func NewQueue[T types.Marshallable](
	region, accessKeyID, secretKey, queueURL string,
) (*Queue[T], error) {
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

	return NewQueueFromConfig[T](awsCfg, queueURL)
}

// Push adds an item to the SQS queue.
func (q *Queue[T]) Push(item T) (string, error) {
	// Marshal the item
	bz, err := item.Marshal()
	if err != nil {
		return "", err
	}

	// Send the message to the SQS queue with the provided context
	str := string(bz)
	output, err := q.svc.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    &q.queueURL,
		MessageBody: &str,
	})

	if err != nil {
		return "", err
	}

	if output == nil || output.MessageId == nil {
		return "", err
	}

	return *output.MessageId, nil
}

// Pop retrieves an item from the SQS queue.
func (q *Queue[T]) Receive() (string, T, bool) {
	var t2 T
	t1 := reflect.TypeOf(t2).Elem()
	newInstance := reflect.New(t1).Interface()
	t, _ := newInstance.(T)

	// Receive a message from the SQS queue
	resp, err := q.svc.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            &q.queueURL,
		MaxNumberOfMessages: 1,
	})
	if err != nil {
		return "", t, false
	}

	// Check if a message was received
	if len(resp.Messages) == 0 {
		return "", t, false
	}

	// Unmarshal the message into a new instance of type T
	if err = t.Unmarshal([]byte(*resp.Messages[0].Body)); err != nil {
		return "", t, false
	}

	// Add to the inProcess MessageID queue, mark the Message as in Process.
	// TODO memory growth atm.
	q.inProcess[*resp.Messages[0].MessageId] = *resp.Messages[0].ReceiptHandle

	return *resp.Messages[0].MessageId, t, true
}

func (q *Queue[T]) ReceiveMany(num int32) ([]string, []T, error) {
	ts := make([]T, 0)
	msgIDs := make([]string, 0)

	// Receive a message from the SQS queue
	resp, err := q.svc.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            &q.queueURL,
		MaxNumberOfMessages: num,
	})

	if err != nil {
		return msgIDs, ts, err
	}
	// Check if a message was received
	if len(resp.Messages) == 0 {
		return msgIDs, ts, nil
	}

	for _, m := range resp.Messages {
		var t2 T
		t1 := reflect.TypeOf(t2).Elem()
		newInstance := reflect.New(t1).Interface()
		t, _ := newInstance.(T)

		// Unmarshal the message into a new instance of type T
		if err = t.Unmarshal([]byte(*m.Body)); err != nil {
			return nil, nil, err
		}

		// Add to the inProcess MessageID queue, mark the Message as in Process.
		// TODO memory growth atm.
		q.inProcessMu.Lock()
		q.inProcess[*m.MessageId] = *m.ReceiptHandle
		q.inProcessMu.Unlock()

		ts = append(ts, t)
		msgIDs = append(msgIDs, *m.MessageId)
	}
	if err != nil {
		return nil, nil, err
	}

	return msgIDs, ts, nil
}

func (q *Queue[T]) Len() int {
	// Delete from the queue to mark as complete.
	resp, err := q.svc.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl: &q.queueURL,
		AttributeNames: []sqstypes.QueueAttributeName{
			"ApproximateNumberOfMessages",
		},
	})
	if err != nil {
		return 0
	}
	anm := resp.Attributes["ApproximateNumberOfMessages"]
	val, _ := strconv.ParseInt(anm, 10, 64)
	return int(val)
}

func (q *Queue[T]) Delete(messageID string) error {
	return q.deleteMessage(messageID)
}

func (q *Queue[T]) deleteMessage(messageID string) error {
	// Grab the latest receipt handle by the messageID.
	q.inProcessMu.RLock()
	receiptHandle := q.inProcess[messageID]
	q.inProcessMu.RUnlock()

	// Delete from the queue to mark as complete.
	_, err := q.svc.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
		QueueUrl:      &q.queueURL,
		ReceiptHandle: &receiptHandle,
	})
	if err != nil {
		return err
	}

	// remove the messageID from the inProcess map.
	q.inProcessMu.Lock()
	defer q.inProcessMu.Unlock()
	delete(q.inProcess, messageID)

	return nil
}
