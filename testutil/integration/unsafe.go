package integration

import (
	"context"
	"testing"
)

type skipUnsafeKey struct{}

// SkipUnsafe skips the test if run-unsafe flag is set to false.
func SkipUnsafe(ctx context.Context, t *testing.T) {
	if GetSkipUnsafe(ctx) {
		t.SkipNow()
	}
}

// WithSkipUnsafe sets the skip unsafe to the context.
func WithSkipUnsafe(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipUnsafeKey{}, true)
}

// GetSkipUnsafe returns  the skip unsafe from the context.
func GetSkipUnsafe(ctx context.Context) bool {
	v, ok := ctx.Value(skipUnsafeKey{}).(bool)
	if !ok {
		return ok
	}
	return v
}
