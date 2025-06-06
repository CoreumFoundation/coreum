package coreum

import (
	"context"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/CoreumFoundation/crust/build/golang"
	"github.com/CoreumFoundation/crust/build/types"
)

const (
	cosmosSDKModule    = "github.com/cosmos/cosmos-sdk"
	cosmosIBCModule    = "github.com/cosmos/ibc-go/v10"
	cosmosProtoModule  = "github.com/cosmos/cosmos-proto"
	cosmWASMModule     = "github.com/CosmWasm/wasmd"
	gogoProtobufModule = "github.com/cosmos/gogoproto"
	grpcGatewayModule  = "github.com/grpc-ecosystem/grpc-gateway"
)

// Generate regenerates everything in coreum.
func Generate(ctx context.Context, deps types.DepsFunc) error {
	deps(generateProtoDocs, generateProtoGo, generateProtoOpenAPI)

	return golang.Generate(ctx, deps)
}

func protoCDirectories(ctx context.Context, repoPath string, deps types.DepsFunc) (map[string]string, []string, error) {
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	moduleDirs, err := golang.ModuleDirs(ctx, deps, repoPath,
		cosmosSDKModule,
		cosmosIBCModule,
		cosmWASMModule,
		cosmosProtoModule,
		gogoProtobufModule,
		grpcGatewayModule,
	)
	if err != nil {
		return nil, nil, err
	}

	return moduleDirs, []string{
		filepath.Join(absPath, "proto"),
		filepath.Join(absPath, "third_party", "proto"),
		filepath.Join(moduleDirs[cosmosSDKModule], "proto"),
		filepath.Join(moduleDirs[cosmosIBCModule], "proto"),
		filepath.Join(moduleDirs[cosmWASMModule], "proto"),
		filepath.Join(moduleDirs[cosmosProtoModule], "proto"),
		moduleDirs[gogoProtobufModule],
		filepath.Join(moduleDirs[grpcGatewayModule], "third_party", "googleapis"),
	}, nil
}
