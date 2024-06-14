package ante

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/CoreumFoundation/coreum/v4/x/wgov/types"
)

const (
	untrackedMaxGasForQueries = uint64(50_000)
)

// GovDepositDecorator is the ante handler which blocks tokens that are part of the
// min deposit, from being depositted into the proposal. Despositing such tokens into
// proposals can lead to problems when they are being refunded.
type GovDepositDecorator struct {
	keeper types.GovKeeper
}

// NewGovDepositDecorator returns a new instance of GovDepositDecorator.
func NewGovDepositDecorator(keeper types.GovKeeper) GovDepositDecorator {
	return GovDepositDecorator{
		keeper: keeper,
	}
}

// AnteHandle rejects gov deposits that contains tokens which are not part of
// the min deposit variable.
func (d GovDepositDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler,
) (sdk.Context, error) {
	for _, msg := range tx.GetMsgs() {
		var deposit sdk.Coins
		switch typedMsg := msg.(type) {
		case *govv1.MsgDeposit:
			deposit = typedMsg.Amount
		case *govv1beta1.MsgDeposit:
			deposit = typedMsg.Amount
		case *govv1.MsgSubmitProposal:
			deposit = typedMsg.InitialDeposit
		case *govv1beta1.MsgSubmitProposal:
			deposit = typedMsg.InitialDeposit
		default:
			continue
		}

		ctxWithUntrackedGas := ctx.WithGasMeter(sdk.NewGasMeter(untrackedMaxGasForQueries))
		govParams := d.keeper.GetParams(ctxWithUntrackedGas)
		minDeposit := sdk.NewCoins(govParams.MinDeposit...)
		for _, coin := range deposit {
			if !minDeposit.AmountOf(coin.Denom).IsPositive() {
				return ctx, sdkerrors.Wrapf(
					cosmoserrors.ErrInvalidCoins,
					"you can only provide denoms which are allowed, invalid denom %s",
					coin.Denom,
				)
			}
		}
	}

	return next(ctx, tx, simulate)
}
