package types

import (
	"encoding/binary"
	"math/big"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	maxPricePartUint64     = 2 // how many uin64 we allow for the whole and decimal part
	uin64MaxTenRangeNumber = 19
	// max uint64 is 2^64, so we can save any number from 0 to the 10^19.
	maxPricePartLen = uin64MaxTenRangeNumber * maxPricePartUint64
)

var tenPowerMaxDecLength = (&big.Int{}).Exp(big.NewInt(10), big.NewInt(maxPricePartLen), nil)

// Price is the price type.
type Price struct {
	i *big.Int
}

// NewPriceFromString returns new instance of the Price from string.
func NewPriceFromString(str string) (Price, error) {
	i, err := priceStringToInternalBigInt(str)
	if err != nil {
		return Price{}, err
	}

	return Price{
		i: i,
	}, nil
}

// Rat returns the stored base Rat type.
func (p Price) Rat() *big.Rat {
	return (&big.Rat{}).SetFrac(p.i, tenPowerMaxDecLength)
}

// MarshallToEndianBytes returns the bytes representation of the Price.
func (p Price) MarshallToEndianBytes() ([]byte, error) {
	var wholePartStr, decPartStr string
	if intStr := p.i.String(); len(intStr) > maxPricePartLen {
		// whole use used
		partIndex := len(intStr) - maxPricePartLen
		wholePartStr = intStr[:partIndex]
		decPartStr = intStr[partIndex:]
	} else {
		// dec only is used
		wholePartStr = "0"
		decPartStr = intStr
	}
	bytes := make([]byte, 0, 2*maxPricePartUint64*bigEndianUint64ByteSize)
	// encode whole part
	var err error
	if bytes, err = bigIntToUint64BigEndianSlice(wholePartStr, bytes); err != nil {
		return nil, err
	}
	// encode dec part
	if bytes, err = bigIntToUint64BigEndianSlice(decPartStr, bytes); err != nil {
		return nil, err
	}

	return bytes, nil
}

// UnmarshallFromEndianBytes unmarshalls endian bytes to Price type and returns remaining bytes.
func (p *Price) UnmarshallFromEndianBytes(bytes []byte) ([]byte, error) {
	const priceBytesLen = 2 * maxPricePartUint64 * bigEndianUint64ByteSize
	if len(bytes) < priceBytesLen {
		return nil, errors.Errorf("failed to convert bytes to Price, invalid length")
	}

	var combinedStr string
	for i := 0; i < 2*maxPricePartUint64; i++ {
		partSrt := strconv.FormatUint(
			binary.BigEndian.Uint64(bytes[i*bigEndianUint64ByteSize:(i+1)*bigEndianUint64ByteSize]), 10,
		)
		combinedStr += strings.Repeat("0", uin64MaxTenRangeNumber-len(partSrt)) + partSrt
	}

	var ok bool
	p.i, ok = (&big.Int{}).SetString(combinedStr, 10)
	if !ok {
		return nil, errors.Errorf("failed to convert %s to big.Int to unmarshall Price", combinedStr)
	}

	return bytes[priceBytesLen:], nil
}

// String returns string representation of the Price.
func (p Price) String() string {
	if p.i == nil {
		return "<nil>"
	}
	return strings.TrimRight(strings.TrimRight(p.Rat().FloatString(maxPricePartLen), "0"), ".")
}

// MarshalTo implements the gogo proto custom type interface.
func (p *Price) MarshalTo(data []byte) (n int, err error) {
	bz, err := p.Marshal()
	if err != nil {
		return 0, err
	}

	copy(data, bz)
	return len(bz), nil
}

// Unmarshal implements the gogo proto custom type interface.
func (p *Price) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	i, err := priceStringToInternalBigInt(string(data))
	if err != nil {
		return err
	}
	p.i = i
	return nil
}

// Size implements the gogo proto custom type interface.
func (p *Price) Size() int {
	bz, _ := p.Marshal()
	return len(bz)
}

// Marshal implements the gogo proto custom type interface.
func (p *Price) Marshal() ([]byte, error) {
	return []byte(p.String()), nil
}

func priceStringToInternalBigInt(str string) (*big.Int, error) {
	if len(str) == 0 {
		return nil, errors.Errorf("failed to create Price from empty string")
	}

	strParts := strings.Split(str, ".")
	var wholePartStr, decPartStr string
	strLen := len(strParts)
	if strLen > 2 {
		return nil, errors.Errorf("failed to create Price from string: %s, invalid format", str)
	}

	wholePartStr = strParts[0]
	if len(wholePartStr) > maxPricePartLen {
		return nil, errors.Errorf(
			"failed to create whole Price part from, string: %s, too long, max: %d",
			wholePartStr, maxPricePartLen,
		)
	}
	if strLen == 2 {
		decPartStr = strParts[1]
		if len(decPartStr) > maxPricePartLen {
			return nil, errors.Errorf(
				"failed to create decimal Price part from, string: %s, too long, max: %d",
				decPartStr, maxPricePartLen,
			)
		}
	}

	// append zero to always determine the easies way of how to convert to Rat or float
	combinedStr := wholePartStr + decPartStr + strings.Repeat("0", maxPricePartLen-len(decPartStr))

	i, ok := (&big.Int{}).SetString(combinedStr, 10)
	if !ok {
		return nil, errors.Errorf(
			"failed to create Price from, string: %s, invalid format", str,
		)
	}
	return i, nil
}

// bigIntToUint64BigEndianSlice converts the str into BigEndian bytes and appends it to bytes slice.
func bigIntToUint64BigEndianSlice(str string, bytes []byte) ([]byte, error) {
	// append zero to the head as placeholder for the empty uint64 to keep same size of the final byte slice
	uint64SliceLen := ((len(str) - 1) / uin64MaxTenRangeNumber) + 1
	for i := 0; i < maxPricePartUint64-uint64SliceLen; i++ {
		bytes = binary.BigEndian.AppendUint64(bytes, 0)
	}

	for _, chunk := range chunkStringFromTail(str, uin64MaxTenRangeNumber) {
		var err error
		bytes, err = appendUint64BigEndian(bytes, chunk)
		if err != nil {
			return bytes, err
		}
	}

	return bytes, nil
}

func chunkStringFromTail(str string, size int) []string {
	currentLen := 0
	currentStart := len(str)
	chunks := make([]string, 0)
	for i := len(str); i > 0; i-- {
		if currentLen == size {
			chunks = append(chunks, str[i:currentStart])
			currentLen = 0
			currentStart = i
		}
		currentLen++
	}
	chunks = append(chunks, str[:currentStart])

	// reverse
	length := len(chunks)
	half := length / 2
	for i := 0; i < half; i++ {
		j := length - 1 - i
		chunks[i], chunks[j] = chunks[j], chunks[i]
	}

	return chunks
}

func appendUint64BigEndian(bytes []byte, str string) ([]byte, error) {
	uint64Value, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(
			err, "failed to convert %s into uint64", str,
		)
	}

	return binary.BigEndian.AppendUint64(bytes, uint64Value), nil
}
