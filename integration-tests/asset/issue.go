package asset

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	assettypes "github.com/CoreumFoundation/coreum/x/asset/types"
)

// TestIssueFTAsset checks that FT asset is issued.
func TestIssueFTAsset(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	chainContext := chain.ClientContext

	assetClient := assettypes.NewQueryClient(chainContext)
	bankClient := banktypes.NewQueryClient(chainContext)

	issuer := chain.RandomWallet()
	requireT.NoError(chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(
			issuer,
			chain.NewCoin(testing.ComputeNeededBalance(
				chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
				chain.GasLimitByMsgs(&assettypes.MsgIssueAsset{}),
				1,
				sdk.NewInt(0),
			)),
		),
	))

	// Issue the new asset
	msg := &assettypes.MsgIssueAsset{
		From: issuer.String(),
		Definition: &assettypes.AssetDefinition{
			Recipient:   issuer.String(),
			Type:        assettypes.AssetType_FT, //nolint:nosnakecase // protogen
			Code:        "BTC",
			Description: "BTC Description",
			Ft: &assettypes.FTCustomDefinition{
				Precision:     6,
				InitialAmount: sdk.NewInt(777),
			},
		},
	}

	res, err := tx.BroadcastTx(
		ctx,
		chainContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)
	requireT.NoError(err)

	eventAssetIssuedName := proto.MessageName(&assettypes.EventAssetIssued{})
	assetIDStr, ok := client.FindEventAttribute(sdk.StringifyEvents(res.Events), eventAssetIssuedName, "id")
	// the typed events are decoded with the strings escape
	assetIDStr = strings.ReplaceAll(assetIDStr, "\"", "")
	requireT.True(ok)
	assetID, err := strconv.ParseUint(assetIDStr, 10, 64)
	requireT.NoError(err)
	requireT.True(assetID > 0)

	// query for the asset the check what is stored
	assetRes, err := assetClient.Asset(ctx, &assettypes.QueryAssetRequest{
		Id: assetID,
	})
	requireT.NoError(err)

	expectedDefinition := msg.Definition
	denomName := fmt.Sprintf("%s%s%d", assettypes.ModuleName, msg.Definition.Code, assetID)
	denomBaseName := fmt.Sprintf("b%s", denomName)
	expectedDefinition.Ft.DenomName = denomName
	expectedDefinition.Ft.DenomBaseName = denomBaseName
	expectedAsset := assettypes.Asset{
		Id:         assetID,
		Definition: expectedDefinition,
	}
	requireT.Equal(expectedAsset, assetRes.Asset)

	// query balance
	assetBalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: issuer.String(),
		Denom:   denomBaseName,
	})
	requireT.NoError(err)
	assetBalance := assetBalanceRes.Balance

	expectedBalance := msg.Definition.Ft.InitialAmount.Mul(sdk.NewIntWithDecimal(1, int(msg.Definition.Ft.Precision)))
	requireT.Equal(sdk.NewCoin(denomBaseName, expectedBalance).String(), assetBalance.String())
}
