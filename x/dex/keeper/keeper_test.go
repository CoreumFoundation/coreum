package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func TestMatching(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	dexKeeper := testApp.DEXKeeper

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	denomA, err := ftKeeper.Issue(ctx, assetfttypes.IssueSettings{
		Issuer:        addr,
		Symbol:        "AAA",
		Description:   "A",
		Subunit:       "a",
		Precision:     6,
		InitialAmount: sdkmath.NewInt(1000),
	})
	requireT.NoError(err)

	denomB, err := ftKeeper.Issue(ctx, assetfttypes.IssueSettings{
		Issuer:        addr,
		Symbol:        "BBB",
		Description:   "B",
		Subunit:       "b",
		Precision:     6,
		InitialAmount: sdkmath.NewInt(1000),
	})
	requireT.NoError(err)

	denomC, err := ftKeeper.Issue(ctx, assetfttypes.IssueSettings{
		Issuer:        addr,
		Symbol:        "CCC",
		Description:   "C",
		Subunit:       "c",
		Precision:     6,
		InitialAmount: sdkmath.NewInt(1000),
	})
	requireT.NoError(err)

	denomD, err := ftKeeper.Issue(ctx, assetfttypes.IssueSettings{
		Issuer:        addr,
		Symbol:        "DDD",
		Description:   "D",
		Subunit:       "d",
		Precision:     6,
		InitialAmount: sdkmath.NewInt(1000),
	})
	requireT.NoError(err)

	testCases := []struct {
		Name   string
		Input  [][]types.Order
		Output []types.Order
	}{
		{
			Name: "single_order",
			Input: [][]types.Order{
				{
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomA, sdkmath.NewIntFromUint64(10)),
						SellPrice: sdk.NewDecCoinFromDec(denomB, sdk.MustNewDecFromStr("2")),
					},
				},
			},
			Output: []types.Order{
				&types.OrderLimit{
					Sender:    addr.String(),
					Amount:    sdk.NewCoin(denomA, sdkmath.NewIntFromUint64(10)),
					SellPrice: sdk.NewDecCoinFromDec(denomB, sdk.MustNewDecFromStr("2")),
				},
			},
		},
		{
			Name: "two_matching_orders_in_one_block",
			Input: [][]types.Order{
				{
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomA, sdkmath.NewIntFromUint64(10)),
						SellPrice: sdk.NewDecCoinFromDec(denomB, sdk.MustNewDecFromStr("2")),
					},
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomB, sdkmath.NewIntFromUint64(20)),
						SellPrice: sdk.NewDecCoinFromDec(denomA, sdk.MustNewDecFromStr("0.5")),
					},
				},
			},
			Output: []types.Order{},
		},
		{
			Name: "two_matching_orders_in_two_blocks",
			Input: [][]types.Order{
				{
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomA, sdkmath.NewIntFromUint64(10)),
						SellPrice: sdk.NewDecCoinFromDec(denomB, sdk.MustNewDecFromStr("2")),
					},
				},
				{
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomB, sdkmath.NewIntFromUint64(20)),
						SellPrice: sdk.NewDecCoinFromDec(denomA, sdk.MustNewDecFromStr("0.5")),
					},
				},
			},
			Output: []types.Order{},
		},
		{
			Name: "two_non_matching_orders_in_one_block",
			Input: [][]types.Order{
				{
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomA, sdkmath.NewIntFromUint64(10)),
						SellPrice: sdk.NewDecCoinFromDec(denomB, sdk.MustNewDecFromStr("2")),
					},
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomB, sdkmath.NewIntFromUint64(20)),
						SellPrice: sdk.NewDecCoinFromDec(denomA, sdk.MustNewDecFromStr("1")),
					},
				},
			},
			Output: []types.Order{
				&types.OrderLimit{
					Sender:    addr.String(),
					Amount:    sdk.NewCoin(denomA, sdkmath.NewIntFromUint64(10)),
					SellPrice: sdk.NewDecCoinFromDec(denomB, sdk.MustNewDecFromStr("2")),
				},
				&types.OrderLimit{
					Sender:    addr.String(),
					Amount:    sdk.NewCoin(denomB, sdkmath.NewIntFromUint64(20)),
					SellPrice: sdk.NewDecCoinFromDec(denomA, sdk.MustNewDecFromStr("1")),
				},
			},
		},
		{
			Name: "two_non_matching_orders_in_one_block_reversed",
			Input: [][]types.Order{
				{
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomB, sdkmath.NewIntFromUint64(20)),
						SellPrice: sdk.NewDecCoinFromDec(denomA, sdk.MustNewDecFromStr("1")),
					},
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomA, sdkmath.NewIntFromUint64(10)),
						SellPrice: sdk.NewDecCoinFromDec(denomB, sdk.MustNewDecFromStr("2")),
					},
				},
			},
			Output: []types.Order{
				&types.OrderLimit{
					Sender:    addr.String(),
					Amount:    sdk.NewCoin(denomB, sdkmath.NewIntFromUint64(20)),
					SellPrice: sdk.NewDecCoinFromDec(denomA, sdk.MustNewDecFromStr("1")),
				},
				&types.OrderLimit{
					Sender:    addr.String(),
					Amount:    sdk.NewCoin(denomA, sdkmath.NewIntFromUint64(10)),
					SellPrice: sdk.NewDecCoinFromDec(denomB, sdk.MustNewDecFromStr("2")),
				},
			},
		},
		{
			Name: "two_non_matching_orders_in_two_blocks",
			Input: [][]types.Order{
				{
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomA, sdkmath.NewIntFromUint64(10)),
						SellPrice: sdk.NewDecCoinFromDec(denomB, sdk.MustNewDecFromStr("2")),
					},
				},
				{
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomB, sdkmath.NewIntFromUint64(20)),
						SellPrice: sdk.NewDecCoinFromDec(denomA, sdk.MustNewDecFromStr("1")),
					},
				},
			},
			Output: []types.Order{
				&types.OrderLimit{
					Sender:    addr.String(),
					Amount:    sdk.NewCoin(denomA, sdkmath.NewIntFromUint64(10)),
					SellPrice: sdk.NewDecCoinFromDec(denomB, sdk.MustNewDecFromStr("2")),
				},
				&types.OrderLimit{
					Sender:    addr.String(),
					Amount:    sdk.NewCoin(denomB, sdkmath.NewIntFromUint64(20)),
					SellPrice: sdk.NewDecCoinFromDec(denomA, sdk.MustNewDecFromStr("1")),
				},
			},
		},
		{
			Name: "two_non_matching_orders_in_two_blocks_reversed",
			Input: [][]types.Order{
				{
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomB, sdkmath.NewIntFromUint64(20)),
						SellPrice: sdk.NewDecCoinFromDec(denomA, sdk.MustNewDecFromStr("1")),
					},
				},
				{
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomA, sdkmath.NewIntFromUint64(10)),
						SellPrice: sdk.NewDecCoinFromDec(denomB, sdk.MustNewDecFromStr("2")),
					},
				},
			},
			Output: []types.Order{
				&types.OrderLimit{
					Sender:    addr.String(),
					Amount:    sdk.NewCoin(denomB, sdkmath.NewIntFromUint64(20)),
					SellPrice: sdk.NewDecCoinFromDec(denomA, sdk.MustNewDecFromStr("1")),
				},
				&types.OrderLimit{
					Sender:    addr.String(),
					Amount:    sdk.NewCoin(denomA, sdkmath.NewIntFromUint64(10)),
					SellPrice: sdk.NewDecCoinFromDec(denomB, sdk.MustNewDecFromStr("2")),
				},
			},
		},
		{
			Name: "better_price_is_used_1",
			Input: [][]types.Order{
				{
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomA, sdkmath.NewIntFromUint64(15)),
						SellPrice: sdk.NewDecCoinFromDec(denomB, sdk.MustNewDecFromStr("2")),
					},
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomB, sdkmath.NewIntFromUint64(20)),
						SellPrice: sdk.NewDecCoinFromDec(denomA, sdk.MustNewDecFromStr("0.25")),
					},
				},
			},
			Output: []types.Order{
				&types.OrderLimit{
					Sender:    addr.String(),
					Amount:    sdk.NewCoin(denomA, sdkmath.NewIntFromUint64(5)),
					SellPrice: sdk.NewDecCoinFromDec(denomB, sdk.MustNewDecFromStr("2")),
				},
			},
		},
		{
			Name: "better_price_is_used_2",
			Input: [][]types.Order{
				{
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomB, sdkmath.NewIntFromUint64(25)),
						SellPrice: sdk.NewDecCoinFromDec(denomA, sdk.MustNewDecFromStr("0.5")),
					},
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomA, sdkmath.NewIntFromUint64(10)),
						SellPrice: sdk.NewDecCoinFromDec(denomB, sdk.MustNewDecFromStr("1")),
					},
				},
			},
			Output: []types.Order{
				&types.OrderLimit{
					Sender:    addr.String(),
					Amount:    sdk.NewCoin(denomB, sdkmath.NewIntFromUint64(5)),
					SellPrice: sdk.NewDecCoinFromDec(denomA, sdk.MustNewDecFromStr("0.5")),
				},
			},
		},
		{
			Name: "two_order_books",
			Input: [][]types.Order{
				{
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomA, sdkmath.NewIntFromUint64(15)),
						SellPrice: sdk.NewDecCoinFromDec(denomB, sdk.MustNewDecFromStr("2")),
					},
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomD, sdkmath.NewIntFromUint64(10)),
						SellPrice: sdk.NewDecCoinFromDec(denomC, sdk.MustNewDecFromStr("1")),
					},
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomB, sdkmath.NewIntFromUint64(20)),
						SellPrice: sdk.NewDecCoinFromDec(denomA, sdk.MustNewDecFromStr("0.25")),
					},
					&types.OrderLimit{
						Sender:    addr.String(),
						Amount:    sdk.NewCoin(denomC, sdkmath.NewIntFromUint64(25)),
						SellPrice: sdk.NewDecCoinFromDec(denomD, sdk.MustNewDecFromStr("0.5")),
					},
				},
			},
			Output: []types.Order{
				&types.OrderLimit{
					Sender:    addr.String(),
					Amount:    sdk.NewCoin(denomA, sdkmath.NewIntFromUint64(5)),
					SellPrice: sdk.NewDecCoinFromDec(denomB, sdk.MustNewDecFromStr("2")),
				},
				&types.OrderLimit{
					Sender:    addr.String(),
					Amount:    sdk.NewCoin(denomC, sdkmath.NewIntFromUint64(15)),
					SellPrice: sdk.NewDecCoinFromDec(denomD, sdk.MustNewDecFromStr("0.5")),
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			requireT := require.New(t)
			assertT := assert.New(t)

			store := ctx.TransientStore(testApp.GetKey(types.StoreKey))
			it := store.Iterator(nil, nil)
			for ; it.Valid(); it.Next() {
				store.Delete(it.Key())
			}

			tStore := ctx.TransientStore(testApp.GetTKey(types.TransientStoreKey))
			for _, orderSet := range tc.Input {
				it := tStore.Iterator(nil, nil)
				for ; it.Valid(); it.Next() {
					tStore.Delete(it.Key())
				}

				for _, order := range orderSet {
					requireT.NoError(dexKeeper.StoreTransientOrder(ctx, order))
				}
				requireT.NoError(dexKeeper.ProcessTransientQueue(ctx))
			}

			orders, err := dexKeeper.ExportOrders(ctx)
			requireT.NoError(err)

			if !assertT.Equal(tc.Output, orders) {
				for _, o := range orders {
					t.Log(o)
				}
			}
		})
	}
}
