package dex

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/v5/x/dex/keeper"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

// InitGenesis initializes the dex module's state from a provided genesis state.
func InitGenesis(
	ctx sdk.Context,
	dexKeeper keeper.Keeper,
	accountKeeper types.AccountKeeper,
	genState types.GenesisState,
) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	if err := dexKeeper.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	maxOrderID := uint32(0)
	for _, orderBook := range genState.OrderBooks {
		if err := dexKeeper.SaveOrderBookIDWithData(ctx, orderBook.ID, orderBook.Data); err != nil {
			panic(errors.Wrap(err, "failed to set order book data"))
		}
		if orderBook.ID > maxOrderID {
			maxOrderID = orderBook.ID
		}
	}
	if maxOrderID != 0 {
		if err := dexKeeper.SetOrderBookSequence(ctx, maxOrderID); err != nil {
			panic(errors.Wrap(err, "failed to set order book sequence"))
		}
	}

	accAddressToNumberCache := make(map[string]uint64)
	for _, order := range genState.Orders {
		// check that the order book exists
		orderBookID, err := dexKeeper.GetOrderBookIDByDenoms(ctx, order.BaseDenom, order.QuoteDenom)
		if err != nil {
			panic(
				errors.Wrapf(
					err,
					"failed to get order book ID by denoms, base: %s, quote: %s",
					order.BaseDenom, order.QuoteDenom,
				),
			)
		}

		creator, err := sdk.AccAddressFromBech32(order.Creator)
		if err != nil {
			panic(sdkerrors.Wrapf(types.ErrInvalidInput, "invalid address: %s", order.Creator))
		}

		accNumber, ok := accAddressToNumberCache[order.Creator]
		if !ok {
			var err error
			acc := accountKeeper.GetAccount(ctx, creator)
			if acc == nil {
				panic(errors.Wrap(err, "account not fond: "+creator.String()))
			}
			accNumber = acc.GetAccountNumber()
			accAddressToNumberCache[order.Creator] = accNumber
		}

		record := types.OrderBookRecord{
			OrderBookID:               orderBookID,
			Side:                      order.Side,
			Price:                     *order.Price,
			OrderSequence:             order.Sequence,
			OrderID:                   order.ID,
			AccountNumber:             accNumber,
			RemainingBaseQuantity:     order.RemainingBaseQuantity,
			RemainingSpendableBalance: order.RemainingSpendableBalance,
		}
		if err := dexKeeper.SaveOrderWithOrderBookRecord(ctx, order, record); err != nil {
			panic(errors.Wrap(err, "failed to set order with order book record"))
		}
	}
	if err := dexKeeper.SetOrderSequence(ctx, genState.OrderSequence); err != nil {
		panic(errors.Wrap(err, "failed to set order sequence"))
	}

	for _, accountDenomOrdersCount := range genState.AccountsDenomsOrdersCounts {
		if err := dexKeeper.SetAccountDenomOrdersCount(ctx, accountDenomOrdersCount); err != nil {
			panic(errors.Wrap(err, "failed to set accounts denoms orders counts"))
		}
	}
}

// ExportGenesis returns the dex module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	orders, _, err := k.GetAccountsOrders(ctx, &query.PageRequest{Limit: query.PaginationMaxLimit})
	if err != nil {
		panic(errors.Wrap(err, "failed to get orders with sequence"))
	}

	orderBooksWithID, _, err := k.GetOrderBooksWithID(ctx, &query.PageRequest{Limit: query.PaginationMaxLimit})
	if err != nil {
		panic(errors.Wrap(err, "failed to get order books with ID"))
	}

	accountsDenomsOrdersCounts, _, err := k.GetAccountsDenomsOrdersCounts(
		ctx, &query.PageRequest{Limit: query.PaginationMaxLimit},
	)
	if err != nil {
		panic(errors.Wrap(err, "failed to get accounts denoms orders counts"))
	}

	orderSequence, err := k.GetOrderSequence(ctx)
	if err != nil {
		panic(errors.Wrap(err, "failed to get order sequence"))
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		panic(errors.Wrap(err, "failed to get params"))
	}

	return &types.GenesisState{
		Params:                     params,
		Orders:                     orders,
		OrderBooks:                 orderBooksWithID,
		OrderSequence:              orderSequence,
		AccountsDenomsOrdersCounts: accountsDenomsOrdersCounts,
	}
}
