package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	"github.com/CoreumFoundation/coreum/pkg/config"
)

// SetupGasMeterDecorator sets the infinite gas limit for ante handler
// CONTRACT: Must be the first decorator in the chain
// CONTRACT: Tx must implement GasTx interface
// FIXME (wojtek): THIS IS BAD, used only for testing
type SetupGasMeterDecorator struct{}

// NewSetupGasMeterDecorator creates new SetupGasMeterDecorator
func NewSetupGasMeterDecorator() SetupGasMeterDecorator {
	return SetupGasMeterDecorator{}
}

// AnteHandle resets the gas limit inside GasMeter
func (sgmd SetupGasMeterDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// Set infinite gas meter for ante handler
	return next(ctx.WithGasMeter(sdk.NewInfiniteGasMeter()), tx, simulate)
}

// FreeGasDecorator adds free gas to gas meter
// CONTRACT: Tx must implement GasTx interface
type FreeGasDecorator struct {
	ak                           authante.AccountKeeper
	deterministicGasRequirements config.DeterministicGasRequirements
}

// NewFreeGasDecorator creates new FreeGasDecorator
func NewFreeGasDecorator(ak authante.AccountKeeper, deterministicGasRequirements config.DeterministicGasRequirements) FreeGasDecorator {
	return FreeGasDecorator{
		ak:                           ak,
		deterministicGasRequirements: deterministicGasRequirements,
	}
}

// AnteHandle resets the gas limit inside GasMeter
func (fgd FreeGasDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	var gasMeter sdk.GasMeter
	if simulate || ctx.BlockHeight() == 0 {
		// During simulation and genesis initialization infinite gas meter is set inside context by `SetUpContextDecorator`.
		// Here, we reset it to initial state with 0 gas consumed.
		gasMeter = sdk.NewInfiniteGasMeter()
	} else {
		params := fgd.ak.GetParams(ctx)

		// It is not needed to verify that tx really implements `GasTx` interface because it has been already done by
		// `SetUpContextDecorator`
		gasTx := tx.(authante.GasTx)

		gasMeter = sdk.NewGasMeter(gasTx.GetGas() + fgd.deterministicGasRequirements.TxBonusGas(params))
	}
	return next(ctx.WithGasMeter(gasMeter), tx, simulate)
}

// FinalGasDecorator sets gas meter for message handlers
// CONTRACT: Tx must implement GasTx interface
type FinalGasDecorator struct {
	ak                           authante.AccountKeeper
	deterministicGasRequirements config.DeterministicGasRequirements
}

// NewFinalGasDecorator creates new FinalGasDecorator
func NewFinalGasDecorator(ak authante.AccountKeeper, deterministicGasRequirements config.DeterministicGasRequirements) FinalGasDecorator {
	return FinalGasDecorator{
		ak:                           ak,
		deterministicGasRequirements: deterministicGasRequirements,
	}
}

// AnteHandle resets the gas limit inside GasMeter
func (fgd FinalGasDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// It is not needed to verify that tx really implements `GasTx` interface because it has been already done by
	// `SetUpContextDecorator`
	gasTx := tx.(authante.GasTx)

	params := fgd.ak.GetParams(ctx)

	var gasMeter sdk.GasMeter
	if simulate || ctx.BlockHeight() == 0 {
		// During simulation and genesis initialization infinite gas meter is set inside context by `SetUpContextDecorator`.
		// We reset it to initial state with 0 gas consumed.
		gasMeter = sdk.NewInfiniteGasMeter()
	} else {
		gasMeter = sdk.NewGasMeter(gasTx.GetGas())
	}

	gasConsumed := ctx.GasMeter().GasConsumed()
	bonus := fgd.deterministicGasRequirements.TxBonusGas(params)
	if gasConsumed > bonus {
		gasMeter.ConsumeGas(gasConsumed-bonus, "OverBonus")
	}
	gasMeter.ConsumeGas(fgd.deterministicGasRequirements.FixedGas, "Fixed")

	return next(ctx.WithGasMeter(gasMeter), tx, simulate)
}
