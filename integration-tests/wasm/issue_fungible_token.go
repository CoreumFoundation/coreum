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
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

var (
	//go:embed testdata/issue-fungible-token/artifacts/issue_fungible_token.wasm
	issueFungibleTokenWASM []byte
)

type issueFungibleTokenRequest struct {
	Symbol    string `json:"symbol"`
	Subunit   string `json:"subunit"`
	Precision uint32 `json:"precision"`
	Amount    string `json:"amount"`
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
	ftClient := assetfttypes.NewQueryClient(clientCtx)

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

	symbol := "mytoken"
	subunit := "mysatoshi"
	subunit1 := subunit + "1"
	subunit2 := subunit + "2"
	precision := uint32(8)
	denom1 := assetfttypes.BuildDenom(subunit1, sdk.MustAccAddressFromBech32(contractAddr))
	denom2 := assetfttypes.BuildDenom(subunit2, sdk.MustAccAddressFromBech32(contractAddr))
	initialAmount := sdk.NewInt(5000)

	// issue fungible token by smart contract
	createPayload, err := json.Marshal(map[fungibleTokenMethod]issueFungibleTokenRequest{
		ftIssue: {
			Symbol:    symbol,
			Subunit:   subunit,
			Precision: precision,
			Amount:    initialAmount.String(),
		},
	})
	requireT.NoError(err)

	txf = txf.WithSimulateAndExecute(true)
	gasUsed, err := Execute(ctx, clientCtx, txf, contractAddr, createPayload, sdk.Coin{})
	requireT.NoError(err)

	logger.Get(ctx).Info("Fungible token issued by smart contract", zap.Int64("gasUsed", gasUsed))

	// check balance of recipient
	balance, err := bankClient.AllBalances(ctx,
		&banktypes.QueryAllBalancesRequest{
			Address: contractAddr,
		})
	requireT.NoError(err)

	assertT := assert.New(t)
	assertT.Equal(initialAmount.String(), balance.Balances.AmountOf(denom1).String())
	assertT.Equal(initialAmount.String(), balance.Balances.AmountOf(denom2).String())

	ft, err := ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{Denom: denom1})
	requireT.NoError(err)
	requireT.EqualValues(assetfttypes.FT{
		Denom:     denom1,
		Issuer:    contractAddr,
		Symbol:    symbol + "1",
		Subunit:   subunit1,
		Precision: precision,
		BurnRate:  sdk.NewDec(0),
	}, ft.GetToken())

	ft, err = ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{Denom: denom2})
	requireT.NoError(err)
	requireT.EqualValues(assetfttypes.FT{
		Denom:     denom2,
		Issuer:    contractAddr,
		Symbol:    symbol + "2",
		Subunit:   subunit2,
		Precision: precision,
		BurnRate:  sdk.NewDec(0),
	}, ft.GetToken())

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
