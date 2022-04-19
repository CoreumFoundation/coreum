module build

// 1.16 is used here because still not all distros deliver 1.17 or 1.18.
// Build tool installs newer go, but the tool itself must be built using a preexisting version.
// go 1.18 applies different package grouping in go.mod, so it can't be mixed with earlier versions.
go 1.16

require (
	github.com/CoreumFoundation/coreum-build-tools v0.1.3
	go.uber.org/zap v1.21.0
)
