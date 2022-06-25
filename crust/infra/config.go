package infra

import (
	"regexp"
)

// Config stores configuration
type Config struct {
	// EnvName is the name of created environment
	EnvName string

	// ModeName is the name of the mode
	ModeName string

	// HomeDir is the path where all the files are kept
	HomeDir string

	// AppDir is the path where app data are stored
	AppDir string

	// WrapperDir is the path where wrappers are stored
	WrapperDir string

	// BinDir is the path where all binaries are present
	BinDir string

	// TestFilters are regular expressions used to filter tests to run
	TestFilters []*regexp.Regexp

	// VerboseLogging turns on verbose logging
	VerboseLogging bool
}
