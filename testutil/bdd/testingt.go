package bdd

// TestingT is a subset of the public methods implemented by go's testing.T. It allows assertion
// libraries to be used with godog, provided they depend only on this subset of methods.
type TestingT interface {
	// Name returns the name of the current pickle under test
	Name() string
	// Log will log to the current testing.T log if set, otherwise it will log to stdout
	Log(args ...interface{})
	// Logf will log a formatted string to the current testing.T log if set, otherwise it will log
	// to stdout
	Logf(format string, args ...interface{})
	// Error fails the current test and logs the provided arguments. Equivalent to calling Log then
	// Fail.
	Error(args ...interface{})
	// Errorf fails the current test and logs the formatted message. Equivalent to calling Logf then
	// Fail.
	Errorf(format string, args ...interface{})
	// Fail marks the current test as failed, but does not halt execution of the step.
	Fail()
	// FailNow marks the current test as failed and halts execution of the step.
	FailNow()
	// Fatal logs the provided arguments, marks the test as failed and halts execution of the step.
	Fatal(args ...interface{})
	// Fatalf logs the formatted message, marks the test as failed and halts execution of the step.
	Fatalf(format string, args ...interface{})
	// Skip logs the provided arguments and marks the test as skipped but does not halt execution
	// of the step.
	Skip(args ...interface{})
	// Skipf logs the formatted message and marks the test as skipped but does not halt execution
	// of the step.
	Skipf(format string, args ...interface{})
	// SkipNow marks the current test as skipped and halts execution of the step.
	SkipNow()
	// Skipped returns true if the test has been marked as skipped.
	Skipped() bool
}
