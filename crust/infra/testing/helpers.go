package testing

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/crust/infra"
)

// WaitUntilHealthy waits until all apps are healthy or context expires
func WaitUntilHealthy(ctx context.Context, t *T, timeout time.Duration, apps ...infra.HealthCheckCapable) {
	waitCtx, waitCancel := context.WithTimeout(ctx, timeout)
	defer waitCancel()
	for _, app := range apps {
		require.NoError(t, infra.WaitUntilHealthy(waitCtx, app))
	}
}
