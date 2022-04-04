module dexapp

go 1.16

require (
	github.com/cosmos/cosmos-sdk v0.44.5
)

replace (
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
