package types

import (
	"encoding/json"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/cosmos/gogoproto/proto"
	"github.com/pkg/errors"

	cbig "github.com/CoreumFoundation/coreum/v5/pkg/math/big"
	"github.com/CoreumFoundation/coreum/v5/pkg/store"
)

var (
	_ proto.Marshaler   = Price{}
	_ proto.Unmarshaler = &Price{}
	_ proto.Sizer       = &Price{}
)

const (
	// MaxNumLen is max allowed num part length.
	MaxNumLen = 19
	// MinExp is the min allowed exponent. Technically it's limited by MinInt8 (-128) + `MaxNumLen` (required for
	//	normalization). But to make the range value easier for understanding and still keeping enough precision we set
	//	it to -100.
	MinExp = int8(-100)
	// MaxExp is the max allowed exponent. Technically it's limited by MaxInt8 (127) but to make it similar to
	// mixExp we set it to 100.
	MaxExp = int8(100)
	// ExponentSymbol is symbol used to represent the exponent in the string price.
	ExponentSymbol        = "e"
	orderedBytesPriceSize = store.Int8OrderedBytesSize + store.Uint64OrderedBytesSize
)

var priceRegex = regexp.MustCompile(`^(([1-9])|([1-9]\d*[1-9]))(e-?[1-9]\d*)?$`)

// Price is the price type.
//
//nolint:recvcheck // we are ok with combination of pointer and non-pointer receiver.
type Price struct {
	exp int8
	num uint64
}

// NewPrice returns new instance of the Price.
func NewPrice(num uint64, exp int8) (Price, error) {
	// the range check is required for the normalization
	if exp < MinExp || exp > MaxExp {
		return Price{}, errors.Errorf("invalid exponent %d, must be in the rage %d:%d", exp, MinExp, MaxExp)
	}
	return Price{
		exp: exp,
		num: num,
	}, nil
}

// NewPriceFromString returns new instance of the Price from string.
func NewPriceFromString(str string) (Price, error) {
	if !priceRegex.MatchString(str) {
		return Price{}, errors.Errorf("invalid price %s, must match %s", str, priceRegex.String())
	}

	parts := strings.Split(str, ExponentSymbol)
	numPart := parts[0]
	if len(numPart) > MaxNumLen {
		return Price{}, errors.Errorf("invalid price num part length, max %d", MaxNumLen)
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
		expPart := parts[1]
		var intExp int64
		intExp, err = strconv.ParseInt(expPart, 10, 8)
		if err != nil {
			return Price{}, errors.Errorf("invalid price exponent part %s", expPart)
		}
		if intExp == 0 {
			return Price{}, errors.New("zero exponent is prohibited")
		}
		// the range check is required for the normalization
		if int8(intExp) < MinExp || int8(intExp) > MaxExp {
			return Price{}, errors.Errorf("invalid exponent %d, must be in the rage %d:%d", intExp, MinExp, MaxExp)
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

// MustNewPriceFromString creates new instance of price from string or panics.
func MustNewPriceFromString(str string) Price {
	price, err := NewPriceFromString(str)
	if err != nil {
		panic(err)
	}

	return price
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

// Equal returns true if prices are equal.
func (p Price) Equal(p2 Price) bool {
	return p.exp == p2.exp && p.num == p2.num
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
		expPart = ExponentSymbol + strconv.Itoa(int(p.exp))
	}
	return strconv.FormatUint(p.num, 10) + expPart
}

// MarshalTo implements the gogo proto custom type interface.
func (p Price) MarshalTo(data []byte) (n int, err error) {
	bz, err := p.Marshal()
	if err != nil {
		return 0, err
	}

	n = copy(data, bz)
	return n, nil
}

// Size implements the gogo proto custom type interface.
func (p *Price) Size() int {
	bz, _ := p.Marshal()
	return len(bz)
}

// Marshal implements the gogo proto custom type interface.
func (p Price) Marshal() ([]byte, error) {
	return []byte(p.String()), nil
}

// Unmarshal implements the gogo proto custom type interface.
func (p *Price) Unmarshal(data []byte) error {
	price, err := NewPriceFromString(string(data))
	if err != nil {
		return err
	}
	p.num = price.num
	p.exp = price.exp

	return nil
}

// MarshalJSON defines custom encoding scheme.
func (p Price) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// UnmarshalJSON defines custom decoding scheme.
func (p *Price) UnmarshalJSON(bz []byte) error {
	var text string
	if err := json.Unmarshal(bz, &text); err != nil {
		return err
	}

	return p.Unmarshal([]byte(text))
}

// MarshalAmino overrides Amino binary marshalling.
func (p Price) MarshalAmino() ([]byte, error) { return p.Marshal() }

// UnmarshalAmino overrides Amino binary unmarshalling.
func (p *Price) UnmarshalAmino(bz []byte) error { return p.Unmarshal(bz) }

// normalizeForOrderedBytes normalizes the num part to have the same uint64 length for all prices stored and updates the
// exp correspondingly.
func normalizeForOrderedBytes(exp int8, num uint64) (int8, uint64, error) {
	srtNum := strconv.FormatUint(num, 10)
	offset := MaxNumLen - len(srtNum)
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
	offset := MaxNumLen - len(srtNum)
	num, err := strconv.ParseUint(srtNum, 10, 64)
	if err != nil {
		return 0, 0, errors.Wrapf(err, "failed to parse uint64 from %s", srtNum)
	}
	exp += int8(offset)

	return exp, num, nil
}
