package big

import (
	"math/big"
)

// NewRatFromInt64 returns *big.Rat from provided int64 nom.
func NewRatFromInt64(nom int64) *big.Rat {
	return (&big.Rat{}).SetFrac(big.NewInt(nom), big.NewInt(1))
}

// NewRatFromBigInt returns *big.Rat from provided nom.
func NewRatFromBigInt(nom *big.Int) *big.Rat {
	return (&big.Rat{}).SetFrac(nom, big.NewInt(1))
}

// NewRatFromBigInts returns *big.Rat from provided nom and denom.
func NewRatFromBigInts(nom, denom *big.Int) *big.Rat {
	return (&big.Rat{}).SetFrac(nom, denom)
}

// RatMul multiplies *big.Rat x by y.
func RatMul(x, y *big.Rat) *big.Rat {
	return (&big.Rat{}).Mul(x, y)
}

// RatQuoWithIntRemainder divides x *big.Rat by y *big.Rat and returns *big.Int result with the remainder.
func RatQuoWithIntRemainder(x, y *big.Rat) (*big.Int, *big.Int) {
	num := IntMul(x.Num(), y.Denom())
	denom := IntMul(x.Denom(), y.Num())
	intPart := IntQuo(num, denom)
	return intPart, IntSub(num, IntMul(intPart, denom))
}

// RatGTE returns true if x is greater or equal to y.
func RatGTE(x, y *big.Rat) bool {
	return x.Cmp(y) != -1
}

// RatLTE returns true if x is lower or equal to y.
func RatLTE(x, y *big.Rat) bool {
	return x.Cmp(y) != 1
}

// RatLT returns true if x is lower than to y.
func RatLT(x, y *big.Rat) bool {
	return x.Cmp(y) == -1
}

// RatInv returns 1/x.
func RatInv(x *big.Rat) *big.Rat {
	return (&big.Rat{}).Inv(x)
}
