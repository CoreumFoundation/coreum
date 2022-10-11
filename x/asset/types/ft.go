package types

import (
	"fmt"
	"github.com/sigurn/crc8"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IssueFungibleTokenSettings is the model which represents the params for the fungible token issuance.
type IssueFungibleTokenSettings struct {
	Issuer        sdk.AccAddress
	Symbol        string
	Description   string
	Recipient     sdk.AccAddress
	InitialAmount sdk.Int
}

// BuildFungibleTokenDenom builds the denom string from the symbol and issuer address.
func BuildFungibleTokenDenom(symbol string, issuer sdk.AccAddress) string {
	base := fmt.Sprintf("%s-%s", symbol, issuer)
	return fmt.Sprintf("%s-%x", base, crc8.Checksum([]byte(base), crc8.MakeTable(crc8.CRC8)))
}
