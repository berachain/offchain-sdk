package baseapp

import (
	"context"
	"errors"
	"time"

	"github.com/berachain/offchain-sdk/job"
	workertypes "github.com/berachain/offchain-sdk/job/types"
)

// producerTask returns a execution task for the given HasProducer job.
func (jm *JobManager) producerTask(ctx context.Context, wrappedJob job.HasProducer) func() {
	return func() {
		err := wrappedJob.Producer(ctx, jm.jobExecutors)
		if err != nil && !errors.Is(err, context.Canceled) {
			jm.Logger(ctx).Error(
				"error in job producer", "job", wrappedJob.RegistryKey(), "err", err,
			)
		}
	}
}

// retryableSubscriber returns a retryable, execution task for the given Subscribable job.
func (jm *JobManager) retryableSubscriber(
	ctx context.Context, subJob job.Subscribable,
) func() bool {
	numRetries := 1

	return func() bool {
		ch := subJob.Subscribe(ctx)
		jm.Logger(ctx).Info(
			"(re)subscribed to subscription", "job", subJob.RegistryKey(), "retries", numRetries,
		)

		// Ensure that the subscription does not drop due to no messages received.
		staleSubscription := time.After(subscriptionStaleTimeout)

		for {
			select {
			case <-ctx.Done():
				return false // no need to retry
			case staleTime := <-staleSubscription:
				jm.Logger(ctx).Warn(
					"subscription went stale, reconnecting...",
					"time", staleTime, "job", subJob.RegistryKey(),
				)
				numRetries++
				return true // should retry again
			case val := <-ch:
				jm.jobExecutors.Submit(workertypes.NewPayload(ctx, subJob, val).Execute)
				staleSubscription = time.After(subscriptionStaleTimeout)
			}
		}
	}
}

// retryableEthSubscriber returns a retryable, execution task for the given EthSubscribable job.
//
// TODO: cleanup the job types interfaces to avoid overlap using generics.
//
//nolint:dupl // refer to TODO above.
func (jm *JobManager) retryableEthSubscriber(
	ctx context.Context, ethSubJob job.EthSubscribable,
) func() bool {
	numRetries := 1

	return func() bool {
		sub, ch, err := ethSubJob.Subscribe(ctx)
		if err != nil {
			jm.Logger(ctx).Error(
				"error subscribing to filter logs, retrying...",
				"job", ethSubJob.RegistryKey(), "err", err,
			)
			numRetries++
			return true // should retry again
		}
		jm.Logger(ctx).Info(
			"(re)subscribed to eth subscription",
			"job", ethSubJob.RegistryKey(), "retries", numRetries,
		)

		// Ensure that the subscription does not drop due to no messages received.
		staleSubscription := time.After(subscriptionStaleTimeout)

		for {
			select {
			case <-ctx.Done():
				ethSubJob.Unsubscribe(ctx)
				return false // no need to retry
			case err = <-sub.Err():
				jm.Logger(ctx).Error(
					"error in subscription, retrying...",
					"job", ethSubJob.RegistryKey(), "err", err,
				)
				ethSubJob.Unsubscribe(ctx)
				numRetries++
				return true // should retry again
			case staleTime := <-staleSubscription:
				jm.Logger(ctx).Warn(
					"subscription went stale, reconnecting...",
					"time", staleTime, "job", ethSubJob.RegistryKey(),
				)
				numRetries++
				return true // should retry again
			case val := <-ch:
				jm.jobExecutors.Submit(workertypes.NewPayload(ctx, ethSubJob, val).Execute)
				staleSubscription = time.After(subscriptionStaleTimeout)
			}
		}
	}
}

// retryableEthSubscriber returns a retryable, execution task for the given BlockHeaderSub job.
//
// TODO: cleanup the job types interfaces to avoid overlap using generics.
//
//nolint:dupl // refer to TODO above.
func (jm *JobManager) retryableHeaderSubscriber(
	ctx context.Context, blockHeaderJob job.BlockHeaderSub,
) func() bool {
	numRetries := 1

	return func() bool {
		sub, ch, err := blockHeaderJob.Subscribe(ctx)
		if err != nil {
			jm.Logger(ctx).Error(
				"error subscribing block header",
				"job", blockHeaderJob.RegistryKey(), "err", err,
			)
			numRetries++
			return true // should retry again
		}
		jm.Logger(ctx).Info(
			"(re)subscribed to block header sub",
			"job", blockHeaderJob.RegistryKey(), "retries", numRetries,
		)

		// Ensure that the subscription does not drop due to no messages received.
		staleSubscription := time.After(subscriptionStaleTimeout)

		for {
			select {
			case <-ctx.Done():
				blockHeaderJob.Unsubscribe(ctx)
				return false // no need to retry
			case err = <-sub.Err():
				jm.Logger(ctx).Error(
					"error in subscription, retrying...",
					"job", blockHeaderJob.RegistryKey(), "err", err,
				)
				blockHeaderJob.Unsubscribe(ctx)
				numRetries++
				return true // should retry again
			case staleTime := <-staleSubscription:
				jm.Logger(ctx).Warn(
					"subscription went stale, reconnecting...",
					"time", staleTime, "job", blockHeaderJob.RegistryKey(),
				)
				numRetries++
				return true // should retry again
			case val := <-ch:
				jm.jobExecutors.Submit(workertypes.NewPayload(ctx, blockHeaderJob, val).Execute)
				staleSubscription = time.After(subscriptionStaleTimeout)
			}
		}
	}
}
