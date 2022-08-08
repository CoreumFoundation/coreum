package types

import (
	"encoding/json"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

// Int is a simple wrapper around big.Int, simplifying the way math is done and also ensuring that each operation
// creates a copy of the big.Int
type Int struct {
	i *big.Int
}

// NewInt creates Int from int64
func NewInt(number int64) Int {
	return Int{i: big.NewInt(number)}
}

// NewIntFromString creates Int from string
func NewIntFromString(number string) (Int, error) {
	i, ok := new(big.Int).SetString(number, 10)
	if !ok {
		return Int{}, errors.Errorf("string '%s' does not represent valid integer", number)
	}
	return Int{i: i}, nil
}

// NewIntFromSDK converts sdk.Int to Int
func NewIntFromSDK(number sdk.Int) Int {
	return Int{i: new(big.Int).Set(number.BigInt())}
}

// BigInt returns internal *big.Int value
// FIXME (wojtek): Remove once crust uses Int type
func (i1 Int) BigInt() *big.Int {
	return i1.i
}

// Add returns i1+i2
func (i1 Int) Add(i2 Int) Int {
	return Int{i: new(big.Int).Add(i1.i, i2.i)}
}

// Sub returns i1-i2
func (i1 Int) Sub(i2 Int) Int {
	return Int{i: new(big.Int).Sub(i1.i, i2.i)}
}

// Mul returns i1*i2
func (i1 Int) Mul(i2 Int) Int {
	return Int{i: new(big.Int).Mul(i1.i, i2.i)}
}

// IsDefault returns true if object was created without initialization
func (i1 Int) IsDefault() bool {
	return i1.i == nil
}

// Equal compares two Ints
func (i1 Int) Equal(i2 Int) bool {
	return i1.i.Cmp(i2.i) == 0
}

// GT returns true if first Int is greater than second
func (i1 Int) GT(i2 Int) bool {
	return i1.i.Cmp(i2.i) == 1
}

// GTE returns true if receiver Int is greater than or equal to the parameter
// Int.
func (i1 Int) GTE(i2 Int) bool {
	return i1.i.Cmp(i2.i) >= 0
}

// LT returns true if first Int is lesser than second
func (i1 Int) LT(i2 Int) bool {
	return i1.i.Cmp(i2.i) == -1
}

// LTE returns true if first Int is less than or equal to second
func (i1 Int) LTE(i2 Int) bool {
	return i1.i.Cmp(i2.i) <= 0
}

// String returns string representation of the number
func (i1 Int) String() string {
	return i1.i.String()
}

// MarshalJSON defines custom encoding scheme
func (i1 Int) MarshalJSON() ([]byte, error) {
	if i1.i == nil {
		return nil, errors.New("internal value is nil")
	}

	text, err := i1.i.MarshalText()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	marshaled, err := json.Marshal(string(text))
	return marshaled, errors.WithStack(err)
}

// UnmarshalJSON defines custom decoding scheme
func (i1 *Int) UnmarshalJSON(bz []byte) error {
	if i1.i == nil {
		i1.i = new(big.Int)
	}

	var text string
	if err := json.Unmarshal(bz, &text); err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(i1.i.UnmarshalText([]byte(text)))
}

// IntToSDK converts Int to sdk.Int
func IntToSDK(number Int) sdk.Int {
	return sdk.NewIntFromBigInt(number.i)
}
