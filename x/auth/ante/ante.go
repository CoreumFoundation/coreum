// This content was copied and modified based on github.com/cosmos/cosmos-sdk/x/auth/ante/ante.go
// Original content: https://github.com/cosmos/cosmos-sdk/blob/ad9e5620fb3445c716e9de45cfcdb56e8f1745bf/x/auth/ante/ante.go

package ante

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/CoreumFoundation/coreum/x/auth/keeper"
	"github.com/CoreumFoundation/coreum/x/auth/types"
	feemodelante "github.com/CoreumFoundation/coreum/x/feemodel/ante"
)

// HandlerOptions are the options required for constructing a default SDK AnteHandler.
type HandlerOptions struct {
	DeterministicGasRequirements types.DeterministicGasRequirements
	AccountKeeper                authante.AccountKeeper
	BankKeeper                   authtypes.BankKeeper
	FeegrantKeeper               authante.FeegrantKeeper
	FeeModelKeeper               feemodelante.Keeper
	SignModeHandler              authsigning.SignModeHandler
	SigGasConsumer               func(meter sdk.GasMeter, sig signing.SignatureV2, params authtypes.Params) error
	WasmTXCounterStoreKey        sdk.StoreKey
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}

	if options.FeeModelKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "fee mdoel keeper keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	if options.SigGasConsumer == nil {
		options.SigGasConsumer = authante.DefaultSigVerificationGasConsumer
	}

	if options.WasmTXCounterStoreKey == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "tx counter key is required for ante builder")
	}

	infiniteAccountKeeper := keeper.NewInfiniteAccountKeeper(options.AccountKeeper)

	anteDecorators := []sdk.AnteDecorator{
		authante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		NewSetupGasMeterDecorator(),
		authante.NewRejectExtensionOptionsDecorator(),
		authante.NewValidateBasicDecorator(),
		authante.NewTxTimeoutHeightDecorator(),
		wasmkeeper.NewCountTXDecorator(options.WasmTXCounterStoreKey),
		authante.NewValidateMemoDecorator(options.AccountKeeper),
		feemodelante.NewFeeDecorator(options.FeeModelKeeper),
		authante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper),
		authante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		authante.NewValidateSigCountDecorator(options.AccountKeeper),
		authante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		authante.NewIncrementSequenceDecorator(options.AccountKeeper),
		NewFreeGasDecorator(infiniteAccountKeeper, options.DeterministicGasRequirements),
		authante.NewConsumeGasForTxSizeDecorator(infiniteAccountKeeper),
		authante.NewSigGasConsumeDecorator(infiniteAccountKeeper, options.SigGasConsumer),
		NewFinalGasDecorator(infiniteAccountKeeper, options.DeterministicGasRequirements),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
