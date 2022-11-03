package wasm

import (
	"context"
	_ "embed"
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	assettypes "github.com/CoreumFoundation/coreum/x/asset/types"
)

var (
	//go:embed testdata/fungible-token/artifacts/fungible_token.wasm
	fungibleTokenWASM []byte
)

type fungibleTokenInstantiatePayload struct{}

type fungibleTokenCreateRequest struct {
	Symbol    string `json:"symbol"`
	Amount    string `json:"amount"`
	Recipient string `json:"recipient"`
}

type fungibleTokenMethod string

const (
	create fungibleTokenMethod = "create"
)

// TestFungibleTokenWasmContract runs a contract deployment flow and tests that the contract is able to use Bank module
// to disperse the native coins.
func TestFungibleTokenWasmContract(ctx context.Context, t testing.T, chain testing.Chain) { //nolint:funlen // The test covers step-by step use case, no need split it
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
	initialPayload, err := json.Marshal(fungibleTokenInstantiatePayload{})
	requireT.NoError(err)
	contractAddr, _, err := DeployAndInstantiate(
		ctx,
		clientCtx,
		txf,
		fungibleTokenWASM,
		InstantiateConfig{
			accessType: wasmtypes.AccessTypeUnspecified,
			payload:    initialPayload,
			label:      "fungible_token",
		},
	)
	requireT.NoError(err)

	recipient := chain.GenAccount()

	symbol := "mytoken"
	denom := assettypes.BuildFungibleTokenDenom(symbol, sdk.MustAccAddressFromBech32(contractAddr))

	// create fungible token by smart contract
	createPayload, err := json.Marshal(map[fungibleTokenMethod]fungibleTokenCreateRequest{
		create: {
			Symbol:    symbol,
			Amount:    "5000",
			Recipient: recipient.String(),
		},
	})
	requireT.NoError(err)

	txf = txf.WithSimulateAndExecute(true)
	_, err = Execute(ctx, clientCtx, txf, contractAddr, createPayload, sdk.Coin{})
	requireT.NoError(err)

	// check balance of recipient
	recipientBalance, err := bankClient.Balance(ctx,
		&banktypes.QueryBalanceRequest{
			Address: recipient.String(),
			Denom:   denom,
		})
	requireT.NoError(err)
	requireT.NotNil(recipientBalance.Balance)
	requireT.Equal(sdk.NewInt64Coin(denom, 5000).String(), recipientBalance.Balance.String())
}
