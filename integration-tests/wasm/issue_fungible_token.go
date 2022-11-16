package wasm

import (
	"context"
	_ "embed"
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	assettypes "github.com/CoreumFoundation/coreum/x/asset/types"
)

var (
	//go:embed testdata/issue-fungible-token/artifacts/issue_fungible_token.wasm
	issueFungibleTokenWASM []byte
)

type issueFungibleTokenRequest struct {
	Symbol    string `json:"symbol"`
	Amount    string `json:"amount"`
	Recipient string `json:"recipient"`
}

type fungibleTokenMethod string

const (
	ftIssue    fungibleTokenMethod = "issue"
	ftGetCount fungibleTokenMethod = "get_count"
	ftGetInfo  fungibleTokenMethod = "get_info"
)

// TestIssueFungibleTokenInWASMContract verifies that smart contract is able to issue fungible token
func TestIssueFungibleTokenInWASMContract(ctx context.Context, t testing.T, chain testing.Chain) {
	admin := chain.GenAccount()

	requireT := require.New(t)
	requireT.NoError(chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(admin, chain.NewCoin(sdk.NewInt(5000000000))),
	))

	clientCtx := chain.ClientContext.WithFromAddress(admin)
	txf := chain.TxFactory().
		WithSimulateAndExecute(true)
	bankClient := banktypes.NewQueryClient(clientCtx)

	// deploy and init contract with the initial coins amount
	initialPayload, err := json.Marshal(struct{}{})
	requireT.NoError(err)
	contractAddr, _, err := DeployAndInstantiate(
		ctx,
		clientCtx,
		txf,
		issueFungibleTokenWASM,
		InstantiateConfig{
			accessType: wasmtypes.AccessTypeUnspecified,
			payload:    initialPayload,
			label:      "fungible_token",
		},
	)
	requireT.NoError(err)

	recipient := chain.GenAccount()

	symbol := "mytoken"
	denom1 := assettypes.BuildFungibleTokenDenom(symbol+"1", sdk.MustAccAddressFromBech32(contractAddr))
	denom2 := assettypes.BuildFungibleTokenDenom(symbol+"2", sdk.MustAccAddressFromBech32(contractAddr))
	initialAmount := sdk.NewInt(5000)

	// issue fungible token by smart contract
	createPayload, err := json.Marshal(map[fungibleTokenMethod]issueFungibleTokenRequest{
		ftIssue: {
			Symbol:    symbol,
			Amount:    initialAmount.String(),
			Recipient: recipient.String(),
		},
	})
	requireT.NoError(err)

	txf = txf.WithSimulateAndExecute(true)
	gasUsed, err := Execute(ctx, clientCtx, txf, contractAddr, createPayload, sdk.Coin{})
	requireT.NoError(err)

	logger.Get(ctx).Info("Fungible token issued by smart contract", zap.Int64("gasUsed", gasUsed))

	// check balance of recipient
	recipientBalance, err := bankClient.AllBalances(ctx,
		&banktypes.QueryAllBalancesRequest{
			Address: recipient.String(),
		})
	requireT.NoError(err)

	assertT := assert.New(t)
	assertT.Equal(initialAmount.String(), recipientBalance.Balances.AmountOf(denom1).String())
	assertT.Equal(initialAmount.String(), recipientBalance.Balances.AmountOf(denom2).String())

	// check the counter
	getCountPayload, err := json.Marshal(map[fungibleTokenMethod]struct{}{
		ftGetCount: {},
	})
	requireT.NoError(err)
	queryOut, err := Query(ctx, clientCtx, contractAddr, getCountPayload)
	requireT.NoError(err)

	counterResponse := struct {
		Count int `json:"count"`
	}{}
	err = json.Unmarshal(queryOut, &counterResponse)
	requireT.NoError(err)
	assertT.Equal(2, counterResponse.Count)

	// query from smart contract
	getInfoPayload, err := json.Marshal(map[fungibleTokenMethod]interface{}{
		ftGetInfo: struct {
			Denom string `json:"denom"`
		}{
			Denom: denom1,
		},
	})
	requireT.NoError(err)
	queryOut, err = Query(ctx, clientCtx, contractAddr, getInfoPayload)
	requireT.NoError(err)

	infoResponse := struct {
		Issuer string `json:"issuer"`
	}{}
	requireT.NoError(json.Unmarshal(queryOut, &infoResponse))
	assertT.Equal(contractAddr, infoResponse.Issuer)
}
