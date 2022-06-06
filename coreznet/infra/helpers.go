package infra

import (
	"context"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/pkg/errors"
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

// AppWithInfo represents application which is able to return information about its deployment
type AppWithInfo interface {
	// Name returns name of app
	Name() string

	// Info returns information about application's deployment
	Info() DeploymentInfo
}

// IsRunning returns a health check which succeeds if application is running
func IsRunning(app AppWithInfo) HealthCheckCapable {
	return isRunningHealthCheck{app: app}
}

type isRunningHealthCheck struct {
	app AppWithInfo
}

// Name returns name of app
func (hc isRunningHealthCheck) Name() string {
	return hc.app.Name()
}

// HealthCheck runs single health check
func (hc isRunningHealthCheck) HealthCheck(ctx context.Context) error {
	if hc.app.Info().Status == AppStatusRunning {
		return nil
	}
	return retry.Retryable(errors.New("application hasn't been started yet"))
}
