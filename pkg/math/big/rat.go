package big

import (
	"math/big"
)

// NewRatFromInt64 returns a *big.Rat from the provided int64 numerator.
func NewRatFromInt64(nom int64) *big.Rat {
	return (&big.Rat{}).SetFrac(big.NewInt(nom), big.NewInt(1))
}

// NewRatFromBigInt returns a *big.Rat from the provided numerator.
func NewRatFromBigInt(nom *big.Int) *big.Rat {
	return (&big.Rat{}).SetFrac(nom, big.NewInt(1))
}

// NewRatFromBigInts returns a *big.Rat from the provided numerator and denominator.
func NewRatFromBigInts(nom, denom *big.Int) *big.Rat {
	return (&big.Rat{}).SetFrac(nom, denom)
}

// RatMul multiplies *big.Rat x by y and returns the result.
func RatMul(x, y *big.Rat) *big.Rat {
	return (&big.Rat{}).Mul(x, y)
}

// RatQuoWithIntRemainder divides x by y and returns the integer quotient and remainder as *big.Int.
func RatQuoWithIntRemainder(x, y *big.Rat) (*big.Int, *big.Int) {
	num := IntMul(x.Num(), y.Denom())
	denom := IntMul(x.Denom(), y.Num())
	intPart := IntQuo(num, denom)
	return intPart, IntSub(num, IntMul(intPart, denom))
}

// RatEQ returns true if x is equal to y.
func RatEQ(x, y *big.Rat) bool {
	return x.Cmp(y) == 0
}

// RatGTE returns true if x is greater than or equal to y.
func RatGTE(x, y *big.Rat) bool {
	return x.Cmp(y) != -1
}

// RatGT returns true if x is greater than y.
func RatGT(x, y *big.Rat) bool {
	return x.Cmp(y) == 1
}

// RatLTE returns true if x is less than or equal to y.
func RatLTE(x, y *big.Rat) bool {
	return x.Cmp(y) != 1
}

// RatLT returns true if x is less than y.
func RatLT(x, y *big.Rat) bool {
	return x.Cmp(y) == -1
}

// RatInv returns the inverse of x (1/x).
func RatInv(x *big.Rat) *big.Rat {
	return (&big.Rat{}).Inv(x)
}

// RatIsZero returns true if x is equal to zero.
func RatIsZero(x *big.Rat) bool {
	return x.Cmp(big.NewRat(0, 1)) == 0
}

func RatMin(x, y *big.Rat) *big.Rat {
	if x.Cmp(y) < 0 {
		return x
	}
	return y
}
