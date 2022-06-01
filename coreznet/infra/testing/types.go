package testing

import (
	"context"
	"reflect"
	"regexp"
	"runtime"

	"github.com/pkg/errors"
)

// PrepareFunc defines function which is executed before environment is deployed
type PrepareFunc func(ctx context.Context) error

// RunFunc defines function which is responsible for running the test
type RunFunc func(ctx context.Context, t *T)

// T is the test
type T struct {
	name    string
	prepare PrepareFunc
	run     RunFunc

	errors []error
	failed bool
}

// Errorf stores test error and mark test as failed
func (t *T) Errorf(format string, args ...interface{}) {
	t.failed = true
	t.errors = append(t.errors, errors.Errorf(format, args...))
}

// FailNow marks test as failed and breaks immediately
// This function is called by require.* to break the flow after first unmet condition
func (t *T) FailNow() {
	t.failed = true

	// This panic is used to exit the test immediately. It is neither logged nor breaks the app, test executor recovers from it.
	panic(t)
}

// New creates new test from functions
func New(prepare PrepareFunc, run RunFunc) *T {
	return &T{
		name:    funcToName(prepare),
		prepare: prepare,
		run:     run,
	}
}

var funcToDescriptionRegEx = regexp.MustCompile(`(^.+/|\.func1$)`)

func funcToName(f interface{}) string {
	return funcToDescriptionRegEx.ReplaceAllString(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), "")
}
