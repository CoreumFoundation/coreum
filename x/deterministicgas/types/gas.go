package types

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/gogoproto/grpc"
	"github.com/cosmos/gogoproto/proto"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/hashicorp/go-metrics"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	googlegrpc "google.golang.org/grpc"

	testutilconstant "github.com/CoreumFoundation/coreum/v5/testutil/constant"
	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v5/x/deterministicgas"
)

const (
	fuseGasMultiplier         = 10
	simFuseGasMultiplier      = 1000
	expectedMaxGasFactor      = 5
	untrackedMaxGasForQueries = uint64(5_000)
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

	methods := make([]googlegrpc.MethodDesc, len(sd.Methods))
	copy(methods, sd.Methods)
	newSD := *sd
	newSD.Methods = methods

	for i, method := range newSD.Methods {
		newSD.Methods[i].Handler = func(
			srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor googlegrpc.UnaryServerInterceptor,
		) (interface{}, error) {
			return method.Handler(srv, ctx, dec, func(
				ctx context.Context, req interface{}, info *googlegrpc.UnaryServerInfo, handler googlegrpc.UnaryHandler,
			) (resp interface{}, err error) {
				return interceptor(ctx, req, info, func(ctx context.Context, req interface{}) (interface{}, error) {
					sdkCtx := sdk.UnwrapSDKContext(ctx)
					msg := req.(sdk.Msg)
					newSDKCtx, gasBefore, isDeterministic, err := s.ctxForDeterministicGas(
						sdkCtx,
						msg,
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
							_, isOutOfGasError := recoveryObj.(storetypes.ErrorOutOfGas)
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
					//nolint:contextcheck // we consider this correct context passing.
					res, err := handler(newSDKCtx, req)
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
	s.baseServer.RegisterService(&newSD, handler)
}

func (s *deterministicMsgServer) ctxForDeterministicGas(
	ctx sdk.Context,
	msg sdk.Msg,
) (sdk.Context, storetypes.Gas, bool, error) {
	gasRequired, isDeterministic := s.deterministicGasConfig.GasRequiredByMessage(msg)
	gasBefore := ctx.GasMeter().GasConsumed()
	if isDeterministic {
		hasExtension, err := hasExtensionCall(ctx, msg, s.assetFTKeeper)
		if err != nil {
			return sdk.Context{}, 0, false, err
		}

		// we consider extensions to be nondeterministic.
		if hasExtension {
			isDeterministic = false
		}
	}

	if isDeterministic {
		// Fixed gas is consumed on original gas meter to require and report deterministic gas amount
		ctx.GasMeter().ConsumeGas(
			gasRequired,
			fmt.Sprintf("DeterministicGas (gas required: %d, message type: %T)", gasRequired, msg),
		)

		// We pass much higher amount of gas to handler to be sure that it succeeds.
		// We want to avoid passing infinite gas meter to always have a limit in case of mistake.
		gasMultiplier := uint64(fuseGasMultiplier)
		if ctx.ChainID() == testutilconstant.SimAppChainID {
			// simulation fuse gas multiplier is different since during the simulation the modules uses the assetft denom
			// for the cases which are possible for the simulation only and require more gas
			gasMultiplier = simFuseGasMultiplier
		}
		ctx = ctx.WithGasMeter(storetypes.NewGasMeter(gasMultiplier * gasRequired))
	}
	return ctx, gasBefore, isDeterministic, nil
}

// TypeAssertMessages type checks the message to find out that it might invoke asset extensions.
func TypeAssertMessages(msg sdk.Msg) (msgCoins sdk.Coins, hasExtension, notExtensionMsg bool, err error) {
	coins := sdk.NewCoins()
	switch typedMsg := msg.(type) {
	case *banktypes.MsgSend:
		coins = typedMsg.Amount
	case *banktypes.MsgMultiSend:
		for _, input := range typedMsg.Inputs {
			coins = coins.Add(input.Coins...)
		}
	case *distributiontypes.MsgCommunityPoolSpend:
		coins = typedMsg.Amount
	case *distributiontypes.MsgFundCommunityPool:
		coins = typedMsg.Amount
	case *ibctransfertypes.MsgTransfer:
		if typedMsg.Token.IsValid() {
			coins = sdk.NewCoins(typedMsg.Token)
		}
	case *assetfttypes.MsgIssue:
		if lo.Contains(typedMsg.Features, assetfttypes.Feature_extension) {
			return nil, true, false, nil
		}
	case *vestingtypes.MsgCreateVestingAccount:
		coins = typedMsg.Amount
	case *vestingtypes.MsgCreatePermanentLockedAccount:
		coins = typedMsg.Amount
	case *vestingtypes.MsgCreatePeriodicVestingAccount:
		for _, period := range typedMsg.VestingPeriods {
			coins = coins.Add(period.Amount...)
		}
	case *govv1.MsgSubmitProposal:
		coins = typedMsg.InitialDeposit
	case *govv1beta1.MsgSubmitProposal:
		coins = typedMsg.InitialDeposit
	case *govv1.MsgDeposit:
		coins = typedMsg.Amount
	case *govv1beta1.MsgDeposit:
		coins = typedMsg.Amount
	case *authz.MsgExec:
		msgs, err := typedMsg.GetMessages()
		if err != nil {
			return nil, false, true, err
		}
		for _, m := range msgs {
			msgCoins, hasExtension, _, err = TypeAssertMessages(m)
			if err != nil || hasExtension {
				return nil, hasExtension, false, err
			}
			coins = coins.Add(msgCoins...)
		}
	default:
		return nil, false, true, nil
	}

	return coins, false, false, nil
}

func hasExtensionCall(ctx sdk.Context, msg sdk.Msg, assetFTKeeper AssetFTKeeper) (bool, error) {
	coins, hasExtension, _, err := TypeAssertMessages(msg)
	if err != nil || hasExtension {
		return hasExtension, err
	}

	for _, coin := range coins {
		// we should not count the used for this query, otherwise it will mess up the gas
		// requirements of the message with deterministic gas.
		ctxWithUntrackedGas := ctx.WithGasMeter(storetypes.NewGasMeter(fuseGasMultiplier * untrackedMaxGasForQueries))
		def, err := assetFTKeeper.GetDefinition(ctxWithUntrackedGas, coin.Denom)
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
	return false, nil
}

func reportDeterministicGas(oldCtx, newCtx sdk.Context, gasBefore storetypes.Gas, msgURL string) error {
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
