package types

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	cbig "github.com/CoreumFoundation/coreum/v4/pkg/math/big"
	"github.com/CoreumFoundation/coreum/v4/pkg/store"
)

const (
	maxNumLen = 19
	// maxExp is the max allowed exponent. Technically it's limited by MinInt8 (-128) + `maxNumLen` (required for
	//	normalization). But to make the range value easier for understanding and still keeping enough precision we set
	//	it to -100.
	minExt = -100
	// maxExp is the max allowed exponent. Technically it's limited by MaxInt8 (127) but to make it similar to
	// mixExp we set it to 100.
	maxExp                = 100
	exponentStr           = "e"
	orderedBytesPriceSize = store.Int8OrderedBytesSize + store.Uint64OrderedBytesSize
)

// Price is the price type.
type Price struct {
	exp int8
	num uint64
}

// NewPriceFromString returns new instance of the Price from string.
func NewPriceFromString(str string) (Price, error) {
	if len(str) == 0 {
		return Price{}, errors.New("price can't be empty")
	}
	parts := strings.Split(str, exponentStr)
	numPart := parts[0]
	// the price must be represented with exponent if it's possible to use the exponent.
	if numPart != "0" && (strings.HasPrefix(numPart, "0") || strings.HasSuffix(numPart, "0")) {
		return Price{}, errors.Errorf("invalid price num part %s, can't start or end with 0", numPart)
	}
	if len(numPart) > maxNumLen {
		return Price{}, errors.Errorf("invalid price num part length, max %d", maxNumLen)
	}

	var (
		exp int8
		num uint64
		err error
	)
	num, err = strconv.ParseUint(numPart, 10, 64)
	if err != nil {
		return Price{}, errors.Errorf("invalid price num part %s", numPart)
	}

	switch len(parts) {
	case 1:
		exp = 0
	case 2:
		if numPart == "0" {
			return Price{}, errors.New("the exponent can't be provided for the zero num")
		}
		var intExp int64
		intExp, err = strconv.ParseInt(parts[1], 10, 8)
		if err != nil {
			return Price{}, errors.Errorf("invalid price exp part %s", parts[1])
		}
		// the range check is required for the normalization
		if intExp < minExt || intExp > maxExp {
			return Price{}, errors.Errorf("invalid exp %d, must be in the rage %d:%d", intExp, minExt, maxExp)
		}
		exp = int8(intExp)
	default:
		return Price{}, errors.Errorf("invalid price string %s", str)
	}

	return Price{
		exp: exp,
		num: num,
	}, nil
}

// Rat creates Rat type from Price.
func (p Price) Rat() *big.Rat {
	if p.exp > 0 {
		// num * 10^exp
		return cbig.NewRatFromBigInt(
			cbig.IntMul(cbig.NewBigIntFromUint64(p.num), cbig.IntTenToThePower(big.NewInt(int64(p.exp)))),
		)
	}
	// num / 10^exp
	return cbig.NewRatFromBigInts(
		cbig.NewBigIntFromUint64(p.num), cbig.IntTenToThePower(big.NewInt(int64(-p.exp))),
	)
}

// MarshallToOrderedBytes returns the ordered bytes representation of the Price.
func (p Price) MarshallToOrderedBytes() ([]byte, error) {
	exp, num, err := normalizeForOrderedBytes(p.exp, p.num)
	if err != nil {
		return nil, err
	}

	b := make([]byte, 0, orderedBytesPriceSize)
	b = store.AppendInt8ToOrderedBytes(b, exp)
	b = store.AppendUint64ToOrderedBytes(b, num)

	return b, nil
}

// UnmarshallFromOrderedBytes unmarshalls ordered bytes to Price type and returns remaining bytes.
func (p *Price) UnmarshallFromOrderedBytes(bytes []byte) ([]byte, error) {
	exp, bRem, err := store.ReadOrderedBytesToInt8(bytes)
	if err != nil {
		return nil, err
	}
	num, bRem, err := store.ReadOrderedBytesToUint64(bRem)
	if err != nil {
		return nil, err
	}

	exp, num, err = denormalizeForOrderedBytes(exp, num)
	if err != nil {
		return nil, err
	}

	p.exp = exp
	p.num = num

	return bRem, nil
}

// String returns string representation of the Price.
func (p Price) String() string {
	var expPart string
	if p.exp != 0 {
		expPart = exponentStr + strconv.Itoa(int(p.exp))
	}
	return strconv.FormatUint(p.num, 10) + expPart
}

// MarshalTo implements the gogo proto custom type interface.
func (p *Price) MarshalTo(data []byte) (n int, err error) {
	// TODO(dzmitryhil) implement
	return 0, nil
}

// Unmarshal implements the gogo proto custom type interface.
func (p *Price) Unmarshal(data []byte) error {
	// TODO(dzmitryhil) implement
	return nil
}

// Size implements the gogo proto custom type interface.
func (p *Price) Size() int {
	// TODO(dzmitryhil) implement
	return 0
}

// Marshal implements the gogo proto custom type interface.
func (p *Price) Marshal() ([]byte, error) {
	// TODO(dzmitryhil) implement
	return nil, nil
}

// normalizeForOrderedBytes normalizes the num part to have the same uint64 length for all prices stored and updates the
// exp correspondingly.
func normalizeForOrderedBytes(exp int8, num uint64) (int8, uint64, error) {
	srtNum := strconv.FormatUint(num, 10)
	offset := maxNumLen - len(srtNum)
	srtNum += strings.Repeat("0", offset)
	num, err := strconv.ParseUint(srtNum, 10, 64)
	if err != nil {
		return 0, 0, errors.Wrapf(err, "failed to parse uint64 from %s", srtNum)
	}
	exp -= int8(offset)

	return exp, num, nil
}

// denormalizeForOrderedBytes denormalizes the num part to initial (before normalization) and updates
// the exponent correspondingly.
func denormalizeForOrderedBytes(exp int8, num uint64) (int8, uint64, error) {
	srtNum := strconv.FormatUint(num, 10)
	srtNum = strings.TrimRight(srtNum, "0")
	offset := maxNumLen - len(srtNum)
	num, err := strconv.ParseUint(srtNum, 10, 64)
	if err != nil {
		return 0, 0, errors.Wrapf(err, "failed to parse uint64 from %s", srtNum)
	}
	exp += int8(offset)

	return exp, num, nil
}
