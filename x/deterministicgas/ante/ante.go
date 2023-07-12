package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	"github.com/CoreumFoundation/coreum/v2/x/deterministicgas"
)

// SetInfiniteGasMeterDecorator sets the infinite gas limit for ante handler
// CONTRACT: Must be the first decorator in the chain.
// CONTRACT: Tx must implement GasTx interface.
type SetInfiniteGasMeterDecorator struct {
	deterministicGasConfig deterministicgas.Config
}

// NewSetInfiniteGasMeterDecorator creates new SetInfiniteGasMeterDecorator.
func NewSetInfiniteGasMeterDecorator(deterministicGasConfig deterministicgas.Config) SetInfiniteGasMeterDecorator {
	return SetInfiniteGasMeterDecorator{
		deterministicGasConfig: deterministicGasConfig,
	}
}

// AnteHandle resets the gas limit inside GasMeter.
func (sigmd SetInfiniteGasMeterDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// This is done to return an error early if user provided gas amount which can't even cover the constant fee charged on the real
	// gas meter in `ChargeFixedGasDecorator`. This will save resources on running preliminary ante decorators.
	ctx.GasMeter().ConsumeGas(sigmd.deterministicGasConfig.FixedGas, "Fixed")

	// Set infinite gas meter for ante handler
	return next(ctx.WithGasMeter(sdk.NewInfiniteGasMeter()), tx, simulate)
}

// AddBaseGasDecorator adds free gas to gas meter.
// CONTRACT: Tx must implement GasTx interface.
type AddBaseGasDecorator struct {
	ak                     authante.AccountKeeper
	deterministicGasConfig deterministicgas.Config
}

// NewAddBaseGasDecorator creates new AddBaseGasDecorator.
func NewAddBaseGasDecorator(ak authante.AccountKeeper, deterministicGasConfig deterministicgas.Config) AddBaseGasDecorator {
	return AddBaseGasDecorator{
		ak:                     ak,
		deterministicGasConfig: deterministicGasConfig,
	}
}

// AnteHandle resets the gas limit inside GasMeter.
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
		gasMeter = sdk.NewGasMeter(gasTx.GetGas() + abgd.deterministicGasConfig.TxBaseGas(params))
	}
	return next(ctx.WithGasMeter(gasMeter), tx, simulate)
}

// ChargeFixedGasDecorator sets gas meter for message handlers.
// CONTRACT: Tx must implement GasTx interface.
type ChargeFixedGasDecorator struct {
	ak                     authante.AccountKeeper
	deterministicGasConfig deterministicgas.Config
}

// NewChargeFixedGasDecorator creates new ChargeFixedGasDecorator.
func NewChargeFixedGasDecorator(ak authante.AccountKeeper, deterministicGasConfig deterministicgas.Config) ChargeFixedGasDecorator {
	return ChargeFixedGasDecorator{
		ak:                     ak,
		deterministicGasConfig: deterministicGasConfig,
	}
}

// AnteHandle resets the gas limit inside GasMeter.
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
	bonus := cfgd.deterministicGasConfig.TxBaseGas(params)
	if gasConsumed > bonus {
		gasMeter.ConsumeGas(gasConsumed-bonus, "OverBonus")
	}
	gasMeter.ConsumeGas(cfgd.deterministicGasConfig.FixedGas, "Fixed")

	return next(ctx.WithGasMeter(gasMeter), tx, simulate)
}
