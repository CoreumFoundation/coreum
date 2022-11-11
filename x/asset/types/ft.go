package types

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// checksumCharset is the set of characters used for the hash
const checksumCharset = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// IssueFungibleTokenSettings is the model which represents the params for the fungible token issuance.
type IssueFungibleTokenSettings struct {
	Issuer        sdk.AccAddress
	Symbol        string
	Description   string
	Recipient     sdk.AccAddress
	InitialAmount sdk.Int
	Features      []FungibleTokenFeature
}

// BuildFungibleTokenDenom builds the denom string from the symbol and issuer address.
func BuildFungibleTokenDenom(symbol string, issuer sdk.AccAddress) string {
	base := fmt.Sprintf("%s-%s", symbol, issuer)
	return fmt.Sprintf("%s-%s", base, checksum(base))
}

// TODO(dhil) revise the func later, probably it should be implemented in a different way
func checksum(data string) string {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, crc32.ChecksumIEEE([]byte(data)))
	for i, b := range buf {
		buf[i] = checksumCharset[int(b)%len(checksumCharset)]
	}

	return string(buf)
}
