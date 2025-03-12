package keeper

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputePriceTick(t *testing.T) {
	// tests := []struct {
	// 	name  string
	// 	base  float64
	// 	quote float64
	// }{
	// 	{
	// 		name:  "3.0/27.123",
	// 		base:  3.0,
	// 		quote: 27.123,
	// 	},
	// 	{
	// 		name:  "10000.0/10000.0",
	// 		base:  10000.0,
	// 		quote: 10000.0,
	// 	},
	// 	{
	// 		name:  "3000.0/20.0",
	// 		base:  3000.0,
	// 		quote: 20.0,
	// 	},
	// 	{
	// 		name:  "300000.0/20.0",
	// 		base:  300000.0,
	// 		quote: 20.0,
	// 	},
	// 	{
	// 		name:  "2.0/2.0",
	// 		base:  2.0,
	// 		quote: 2.0,
	// 	},
	// 	{
	// 		name:  "100.0/1.0",
	// 		base:  100.0,
	// 		quote: 1.0,
	// 	},
	// 	{
	// 		name:  "3.0/1.0",
	// 		base:  3.0,
	// 		quote: 1.0,
	// 	},
	// 	{
	// 		name:  "3100000.0/8.0",
	// 		base:  3100000.0,
	// 		quote: 8.0,
	// 	},
	// 	{
	// 		name:  "0.00017/100",
	// 		base:  0.00017,
	// 		quote: 100,
	// 	},
	// 	{
	// 		name:  "0.000001/10000000",
	// 		base:  0.000001,
	// 		quote: 10000000,
	// 	},
	// 	{
	// 		name:  "100/1000000000000",
	// 		base:  100,
	// 		quote: 1000000000000,
	// 	},
	// }

	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		assertTickCalculations(t, tt.base, tt.quote)
	// 		assertTickCalculations(t, tt.quote, tt.base)
	// 	})
	// }

	type args struct {
		baseURA           int64
		quoteURA          int64
		priceTickExponent int32
	}
	tests := []struct {
		args args
		want *big.Rat
	}{
		{
			args: args{
				baseURA:           1,
				quoteURA:          1,
				priceTickExponent: -6,
			},
			want: big.NewRat(1, 1_000_000),
		},
		{
			args: args{
				baseURA:           1,
				quoteURA:          1,
				priceTickExponent: -8,
			},
			want: big.NewRat(1, 100_000_000),
		},
		{
			args: args{
				baseURA:           900_000,
				quoteURA:          300_000,
				priceTickExponent: -6,
			},
			want: big.NewRat(1, 1_000_000), // RatLog10RoundUp(300_000/900_000) = 0
		},
		{
			args: args{
				baseURA:           300_000,
				quoteURA:          900_000,
				priceTickExponent: -6,
			},
			want: big.NewRat(1, 100_000), // RatLog10RoundUp(900_000/300_000) = 1
		},
		{
			args: args{
				baseURA:           300_000,
				quoteURA:          30_000,
				priceTickExponent: -6,
			},
			want: big.NewRat(1, 10_000_000), // RatLog10RoundUp(30_000/300_000) = -1
		},
		{
			args: args{
				baseURA:           300_000,
				quoteURA:          30_000 - 1,
				priceTickExponent: -6,
			},
			want: big.NewRat(1, 10_000_000), // RatLog10RoundUp(29_999/300_000) = -1
		},
		{
			args: args{
				baseURA:           300_000 - 1,
				quoteURA:          30_000,
				priceTickExponent: -6,
			},
			want: big.NewRat(1, 1_000_000), // RatLog10RoundUp(30_000/299_999) = 0
		},
		{
			args: args{
				baseURA:           1_000_000,
				quoteURA:          1_000_000,
				priceTickExponent: -6,
			},
			want: big.NewRat(1, 1_000_000),
		},
		{
			args: args{
				baseURA:           1_000_000 + 1,
				quoteURA:          1_000_000,
				priceTickExponent: -6,
			},
			want: big.NewRat(1, 1_000_000), // RatLog10RoundUp(1_000_000/1_000_001) = 1
		},
		{
			args: args{
				baseURA:           1_000_000,
				quoteURA:          1_000_000 + 1,
				priceTickExponent: -6,
			},
			want: big.NewRat(1, 100_000), // RatLog10RoundUp(1_000_001/1_000_000) = 2
		},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("baseURA=%d,quoteURA=%d,priceTickExponent=%d", tt.args.baseURA, tt.args.quoteURA, tt.args.priceTickExponent)
		t.Run(name, func(t *testing.T) {
			actual := ComputePriceTick(big.NewInt(tt.args.baseURA), big.NewInt(tt.args.quoteURA), tt.args.priceTickExponent)
			assert.EqualValues(t, tt.want, actual, "want: %v actual: %v", tt.want, actual)
		})
	}
}
