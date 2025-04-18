package types_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v6/x/asset/nft/types"
)

func TestSymbolAndIDValidation(t *testing.T) {
	testCases := []struct {
		name        string
		value       string
		validSymbol bool
		validID     bool
	}{
		{
			name:        "containing dash",
			value:       "ABC-1",
			validSymbol: false,
			validID:     true,
		},
		{
			name:        "containing dot and upper case",
			value:       "ABC.1",
			validSymbol: true,
			validID:     true,
		},
		{
			name:        "containing dot",
			value:       "abc.1",
			validSymbol: true,
			validID:     true,
		},
		{
			name:        "containing colon",
			value:       "ABC:1",
			validSymbol: true,
			validID:     true,
		},
		{
			name:        "similar to ft denom",
			value:       "btc-devcore1phjrez5j2wp5qzp0zvlqavasvw60mkp2zmfe6h",
			validSymbol: false,
			validID:     true,
		},
		{
			name:        "just core",
			value:       "core",
			validSymbol: true,
			validID:     true,
		},
		{
			name:        "31 chars",
			value:       "A123456789012345678901234567890",
			validSymbol: true,
			validID:     true,
		},
		{
			name:        "32 chars",
			value:       "A1234567890123456789012345678901",
			validSymbol: false,
			validID:     true,
		},
		{
			name:        "start with number",
			value:       "3abc",
			validSymbol: false,
			validID:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateClassSymbol(tc.value)
			if tc.validSymbol {
				require.NoError(t, err, "expected no error for symbol %s", tc.value)
			} else {
				require.Error(t, err, "expected error for symbol %s", tc.value)
			}

			err = types.ValidateTokenID(tc.value)
			if tc.validID {
				require.NoError(t, err, "expected no error for nft id %s", tc.value)
			} else {
				require.Error(t, err, "expected error for nft id %s", tc.value)
			}
		})
	}
}

func TestFTDefinition_CheckFeatureAllowed(t *testing.T) {
	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	nonIssuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	type fields struct {
		Denom              string
		Issuer             string
		Features           []types.ClassFeature
		BurnRate           sdkmath.LegacyDec
		SendCommissionRate sdkmath.LegacyDec
	}
	type args struct {
		addr    sdk.AccAddress
		feature types.ClassFeature
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "whitelisting_feature_enabled_for_issuer",
			fields: fields{
				Issuer: issuer.String(),
				Features: []types.ClassFeature{
					types.ClassFeature_whitelisting,
				},
			},
			args: args{
				addr:    issuer,
				feature: types.ClassFeature_whitelisting,
			},
			wantErr: require.NoError,
		},
		{
			name: "burning_feature_always_enabled_for_issuer",
			fields: fields{
				Issuer: issuer.String(),
				Features: []types.ClassFeature{
					types.ClassFeature_burning,
				},
			},
			args: args{
				addr:    nonIssuer,
				feature: types.ClassFeature_burning,
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
				feature: types.ClassFeature_burning,
			},
			wantErr: require.NoError,
		},
		{
			name: "whitelisting_feature_disabled_for_non_issuer",
			fields: fields{
				Issuer: issuer.String(),
				Features: []types.ClassFeature{
					types.ClassFeature_whitelisting,
				},
			},
			args: args{
				addr:    nonIssuer,
				feature: types.ClassFeature_whitelisting,
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				if assert.ErrorIs(t, err, cosmoserrors.ErrUnauthorized) {
					return
				}
				t.FailNow()
			},
		},
		{
			name: "whitelisting_feature_disabled_for_issuer",
			fields: fields{
				Issuer: issuer.String(),
			},
			args: args{
				addr:    issuer,
				feature: types.ClassFeature_whitelisting,
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
		t.Run(tt.name, func(t *testing.T) {
			ftd := types.ClassDefinition{
				Issuer:   tt.fields.Issuer,
				Features: tt.fields.Features,
			}
			tt.wantErr(
				t,
				ftd.CheckFeatureAllowed(tt.args.addr, tt.args.feature),
				fmt.Sprintf("CheckFeatureAllowed(%v, %v)", tt.args.addr, tt.args.feature),
			)
		})
	}
}

func TestValidateClassFeatures(t *testing.T) {
	t.Parallel()

	type testCase struct {
		Name     string
		Features []types.ClassFeature
		Ok       bool
	}

	allFeatures := make([]types.ClassFeature, 0, len(types.ClassFeature_name))
	for f := range types.ClassFeature_name {
		allFeatures = append(allFeatures, types.ClassFeature(f))
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
			Features: []types.ClassFeature{},
			Ok:       true,
		},
		{
			Name: "with one",
			Features: []types.ClassFeature{
				types.ClassFeature_burning,
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
			Features: []types.ClassFeature{1000},
			Ok:       false,
		},
		{
			Name: "one normal + out of scope at the end",
			Features: []types.ClassFeature{
				types.ClassFeature_whitelisting,
				2000,
			},
			Ok: false,
		},
		{
			Name: "one normal + out of scope at the beginning",
			Features: []types.ClassFeature{
				3000,
				types.ClassFeature_whitelisting,
			},
			Ok: false,
		},
		{
			Name: "two normal + out of scope in the middle",
			Features: []types.ClassFeature{
				types.ClassFeature_whitelisting,
				4000,
				types.ClassFeature_freezing,
			},
			Ok: false,
		},
		{
			Name:     "all normal + out of scope in the middle",
			Features: append([]types.ClassFeature{5000}, allFeatures...),
			Ok:       false,
		},
		{
			Name:     "duplicated",
			Features: append([]types.ClassFeature{allFeatures[0]}, allFeatures...),
			Ok:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			requireT := require.New(t)
			err := types.ValidateClassFeatures(tc.Features)
			if tc.Ok {
				requireT.NoError(err)
			} else {
				requireT.Error(err)
			}
		})
	}
}
