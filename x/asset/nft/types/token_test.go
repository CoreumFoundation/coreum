package types_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
)

//nolint:funlen,nosnakecase // this is complex test scenario and breaking it down is not helpful
func TestFTDefinition_CheckFeatureAllowed(t *testing.T) {
	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	nonIssuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	type fields struct {
		Denom              string
		Issuer             string
		Features           []types.ClassFeature
		BurnRate           sdk.Dec
		SendCommissionRate sdk.Dec
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
				if assert.ErrorIs(t, err, sdkerrors.ErrUnauthorized) {
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ftd := types.ClassDefinition{
				Issuer:   tt.fields.Issuer,
				Features: tt.fields.Features,
			}
			tt.wantErr(t, ftd.CheckFeatureAllowed(tt.args.addr, tt.args.feature), fmt.Sprintf("CheckFeatureAllowed(%v, %v)", tt.args.addr, tt.args.feature))
		})
	}
}
