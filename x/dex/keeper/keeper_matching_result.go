package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"

	matchingengine "github.com/CoreumFoundation/coreum/v6/x/dex/matching-engine"
	"github.com/CoreumFoundation/coreum/v6/x/dex/types"
)

// RecordToAddress maps an account address to an order book record.
// TODO: RecordToAddress could potentially be replaced by matchingengine.RecordToAddress.
type RecordToAddress struct {
	Address sdk.AccAddress
	Record  *types.OrderBookRecord
}

func (k Keeper) applyMatchingResult(ctx sdk.Context, mr matchingengine.MatchingResult) error {
	// if matched passed but no changes are applied return
	if mr.FTActions.CreatorExpectedToSpend.IsNil() {
		return nil
	}

	for _, item := range mr.RecordsToRemove {
		if err := k.removeOrderByRecord(ctx, item.Address, *item.Record); err != nil {
			return err
		}
	}

	if mr.RecordToUpdate != nil {
		if err := k.saveOrderBookRecord(ctx, *mr.RecordToUpdate); err != nil {
			return err
		}
	}

	if err := k.publishMatchingEvents(ctx, mr); err != nil {
		return err
	}

	// the call to smart contract is the last call here to avoid reentrancy vulnerability.
	return k.assetFTKeeper.DEXExecuteActions(ctx, mr.FTActions)
}

func (k Keeper) publishMatchingEvents(
	ctx sdk.Context,
	mr matchingengine.MatchingResult,
) error {
	events := mr.MakerOrderReducedEvents
	if !mr.TakerOrderReducedEvent.SentCoin.IsZero() {
		events = append(events, mr.TakerOrderReducedEvent)
	}

	for _, evt := range events {
		if err := ctx.EventManager().EmitTypedEvent(&evt); err != nil {
			return sdkerrors.Wrapf(cosmoserrors.ErrIO, "failed to emit event EventOrderReduced: %s", err)
		}
	}

	return nil
}
