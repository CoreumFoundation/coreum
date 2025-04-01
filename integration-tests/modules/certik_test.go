//go:build integrationtests

package modules

import (
	"context"
	"encoding/json"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v5/integration-tests"
	"github.com/CoreumFoundation/coreum/v5/pkg/client"
	"github.com/CoreumFoundation/coreum/v5/testutil/integration"
	testcontracts "github.com/CoreumFoundation/coreum/v5/x/asset/ft/keeper/test-contracts"
	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	dextypes "github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

func TestCertikPoc(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)

	assetFTClint := assetfttypes.NewQueryClient(chain.ClientContext)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)
	dexReserver := dexParamsRes.Params.OrderReserve

	admin := chain.GenAccount()
	acc1 := chain.GenAccount()
	acc2 := chain.GenAccount()

	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.
			AddRaw(1_000_000_000_000),
	})
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		// message + order reserve
		Amount: sdkmath.NewInt(500_000_000).
			Add(dexReserver.Amount),
	})
	chain.FundAccountWithOptions(ctx, t, acc2, integration.BalancesOptions{
		Amount: sdkmath.NewInt(500_000_000).
			AddRaw(100_000_000).
			Add(dexReserver.Amount), // message  + balance to place an order + order reserve
	})

	codeID, err1 := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactory().WithSimulateAndExecute(true), admin, testcontracts.CertikPocWasm,
	)
	requireT.NoError(err1)

	// issue tokenA
	issueMsgA := &assetfttypes.MsgIssue{
		Issuer:        admin.String(),
		Symbol:        "TKNA",
		Subunit:       "ua",
		Precision:     6,
		Description:   "TKNA Description",
		InitialAmount: sdkmath.NewInt(100000000),
		Features:      []assetfttypes.Feature{},
		URI:           "https://my-class-meta.invalid/1",
		URIHash:       "content-hash",
	}
	denomA := assetfttypes.BuildDenom(issueMsgA.Subunit, admin)
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsgA)),
		issueMsgA,
	)
	requireT.NoError(err)

	//nolint:tagliatelle // these will be exposed to rust and must be snake case.
	issuanceMsg := struct {
		ExtraData string `json:"extra_data"`
	}{
		ExtraData: denomA,
	}

	issuanceMsgBytes, err := json.Marshal(issuanceMsg)
	requireT.NoError(err)

	attachedFund := chain.NewCoin(sdkmath.NewInt(1000000))
	issueMsgB := &assetfttypes.MsgIssue{
		Issuer:        admin.String(),
		Symbol:        "TKNB",
		Subunit:       "ub",
		Precision:     6,
		Description:   "TKNB Description",
		InitialAmount: sdkmath.NewInt(100000000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_extension,
		},
		URI:     "https://my-class-meta.invalid/1",
		URIHash: "content-hash",
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId:      codeID,
			Funds:       sdk.NewCoins(attachedFund),
			Label:       "testing-hack",
			IssuanceMsg: issuanceMsgBytes,
		},
	}

	denomB := assetfttypes.BuildDenom(issueMsgB.Subunit, admin)
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactoryAuto(),
		issueMsgB,
	)
	requireT.NoError(err)
	// get extension contract addr
	denomBTokenRes, err := assetFTClint.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: denomB,
	})
	requireT.NoError(err)
	tokenBExtensionAddress := denomBTokenRes.Token.ExtensionCWAddress

	// send from admin to cw to place an order
	_, err = sendFromAdmin(ctx, chain, admin, tokenBExtensionAddress, "udevcore", sdkmath.NewInt(10000000))
	requireT.NoError(err)
	_, err = sendFromAdmin(ctx, chain, admin, tokenBExtensionAddress, denomA, sdkmath.NewInt(100000))
	requireT.NoError(err)
	_, err = sendFromAdmin(ctx, chain, admin, tokenBExtensionAddress, denomB, sdkmath.NewInt(100000))
	requireT.NoError(err)
	// send from admin to acc1 to place an order
	_, err = sendFromAdmin(ctx, chain, admin, acc1.String(), denomA, sdkmath.NewInt(100000))
	requireT.NoError(err)
	_, err = sendFromAdmin(ctx, chain, admin, acc1.String(), denomB, sdkmath.NewInt(100000))
	requireT.NoError(err)

	// send from admin to acc2 to place an order
	_, err = sendFromAdmin(ctx, chain, admin, acc2.String(), denomA, sdkmath.NewInt(100000))
	requireT.NoError(err)
	_, err = sendFromAdmin(ctx, chain, admin, acc2.String(), denomB, sdkmath.NewInt(100000))
	requireT.NoError(err)
	printBalanceResponse := func(prefix string, res *assetfttypes.QueryBalanceResponse) {
		t.Log(prefix + ":")
		t.Log("  Balance: " + res.Balance.String())
		t.Log("  Whitelisted: " + res.Whitelisted.String())
		t.Log("  Frozen: " + res.Frozen.String())
		t.Log("  Locked: " + res.Locked.String())
		t.Log("  LockedInVesting: " + res.LockedInVesting.String())
		t.Log("  LockedInDEX: " + res.LockedInDEX.String())
		t.Log("  ExpectedToReceiveInDEX: " + res.ExpectedToReceiveInDEX.String())
	}

	// place 3 Buy order from acc1(BUY A(without extension) SELL B(with extension))
	_, err = placeSellOrder(ctx, chain, acc1, "id1", denomA, denomB, "1", sdkmath.NewInt(120))
	requireT.NoError(err)
	_, err = placeSellOrder(ctx, chain, acc1, "id2", denomA, denomB, "1", sdkmath.NewInt(120))
	requireT.NoError(err)
	_, err = placeSellOrder(ctx, chain, acc1, "id3", denomA, denomB, "1e2", sdkmath.NewInt(99))
	requireT.NoError(err)

	acc1ABalanceRes, acc1BBalanceRes, acc2ABalanceRes, acc2BBalanceRes, cwABalanceRes, cwBBalanceRes := checkBalance(
		ctx, assetFTClint, acc1.String(), acc2.String(), tokenBExtensionAddress, denomA, denomB,
	)
	t.Log("--------BALANCE---------")
	t.Log("acc1A:")
	printBalanceResponse("acc1A", acc1ABalanceRes)

	t.Log("acc1B:")
	printBalanceResponse("acc1B", acc1BBalanceRes)

	t.Log("acc2A:")
	printBalanceResponse("acc2A", acc2ABalanceRes)

	t.Log("acc2B:")
	printBalanceResponse("acc2B", acc2BBalanceRes)

	t.Log("cwA:")
	printBalanceResponse("cwA", cwABalanceRes)

	t.Log("cwB:")
	printBalanceResponse("cwB", cwBBalanceRes)

	t.Log("-----------------------")

	// place SELL order from acc2(SELL A(without extension) BUY B(with extension))
	msg3Resp, err := placeBuyOrder(ctx, chain, acc2, "hackid0", denomA, denomB, "1e1", sdkmath.NewInt(100))
	requireT.NoError(err)
	for _, event := range msg3Resp.Events {
		t.Log("-----------")
		t.Log(event.Type)
		for _, attr := range event.Attributes {
			t.Log(attr.Key + ": " + attr.Value)
		}
	}
	requireT.NoError(err)
	acc1ABalanceRes, acc1BBalanceRes, acc2ABalanceRes, acc2BBalanceRes, cwABalanceRes, cwBBalanceRes = checkBalance(
		ctx, assetFTClint, acc1.String(), acc2.String(), tokenBExtensionAddress, denomA, denomB,
	)
	t.Log("--------BALANCE---------")
	t.Log("acc1A:")
	printBalanceResponse("acc1A", acc1ABalanceRes)

	t.Log("acc1B:")
	printBalanceResponse("acc1B", acc1BBalanceRes)

	t.Log("acc2A:")
	printBalanceResponse("acc2A", acc2ABalanceRes)

	t.Log("acc2B:")
	printBalanceResponse("acc2B", acc2BBalanceRes)

	t.Log("cwA:")
	printBalanceResponse("cwA", cwABalanceRes)

	t.Log("cwB:")
	printBalanceResponse("cwB", cwBBalanceRes)

	t.Log("-----------------------")

	requireT.Equal("99", acc1ABalanceRes.LockedInDEX.String())
}

func checkBalance(ctx context.Context, assetFTClint assetfttypes.QueryClient, acc1, acc2, cw, denomA, denomB string) (
	*assetfttypes.QueryBalanceResponse, // acc1A
	*assetfttypes.QueryBalanceResponse, // acc1B
	*assetfttypes.QueryBalanceResponse, // acc2A
	*assetfttypes.QueryBalanceResponse, // acc2B
	*assetfttypes.QueryBalanceResponse, // cwA
	*assetfttypes.QueryBalanceResponse, // cwB
) {
	acc2ABalanceRes, _ := assetFTClint.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc2,
		Denom:   denomA,
	})
	acc1ABalanceRes, _ := assetFTClint.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1,
		Denom:   denomA,
	})
	cwABalanceRes, _ := assetFTClint.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: cw,
		Denom:   denomA,
	})
	acc2BBalanceRes, _ := assetFTClint.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc2,
		Denom:   denomB,
	})
	acc1BBalanceRes, _ := assetFTClint.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1,
		Denom:   denomB,
	})
	cwBBalanceRes, _ := assetFTClint.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: cw,
		Denom:   denomB,
	})

	return acc1ABalanceRes, acc1BBalanceRes, acc2ABalanceRes, acc2BBalanceRes, cwABalanceRes, cwBBalanceRes
}

func placeBuyOrder(
	ctx context.Context,
	chain integration.CoreumChain,
	acc sdk.AccAddress,
	id, denomA, denomB, price string,
	amount sdkmath.Int,
) (*sdk.TxResponse, error) {
	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          id,
		BaseDenom:   denomA,
		QuoteDenom:  denomB,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString(price)),
		Quantity:    amount,
		Side:        dextypes.SIDE_BUY,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}
	resp, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc),
		chain.TxFactoryAuto(),
		placeBuyOrderMsg,
	)
	return resp, err
}

func placeSellOrder(
	ctx context.Context,
	chain integration.CoreumChain,
	acc sdk.AccAddress, id, denomA, denomB, price string,
	amount sdkmath.Int,
) (*sdk.TxResponse, error) {
	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          id,
		BaseDenom:   denomA,
		QuoteDenom:  denomB,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString(price)),
		Quantity:    amount,
		Side:        dextypes.SIDE_SELL,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}
	resp, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	return resp, err
}

func sendFromAdmin(
	ctx context.Context,
	chain integration.CoreumChain,
	admin sdk.AccAddress,
	to string,
	denom string,
	amount sdkmath.Int,
) (*sdk.TxResponse, error) {
	sendMsg := &banktypes.MsgSend{
		FromAddress: admin.String(),
		ToAddress:   to,
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, amount)),
	}
	resp, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactoryAuto(),
		sendMsg,
	)
	return resp, err
}
