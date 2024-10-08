package baseapp

import (
	"context"
	"errors"
	"time"

	"github.com/berachain/offchain-sdk/job"
	workertypes "github.com/berachain/offchain-sdk/job/types"
)

// producerTask returns an execution task for the given HasProducer job.
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

// retryableSubscriber returns a retryable execution task for the given Subscribable job.
func (jm *JobManager) retryableSubscriber(
	ctx context.Context, subJob job.Subscribable,
) func() bool {
	numRetries := 1

	return func() bool {
		ch := subJob.Subscribe(ctx)
		jm.Logger(ctx).Info(
			"(re)subscribed to subscription", "job", subJob.RegistryKey(), "retries", numRetries,
		)

		staleSubscription := time.After(subscriptionStaleTimeout)

		for {
			select {
			case <-ctx.Done():
				return false

			case staleTime := <-staleSubscription:
				numRetries++
				jm.Logger(ctx).Warn(
					"subscription went stale, reconnecting...",
					"time", staleTime, "job", subJob.RegistryKey(),
				)
				return true

			case val := <-ch:
				jm.jobExecutors.Submit(workertypes.NewPayload(ctx, subJob, val).Execute)
				staleSubscription = time.After(subscriptionStaleTimeout)
			}
		}
	}
}

// retryableEthSubscriber returns a retryable execution task for the given EthSubscribable job.
func (jm *JobManager) retryableEthSubscriber(
	ctx context.Context, ethSubJob job.EthSubscribable,
) func() bool {
	numRetries := 1

	return func() bool {
		defer func() {
			if sub != nil {
				sub.Unsubscribe()
			}
		}()

		sub, ch, err := ethSubJob.Subscribe(ctx)
		if err != nil {
			numRetries++
			jm.Logger(ctx).Error(
				"error subscribing to filter logs, retrying...",
				"job", ethSubJob.RegistryKey(), "err", err,
			)
			return true
		}

		jm.Logger(ctx).Info(
			"(re)subscribed to eth subscription",
			"job", ethSubJob.RegistryKey(), "retries", numRetries,
		)

		staleSubscription := time.After(subscriptionStaleTimeout)

		for {
			select {
			case <-ctx.Done():
				return false

			case err = <-sub.Err():
				numRetries++
				jm.Logger(ctx).Error(
					"error in subscription, retrying...",
					"job", ethSubJob.RegistryKey(), "err", err,
				)
				return true

			case staleTime := <-staleSubscription:
				numRetries++
				jm.Logger(ctx).Warn(
					"subscription went stale, reconnecting...",
					"time", staleTime, "job", ethSubJob.RegistryKey(),
				)
				return true

			case val := <-ch:
				jm.jobExecutors.Submit(workertypes.NewPayload(ctx, ethSubJob, val).Execute)
				staleSubscription = time.After(subscriptionStaleTimeout)
			}
		}
	}
}

// retryableHeaderSubscriber returns a retryable execution task for the given BlockHeaderSub job.
func (jm *JobManager) retryableHeaderSubscriber(
	ctx context.Context, blockHeaderJob job.BlockHeaderSub,
) func() bool {
	numRetries := 1

	return func() bool {
		defer func() {
			if sub != nil {
				sub.Unsubscribe()
			}
		}()

		sub, ch, err := blockHeaderJob.Subscribe(ctx)
		if err != nil {
			numRetries++
			jm.Logger(ctx).Error(
				"error subscribing to block headers, retrying...",
				"job", blockHeaderJob.RegistryKey(), "err", err,
			)
			return true
		}

		jm.Logger(ctx).Info(
			"(re)subscribed to block header subscription",
			"job", blockHeaderJob.RegistryKey(), "retries", numRetries,
		)

		staleSubscription := time.After(subscriptionStaleTimeout)

		for {
			select {
			case <-ctx.Done():
				return false

			case err = <-sub.Err():
				numRetries++
				jm.Logger(ctx).Error(
					"error in subscription, retrying...",
					"job", blockHeaderJob.RegistryKey(), "err", err,
				)
				return true

			case staleTime := <-staleSubscription:
				numRetries++
				jm.Logger(ctx).Warn(
					"subscription went stale, reconnecting...",
					"time", staleTime, "job", blockHeaderJob.RegistryKey(),
				)
				return true

			case val := <-ch:
				jm.jobExecutors.Submit(workertypes.NewPayload(ctx, blockHeaderJob, val).Execute)
				staleSubscription = time.After(subscriptionStaleTimeout)
			}
		}
	}
}
