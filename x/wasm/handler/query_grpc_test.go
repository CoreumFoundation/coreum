package handler_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/CoreumFoundation/coreum/v5/testutil/simapp"
	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v5/x/wasm/handler"
)

func TestGRPCQuerier(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	issuer, _ := testApp.GenAccount(sdkCtx)
	settingsWithExtension := assetfttypes.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEFEXT",
		Subunit:       "defext",
		Precision:     6,
		InitialAmount: sdkmath.NewIntWithDecimal(1, 10),
	}
	denom, err := testApp.AssetFTKeeper.Issue(sdkCtx, settingsWithExtension)
	require.NoError(t, err)

	q := handler.NewGRPCQuerier(testApp.GRPCQueryRouter(), testApp.AppCodec())
	queryTokenReq := &assetfttypes.QueryTokenRequest{
		Denom: denom,
	}
	wasmGrpcData, err := testApp.AppCodec().Marshal(queryTokenReq)
	require.NoError(t, err)

	eg, _ := errgroup.WithContext(ctx)
	for range 1000 {
		// rebuild the ctx
		routineSDKCtx := testApp.BaseApp.NewContext(false)
		eg.Go(func() error {
			wasmGrpcReq := &wasmvmtypes.GrpcQuery{
				Data: wasmGrpcData,
				// url which corresponds query token
				Path: "/coreum.asset.ft.v1.Query/Token",
			}
			wasmGrpcRes, err := q.Query(routineSDKCtx, wasmGrpcReq)
			if err != nil {
				return err
			}

			queryTokenResData, err := gogoproto.Marshal(wasmGrpcRes)
			if err != nil {
				return err
			}

			queryTokenRes := &assetfttypes.QueryTokenResponse{}
			if err := testApp.AppCodec().Unmarshal(queryTokenResData, queryTokenRes); err != nil {
				return err
			}

			want := assetfttypes.Token{
				Denom:              denom,
				Issuer:             issuer.String(),
				Symbol:             settingsWithExtension.Symbol,
				Subunit:            settingsWithExtension.Subunit,
				Precision:          settingsWithExtension.Precision,
				BurnRate:           sdkmath.LegacyNewDec(0),
				SendCommissionRate: sdkmath.LegacyNewDec(0),
				Version:            assetfttypes.CurrentTokenVersion,
				Admin:              issuer.String(),
			}
			if !reflect.DeepEqual(want, queryTokenRes.Token) {
				return errors.Errorf("unexpected token, want:%v, got:%v", want, queryTokenRes.Token)
			}
			return nil
		})
	}

	require.NoError(t, eg.Wait())
}
