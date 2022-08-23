// This content was copied and modified based on github.com/cosmos/cosmos-sdk/x/auth/ante/ante.go
// Original content: https://github.com/cosmos/cosmos-sdk/blob/ad9e5620fb3445c716e9de45cfcdb56e8f1745bf/x/auth/ante/ante.go

package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	feeante "github.com/CoreumFoundation/coreum/x/fee/ante"
)

// HandlerOptions are the options required for constructing a default SDK AnteHandler.
type HandlerOptions struct {
	AccountKeeper      authante.AccountKeeper
	BankKeeper         types.BankKeeper
	FeegrantKeeper     authante.FeegrantKeeper
	GasPriceKeeper     feeante.GasPriceKeeper
	GasCollectorKeeper feeante.GasCollectorKeeper
	SignModeHandler    authsigning.SignModeHandler
	SigGasConsumer     func(meter sdk.GasMeter, sig signing.SignatureV2, params types.Params) error
	GasRequirements    DeterministicGasRequirements
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

	if options.GasPriceKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "gas price keeper is required for ante builder")
	}

	if options.GasCollectorKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "gas collector keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	if options.SigGasConsumer == nil {
		options.SigGasConsumer = authante.DefaultSigVerificationGasConsumer
	}

	anteDecorators := []sdk.AnteDecorator{
		authante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		authante.NewRejectExtensionOptionsDecorator(),
		authante.NewValidateBasicDecorator(),
		authante.NewTxTimeoutHeightDecorator(),
		NewDeterministicGasDecorator(options.GasRequirements),
		authante.NewValidateMemoDecorator(options.AccountKeeper),
		feeante.NewFeeDecorator(options.GasPriceKeeper),
		feeante.NewCollectGasDecorator(options.GasCollectorKeeper),
		authante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		authante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper),
		authante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		authante.NewValidateSigCountDecorator(options.AccountKeeper),
		authante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		authante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		authante.NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
