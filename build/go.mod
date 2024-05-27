module github.com/CoreumFoundation/coreum/build

go 1.21

replace (
	github.com/CoreumFoundation/coreum/v4 => ../
	// Make sure to not bump x/exp dependency without cosmos-sdk updated because their breaking change is not compatible
	// with cosmos-sdk v0.47.
	// Details: https://github.com/cosmos/cosmos-sdk/issues/18415
	golang.org/x/exp => golang.org/x/exp v0.0.0-20230711153332-06a737ee72cb
)

require (
	github.com/CoreumFoundation/coreum-tools v0.4.1-0.20240321120602-0a9c50facc68
	github.com/CoreumFoundation/coreum/v4 v4.0.0-20240213123712-d7d6a45ddb8f
	github.com/CoreumFoundation/crust/build v0.0.0-20240527125419-8c85e2cfdda9
	github.com/iancoleman/strcase v0.3.0
	github.com/pkg/errors v0.9.1
)

require (
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/mod v0.17.0 // indirect
)

require (
	github.com/BurntSushi/toml v1.3.2 // indirect
	github.com/samber/lo v1.39.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa // indirect
)
