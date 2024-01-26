package sender

import (
	"context"
	"time"

	coretypes "github.com/ethereum/go-ethereum/core/types"
)

type (
	RetryPolicy func(ctx context.Context,
		tx *coretypes.Transaction, err error) (bool, time.Duration)
)

func DefaultRetryPolicy(
	_ context.Context, _ *coretypes.Transaction, _ error,
) (bool, time.Duration) {
	return false, 30 * time.Second //nolint:gomnd // todo fix later.
}
