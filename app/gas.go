package app

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/gogo/protobuf/grpc"
	googlegrpc "google.golang.org/grpc"
)

// DefaultDeterministicGasRequirements returns default config for deterministic gas
func DefaultDeterministicGasRequirements() DeterministicGasRequirements {
	return DeterministicGasRequirements{
		BankSend: 80000,
	}
}

// DeterministicGasRequirements specifies gas required by some transaction types
type DeterministicGasRequirements struct {
	BankSend uint64
}

// GasRequiredByMessage returns gas required by a sdk.Msg.
// If fixed gas is not specified for the message type it returns 0.
func (dgr DeterministicGasRequirements) GasRequiredByMessage(msg sdk.Msg) uint64 {
	switch msg.(type) {
	case *banktypes.MsgSend:
		return dgr.BankSend
	default:
		return 0
	}
}

// NewDeterministicGasRouter returns wrapped router charging deterministic amount of gas for defined message types
func NewDeterministicGasRouter(baseRouter sdk.Router, deterministicGasRequirements DeterministicGasRequirements) sdk.Router {
	return &deterministicGasRouter{
		baseRouter:                   baseRouter,
		deterministicGasRequirements: deterministicGasRequirements,
	}
}

type deterministicGasRouter struct {
	baseRouter                   sdk.Router
	deterministicGasRequirements DeterministicGasRequirements
}

func (r *deterministicGasRouter) AddRoute(route sdk.Route) sdk.Router {
	fmt.Println(route.Path())
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
func NewDeterministicMsgServer(baseServer grpc.Server, deterministicGasRequirements DeterministicGasRequirements) grpc.Server {
	return &deterministicMsgServer{
		baseServer:                   baseServer,
		deterministicGasRequirements: deterministicGasRequirements,
	}
}

type deterministicMsgServer struct {
	baseServer                   grpc.Server
	deterministicGasRequirements DeterministicGasRequirements
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
					return handler(ctx, req)
				})
			})
		}
	}
	s.baseServer.RegisterService(sd, handler)
}

func ctxForDeterministicGas(ctx sdk.Context, msg sdk.Msg, deterministicGasRequirements DeterministicGasRequirements) sdk.Context {
	gasRequired := deterministicGasRequirements.GasRequiredByMessage(msg)
	if gasRequired > 0 {
		ctx.GasMeter().ConsumeGas(gasRequired, "DeterministicGas")
		ctx = ctx.WithGasMeter(sdk.NewGasMeter(5 * gasRequired))
	}
	return ctx
}
