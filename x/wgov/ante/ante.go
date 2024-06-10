package ante

import (
	sdkerrors "cosmossdk.io/errors"
	"github.com/CoreumFoundation/coreum/v4/x/wgov/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

type GovDepositDecorator struct {
	keeer types.GovKeeper
}

func NewGovDepositDecorator(keeer types.GovKeeper) GovDepositDecorator {
	return GovDepositDecorator{
		keeer: keeer,
	}
}

func (d GovDepositDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler,
) (sdk.Context, error) {
	for _, msg := range tx.GetMsgs() {
		deposit := sdk.NewCoins()
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

		govParams := d.keeer.GetParams(ctx)
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
