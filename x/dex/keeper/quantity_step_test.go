package keeper

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_computeQuantityStep(t *testing.T) {
	type args struct {
		baseURA              int64
		quantityStepExponent int32
	}
	tests := []struct {
		args args
		want *big.Int
	}{
		{
			args: args{baseURA: 1_000_000, quantityStepExponent: -2},
			want: big.NewInt(10_000),
		},
		{
			args: args{baseURA: 1_000_000 - 1, quantityStepExponent: -2},
			want: big.NewInt(10_000),
		},
		{
			args: args{baseURA: 1_000_000 + 1, quantityStepExponent: -2},
			want: big.NewInt(100_000),
		},
		{
			args: args{baseURA: 500_000, quantityStepExponent: -4},
			want: big.NewInt(100),
		},
		{
			args: args{baseURA: 10, quantityStepExponent: -2},
			want: big.NewInt(1),
		},
	}
	for _, tt := range tests {
		name := fmt.Sprintf(
			"baseURA=%d,quantityStepExponent=%d",
			tt.args.baseURA, tt.args.quantityStepExponent,
		)
		t.Run(name, func(t *testing.T) {
			actual := computeQuantityStep(big.NewInt(tt.args.baseURA), tt.args.quantityStepExponent)
			assert.EqualValues(t, tt.want, actual, "want: %v actual: %v", tt.want, actual)
		})
	}
}

func Test_isQuantityStepValid(t *testing.T) {
	type args struct {
		quantity     *big.Int
		quantityStep *big.Int
	}
	tests := []struct {
		args args
		want bool
	}{
		{
			args: args{
				quantity:     big.NewInt(100_000),
				quantityStep: big.NewInt(1_000),
			},
			want: true,
		},
		{
			args: args{
				quantity:     big.NewInt(500),
				quantityStep: big.NewInt(1_000),
			},
			want: false,
		},
		{
			args: args{
				quantity:     big.NewInt(999),
				quantityStep: big.NewInt(1),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("quantity=%s,quantityStep=%s", tt.args.quantity.String(), tt.args.quantityStep.String())
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, isQuantityStepValid(tt.args.quantity, tt.args.quantityStep))
		})
	}
}
