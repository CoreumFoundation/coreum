module github.com/CoreumFoundation/coreum/build

// 1.20 is used here because still not all distros deliver 1.21.
// Build tool installs newer go, but the tool itself must be built using a preexisting version.
go 1.21

replace github.com/CoreumFoundation/coreum/v4 => ../

require (
	github.com/CoreumFoundation/coreum-tools v0.4.1-0.20230627094203-821c6a4eebab
	github.com/CoreumFoundation/coreum/v4 v4.0.0-20231128071941-710434b13177
	github.com/CoreumFoundation/crust/build v0.0.0-20240125130501-e91f8161839f
	github.com/iancoleman/strcase v0.3.0
	github.com/pkg/errors v0.9.1
	go.uber.org/zap v1.24.0
	golang.org/x/mod v0.12.0
)

require (
	github.com/bufbuild/buf v1.26.1 // indirect
	github.com/bufbuild/connect-go v1.9.0 // indirect
	github.com/bufbuild/connect-opentelemetry-go v0.4.0 // indirect
	github.com/bufbuild/protocompile v0.6.0 // indirect
	github.com/cosmos/gogoproto v1.4.10 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/felixge/fgprof v0.9.3 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gofrs/uuid/v5 v5.0.0 // indirect
	github.com/golang/glog v1.1.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/pprof v0.0.0-20230705174524-200ffdc848b8 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.17.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jdxcode/netrc v0.0.0-20221124155335-4616370d1a84 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/pkg/profile v1.7.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/samber/lo v1.38.1 // indirect
	github.com/spf13/cobra v1.7.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.opentelemetry.io/otel v1.16.0 // indirect
	go.opentelemetry.io/otel/metric v1.16.0 // indirect
	go.opentelemetry.io/otel/sdk v1.16.0 // indirect
	go.opentelemetry.io/otel/trace v1.16.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/exp v0.0.0-20230713183714-613f0c0eb8a1 // indirect
	golang.org/x/sync v0.4.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/term v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto v0.0.0-20231120223509-83a465c0220f // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231127180814-3a041ad873d4 // indirect
	google.golang.org/grpc v1.59.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
