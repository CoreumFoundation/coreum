package ibc

import (
	_ "embed"
	"encoding/json"
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
)

//go:embed testdata/wasm/ibc-transfer/artifacts/ibc_transfer.wasm
var ibcTransferWASM []byte

type ibcTimeoutBlock struct {
	Revision uint64 `json:"revision"`
	Height   uint64 `json:"height"`
}

type ibcTimeout struct {
	Block ibcTimeoutBlock `json:"block"`
}

//nolint:tagliatelle // wasm requirements
type ibcTransferRequest struct {
	ChannelID string     `json:"channel_id"`
	ToAddress string     `json:"to_address"`
	Amount    sdk.Coin   `json:"amount"`
	Timeout   ibcTimeout `json:"timeout"`
}

type ibcTransferMethod string

const (
	transfer ibcTransferMethod = "transfer"
)

// TestIBCTransferFromSmartContract runs a contract deployment flow and tests that the contract is able to use IBCTransfer.
func TestIBCTransferFromSmartContract(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t, false)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	osmosisChain := chains.Osmosis

	osmosisToCoreumChannelID, err := osmosisChain.GetIBCChannelID(ctx, coreumChain.ChainSettings.ChainID)
	requireT.NoError(err)
	coreumToOsmosisChannelID, err := coreumChain.GetIBCChannelID(ctx, osmosisChain.ChainSettings.ChainID)
	requireT.NoError(err)

	coreumAdmin := coreumChain.GenAccount()
	osmosisRecipient := osmosisChain.GenAccount()

	requireT.NoError(coreumChain.FundAccountsWithOptions(ctx, coreumAdmin, integrationtests.BalancesOptions{
		Amount: sdk.NewInt(2000000),
	}))
	sendToOsmosisCoin := coreumChain.NewCoin(sdk.NewInt(1000))

	clientCtx := coreumChain.ClientContext
	txf := coreumChain.TxFactory().
		WithSimulateAndExecute(true)
	bankClient := banktypes.NewQueryClient(clientCtx)

	// deploy the contract and fund it
	initialPayload, err := json.Marshal(struct{}{})
	requireT.NoError(err)
	contractAddr, _, err := coreumChain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		txf,
		coreumAdmin,
		ibcTransferWASM,
		integrationtests.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayload,
			Amount:     sendToOsmosisCoin,
			Label:      "ibc_transfer",
		},
	)
	requireT.NoError(err)

	// get the contract balance and check total
	contractBalance, err := bankClient.Balance(ctx,
		&banktypes.QueryBalanceRequest{
			Address: contractAddr,
			Denom:   sendToOsmosisCoin.Denom,
		})
	requireT.NoError(err)
	requireT.Equal(sendToOsmosisCoin.Amount.String(), contractBalance.Balance.Amount.String())

	coreumChainHeight, err := coreumChain.QueryLatestConsensusHeight(
		ctx,
		ibctransfertypes.PortID,
		coreumToOsmosisChannelID,
	)
	requireT.NoError(err)

	transferPayload, err := json.Marshal(map[ibcTransferMethod]ibcTransferRequest{
		transfer: {
			ChannelID: coreumToOsmosisChannelID,
			ToAddress: osmosisChain.ConvertToBech32Address(osmosisRecipient),
			Amount:    sendToOsmosisCoin,
			Timeout: ibcTimeout{
				Block: ibcTimeoutBlock{
					Revision: coreumChainHeight.RevisionNumber,
					Height:   coreumChainHeight.RevisionHeight + 1000,
				},
			},
		},
	})
	requireT.NoError(err)

	_, err = coreumChain.Wasm.ExecuteWASMContract(ctx, txf, coreumAdmin, contractAddr, transferPayload, sdk.Coin{})
	requireT.NoError(err)

	contractBalance, err = bankClient.Balance(ctx,
		&banktypes.QueryBalanceRequest{
			Address: contractAddr,
			Denom:   sendToOsmosisCoin.Denom,
		})
	requireT.NoError(err)
	requireT.Equal(sdk.ZeroInt().String(), contractBalance.Balance.Amount.String())

	expectedOsmosisRecipientBalance := sdk.NewCoin(convertToIBCDenom(osmosisToCoreumChannelID, sendToOsmosisCoin.Denom), sendToOsmosisCoin.Amount)
	requireT.NoError(osmosisChain.AwaitForBalance(ctx, osmosisRecipient, expectedOsmosisRecipientBalance))
}
