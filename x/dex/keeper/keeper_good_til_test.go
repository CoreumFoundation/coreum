package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func TestKeeper_GoodTil(t *testing.T) {
	blockTime := time.Second + time.Second/2
	initialBlockTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		// height to orders
		orders      func(accSet AccSet) map[uint64][]types.Order
		wantOrders  func(accSet AccSet) []types.Order
		startHeight uint64
		endHeight   uint64
	}{
		{
			name: "no_match_no_good_til",
			orders: func(accSet AccSet) map[uint64][]types.Order {
				return map[uint64][]types.Order{
					2: {
						{
							Creator:     accSet.acc1.String(),
							Type:        types.ORDER_TYPE_LIMIT,
							ID:          "id1",
							BaseDenom:   denom1,
							QuoteDenom:  denom2,
							Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
							Quantity:    sdkmath.NewInt(1000),
							Side:        types.SIDE_SELL,
							TimeInForce: types.TIME_IN_FORCE_GTC,
						},
					},
				}
			},
			startHeight: 1,
			endHeight:   10,
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:           accSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
				}
			},
		},
		{
			name: "no_match_with_good_til_block_height",
			orders: func(accSet AccSet) map[uint64][]types.Order {
				return map[uint64][]types.Order{
					301: {
						{
							Creator:     accSet.acc1.String(),
							Type:        types.ORDER_TYPE_LIMIT,
							ID:          "id1",
							BaseDenom:   denom1,
							QuoteDenom:  denom2,
							Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
							Quantity:    sdkmath.NewInt(1000),
							Side:        types.SIDE_SELL,
							GoodTil:     &types.GoodTil{GoodTilBlockHeight: 343},
							TimeInForce: types.TIME_IN_FORCE_GTC,
						},
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:           accSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_SELL,
						GoodTil:           &types.GoodTil{GoodTilBlockHeight: 343},
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
				}
			},
			startHeight: 300,
			endHeight:   310,
		},
		{
			name: "no_match_with_good_til_block_time",
			orders: func(accSet AccSet) map[uint64][]types.Order {
				return map[uint64][]types.Order{
					301: {
						{
							Creator:     accSet.acc1.String(),
							Type:        types.ORDER_TYPE_LIMIT,
							ID:          "id1",
							BaseDenom:   denom1,
							QuoteDenom:  denom2,
							Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
							Quantity:    sdkmath.NewInt(1000),
							Side:        types.SIDE_SELL,
							GoodTil:     &types.GoodTil{GoodTilBlockTime: lo.ToPtr(initialBlockTime.Add(time.Hour))},
							TimeInForce: types.TIME_IN_FORCE_GTC,
						},
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:           accSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_SELL,
						GoodTil:           &types.GoodTil{GoodTilBlockTime: lo.ToPtr(initialBlockTime.Add(time.Hour))},
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
				}
			},
			startHeight: 300,
			endHeight:   310,
		},
		{
			name: "no_match_with_good_til_block_high_and_time",
			orders: func(accSet AccSet) map[uint64][]types.Order {
				return map[uint64][]types.Order{
					301: {
						{
							Creator:    accSet.acc1.String(),
							Type:       types.ORDER_TYPE_LIMIT,
							ID:         "id1",
							BaseDenom:  denom1,
							QuoteDenom: denom2,
							Price:      lo.ToPtr(types.MustNewPriceFromString("376e-3")),
							Quantity:   sdkmath.NewInt(1000),
							Side:       types.SIDE_SELL,
							GoodTil: &types.GoodTil{
								GoodTilBlockHeight: 343,
								GoodTilBlockTime:   lo.ToPtr(initialBlockTime.Add(time.Hour)),
							},
							TimeInForce: types.TIME_IN_FORCE_GTC,
						},
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.SIDE_SELL,
						GoodTil: &types.GoodTil{
							GoodTilBlockHeight: 343,
							GoodTilBlockTime:   lo.ToPtr(initialBlockTime.Add(time.Hour)),
						},
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
				}
			},
			startHeight: 300,
			endHeight:   310,
		},
		{
			name: "partial_taker_match_with_good_til_block_height",
			orders: func(accSet AccSet) map[uint64][]types.Order {
				return map[uint64][]types.Order{
					101: {
						{
							Creator:     accSet.acc1.String(),
							Type:        types.ORDER_TYPE_LIMIT,
							ID:          "id1",
							BaseDenom:   denom1,
							QuoteDenom:  denom2,
							Price:       lo.ToPtr(types.MustNewPriceFromString("1")),
							Quantity:    sdkmath.NewInt(500),
							Side:        types.SIDE_SELL,
							GoodTil:     &types.GoodTil{GoodTilBlockHeight: 454},
							TimeInForce: types.TIME_IN_FORCE_GTC,
						},
					},
					102: {
						{
							Creator:     accSet.acc2.String(),
							Type:        types.ORDER_TYPE_LIMIT,
							ID:          "id2",
							BaseDenom:   denom1,
							QuoteDenom:  denom2,
							Price:       lo.ToPtr(types.MustNewPriceFromString("1")),
							Quantity:    sdkmath.NewInt(1000),
							Side:        types.SIDE_BUY,
							GoodTil:     &types.GoodTil{GoodTilBlockHeight: 123},
							TimeInForce: types.TIME_IN_FORCE_GTC,
						},
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:           accSet.acc2.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id2",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("1")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_BUY,
						GoodTil:           &types.GoodTil{GoodTilBlockHeight: 123},
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(500),
						RemainingBalance:  sdkmath.NewInt(500),
					},
				}
			},
			startHeight: 100,
			endHeight:   110,
		},
		{
			name: "full_taker_match_with_good_til_block_height",
			orders: func(accSet AccSet) map[uint64][]types.Order {
				return map[uint64][]types.Order{
					105: {
						{
							Creator:     accSet.acc2.String(),
							Type:        types.ORDER_TYPE_LIMIT,
							ID:          "id1",
							BaseDenom:   denom1,
							QuoteDenom:  denom2,
							Price:       lo.ToPtr(types.MustNewPriceFromString("1")),
							Quantity:    sdkmath.NewInt(1000),
							Side:        types.SIDE_BUY,
							GoodTil:     &types.GoodTil{GoodTilBlockHeight: 123},
							TimeInForce: types.TIME_IN_FORCE_GTC,
						},
						{
							Creator:     accSet.acc1.String(),
							Type:        types.ORDER_TYPE_LIMIT,
							ID:          "id2",
							BaseDenom:   denom1,
							QuoteDenom:  denom2,
							Price:       lo.ToPtr(types.MustNewPriceFromString("1")),
							Quantity:    sdkmath.NewInt(500),
							Side:        types.SIDE_SELL,
							GoodTil:     &types.GoodTil{GoodTilBlockHeight: 454},
							TimeInForce: types.TIME_IN_FORCE_GTC,
						},
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:           accSet.acc2.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("1")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_BUY,
						GoodTil:           &types.GoodTil{GoodTilBlockHeight: 123},
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(500),
						RemainingBalance:  sdkmath.NewInt(500),
					},
				}
			},
			startHeight: 100,
			endHeight:   110,
		},
		{
			name: "no_match_with_good_til_block_height_keep_to_max_height",
			orders: func(accSet AccSet) map[uint64][]types.Order {
				return map[uint64][]types.Order{
					310: {
						{
							Creator:     accSet.acc1.String(),
							Type:        types.ORDER_TYPE_LIMIT,
							ID:          "id1",
							BaseDenom:   denom1,
							QuoteDenom:  denom2,
							Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
							Quantity:    sdkmath.NewInt(1000),
							Side:        types.SIDE_SELL,
							GoodTil:     &types.GoodTil{GoodTilBlockHeight: 343},
							TimeInForce: types.TIME_IN_FORCE_GTC,
						},
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:           accSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_SELL,
						GoodTil:           &types.GoodTil{GoodTilBlockHeight: 343},
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
				}
			},
			startHeight: 300,
			endHeight:   343,
		},
		{
			name: "no_match_with_good_til_block_height_remove_from_max_height",
			orders: func(accSet AccSet) map[uint64][]types.Order {
				return map[uint64][]types.Order{
					310: {
						// this order will be cancelled by good til
						{
							Creator:     accSet.acc1.String(),
							Type:        types.ORDER_TYPE_LIMIT,
							ID:          "id1",
							BaseDenom:   denom1,
							QuoteDenom:  denom2,
							Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
							Quantity:    sdkmath.NewInt(1000),
							Side:        types.SIDE_SELL,
							GoodTil:     &types.GoodTil{GoodTilBlockHeight: 343}, // same height as in next order
							TimeInForce: types.TIME_IN_FORCE_GTC,
						},
					},
					// this order will be cancelled by good til
					311: {
						{
							Creator:     accSet.acc1.String(),
							Type:        types.ORDER_TYPE_LIMIT,
							ID:          "id2",
							BaseDenom:   denom2,
							QuoteDenom:  denom3,
							Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
							Quantity:    sdkmath.NewInt(1000),
							Side:        types.SIDE_SELL,
							GoodTil:     &types.GoodTil{GoodTilBlockHeight: 343}, // same height as in next order
							TimeInForce: types.TIME_IN_FORCE_GTC,
						},
					},
					// this order will remain in the order book
					314: {
						{
							Creator:     accSet.acc1.String(),
							Type:        types.ORDER_TYPE_LIMIT,
							ID:          "id3",
							BaseDenom:   denom1,
							QuoteDenom:  denom3,
							Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
							Quantity:    sdkmath.NewInt(1000),
							Side:        types.SIDE_SELL,
							GoodTil:     &types.GoodTil{GoodTilBlockHeight: 345},
							TimeInForce: types.TIME_IN_FORCE_GTC,
						},
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:           accSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id3",
						BaseDenom:         denom1,
						QuoteDenom:        denom3,
						Price:             lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_SELL,
						GoodTil:           &types.GoodTil{GoodTilBlockHeight: 345},
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
				}
			},
			startHeight: 300,
			endHeight:   344,
		},
		{
			name: "no_match_with_good_til_block_time_remove_from_max_time",
			orders: func(accSet AccSet) map[uint64][]types.Order {
				return map[uint64][]types.Order{
					101: {
						// this order will be cancelled by good til
						{
							Creator:    accSet.acc1.String(),
							Type:       types.ORDER_TYPE_LIMIT,
							ID:         "id1",
							BaseDenom:  denom1,
							QuoteDenom: denom2,
							Price:      lo.ToPtr(types.MustNewPriceFromString("376e-3")),
							Quantity:   sdkmath.NewInt(1000),
							Side:       types.SIDE_SELL,
							GoodTil: &types.GoodTil{
								GoodTilBlockTime: lo.ToPtr(initialBlockTime.Add(blockTime * time.Duration(3))),
							}, // same height as in next order
							TimeInForce: types.TIME_IN_FORCE_GTC,
						},
					},
					// this order will be cancelled by good til
					102: {
						{
							Creator:    accSet.acc1.String(),
							Type:       types.ORDER_TYPE_LIMIT,
							ID:         "id2",
							BaseDenom:  denom2,
							QuoteDenom: denom3,
							Price:      lo.ToPtr(types.MustNewPriceFromString("376e-3")),
							Quantity:   sdkmath.NewInt(1000),
							Side:       types.SIDE_SELL,
							GoodTil: &types.GoodTil{
								GoodTilBlockTime: lo.ToPtr(initialBlockTime.Add(blockTime * time.Duration(3))),
							}, // same height as in next order
							TimeInForce: types.TIME_IN_FORCE_GTC,
						},
					},
					// this order will remain in the order book
					108: {
						{
							Creator:    accSet.acc1.String(),
							Type:       types.ORDER_TYPE_LIMIT,
							ID:         "id3",
							BaseDenom:  denom1,
							QuoteDenom: denom3,
							Price:      lo.ToPtr(types.MustNewPriceFromString("376e-3")),
							Quantity:   sdkmath.NewInt(1000),
							Side:       types.SIDE_SELL,
							GoodTil: &types.GoodTil{
								GoodTilBlockTime: lo.ToPtr(initialBlockTime.Add(time.Hour)),
							},
							TimeInForce: types.TIME_IN_FORCE_GTC,
						},
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id3",
						BaseDenom:  denom1,
						QuoteDenom: denom3,
						Price:      lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.SIDE_SELL,
						GoodTil: &types.GoodTil{
							GoodTilBlockTime: lo.ToPtr(initialBlockTime.Add(time.Hour)),
						},
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
				}
			},
			startHeight: 100,
			endHeight:   110,
		},
		{
			name: "no_match_with_good_til_block_time_and_height_remove_both",
			orders: func(accSet AccSet) map[uint64][]types.Order {
				return map[uint64][]types.Order{
					101: {
						// this order will be cancelled by good til
						{
							Creator:    accSet.acc1.String(),
							Type:       types.ORDER_TYPE_LIMIT,
							ID:         "id1",
							BaseDenom:  denom1,
							QuoteDenom: denom2,
							Price:      lo.ToPtr(types.MustNewPriceFromString("376e-3")),
							Quantity:   sdkmath.NewInt(1000),
							Side:       types.SIDE_SELL,
							GoodTil: &types.GoodTil{
								GoodTilBlockHeight: 103,
								GoodTilBlockTime:   lo.ToPtr(initialBlockTime.Add(blockTime * time.Duration(3))),
							},
							TimeInForce: types.TIME_IN_FORCE_GTC,
						},
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{}
			},
			startHeight: 100,
			endHeight:   110,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require.NotZero(t, tt.startHeight)
			require.GreaterOrEqual(t, tt.endHeight, tt.startHeight)

			logger := log.NewTestLogger(t)
			testApp := simapp.New(simapp.WithCustomLogger(logger))

			// place all in the start height block
			sdkCtx := testApp.NewContextLegacy(false, cmtproto.Header{
				Time:   initialBlockTime,
				Height: int64(tt.startHeight),
			})
			accSet := getAccSet(sdkCtx, testApp)
			orderBooksIDs := make(map[uint32]struct{})

			// validate height
			heightToOrders := tt.orders(accSet)
			for height := range heightToOrders {
				if height < tt.startHeight || height > tt.endHeight {
					t.Fatalf("Order height must be in the range [%d, %d]", tt.startHeight, tt.endHeight)
				}
			}

			// simulate block processing
			for i := 1; i <= int(tt.endHeight-tt.startHeight); i++ {
				height := tt.startHeight + uint64(i)
				sdkCtx := testApp.NewContextLegacy(false, cmtproto.Header{
					// increase block time every block
					Time:   initialBlockTime.Add(blockTime * time.Duration(i)),
					Height: int64(height),
				})
				_, err := testApp.BeginBlocker(sdkCtx)
				require.NoError(t, err)

				// process orders for specific height
				orders := heightToOrders[height]
				for _, order := range orders {
					balance, err := order.ComputeLimitOrderLockedBalance()
					require.NoError(t, err)
					testApp.MintAndSendCoin(t, sdkCtx, sdk.MustAccAddressFromBech32(order.Creator), sdk.NewCoins(balance))
					require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order))
					orderBooksID, err := testApp.DEXKeeper.GetOrderBookIDByDenoms(sdkCtx, order.BaseDenom, order.QuoteDenom)
					require.NoError(t, err)
					orderBooksIDs[orderBooksID] = struct{}{}
				}

				_, err = testApp.EndBlocker(sdkCtx)
				require.NoError(t, err)
			}

			gotOrders := make([]types.Order, 0)
			for orderBookID := range orderBooksIDs {
				gotOrders = append(gotOrders, getSorterOrderBookOrders(t, testApp, sdkCtx, orderBookID, types.SIDE_BUY)...)
				gotOrders = append(gotOrders, getSorterOrderBookOrders(t, testApp, sdkCtx, orderBookID, types.SIDE_SELL)...)
			}
			require.ElementsMatch(t, tt.wantOrders(accSet), gotOrders)
		})
	}
}
