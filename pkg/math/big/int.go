package big

import (
	"math/big"
)

// NewBigIntFromUint64 returns new *big.Int from uint64.
func NewBigIntFromUint64(x uint64) *big.Int {
	return (&big.Int{}).SetUint64(x)
}

// IntAdd returns sum of x+y.
func IntAdd(x, y *big.Int) *big.Int {
	return (&big.Int{}).Add(x, y)
}

// IntSub subtracts Int by Int.
func IntSub(x, y *big.Int) *big.Int {
	return (&big.Int{}).Sub(x, y)
}

// IntMul multiplies Int by Int.
func IntMul(x, y *big.Int) *big.Int {
	return (&big.Int{}).Mul(x, y)
}

// IntMulRatWithRemainder multiplies x *big.Int by y *big.Rat and returns *big.Int result with the remainder.
func IntMulRatWithRemainder(x *big.Int, y *big.Rat) (*big.Int, *big.Int) {
	num := IntMul(x, y.Num())
	denom := y.Denom()
	intPart := IntQuo(num, denom)
	return intPart, IntSub(num, IntMul(intPart, denom))
}

// IntQuo divides Int by Int.
func IntQuo(x, y *big.Int) *big.Int {
	return (&big.Int{}).Quo(x, y)
}

// IntTenToThePower returns 10 to the power of x.
func IntTenToThePower(x *big.Int) *big.Int {
	return (&big.Int{}).Exp(big.NewInt(10), x, nil)
}

// IntGTE returns true if x is greater or equal to y.
func IntGTE(x, y *big.Int) bool {
	return x.Cmp(y) != -1
}

// IntEqZero returns true if x is equal to zero.
func IntEqZero(x *big.Int) bool {
	return x.Sign() == 0
}
