package integration

import (
	"context"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
)

const (
	// AwaitStateTimeout is duration to await for account to have a specific balance.
	AwaitStateTimeout = 30 * time.Second
	// awaitRecheckTimeout is duration between the state recheck.
	awaitRecheckDelay = 100 * time.Millisecond
	// awaitCheckTimeout is timeout for a single check.
	awaitCheckTimeout = 5 * time.Second
)

// AwaitState waits for stateChecker function to rerun nil and retires in case of failure.
func (c ChainContext) AwaitState(ctx context.Context, stateChecker func(ctx context.Context) error) error {
	retryCtx, retryCancel := context.WithTimeout(ctx, AwaitStateTimeout)
	defer retryCancel()
	err := retry.Do(retryCtx, awaitRecheckDelay, func() error {
		checkCtx, checkCtxCancel := context.WithTimeout(retryCtx, awaitCheckTimeout)
		defer checkCtxCancel()
		if err := stateChecker(checkCtx); err != nil {
			return retry.Retryable(err)
		}

		return nil
	})
	return err
}
