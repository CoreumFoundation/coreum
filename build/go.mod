module github.com/CoreumFoundation/coreum/build

go 1.21

replace github.com/CoreumFoundation/coreum/v4 => ../

require (
	github.com/CoreumFoundation/coreum-tools v0.4.1-0.20240321120602-0a9c50facc68
	github.com/CoreumFoundation/coreum/v4 v4.0.0-20240213123712-d7d6a45ddb8f
	github.com/CoreumFoundation/crust/build v0.0.0-20240412142216-b3a803366e4c
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
	// Make sure to not bump x/exp dependency without cosmos-sdk updated because their breaking change is not compatible
	// with cosmos-sdk v0.47.
	// Details: https://github.com/cosmos/cosmos-sdk/issues/18415
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
)
