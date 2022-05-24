package infra

import (
	"context"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/coreznet/pkg/retry"
)

// HealthCheckCapable represents application exposing health check endpoint
type HealthCheckCapable interface {
	// Name returns name of app
	Name() string

	// HealthCheck runs single health check
	HealthCheck(ctx context.Context) error
}

// WaitUntilHealthy waits until app is healthy or context expires
func WaitUntilHealthy(ctx context.Context, apps ...HealthCheckCapable) error {
	for _, app := range apps {
		app := app
		ctx = logger.With(ctx, zap.String("app", app.Name()))
		if err := retry.Do(ctx, time.Second, func() error {
			return app.HealthCheck(ctx)
		}); err != nil {
			return err
		}
	}
	return nil
}
