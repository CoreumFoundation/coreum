package app

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/grpc"
	googlegrpc "google.golang.org/grpc"

	authtypes "github.com/CoreumFoundation/coreum/x/auth/types"
)

// NewDeterministicGasRouter returns wrapped router charging deterministic amount of gas for defined message types
func NewDeterministicGasRouter(baseRouter sdk.Router, deterministicGasRequirements authtypes.DeterministicGasRequirements) sdk.Router {
	return &deterministicGasRouter{
		baseRouter:                   baseRouter,
		deterministicGasRequirements: deterministicGasRequirements,
	}
}

type deterministicGasRouter struct {
	baseRouter                   sdk.Router
	deterministicGasRequirements authtypes.DeterministicGasRequirements
}

func (r *deterministicGasRouter) AddRoute(route sdk.Route) sdk.Router {
	r.baseRouter.AddRoute(sdk.NewRoute(route.Path(), r.handler(route.Handler())))
	return r
}

func (r *deterministicGasRouter) Route(ctx sdk.Context, path string) sdk.Handler {
	return r.baseRouter.Route(ctx, path)
}

func (r *deterministicGasRouter) handler(baseHandler sdk.Handler) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctxForDeterministicGas(ctx, msg, r.deterministicGasRequirements)
		return baseHandler(ctx, msg)
	}
}

// NewDeterministicMsgServer returns wrapped message server charging deterministic amount of gas for defined message types
func NewDeterministicMsgServer(baseServer grpc.Server, deterministicGasRequirements authtypes.DeterministicGasRequirements) grpc.Server {
	return &deterministicMsgServer{
		baseServer:                   baseServer,
		deterministicGasRequirements: deterministicGasRequirements,
	}
}

type deterministicMsgServer struct {
	baseServer                   grpc.Server
	deterministicGasRequirements authtypes.DeterministicGasRequirements
}

func (s *deterministicMsgServer) RegisterService(sd *googlegrpc.ServiceDesc, handler interface{}) {
	// this is magic
	for i, method := range sd.Methods {
		method := method
		sd.Methods[i].Handler = func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor googlegrpc.UnaryServerInterceptor) (interface{}, error) {
			return method.Handler(srv, ctx, dec, func(ctx context.Context, req interface{}, info *googlegrpc.UnaryServerInfo, handler googlegrpc.UnaryHandler) (resp interface{}, err error) {
				return interceptor(ctx, req, info, func(ctx context.Context, req interface{}) (interface{}, error) {
					sdkCtx := sdk.UnwrapSDKContext(ctx)
					sdkCtx = ctxForDeterministicGas(sdkCtx, req.(sdk.Msg), s.deterministicGasRequirements)
					ctx = sdk.WrapSDKContext(sdkCtx)
					//nolint:contextcheck // Naming sdk functions (sdk.WrapSDKContext) is not our responsibility
					return handler(ctx, req)
				})
			})
		}
	}
	s.baseServer.RegisterService(sd, handler)
}

func ctxForDeterministicGas(ctx sdk.Context, msg sdk.Msg, deterministicGasRequirements authtypes.DeterministicGasRequirements) sdk.Context {
	gasRequired := deterministicGasRequirements.GasRequiredByMessage(msg)
	if gasRequired > 0 {
		ctx.GasMeter().ConsumeGas(gasRequired, fmt.Sprintf("DeterministicGas (gas required: %d, message type: %T)", gasRequired, msg))
		ctx = ctx.WithGasMeter(sdk.NewGasMeter(5 * gasRequired))
	}
	return ctx
}
