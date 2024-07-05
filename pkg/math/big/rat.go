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
