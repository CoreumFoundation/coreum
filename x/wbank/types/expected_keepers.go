package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// FungibleTokenProvider defines an interface to interact with the fungible token functionality.
type FungibleTokenProvider interface {
	BeforeSendCoins(ctx sdk.Context, fromAddress, toAddress sdk.AccAddress, coins sdk.Coins) error
	BeforeInputOutputCoins(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error
}
