package testing

import (
	"context"
	"fmt"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/parallel"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/coreznet/infra"
)

type panicError struct {
	value interface{}
	stack []byte
}

func (err panicError) Error() string {
	return fmt.Sprintf("test panicked: %s\n\n%s", err.value, err.stack)
}

// Run deploys testing environment and runs tests there
func Run(ctx context.Context, target infra.Target, mode infra.Mode, tests []*T, filters []*regexp.Regexp) error {
	toRun := make([]*T, 0, len(tests))
	for _, t := range tests {
		if !matchesAny(t.name, filters) {
			continue
		}
		toRun = append(toRun, t)
		if err := t.prepare(ctx); err != nil {
			return err
		}
	}

	if len(toRun) == 0 {
		return errors.New("there are no tests to run")
	}

	if err := target.Deploy(ctx, mode); err != nil {
		return err
	}

	failed := atomic.NewBool(false)
	err := parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
		// The tests themselves are not computationally expensive, most of the time they spend waiting for
		// transactions to be included in blocks so it should be safe to run more tests in parallel than we have CPus
		// available.
		runners := 2 * runtime.NumCPU()
		if runners > len(toRun) {
			runners = len(toRun)
		}

		queue := make(chan *T)
		for i := 0; i < runners; i++ {
			spawn("runner."+strconv.Itoa(i), parallel.Continue, func(ctx context.Context) error {
				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case t, ok := <-queue:
						if !ok {
							return nil
						}
						runTest(logger.With(ctx, zap.String("test", t.name)), t)
						if t.failed {
							failed.Store(true)
						}
					}
				}
			})
		}
		spawn("enqueue", parallel.Continue, func(ctx context.Context) error {
			defer close(queue)

			for _, t := range toRun {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case queue <- t:
				}
			}
			return nil
		})

		return nil
	})
	if err != nil {
		return err
	}
	if failed.Load() {
		return errors.New("tests failed")
	}
	logger.Get(ctx).Info("All tests succeeded")
	return nil
}

func matchesAny(val string, regs []*regexp.Regexp) bool {
	if len(regs) == 0 {
		return true
	}
	for _, reg := range regs {
		if reg.MatchString(val) {
			return true
		}
	}
	return false
}

func runTest(ctx context.Context, t *T) {
	log := logger.Get(ctx)
	log.Info("Test started")
	defer func() {
		log.Info("Test finished")

		r := recover()
		switch {
		// Panic in tested code causes failure of test.
		// Panic caused by T.FailNow is ignored (r != t) as it is used only to exit the test after first failure.
		case r != nil && r != t:
			t.failed = true
			t.errors = append(t.errors, panicError{value: r, stack: debug.Stack()})
			log.Error("Test panicked", zap.Any("panic", r))
		case t.failed:
			for _, e := range t.errors {
				log.Error("Test failed", zap.Error(e))
			}
		default:
			log.Info("Test succeeded")
		}
	}()
	t.run(ctx, t)
}
