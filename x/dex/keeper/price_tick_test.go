package keeper

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	cbig "github.com/CoreumFoundation/coreum/v5/pkg/math/big"
)

func Test_computePriceTick(t *testing.T) {
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
		// Examples from spec/price-and-amount.md
		// All amounts are represented as subunits on chain so they are multiplied by 10^6.
		{
			// BTC/USDT
			args: args{
				baseURA:           11,
				quoteURA:          1_000_000,
				priceTickExponent: -6,
			},
			want: big.NewRat(1, 10), // RatLog10RoundUp(1_000_000/11) = 5
		},
		{
			// ETH/USDT
			args: args{
				baseURA:           333,
				quoteURA:          1_000_000,
				priceTickExponent: -6,
			},
			want: big.NewRat(1, 100), // RatLog10RoundUp(1_000_000/333) = 4
		},
		{
			// TRX/USDT
			args: args{
				baseURA:           4_500_000,
				quoteURA:          1_000_000,
				priceTickExponent: -6,
			},
			want: big.NewRat(1, 1_000_000), // RatLog10RoundUp(1_000_000/4_500_000) = 0
		},
		{
			// PEPE/USDT
			args: args{
				baseURA:           80_000 * 1_000_000,
				quoteURA:          1_000_000,
				priceTickExponent: -6,
			},
			want: big.NewRat(1, 10_000_000_000), // RatLog10RoundUp(1_000_000/(80_000*1_000_000)) = 0
		},
		{
			// ETH/BTC
			args: args{
				baseURA:           333,
				quoteURA:          11,
				priceTickExponent: -6,
			},
			want: big.NewRat(1, 10_000_000), // RatLog10RoundUp(11/333) = -1
		},
	}
	for _, tt := range tests {
		name := fmt.Sprintf(
			"baseURA=%d,quoteURA=%d,priceTickExponent=%d",
			tt.args.baseURA, tt.args.quoteURA, tt.args.priceTickExponent,
		)
		t.Run(name, func(t *testing.T) {
			actual := computePriceTick(big.NewInt(tt.args.baseURA), big.NewInt(tt.args.quoteURA), tt.args.priceTickExponent)
			assert.EqualValues(t, tt.want, actual, "want: %v actual: %v", tt.want, actual)
		})
	}
}

func Test_validatePriceTick(t *testing.T) {
	type args struct {
		price     *big.Rat
		priceTick *big.Rat
	}
	tests := []struct {
		args      args
		wantError bool
	}{
		{
			args: args{
				price:     cbig.NewRatFromInts(123, 100), // 1.23
				priceTick: cbig.NewRatFromInts(1, 100),   // 0.01
			},
			wantError: false,
		},
		{
			args: args{
				price:     cbig.NewRatFromInts(123, 1), // 123
				priceTick: cbig.NewRatFromInts(1, 100), // 0.01
			},
			wantError: false,
		},
		{
			args: args{
				price:     cbig.NewRatFromInts(9, 100), // 0.09
				priceTick: cbig.NewRatFromInts(1, 10),  // 0.1
			},
			wantError: true,
		},
		{
			args: args{
				price:     cbig.NewRatFromInts(12, 100), // 0.12
				priceTick: cbig.NewRatFromInts(1, 10),   // 0.1
			},
			wantError: true,
		},
		{
			args: args{
				price:     cbig.NewRatFromInts(123, 1), // 123
				priceTick: cbig.NewRatFromInts(10, 1),  // 10
			},
			wantError: true,
		},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("price=%s,priceTick=%s", tt.args.price.String(), tt.args.priceTick.String())
		t.Run(name, func(t *testing.T) {
			err := validatePriceTick(tt.args.price, tt.args.priceTick)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
