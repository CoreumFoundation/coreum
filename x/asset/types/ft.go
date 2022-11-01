package types

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"strings"

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

// ValidateDenom ensures that denom has acceptable bank format and checksum is valid
func ValidateDenom(denom string) error {
	if err := sdk.ValidateDenom(denom); err != nil {
		return ErrInvalidDenomFormat
	}

	return ValidateDenomChecksum(denom)
}

// ValidateDenomChecksum ensures that the checksum value of user created denom is valid
func ValidateDenomChecksum(denom string) error {
	splits := strings.Split(denom, "-")
	if len(splits) != 3 {
		return ErrInvalidDenomFormat
	}

	cs := checksum(splits[0] + "-" + splits[1])
	if cs != splits[2] {
		return ErrInvalidDenomChecksum
	}

	return nil
}

func checksum(data string) string {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, crc32.ChecksumIEEE([]byte(data)))
	for i, b := range buf {
		buf[i] = checksumCharset[int(b)%len(checksumCharset)]
	}

	return string(buf)
}
