package big

import "math/big"

// NewRatFromBigInt returns *big.Rat from provided nom.
func NewRatFromBigInt(nom *big.Int) *big.Rat {
	return (&big.Rat{}).SetFrac(nom, big.NewInt(1))
}

// NewRatFromBigInts returns *big.Rat from provided nom and denom.
func NewRatFromBigInts(nom, denom *big.Int) *big.Rat {
	return (&big.Rat{}).SetFrac(nom, denom)
}

// RatGTE returns true if x is greater or equal to y.
func RatGTE(x, y *big.Rat) bool {
	return x.Cmp(y) != -1
}

// RatLTE returns true if x is lower or equal to y.
func RatLTE(x, y *big.Rat) bool {
	return x.Cmp(y) != 1
}
