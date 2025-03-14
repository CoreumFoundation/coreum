package big

import (
	"math/big"
)

// NewRatFromInt64 returns a *big.Rat from the provided int64 numerator.
func NewRatFromInt64(num int64) *big.Rat {
	return (&big.Rat{}).SetFrac(big.NewInt(num), big.NewInt(1))
}

// NewRatFromInts returns a *big.Rat from the provided int64 numerator.
func NewRatFromInts(num, denom int64) *big.Rat {
	return (&big.Rat{}).SetFrac(big.NewInt(num), big.NewInt(denom))
}

// NewRatFromBigInt returns a *big.Rat from the provided numerator.
func NewRatFromBigInt(num *big.Int) *big.Rat {
	return (&big.Rat{}).SetFrac(num, big.NewInt(1))
}

// NewRatFromBigInts returns a *big.Rat from the provided numerator and denominator.
func NewRatFromBigInts(num, denom *big.Int) *big.Rat {
	return (&big.Rat{}).SetFrac(num, denom)
}

// RatTenToThePower returns 10 to the power of x as *big.Rat.
func RatTenToThePower(power int64) *big.Rat {
	if power >= 0 {
		return (&big.Rat{}).SetFrac(IntTenToThePower(big.NewInt(power)), big.NewInt(1))
	}

	return (&big.Rat{}).SetFrac(big.NewInt(1), IntTenToThePower(big.NewInt(-power)))
}

// RatMul multiplies *big.Rat x by y and returns the result.
func RatMul(x, y *big.Rat) *big.Rat {
	return (&big.Rat{}).Mul(x, y)
}

// RatDiv divides *big.Rat x by y and returns the result.
func RatDiv(x, y *big.Rat) *big.Rat {
	return (&big.Rat{}).Mul(x, RatInv(y))
}

// RatQuoWithIntRemainder divides x by y and returns the integer quotient and remainder as *big.Int.
func RatQuoWithIntRemainder(x, y *big.Rat) (*big.Int, *big.Int) {
	num := IntMul(x.Num(), y.Denom())
	denom := IntMul(x.Denom(), y.Num())
	intPart := IntQuo(num, denom)
	return intPart, IntSub(num, IntMul(intPart, denom))
}

// RatLog10RoundUp returns exponent of the largest power of 10 that is less than or equal to x.
func RatLog10RoundUp(val *big.Rat) int64 {
	num := val.Num()
	denom := val.Denom()

	// exponent is difference between exponents in scientific notation
	// e.g. num: 30=3*10^1, denom: 900=9*10^2 => exponent = 2-1
	// to calculated exponent we use length of a string representation of a number.
	exponent := int64(len(num.String()) - len(denom.String()))

	// special case, when val is already a power of 10, then we keep exponent as it is (since val is already rounded up)
	if RatGTE(RatTenToThePower(exponent), val) {
		return exponent
	}

	// otherwise we increment exponent by 1 to round up
	return exponent + 1
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

// RatMin returns minimal of x and y.
func RatMin(x, y *big.Rat) *big.Rat {
	if x.Cmp(y) < 0 {
		return x
	}
	return y
}

// RatMax returns maximum of x and y.
func RatMax(x, y *big.Rat) *big.Rat {
	if x.Cmp(y) > 0 {
		return x
	}
	return y
}
