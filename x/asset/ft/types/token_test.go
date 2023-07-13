package types_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibctypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

func TestBuildDenom(t *testing.T) {
	subunit := "abc"
	addr, err := sdk.AccAddressFromBech32("devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5")
	require.NoError(t, err)

	denom := types.BuildDenom(subunit, addr)
	require.Equal(t, "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5", denom)
}

func TestValidatePrecision(t *testing.T) {
	testCases := []struct {
		precision   uint32
		expectError bool
	}{
		{precision: 1},
		{precision: 3},
		{precision: 10},
		{precision: types.MaxPrecision},
		{precision: 0, expectError: true},
		{precision: types.MaxPrecision + 1, expectError: true},
		{precision: 100_000, expectError: true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprint(tc), func(t *testing.T) {
			err := types.ValidatePrecision(tc.precision)
			if tc.expectError {
				assert.ErrorIs(t, err, types.ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSubunit(t *testing.T) {
	requireT := require.New(t)
	unacceptableSubunits := []string{
		"",
		"T",
		"ABC1",
		"ABC-1",
		"ABC/1",
		"btc-devcore1phjrez5j2wp5qzp0zvlqavasvw60mkp2zmfe6h",
		"BTC-devcore1phjrez5j2wp5qzp0zvlqavasvw60mkp2zmfe6h",
		"core",
		"ucore",
		"Coreum",
		"uCoreum",
		"COREeum",
		"A1234567890123456789012345678901234567890123456789",
		"Core",
		"uCore",
		"CORE",
		"UCORE",
		"3abc",
		"3ABC",
		"AB123456789012345678901234567890123456789012345678",
		ibctypes.DenomPrefix,
		ibctypes.DenomPrefix + "-",
		ibctypes.DenomPrefix + "/",
		ibctypes.DenomPrefix + "/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
	}

	acceptableSubunits := []string{
		"t",
		"abc1",
		"coreum",
		"ucoreum",
		"coreum",
		"ucoreum",
		"coreeum",
		"a1234567890123456789012345678901234567890123456789",
	}

	assertValidSubunit := func(symbol string, isValid bool) {
		err := types.ValidateSubunit(symbol)
		if isValid {
			requireT.NoError(err)
		} else {
			requireT.True(types.ErrInvalidInput.Is(err))
		}
	}

	for _, symbol := range unacceptableSubunits {
		assertValidSubunit(symbol, false)
	}

	for _, symbol := range acceptableSubunits {
		assertValidSubunit(symbol, true)
	}
}

func TestValidateSymbol(t *testing.T) {
	assertT := assert.New(t)
	unacceptableSymbols := []string{
		"",
		".",
		"-",
		"t$",
		"t ",
		"t",
		"t=",
		"t@",
		"t!",
		"ABC/1",
		"core",
		"ucore",
		"Core",
		"uCore",
		"CORE",
		"UCORE",
		"3abc",
		"3ABC",
		"t-",
		"t.",
	}

	acceptableSymbols := []string{
		"tt-",
		"ABC-1",
		"btc-devcore1phjrez5j2wp5qzp0zvlqavasvw60mkp2zmfe6h",
		"BTC-devcore1phjrez5j2wp5qzp0zvlqavasvw60mkp2zmfe6h",
		"abc1",
		"TTT",
		"ABC1",
		"coreum",
		"ucoreum",
		"Coreum",
		"uCoreum",
		"COREeum",
		"coreum",
		"ucoreum",
		"coreeum",
		"a1234567890123456789012345678901234567890123456789012345678901234567890",
		"AB1234567890123456789012345678901234567890123456789012345678901234567890",
	}

	assertValidSymbol := func(symbol string, isValid bool) {
		err := types.ValidateSymbol(symbol)
		if types.ErrInvalidInput.Is(err) == isValid {
			assertT.Failf("", "case: %s", symbol)
		}
	}

	for _, symbol := range unacceptableSymbols {
		assertValidSymbol(symbol, false)
	}

	for _, symbol := range acceptableSymbols {
		assertValidSymbol(symbol, true)
	}
}

func TestValidateFeatures(t *testing.T) {
	t.Parallel()

	assertT := assert.New(t)

	type testCase struct {
		Name     string
		Features []types.Feature
		Ok       bool
	}

	allFeatures := make([]types.Feature, 0, len(types.Feature_name))
	for f := range types.Feature_name {
		allFeatures = append(allFeatures, types.Feature(f))
	}

	testCases := []testCase{
		// valid cases
		{
			Name:     "nil",
			Features: nil,
			Ok:       true,
		},
		{
			Name:     "empty",
			Features: []types.Feature{},
			Ok:       true,
		},
		{
			Name: "with one",
			Features: []types.Feature{
				types.Feature_burning,
			},
			Ok: true,
		},
		{
			Name: "ibc only",
			Features: []types.Feature{
				types.Feature_ibc,
			},
			Ok: true,
		},
		{
			Name:     "all",
			Features: allFeatures,
			Ok:       true,
		},
		{
			Name:     "all except one",
			Features: allFeatures[1:],
			Ok:       true,
		},

		// invalid cases
		{
			Name:     "single out of scope",
			Features: []types.Feature{1000},
			Ok:       false,
		},
		{
			Name: "one normal + out of scope at the end",
			Features: []types.Feature{
				types.Feature_minting,
				2000,
			},
			Ok: false,
		},
		{
			Name: "one normal + out of scope at the beginning",
			Features: []types.Feature{
				3000,
				types.Feature_minting,
			},
			Ok: false,
		},
		{
			Name: "two normal + out of scope in the middle",
			Features: []types.Feature{
				types.Feature_burning,
				4000,
				types.Feature_minting,
			},
			Ok: false,
		},
		{
			Name:     "all normal + out of scope in the middle",
			Features: append([]types.Feature{5000}, allFeatures...),
			Ok:       false,
		},
		{
			Name:     "duplicated",
			Features: append([]types.Feature{allFeatures[0]}, allFeatures...),
			Ok:       false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			err := types.ValidateFeatures(tc.Features)
			if tc.Ok {
				assertT.NoError(err)
			} else {
				assertT.Error(err)
			}
		})
	}
}

//nolint:dupl // We don't care
func TestValidateBurnRate(t *testing.T) {
	testCases := []struct {
		rate    string
		invalid bool
	}{
		{
			rate: "0",
		},
		{
			rate: "0.00",
		},
		{
			rate: "1.00",
		},
		{
			rate: "0.10",
		},
		{
			rate: "0.10000",
		},
		{
			rate: "0.0001",
		},
		{
			rate:    "0.00001",
			invalid: true,
		},
		{
			rate:    "-0.01",
			invalid: true,
		},
		{
			rate:    "-1.0",
			invalid: true,
		},
		{
			rate:    "1.0002",
			invalid: true,
		},
		{
			rate:    "1.00023",
			invalid: true,
		},
		{
			rate:    "0.12345",
			invalid: true,
		},
		{
			rate:    "0.000000000000000001",
			invalid: true,
		},
		{
			rate:    "0.0000000000000000001",
			invalid: true,
		},
	}

	parseAndValidate := func(in string) error {
		rate, err := sdk.NewDecFromStr(in)
		if err != nil {
			return err
		}

		err = types.ValidateBurnRate(rate)
		return err
	}

	for _, tc := range testCases {
		tc := tc
		name := fmt.Sprintf("%+v", tc)
		t.Run(name, func(t *testing.T) {
			assertT := assert.New(t)
			err := parseAndValidate(tc.rate)
			if tc.invalid {
				assertT.Error(err)
			} else {
				assertT.NoError(err)
			}
		})
	}
}

//nolint:dupl // We don't care
func TestValidateSendCommissionRate(t *testing.T) {
	testCases := []struct {
		rate    string
		invalid bool
	}{
		{
			rate: "0",
		},
		{
			rate: "0.00",
		},
		{
			rate: "1.00",
		},
		{
			rate: "0.10",
		},
		{
			rate: "0.10000",
		},
		{
			rate: "0.0001",
		},
		{
			rate:    "0.00001",
			invalid: true,
		},
		{
			rate:    "-0.01",
			invalid: true,
		},
		{
			rate:    "-1.0",
			invalid: true,
		},
		{
			rate:    "1.0002",
			invalid: true,
		},
		{
			rate:    "1.00023",
			invalid: true,
		},
		{
			rate:    "0.12345",
			invalid: true,
		},
		{
			rate:    "0.000000000000000001",
			invalid: true,
		},
		{
			rate:    "0.0000000000000000001",
			invalid: true,
		},
	}

	parseAndValidate := func(in string) error {
		rate, err := sdk.NewDecFromStr(in)
		if err != nil {
			return err
		}

		err = types.ValidateSendCommissionRate(rate)
		return err
	}

	for _, tc := range testCases {
		tc := tc
		name := fmt.Sprintf("%+v", tc)
		t.Run(name, func(t *testing.T) {
			assertT := assert.New(t)
			err := parseAndValidate(tc.rate)
			if tc.invalid {
				assertT.Error(err)
			} else {
				assertT.NoError(err)
			}
		})
	}
}

func TestDefinition_CheckFeatureAllowed(t *testing.T) {
	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	nonIssuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	type fields struct {
		Denom              string
		Issuer             string
		Features           []types.Feature
		BurnRate           sdk.Dec
		SendCommissionRate sdk.Dec
	}
	type args struct {
		addr    sdk.AccAddress
		feature types.Feature
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "minting_feature_enabled_for_issuer",
			fields: fields{
				Issuer: issuer.String(),
				Features: []types.Feature{
					types.Feature_minting,
				},
			},
			args: args{
				addr:    issuer,
				feature: types.Feature_minting,
			},
			wantErr: require.NoError,
		},
		{
			name: "burning_feature_always_enabled_for_issuer",
			fields: fields{
				Issuer: issuer.String(),
				Features: []types.Feature{
					types.Feature_burning,
				},
			},
			args: args{
				addr:    nonIssuer,
				feature: types.Feature_burning,
			},
			wantErr: require.NoError,
		},
		{
			name: "burning_feature_enabled_for_non_issuer",
			fields: fields{
				Issuer: issuer.String(),
			},
			args: args{
				addr:    issuer,
				feature: types.Feature_burning,
			},
			wantErr: require.NoError,
		},
		{
			name: "minting_feature_disabled_for_non_issuer",
			fields: fields{
				Issuer: issuer.String(),
				Features: []types.Feature{
					types.Feature_minting,
				},
			},
			args: args{
				addr:    nonIssuer,
				feature: types.Feature_minting,
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				if assert.ErrorIs(t, err, sdkerrors.ErrUnauthorized) {
					return
				}
				t.FailNow()
			},
		},
		{
			name: "minting_feature_disabled_for_issuer",
			fields: fields{
				Issuer: issuer.String(),
			},
			args: args{
				addr:    issuer,
				feature: types.Feature_minting,
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				if assert.ErrorIs(t, err, types.ErrFeatureDisabled) {
					return
				}
				t.FailNow()
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			def := types.Definition{
				Denom:              tt.fields.Denom,
				Issuer:             tt.fields.Issuer,
				Features:           tt.fields.Features,
				BurnRate:           tt.fields.BurnRate,
				SendCommissionRate: tt.fields.SendCommissionRate,
			}
			tt.wantErr(t, def.CheckFeatureAllowed(tt.args.addr, tt.args.feature), fmt.Sprintf("CheckFeatureAllowed(%v, %v)", tt.args.addr, tt.args.feature))
		})
	}
}
