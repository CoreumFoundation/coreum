package big

import "math/big"

// NewBigIntFromUint64 returns new *big.Int from uint64.
func NewBigIntFromUint64(x uint64) *big.Int {
	return (&big.Int{}).SetUint64(x)
}

// IntMul multiplies Int by Int.
func IntMul(x, y *big.Int) *big.Int {
	return (&big.Int{}).Mul(x, y)
}

// IntTenToThePower returns 10 to the power of x.
func IntTenToThePower(x *big.Int) *big.Int {
	return (&big.Int{}).Exp(big.NewInt(10), x, nil)
}
