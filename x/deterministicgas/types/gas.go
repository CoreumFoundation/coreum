package types

import (
	"context"
	"fmt"

	"github.com/armon/go-metrics"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/grpc"
	"github.com/cosmos/gogoproto/proto"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	googlegrpc "google.golang.org/grpc"

	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v4/x/deterministicgas"
)

const (
	// TODO(dzmitryhil) update to 10 once if fix simulation issue.
	fuseGasMultiplier    = 1000
	expectedMaxGasFactor = 5
)

// NewDeterministicMsgServer returns wrapped message server charging deterministic amount of gas for
// defined message types.
func NewDeterministicMsgServer(
	baseServer grpc.Server,
	deterministicGasConfig deterministicgas.Config,
	assetFTKeeper AssetFTKeeper,
) grpc.Server {
	return &deterministicMsgServer{
		baseServer:             baseServer,
		deterministicGasConfig: deterministicGasConfig,
		assetFTKeeper:          assetFTKeeper,
	}
}

type deterministicMsgServer struct {
	baseServer             grpc.Server
	deterministicGasConfig deterministicgas.Config
	assetFTKeeper          AssetFTKeeper
}

func (s *deterministicMsgServer) RegisterService(sd *googlegrpc.ServiceDesc, handler interface{}) {
	//nolint:lll // the comment contains multiple URLs that cannot be broken down.
	// To understand this implementation it is recommended to study the code in
	// https://github.com/cosmos/cosmos-sdk/blob/ff416ee63d32da5d520a8b2d16b00da762416146/baseapp/msg_service_router.go#L109
	//
	// `sd` argument contains service description generated by protobuf. An example of simple description might be found here:
	// https://github.com/cosmos/cosmos-sdk/blob/ff416ee63d32da5d520a8b2d16b00da762416146/x/crisis/types/tx.pb.go#L208
	// Below, we replace original `Handler` of every message with our wrapper charging constant gas amount.
	// The signature of handler is
	//
	// func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor googlegrpc.UnaryServerInterceptor) (interface{}, error)
	//
	// Handler is called by GRPC framework passing an `interceptor`. The signature of `interceptor` is:
	//
	// func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (resp interface{}, err error)
	//
	// The last argument (`handler`) is the final function which must be called to handle the request.
	// We must call it passing message object as an argument (here called `req`).
	//
	// In original code, Cosmos SDK creates special interceptor which configures sdk context object: https://github.com/cosmos/cosmos-sdk/blob/ff416ee63d32da5d520a8b2d16b00da762416146/baseapp/msg_service_router.go#L111
	// We need to replace gas meter inside that object.
	// To do that we replace the original `Handler` with a function, which receives original `interceptor` created by Cosmos SDK.
	// But we don't call it directly. Instead, we pass our own interceptor function which calls the original one.
	// That interceptor wrapper receives the `handler` argument. But again, instead of calling it directly we pass our function
	// which is called by Cosmos SDK here: https://github.com/cosmos/cosmos-sdk/blob/ff416ee63d32da5d520a8b2d16b00da762416146/baseapp/msg_service_router.go#L113
	// giving us `ctx` containing cosmos context.
	//
	// Then we extract cosmos context from `ctx` replace gas meter, pack it into `ctx` again and hall final handler.

	for i, method := range sd.Methods {
		method := method
		sd.Methods[i].Handler = func(
			srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor googlegrpc.UnaryServerInterceptor,
		) (interface{}, error) {
			return method.Handler(srv, ctx, dec, func(
				ctx context.Context, req interface{}, info *googlegrpc.UnaryServerInfo, handler googlegrpc.UnaryHandler,
			) (resp interface{}, err error) {
				return interceptor(ctx, req, info, func(ctx context.Context, req interface{}) (interface{}, error) {
					sdkCtx := sdk.UnwrapSDKContext(ctx)
					msg := req.(sdk.Msg)
					newSDKCtx, gasBefore, isDeterministic, err := ctxForDeterministicGas(
						sdkCtx,
						msg,
						s.deterministicGasConfig,
						s.assetFTKeeper,
					)
					if err != nil {
						return nil, err
					}

					// gas metrics are reported only if message type is deterministic, and was successful
					// CheckTx and ReCheckTx phases are ignored, since are only interested in the real execution
					// of the message at DeliverTx phase.
					isDeterministicDeliverTx := isDeterministic && !newSDKCtx.IsCheckTx() && !newSDKCtx.IsReCheckTx()
					defer func() {
						// handle case when the expected deterministic message gas multiplied by fuseGasMultiplier exceeded spent gas
						if recoveryObj := recover(); recoveryObj != nil {
							_, isOutOfGasError := recoveryObj.(sdk.ErrorOutOfGas)
							if isOutOfGasError && isDeterministicDeliverTx {
								metrics.AddSampleWithLabels(
									[]string{"deterministic_gas_exceed_fuse_gas_multiplier"},
									1,
									[]metrics.Label{
										{Name: "msg_name", Value: proto.MessageName(msg)},
									})
							}
							// panic one more time to be handled by base app middleware
							panic(recoveryObj)
						}
					}()
					//nolint:contextcheck // Naming sdk functions (sdk.WrapSDKContext) is not our responsibility
					res, err := handler(sdk.WrapSDKContext(newSDKCtx), req)
					if err == nil && isDeterministicDeliverTx {
						if err := reportDeterministicGas(sdkCtx, newSDKCtx, gasBefore, proto.MessageName(msg)); err != nil {
							return nil, err
						}
					}
					return res, err
				})
			})
		}
	}
	s.baseServer.RegisterService(sd, handler)
}

func ctxForDeterministicGas(
	ctx sdk.Context,
	msg sdk.Msg,
	deterministicGasConfig deterministicgas.Config,
	assetFTKeeper AssetFTKeeper,
) (sdk.Context, sdk.Gas, bool, error) {
	gasRequired, exists := deterministicGasConfig.GasRequiredByMessage(msg)
	gasBefore := ctx.GasMeter().GasConsumed()
	if exists {
		hasExtension, err := hasExtensionCall(ctx, msg, assetFTKeeper)
		if err != nil {
			return sdk.Context{}, 0, false, err
		}

		// we consider extensions to be nondeterministic.
		if !hasExtension {
			// Fixed gas is consumed on original gas meter to require and report deterministic gas amount
			ctx.GasMeter().ConsumeGas(
				gasRequired,
				fmt.Sprintf("DeterministicGas (gas required: %d, message type: %T)", gasRequired, msg),
			)

			// We pass much higher amount of gas to handler to be sure that it succeeds.
			// We want to avoid passing infinite gas meter to always have a limit in case of mistake.
			ctx = ctx.WithGasMeter(sdk.NewGasMeter(fuseGasMultiplier * gasRequired))
		}
	}
	return ctx, gasBefore, exists, nil
}

func hasExtensionCall(ctx sdk.Context, msg sdk.Msg, assetFTKeeper AssetFTKeeper) (bool, error) {
	switch typedMsg := msg.(type) {
	case *banktypes.MsgSend:
		for _, coin := range typedMsg.Amount {
			def, err := assetFTKeeper.GetDefinition(ctx, coin.Denom)
			if assetfttypes.ErrInvalidDenom.Is(err) || assetfttypes.ErrTokenNotFound.Is(err) {
				// if the token is not defined in asset ft module, we assume this is different
				// type of token (e.g core, ibc, etc) and don't apply asset ft rules.
				continue
			} else if err != nil {
				return false, err
			}
			if def.IsFeatureEnabled(assetfttypes.Feature_extension) {
				return true, nil
			}
		}
	case *assetfttypes.MsgIssue:
		if lo.Contains(typedMsg.Features, assetfttypes.Feature_extension) {
			return true, nil
		}
	}
	return false, nil
}

func reportDeterministicGas(oldCtx, newCtx sdk.Context, gasBefore sdk.Gas, msgURL string) error {
	deterministicGas := oldCtx.GasMeter().GasConsumed() - gasBefore
	if deterministicGas == 0 {
		return nil
	}

	realGas := newCtx.GasMeter().GasConsumed()

	gasFactor := float32(realGas) / float32(deterministicGas)
	metrics.AddSampleWithLabels([]string{"deterministic_gas_factor"}, gasFactor, []metrics.Label{
		{Name: "msg_name", Value: msgURL},
	})
	if gasFactor > expectedMaxGasFactor {
		metrics.AddSampleWithLabels([]string{"deterministic_gas_factor_exceed_expected_max"}, gasFactor, []metrics.Label{
			{Name: "msg_name", Value: msgURL},
		})
	}

	return errors.WithStack(oldCtx.EventManager().EmitTypedEvent(&EventGas{
		MsgURL:           msgURL,
		RealGas:          realGas,
		DeterministicGas: deterministicGas,
	}))
}
