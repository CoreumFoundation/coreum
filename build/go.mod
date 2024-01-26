module github.com/CoreumFoundation/coreum/build

// 1.20 is used here because still not all distros deliver 1.21.
// Build tool installs newer go, but the tool itself must be built using a preexisting version.
go 1.21

replace github.com/CoreumFoundation/coreum/v4 => ../

require (
	github.com/CoreumFoundation/coreum-tools v0.4.1-0.20230627094203-821c6a4eebab
	github.com/CoreumFoundation/coreum/v4 v4.0.0-20231128071941-710434b13177
	github.com/CoreumFoundation/crust/build v0.0.0-20240126180544-9d6145e83184
	github.com/iancoleman/strcase v0.3.0
	github.com/pkg/errors v0.9.1
	go.uber.org/zap v1.24.0
	golang.org/x/mod v0.12.0
)

require (
	github.com/samber/lo v1.38.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20230713183714-613f0c0eb8a1 // indirect
)
