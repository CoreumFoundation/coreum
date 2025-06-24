package bdd

import (
	"cosmossdk.io/log"
)

type THelper = interface {
	Helper()
}
type logWrapper struct {
	innerLogger TestingT
}

func NewLogWrapper(logger TestingT) log.TestingT {
	return logWrapper{logger}
}

func (a logWrapper) Log(args ...interface{}) {
	a.innerLogger.Log(args...)
}

func (a logWrapper) Logf(format string, args ...interface{}) {
	a.innerLogger.Logf(format, args...)
}

func (a logWrapper) Helper() {
	if h, ok := a.innerLogger.(THelper); ok {
		h.Helper()
	}
}
