//go:build integrationtests

package modules

import (
	"context"
	_ "embed"
	"encoding/json"
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/testutil/event"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

var (
	//go:embed testdata/wasm/bank-send/artifacts/bank_send.wasm
	bankSendWASM []byte
	//go:embed testdata/wasm/simple-state/artifacts/simple_state.wasm
	simpleStateWASM []byte
	//go:embed testdata/wasm/issue-fungible-token/artifacts/issue_fungible_token.wasm
	issueFungibleTokenWASM []byte
)

type bankWithdrawRequest struct {
	Amount    string `json:"amount"`
	Denom     string `json:"denom"`
	Recipient string `json:"recipient"`
}

type bankMethod string

const (
	withdraw bankMethod = "withdraw"
)

type simpleState struct {
	Count int `json:"count"`
}

type simpleStateMethod string

const (
	simpleGetCount  simpleStateMethod = "get_count"
	simpleIncrement simpleStateMethod = "increment"
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

// TestWASMBankSendContract runs a contract deployment flow and tests that the contract is able to use Bank module
// to disperse the native coins.
func TestWASMBankSendContract(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	admin := chain.GenAccount()
	nativeDenom := chain.NetworkConfig.Denom

	requireT := require.New(t)
	requireT.NoError(chain.Faucet.FundAccounts(ctx,
		integrationtests.NewFundedAccount(admin, chain.NewCoin(sdk.NewInt(5000000000))),
	))

	clientCtx := chain.ClientContext.WithFromAddress(admin)
	txf := chain.TxFactory().
		WithSimulateAndExecute(true)
	bankClient := banktypes.NewQueryClient(clientCtx)

	// deployWASMContract and init contract with the initial coins amount
	initialPayload, err := json.Marshal(struct{}{})
	requireT.NoError(err)
	contractAddr, _, err := deployAndInstantiateWASMContract(
		ctx,
		clientCtx,
		txf,
		bankSendWASM,
		instantiateConfig{
			accessType: wasmtypes.AccessTypeUnspecified,
			payload:    initialPayload,
			amount:     chain.NewCoin(sdk.NewInt(10000)),
			label:      "bank_send",
		},
	)
	requireT.NoError(err)

	// send additional coins to contract directly
	sdkContractAddress, err := sdk.AccAddressFromBech32(contractAddr)
	requireT.NoError(err)

	msg := &banktypes.MsgSend{
		FromAddress: admin.String(),
		ToAddress:   sdkContractAddress.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdk.NewInt(5000))),
	}

	_, err = tx.BroadcastTx(ctx, clientCtx, txf, msg)
	requireT.NoError(err)

	// get the contract balance and check total
	contractBalance, err := bankClient.Balance(ctx,
		&banktypes.QueryBalanceRequest{
			Address: contractAddr,
			Denom:   nativeDenom,
		})
	requireT.NoError(err)
	requireT.NotNil(contractBalance.Balance)
	requireT.Equal(sdk.NewInt64Coin(nativeDenom, 15000).String(), contractBalance.Balance.String())

	recipient := chain.GenAccount()
	// try to exceed the contract limit
	withdrawPayload, err := json.Marshal(map[bankMethod]bankWithdrawRequest{
		withdraw: {
			Amount:    "16000",
			Denom:     nativeDenom,
			Recipient: recipient.String(),
		},
	})
	requireT.NoError(err)

	// try to withdraw more than the admin has
	txf = txf.
		WithSimulateAndExecute(false).
		// the gas here is to try to execute the tx and don't fail on the gas estimation
		WithGas(uint64(chain.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas))
	_, err = executeWASMContract(ctx, clientCtx, txf, contractAddr, withdrawPayload, sdk.Coin{})
	requireT.True(cosmoserrors.ErrInsufficientFunds.Is(err))

	// send coin from the contract to test wallet
	withdrawPayload, err = json.Marshal(map[bankMethod]bankWithdrawRequest{
		withdraw: {
			Amount:    "5000",
			Denom:     nativeDenom,
			Recipient: recipient.String(),
		},
	})
	requireT.NoError(err)

	txf = txf.WithSimulateAndExecute(true)
	_, err = executeWASMContract(ctx, clientCtx, txf, contractAddr, withdrawPayload, sdk.Coin{})
	requireT.NoError(err)

	// check contract and wallet balances
	contractBalance, err = bankClient.Balance(ctx,
		&banktypes.QueryBalanceRequest{
			Address: contractAddr,
			Denom:   nativeDenom,
		})
	requireT.NoError(err)
	requireT.NotNil(contractBalance.Balance)
	requireT.Equal(sdk.NewInt64Coin(nativeDenom, 10000).String(), contractBalance.Balance.String())

	recipientBalance, err := bankClient.Balance(ctx,
		&banktypes.QueryBalanceRequest{
			Address: recipient.String(),
			Denom:   nativeDenom,
		})
	requireT.NoError(err)
	requireT.NotNil(recipientBalance.Balance)
	requireT.Equal(sdk.NewInt64Coin(nativeDenom, 5000).String(), recipientBalance.Balance.String())
}

// TestWASMGasBankSendAndBankSend checks that a message containing a deterministic and a
// non-deterministic transaction takes gas within appropriate limits.
func TestWASMGasBankSendAndBankSend(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	requireT := require.New(t)
	admin := chain.GenAccount()

	requireT.NoError(chain.Faucet.FundAccounts(ctx,
		integrationtests.NewFundedAccount(admin, chain.NewCoin(sdk.NewInt(5000000000))),
	))

	// deployWASMContract and init contract with the initial coins amount
	initialPayload, err := json.Marshal(struct{}{})
	requireT.NoError(err)

	clientCtx := chain.ClientContext.WithFromAddress(admin)
	txf := chain.TxFactory().
		WithSimulateAndExecute(true)

	contractAddr, _, err := deployAndInstantiateWASMContract(
		ctx,
		clientCtx,
		txf,
		bankSendWASM,
		instantiateConfig{
			accessType: wasmtypes.AccessTypeUnspecified,
			payload:    initialPayload,
			amount:     chain.NewCoin(sdk.NewInt(10000)),
			label:      "bank_send",
		},
	)
	requireT.NoError(err)

	// Send tokens
	receiver := chain.GenAccount()
	withdrawPayload, err := json.Marshal(map[bankMethod]bankWithdrawRequest{
		withdraw: {
			Amount:    "5000",
			Denom:     chain.NetworkConfig.Denom,
			Recipient: receiver.String(),
		},
	})
	requireT.NoError(err)

	wasmBankSend := &wasmtypes.MsgExecuteContract{
		Sender:   admin.String(),
		Contract: contractAddr,
		Msg:      wasmtypes.RawContractMessage(withdrawPayload),
		Funds:    sdk.Coins{},
	}

	bankSend := &banktypes.MsgSend{
		FromAddress: admin.String(),
		ToAddress:   receiver.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(chain.NetworkConfig.Denom, sdk.NewInt(1000))),
	}

	minGasExpected := chain.GasLimitByMsgs(&banktypes.MsgSend{}, &banktypes.MsgSend{})
	maxGasExpected := minGasExpected * 10

	clientCtx = chain.ChainContext.ClientContext.WithFromAddress(admin)
	txf = chain.ChainContext.TxFactory().WithGas(maxGasExpected)
	result, err := tx.BroadcastTx(ctx, clientCtx, txf, wasmBankSend, bankSend)
	require.NoError(t, err)

	require.NoError(t, err)
	assert.Greater(t, uint64(result.GasUsed), minGasExpected)
	assert.Less(t, uint64(result.GasUsed), maxGasExpected)
}

// TestWASMPinningAndUnpinningSmartContractUsingGovernance deploys simple smart contract, verifies that it works properly and then tests that
// pinning and unpinning through proposals works correctly. We also verify that pinned smart contract consumes less gas.
func TestWASMPinningAndUnpinningSmartContractUsingGovernance(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	admin := chain.GenAccount()
	proposer := chain.GenAccount()

	requireT := require.New(t)

	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	requireT.NoError(err)
	proposerBalance.Amount = proposerBalance.Amount.MulRaw(2)

	requireT.NoError(chain.Faucet.FundAccounts(ctx,
		integrationtests.NewFundedAccount(admin, chain.NewCoin(sdk.NewInt(5000000000))),
		integrationtests.NewFundedAccount(proposer, proposerBalance),
	))

	// instantiateWASMContract the contract and set the initial counter state.
	initialPayload, err := json.Marshal(simpleState{
		Count: 1337,
	})
	requireT.NoError(err)

	clientCtx := chain.ClientContext.WithFromAddress(admin)
	txf := chain.TxFactory().
		WithSimulateAndExecute(true)

	contractAddr, codeID, err := deployAndInstantiateWASMContract(
		ctx,
		clientCtx,
		txf,
		simpleStateWASM,
		instantiateConfig{
			accessType: wasmtypes.AccessTypeUnspecified,
			payload:    initialPayload,
			label:      "simple_state",
		},
	)
	requireT.NoError(err)

	// get the current counter state
	getCountPayload, err := methodToEmptyBodyPayload(simpleGetCount)
	requireT.NoError(err)
	queryOut, err := queryWASMContract(ctx, clientCtx, contractAddr, getCountPayload)
	requireT.NoError(err)
	var response simpleState
	err = json.Unmarshal(queryOut, &response)
	requireT.NoError(err)
	requireT.Equal(1337, response.Count)

	// execute contract to increment the count
	gasUsedBeforePinning := incrementAndVerify(ctx, clientCtx, txf, contractAddr, requireT, 1338)

	// verify that smart contract is not pinned
	requireT.False(isWASMContractPinned(ctx, clientCtx, codeID))

	// pin smart contract
	proposalMsg, err := chain.Governance.NewMsgSubmitProposal(ctx, proposer, &wasmtypes.PinCodesProposal{
		Title:       "Pin smart contract",
		Description: "Testing smart contract pinning",
		CodeIDs:     []uint64{codeID},
	})
	requireT.NoError(err)
	proposalID, err := chain.Governance.Propose(ctx, proposalMsg)
	requireT.NoError(err)

	proposal, err := chain.Governance.GetProposal(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypes.StatusVotingPeriod, proposal.Status)

	err = chain.Governance.VoteAll(ctx, govtypes.OptionYes, proposal.ProposalId)
	requireT.NoError(err)

	// Wait for proposal result.
	finalStatus, err := chain.Governance.WaitForVotingToFinalize(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypes.StatusPassed, finalStatus)

	requireT.True(isWASMContractPinned(ctx, clientCtx, codeID))

	gasUsedAfterPinning := incrementAndVerify(ctx, clientCtx, txf, contractAddr, requireT, 1339)

	// unpin smart contract
	proposalMsg, err = chain.Governance.NewMsgSubmitProposal(ctx, proposer, &wasmtypes.UnpinCodesProposal{
		Title:       "Unpin smart contract",
		Description: "Testing smart contract unpinning",
		CodeIDs:     []uint64{codeID},
	})
	requireT.NoError(err)
	proposalID, err = chain.Governance.Propose(ctx, proposalMsg)
	requireT.NoError(err)

	proposal, err = chain.Governance.GetProposal(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypes.StatusVotingPeriod, proposal.Status)

	err = chain.Governance.VoteAll(ctx, govtypes.OptionYes, proposal.ProposalId)
	requireT.NoError(err)
	finalStatus, err = chain.Governance.WaitForVotingToFinalize(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypes.StatusPassed, finalStatus)

	requireT.False(isWASMContractPinned(ctx, clientCtx, codeID))

	gasUsedAfterUnpinning := incrementAndVerify(ctx, clientCtx, txf, contractAddr, requireT, 1340)

	logger.Get(ctx).Info("Gas saved on pinned contract",
		zap.Int64("gasBeforePinning", gasUsedBeforePinning),
		zap.Int64("gasAfterPinning", gasUsedAfterPinning))

	assertT := assert.New(t)
	assertT.Less(gasUsedAfterPinning, gasUsedBeforePinning)
	assertT.Greater(gasUsedAfterUnpinning, gasUsedAfterPinning)
}

// TestWASMIssueFungibleTokenInContract verifies that smart contract is able to issue fungible token
func TestWASMIssueFungibleTokenInContract(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	admin := chain.GenAccount()

	requireT := require.New(t)
	requireT.NoError(chain.Faucet.FundAccounts(ctx,
		integrationtests.NewFundedAccount(admin, chain.NewCoin(sdk.NewInt(5000000000))),
	))

	clientCtx := chain.ClientContext.WithFromAddress(admin)
	txf := chain.TxFactory().
		WithSimulateAndExecute(true)
	bankClient := banktypes.NewQueryClient(clientCtx)
	ftClient := assetfttypes.NewQueryClient(clientCtx)

	// deployWASMContract and init contract with the initial coins amount
	initialPayload, err := json.Marshal(struct{}{})
	requireT.NoError(err)
	contractAddr, _, err := deployAndInstantiateWASMContract(
		ctx,
		clientCtx,
		txf,
		issueFungibleTokenWASM,
		instantiateConfig{
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
	gasUsed, err := executeWASMContract(ctx, clientCtx, txf, contractAddr, createPayload, sdk.Coin{})
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
		Denom:              denom1,
		Issuer:             contractAddr,
		Symbol:             symbol + "1",
		Subunit:            subunit1,
		Precision:          precision,
		BurnRate:           sdk.NewDec(0),
		SendCommissionRate: sdk.NewDec(0),
	}, ft.GetToken())

	ft, err = ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{Denom: denom2})
	requireT.NoError(err)
	requireT.EqualValues(assetfttypes.FT{
		Denom:              denom2,
		Issuer:             contractAddr,
		Symbol:             symbol + "2",
		Subunit:            subunit2,
		Precision:          precision,
		BurnRate:           sdk.NewDec(0),
		SendCommissionRate: sdk.NewDec(0),
	}, ft.GetToken())

	// check the counter
	getCountPayload, err := json.Marshal(map[fungibleTokenMethod]struct{}{
		ftGetCount: {},
	})
	requireT.NoError(err)
	queryOut, err := queryWASMContract(ctx, clientCtx, contractAddr, getCountPayload)
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
	queryOut, err = queryWASMContract(ctx, clientCtx, contractAddr, getInfoPayload)
	requireT.NoError(err)

	infoResponse := struct {
		Issuer string `json:"issuer"`
	}{}
	requireT.NoError(json.Unmarshal(queryOut, &infoResponse))
	assertT.Equal(contractAddr, infoResponse.Issuer)
}

func methodToEmptyBodyPayload(methodName simpleStateMethod) (json.RawMessage, error) {
	return json.Marshal(map[simpleStateMethod]struct{}{
		methodName: {},
	})
}

func incrementAndVerify(
	ctx context.Context,
	clientCtx tx.ClientContext,
	txf tx.Factory,
	contractAddr string,
	requireT *require.Assertions,
	expectedValue int,
) int64 {
	// execute contract to increment the count
	incrementPayload, err := methodToEmptyBodyPayload(simpleIncrement)
	requireT.NoError(err)
	gasUsed, err := executeWASMContract(ctx, clientCtx, txf, contractAddr, incrementPayload, sdk.Coin{})
	requireT.NoError(err)

	// check the update count
	getCountPayload, err := methodToEmptyBodyPayload(simpleGetCount)
	requireT.NoError(err)
	queryOut, err := queryWASMContract(ctx, clientCtx, contractAddr, getCountPayload)
	requireT.NoError(err)

	var response simpleState
	err = json.Unmarshal(queryOut, &response)
	requireT.NoError(err)
	requireT.Equal(expectedValue, response.Count)

	return gasUsed
}

// --------------------------- Client ---------------------------

var gasMultiplier = 1.5

// instantiateConfig contains params specific to contract instantiation.
type instantiateConfig struct {
	accessType wasmtypes.AccessType
	payload    json.RawMessage
	amount     sdk.Coin
	label      string
	CodeID     uint64
}

// deployAndInstantiateWASMContract deploys, instantiateWASMContract the wasm contract and returns its address.
func deployAndInstantiateWASMContract(ctx context.Context, clientCtx tx.ClientContext, txf tx.Factory, wasmData []byte, initConfig instantiateConfig) (string, uint64, error) {
	codeID, err := deployWASMContract(ctx, clientCtx, txf, wasmData)
	if err != nil {
		return "", 0, err
	}

	initConfig.CodeID = codeID
	contractAddr, err := instantiateWASMContract(ctx, clientCtx, txf, initConfig)
	if err != nil {
		return "", 0, err
	}

	return contractAddr, codeID, nil
}

// executeWASMContract executes the wasm contract with the payload and optionally funding amount.
func executeWASMContract(ctx context.Context, clientCtx tx.ClientContext, txf tx.Factory, contractAddr string, payload json.RawMessage, fundAmt sdk.Coin) (int64, error) {
	funds := sdk.NewCoins()
	if !fundAmt.Amount.IsNil() {
		funds = funds.Add(fundAmt)
	}

	msg := &wasmtypes.MsgExecuteContract{
		Sender:   clientCtx.FromAddress().String(),
		Contract: contractAddr,
		Msg:      wasmtypes.RawContractMessage(payload),
		Funds:    funds,
	}

	txf = txf.
		WithGasAdjustment(gasMultiplier)

	res, err := tx.BroadcastTx(ctx, clientCtx, txf, msg)
	if err != nil {
		return 0, err
	}
	return res.GasUsed, nil
}

// queryWASMContract queries the contract with the requested payload.
func queryWASMContract(ctx context.Context, clientCtx tx.ClientContext, contractAddr string, payload json.RawMessage) (json.RawMessage, error) {
	query := &wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddr,
		QueryData: wasmtypes.RawContractMessage(payload),
	}

	wasmClient := wasmtypes.NewQueryClient(clientCtx)
	resp, err := wasmClient.SmartContractState(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "WASMQueryClient returned an error after smart contract state queryWASMContract")
	}

	return json.RawMessage(resp.Data), nil
}

// isWASMContractPinned returns true if smart contract is pinned
func isWASMContractPinned(ctx context.Context, clientCtx tx.ClientContext, codeID uint64) (bool, error) {
	wasmClient := wasmtypes.NewQueryClient(clientCtx)
	resp, err := wasmClient.PinnedCodes(ctx, &wasmtypes.QueryPinnedCodesRequest{})
	if err != nil {
		return false, errors.Wrap(err, "WASMQueryClient returned an error after querying pinned contracts")
	}
	for _, c := range resp.CodeIDs {
		if c == codeID {
			return true, nil
		}
	}
	return false, nil
}

// deploys the wasm contract and returns its codeID.
func deployWASMContract(ctx context.Context, clientCtx tx.ClientContext, txf tx.Factory, wasmData []byte) (uint64, error) {
	msgStoreCode := &wasmtypes.MsgStoreCode{
		Sender:       clientCtx.FromAddress().String(),
		WASMByteCode: wasmData,
	}

	txf = txf.
		WithGasAdjustment(gasMultiplier)

	res, err := tx.BroadcastTx(ctx, clientCtx, txf, msgStoreCode)
	if err != nil {
		return 0, err
	}

	codeID, err := event.FindUint64EventAttribute(res.Events, wasmtypes.EventTypeStoreCode, wasmtypes.AttributeKeyCodeID)
	if err != nil {
		return 0, err
	}

	return codeID, nil
}

// instantiates the contract and returns the contract address.
func instantiateWASMContract(ctx context.Context, clientCtx tx.ClientContext, txf tx.Factory, req instantiateConfig) (string, error) {
	funds := sdk.NewCoins()
	if amount := req.amount; !amount.Amount.IsNil() {
		funds = funds.Add(amount)
	}
	msg := &wasmtypes.MsgInstantiateContract{
		Sender: clientCtx.FromAddress().String(),
		CodeID: req.CodeID,
		Label:  req.label,
		Msg:    wasmtypes.RawContractMessage(req.payload),
		Funds:  funds,
	}

	txf = txf.
		WithGasAdjustment(gasMultiplier)

	res, err := tx.BroadcastTx(ctx, clientCtx, txf, msg)
	if err != nil {
		return "", err
	}

	contractAddr, err := event.FindStringEventAttribute(res.Events, wasmtypes.EventTypeInstantiate, wasmtypes.AttributeKeyContractAddr)
	if err != nil {
		return "", err
	}

	return contractAddr, nil
}
