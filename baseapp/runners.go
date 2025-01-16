package baseapp

import (
	"context"
	"errors"
	"time"

	"github.com/berachain/offchain-sdk/v2/job"
	workertypes "github.com/berachain/offchain-sdk/v2/job/types"
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
				// Execute the job with the received value.
				jm.jobExecutors.Submit(workertypes.NewPayload(ctx, subJob, val).Execute)

				// Reset the stale subscription timer since we received a message.
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
		var (
			shouldRetry  bool
			sub, ch, err = ethSubJob.Subscribe(ctx)
		)

		// If retrying update the retry count and unsubscribe the previous subscription.
		defer func() {
			if sub != nil {
				sub.Unsubscribe()
			}
			if shouldRetry {
				numRetries++
			}
		}()

		// Handle error while subscribing.
		if shouldRetry = (err != nil); shouldRetry {
			jm.Logger(ctx).Error(
				"error subscribing to filter logs, retrying...",
				"job", ethSubJob.RegistryKey(), "err", err,
			)
			return shouldRetry
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
				return false // no need to retry

			case err = <-sub.Err():
				jm.Logger(ctx).Error(
					"error in subscription, retrying...",
					"job", ethSubJob.RegistryKey(), "err", err,
				)
				shouldRetry = err != nil
				return shouldRetry

			case staleTime := <-staleSubscription:
				jm.Logger(ctx).Warn(
					"subscription went stale, reconnecting...",
					"time", staleTime, "job", ethSubJob.RegistryKey(),
				)
				shouldRetry = true
				return shouldRetry

			case val := <-ch:
				// Execute the job with the received value.
				jm.jobExecutors.Submit(workertypes.NewPayload(ctx, ethSubJob, val).Execute)

				// Reset the stale subscription timer since we received a message.
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
		var (
			shouldRetry  bool
			sub, ch, err = blockHeaderJob.Subscribe(ctx)
		)

		// If retrying update the retry count and unsubscribe the previous subscription.
		defer func() {
			if sub != nil {
				sub.Unsubscribe()
			}
			if shouldRetry {
				numRetries++
			}
		}()

		// Handle error while subscribing.
		if shouldRetry = (err != nil); shouldRetry {
			jm.Logger(ctx).Error(
				"error subscribing to block headers, retrying...",
				"job", blockHeaderJob.RegistryKey(), "err", err,
			)
			return shouldRetry
		}

		jm.Logger(ctx).Info(
			"(re)subscribed to eth subscription",
			"job", blockHeaderJob.RegistryKey(), "retries", numRetries,
		)

		// Ensure that the subscription does not drop due to no messages received.
		staleSubscription := time.After(subscriptionStaleTimeout)

		for {
			select {
			case <-ctx.Done():
				return false // no need to retry

			case err = <-sub.Err():
				jm.Logger(ctx).Error(
					"error in subscription, retrying...",
					"job", blockHeaderJob.RegistryKey(), "err", err,
				)
				shouldRetry = err != nil
				return shouldRetry

			case staleTime := <-staleSubscription:
				jm.Logger(ctx).Warn(
					"subscription went stale, reconnecting...",
					"time", staleTime, "job", blockHeaderJob.RegistryKey(),
				)
				shouldRetry = true
				return shouldRetry

			case val := <-ch:
				// Execute the job with the received value.
				jm.jobExecutors.Submit(workertypes.NewPayload(ctx, blockHeaderJob, val).Execute)

				// Reset the stale subscription timer since we received a message.
				staleSubscription = time.After(subscriptionStaleTimeout)
			}
		}
	}
}
