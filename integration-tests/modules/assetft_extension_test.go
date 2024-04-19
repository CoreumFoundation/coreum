package modules

import (
	"encoding/json"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v4/integration-tests"
	"github.com/CoreumFoundation/coreum/v4/pkg/client"
	"github.com/CoreumFoundation/coreum/v4/testutil/integration"
	assetftkeeper "github.com/CoreumFoundation/coreum/v4/x/asset/ft/keeper"
	testcontracts "github.com/CoreumFoundation/coreum/v4/x/asset/ft/keeper/test-contracts"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
)

// TestAssetFTIssue tests issue functionality of fungible tokens.
func TestAssetFTExtensionIssue(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	assetFTClient := assetfttypes.NewQueryClient(chain.ClientContext)
	requireT := require.New(t)

	issuer := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.
			Add(sdk.NewInt(1000_000)), // one million added for uploading wasm code
	})

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactory().WithSimulateAndExecute(true), issuer, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdkmath.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_extensions,
		},
		URI:     "https://my-class-meta.invalid/1",
		URIHash: "content-hash",
		ExtensionSettings: &assetfttypes.ExtensionSettings{
			CodeId:           codeID,
			InstantiationMsg: []byte("{}"),
		},
	}

	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)

	receiver := chain.GenAccount()
	// sending 1 will succeed
	sendMsg := &banktypes.MsgSend{
		FromAddress: issueMsg.Issuer,
		ToAddress:   receiver.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	// sending 7 will fail
	sendMsg.Amount = sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(7)))
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.ErrorIs(err, assetfttypes.ErrExtensionCallFailed)

	// call directly from the user
	// sending 1 will succeed
	chain.FundAccountWithOptions(ctx, t, receiver, integration.BalancesOptions{
		Amount: sdk.NewInt(1000_000), // one million added for uploading wasm code
	})

	receiver2 := chain.GenAccount()
	token, err := assetFTClient.Token(ctx, &assetfttypes.QueryTokenRequest{Denom: denom})
	requireT.NoError(err)
	contractMsg := map[string]interface{}{
		assetftkeeper.ExtenstionTransferMethod: assetftkeeper.ExtensionTransferMsg{
			Amount: sdk.NewInt(1),
			Recipients: map[string]sdkmath.Int{
				receiver2.String(): sdk.NewInt(1),
			},
		},
	}
	contractMsgBytes, err := json.Marshal(contractMsg)
	requireT.NoError(err)
	_, err = chain.Wasm.ExecuteWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		receiver,
		token.Token.ExtensionCwAddress,
		contractMsgBytes,
		sdk.NewCoin(denom, sdk.NewInt(1)),
	)

	requireT.NoError(err)

	// sending 7 will fail
	requireT.NoError(err)
	contractMsg = map[string]interface{}{
		assetftkeeper.ExtenstionTransferMethod: assetftkeeper.ExtensionTransferMsg{
			Amount: sdk.NewInt(7),
			Recipients: map[string]sdkmath.Int{
				receiver2.String(): sdk.NewInt(7),
			},
		},
	}
	contractMsgBytes, err = json.Marshal(contractMsg)
	requireT.NoError(err)
	_, err = chain.Wasm.ExecuteWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		receiver,
		token.Token.ExtensionCwAddress,
		contractMsgBytes,
		sdk.NewCoin(denom, sdk.NewInt(7)),
	)
	requireT.ErrorIs(err, assetfttypes.ErrExtensionCallFailed)
}
