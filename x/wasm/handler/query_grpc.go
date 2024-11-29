package handler

import (
	"fmt"

	msgv1 "cosmossdk.io/api/cosmos/msg/v1"
	queryv1 "cosmossdk.io/api/cosmos/query/v1"
	nfttypes "cosmossdk.io/x/nft"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// GRPCQuerier is a WASM grpc querier.
type GRPCQuerier struct {
	gRPCQueryRouter *baseapp.GRPCQueryRouter
	codec           codec.Codec
	// map[query proto URL]proto response type
	acceptedQueries map[string]func() gogoproto.Message
}

// NewGRPCQuerier returns a new instance of GRPCQuerier.
func NewGRPCQuerier(gRPCQueryRouter *baseapp.GRPCQueryRouter, codec codec.Codec) *GRPCQuerier {
	acceptedQueries := newModuleQuerySafeAllowList()
	// "/cosmos.nft.v1beta1.Query/Owner" is not marked as module_query_safe in cosmos, but we need it
	acceptedQueries["/cosmos.nft.v1beta1.Query/Owner"] = func() gogoproto.Message {
		return &nfttypes.QueryOwnerResponse{}
	}

	return &GRPCQuerier{
		gRPCQueryRouter: gRPCQueryRouter,
		codec:           codec,
		acceptedQueries: acceptedQueries,
	}
}

// Query returns WASM GRPC query handler.
func (q GRPCQuerier) Query(ctx sdk.Context, request *wasmvmtypes.GrpcQuery) (gogoproto.Message, error) {
	protoResponseBuilder, accepted := q.acceptedQueries[request.Path]
	if !accepted {
		return nil, wasmvmtypes.UnsupportedRequest{
			Kind: fmt.Sprintf("'%s' path is not allowed from the contract", request.Path),
		}
	}
	protoResponse := protoResponseBuilder()

	handler := q.gRPCQueryRouter.Route(request.Path)
	if handler == nil {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("No route to query '%s'", request.Path)}
	}

	res, err := handler(ctx, &abci.RequestQuery{
		Data: request.Data,
		Path: request.Path,
	})
	if err != nil {
		return nil, err
	}

	// decode the query response into the expected protobuf message
	err = q.codec.Unmarshal(res.Value, protoResponse)
	if err != nil {
		return nil, err
	}

	return protoResponse, nil
}

// newModuleQuerySafeAllowList returns a map of all query paths labeled with module_query_safe in the proto files to
// their response proto.
func newModuleQuerySafeAllowList() map[string]func() gogoproto.Message {
	fds, err := gogoproto.MergedGlobalFileDescriptors()
	if err != nil {
		panic(err)
	}
	// create the files using 'AllowUnresolvable' to avoid
	// unnecessary panic: https://github.com/cosmos/ibc-go/issues/6435
	protoFiles, err := protodesc.FileOptions{
		AllowUnresolvable: true,
	}.NewFiles(fds)
	if err != nil {
		panic(err)
	}

	allowList := make(map[string]func() gogoproto.Message)
	protoFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		for i := range fd.Services().Len() {
			// Get the service descriptor
			sd := fd.Services().Get(i)

			// Skip services that are annotated with the "cosmos.msg.v1.service" option.
			if ext := proto.GetExtension(sd.Options(), msgv1.E_Service); ext != nil && ext.(bool) {
				continue
			}

			for j := range sd.Methods().Len() {
				// Get the method descriptor
				md := sd.Methods().Get(j)

				// Skip methods that are not annotated with the "cosmos.query.v1.module_query_safe" option.
				if ext := proto.GetExtension(md.Options(), queryv1.E_ModuleQuerySafe); ext == nil || !ext.(bool) {
					continue
				}

				// Add the method to the whitelist
				path := fmt.Sprintf("/%s/%s", sd.FullName(), md.Name())
				out := md.Output()
				allowList[path] = func() gogoproto.Message {
					return dynamicpb.NewMessage(out)
				}
			}
		}
		return true
	})

	return allowList
}
