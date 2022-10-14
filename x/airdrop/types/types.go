package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type AirdropInfo struct {
	Sender        string
	Height        int64
	Description   string
	RequiredDenom string
	Offer         []sdk.DecCoin
}
