package wasm

import (
	"context"
	_ "embed"
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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
	Subunit   string `json:"subunit"`
	Precision uint32 `json:"precision"`
	Amount    string `json:"amount"`
	Recipient string `json:"recipient"`
}

type fungibleTokenMethod string

const (
	issue fungibleTokenMethod = "issue"
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
	assetClient := assettypes.NewQueryClient(clientCtx)

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

	subunit := "mysatoshi"
	var precision uint32 = 8
	denom := assettypes.BuildFungibleTokenDenom(subunit, sdk.MustAccAddressFromBech32(contractAddr))
	initialAmount := sdk.NewInt(5000)
	symbol := "mytoken"

	// issue fungible token by smart contract
	createPayload, err := json.Marshal(map[fungibleTokenMethod]issueFungibleTokenRequest{
		issue: {
			Symbol:    symbol,
			Subunit:   subunit,
			Precision: precision,
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
	recipientBalance, err := bankClient.Balance(ctx,
		&banktypes.QueryBalanceRequest{
			Address: recipient.String(),
			Denom:   denom,
		})
	requireT.NoError(err)
	requireT.NotNil(recipientBalance.Balance)
	requireT.Equal(sdk.NewCoin(denom, initialAmount).String(), recipientBalance.Balance.String())

	ft, err := assetClient.FungibleToken(ctx, &assettypes.QueryFungibleTokenRequest{Denom: denom})
	requireT.NoError(err)
	requireT.EqualValues(assettypes.FungibleToken{
		Denom:     denom,
		Issuer:    contractAddr,
		Symbol:    symbol,
		SubUnit:   subunit,
		Precision: precision,
	}, ft.GetFungibleToken())
}
