module github.com/CoreumFoundation/coreum

go 1.16

replace (
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)

require (
	github.com/CoreumFoundation/coreum-tools v0.2.1
	github.com/cosmos/cosmos-sdk v0.45.4
	github.com/cosmos/go-bip39 v1.0.0
	github.com/cosmos/ibc-go/v3 v3.0.0
	github.com/ignite-hq/cli v0.22.1-0.20220610070456-1b33c09fceb7
	github.com/ignite/cli v0.22.2
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.26.1
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.4.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.1
	github.com/tendermint/spn v0.2.1-0.20220610090138-44b136f042c4
	github.com/tendermint/tendermint v0.34.19
	github.com/tendermint/tm-db v0.6.7
	google.golang.org/grpc v1.46.2
)
