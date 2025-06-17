package matchingengine

import (
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v6/x/dex/types"
)

// MatchingEngine takes an incoming order and returns a MatchingResult.
type MatchingEngine struct {
	obq       OrderBookQueue
	dexKeeper DEXKeeper
	ak        AccountKeeper
	logger    log.Logger
}

// NewMatchingEngine returns a new instance of MatchingEngine.
func NewMatchingEngine(
	obq OrderBookQueue,
	ak AccountKeeper,
	logger log.Logger,
	dexKeeper DEXKeeper,
) MatchingEngine {
	return MatchingEngine{
		obq:       obq,
		ak:        ak,
		logger:    logger,
		dexKeeper: dexKeeper,
	}
}

// RecordToAddress maps an account address to an order book record.
type RecordToAddress struct {
	Address sdk.AccAddress
	Record  *types.OrderBookRecord
}

func convertOrderToOrderBookRecord(
	accNumber uint64,
	orderBookID uint32,
	order types.Order,
	remainingBalance sdkmath.Int,
) types.OrderBookRecord {
	var price types.Price
	if order.Price != nil {
		price = *order.Price
	}

	return types.OrderBookRecord{
		OrderBookID:               orderBookID,
		Side:                      order.Side,
		Price:                     price,
		OrderSequence:             order.Sequence,
		OrderID:                   order.ID,
		AccountNumber:             accNumber,
		RemainingBaseQuantity:     order.Quantity,
		RemainingSpendableBalance: remainingBalance,
	}
}
