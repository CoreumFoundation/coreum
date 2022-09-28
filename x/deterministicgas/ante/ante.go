package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	"github.com/CoreumFoundation/coreum/pkg/config"
)

// SetInfiniteGasMeterDecorator sets the infinite gas limit for ante handler
// CONTRACT: Must be the first decorator in the chain
// CONTRACT: Tx must implement GasTx interface
type SetInfiniteGasMeterDecorator struct {
	deterministicGasRequirements config.DeterministicGasRequirements
}

// NewSetInfiniteGasMeterDecorator creates new SetInfiniteGasMeterDecorator
func NewSetInfiniteGasMeterDecorator(deterministicGasRequirements config.DeterministicGasRequirements) SetInfiniteGasMeterDecorator {
	return SetInfiniteGasMeterDecorator{
		deterministicGasRequirements: deterministicGasRequirements,
	}
}

// AnteHandle resets the gas limit inside GasMeter
func (sigmd SetInfiniteGasMeterDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// This is done to return an error early if user provided gas amount which can't even cover the constant fee charged on the real
	// gas meter in `ChargeFixedGasDecorator`. This will save resources on running preliminary ante decorators.
	ctx.GasMeter().ConsumeGas(sigmd.deterministicGasRequirements.FixedGas, "Fixed")

	// Set infinite gas meter for ante handler
	return next(ctx.WithGasMeter(sdk.NewInfiniteGasMeter()), tx, simulate)
}

// AddBaseGasDecorator adds free gas to gas meter
// CONTRACT: Tx must implement GasTx interface
type AddBaseGasDecorator struct {
	ak                           authante.AccountKeeper
	deterministicGasRequirements config.DeterministicGasRequirements
}

// NewAddBaseGasDecorator creates new AddBaseGasDecorator
func NewAddBaseGasDecorator(ak authante.AccountKeeper, deterministicGasRequirements config.DeterministicGasRequirements) AddBaseGasDecorator {
	return AddBaseGasDecorator{
		ak:                           ak,
		deterministicGasRequirements: deterministicGasRequirements,
	}
}

// AnteHandle resets the gas limit inside GasMeter
func (abgd AddBaseGasDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	var gasMeter sdk.GasMeter
	if simulate || ctx.BlockHeight() == 0 {
		// During simulation and genesis initialization infinite gas meter is set inside context by `SetUpContextDecorator`.
		// Here, we reset it to initial state with 0 gas consumed.
		gasMeter = sdk.NewInfiniteGasMeter()
	} else {
		params := abgd.ak.GetParams(ctx)

		// It is not needed to verify that tx really implements `GasTx` interface because it has been already done by
		// `SetUpContextDecorator`
		gasTx := tx.(authante.GasTx)
		gasMeter = sdk.NewGasMeter(gasTx.GetGas() + abgd.deterministicGasRequirements.TxBaseGas(params))
	}
	return next(ctx.WithGasMeter(gasMeter), tx, simulate)
}

// ChargeFixedGasDecorator sets gas meter for message handlers
// CONTRACT: Tx must implement GasTx interface
type ChargeFixedGasDecorator struct {
	ak                           authante.AccountKeeper
	deterministicGasRequirements config.DeterministicGasRequirements
}

// NewChargeFixedGasDecorator creates new ChargeFixedGasDecorator
func NewChargeFixedGasDecorator(ak authante.AccountKeeper, deterministicGasRequirements config.DeterministicGasRequirements) ChargeFixedGasDecorator {
	return ChargeFixedGasDecorator{
		ak:                           ak,
		deterministicGasRequirements: deterministicGasRequirements,
	}
}

// AnteHandle resets the gas limit inside GasMeter
func (cfgd ChargeFixedGasDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// It is not needed to verify that tx really implements `GasTx` interface because it has been already done by
	// `SetUpContextDecorator`
	gasTx := tx.(authante.GasTx)

	params := cfgd.ak.GetParams(ctx)

	var gasMeter sdk.GasMeter
	if simulate || ctx.BlockHeight() == 0 {
		// During simulation and genesis initialization infinite gas meter is set inside context by `SetUpContextDecorator`.
		// We reset it to initial state with 0 gas consumed.
		gasMeter = sdk.NewInfiniteGasMeter()
	} else {
		gasMeter = sdk.NewGasMeter(gasTx.GetGas())
	}

	gasConsumed := ctx.GasMeter().GasConsumed()
	bonus := cfgd.deterministicGasRequirements.TxBaseGas(params)
	if gasConsumed > bonus {
		gasMeter.ConsumeGas(gasConsumed-bonus, "OverBonus")
	}
	gasMeter.ConsumeGas(cfgd.deterministicGasRequirements.FixedGas, "Fixed")

	return next(ctx.WithGasMeter(gasMeter), tx, simulate)
}
