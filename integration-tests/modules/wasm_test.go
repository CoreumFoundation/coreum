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
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/testutil/event"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

var (
	//go:embed testdata/wasm/bank-send/artifacts/bank_send.wasm
	bankSendWASM []byte
	//go:embed testdata/wasm/simple-state/artifacts/simple_state.wasm
	simpleStateWASM []byte
	//go:embed testdata/wasm/ft/artifacts/fungible_token.wasm
	fungibleTokenWASM []byte
)

// bank wasm models

type bankWithdrawRequest struct {
	Amount    string `json:"amount"`
	Denom     string `json:"denom"`
	Recipient string `json:"recipient"`
}

type bankMethod string

const (
	withdraw bankMethod = "withdraw"
)

// simple state models

type simpleState struct {
	Count int `json:"count"`
}

type simpleStateMethod string

const (
	simpleGetCount  simpleStateMethod = "get_count"
	simpleIncrement simpleStateMethod = "increment"
)

// fungible token wasm models
//
//nolint:tagliatelle
type issueFungibleTokenRequest struct {
	Symbol             string                 `json:"symbol"`
	Subunit            string                 `json:"subunit"`
	Precision          uint32                 `json:"precision"`
	InitialAmount      string                 `json:"initial_amount"`
	Description        string                 `json:"description"`
	Features           []assetfttypes.Feature `json:"features"`
	BurnRate           string                 `json:"burn_rate"`
	SendCommissionRate string                 `json:"send_commission_rate"`
}

type amountBodyFungibleTokenRequest struct {
	Amount string `json:"amount"`
}

type accountAmountBodyFungibleTokenRequest struct {
	Account string `json:"account"`
	Amount  string `json:"amount"`
}

type accountBodyFungibleTokenRequest struct {
	Account string `json:"account"`
}

type fungibleTokenMethod string

const (
	// tx.
	fungibleTokenMethodMint                fungibleTokenMethod = "mint"
	fungibleTokenMethodBurn                fungibleTokenMethod = "burn"
	fungibleTokenMethodFreeze              fungibleTokenMethod = "freeze"
	fungibleTokenMethodUnfreeze            fungibleTokenMethod = "unfreeze"
	fungibleTokenMethodGloballyFreeze      fungibleTokenMethod = "globally_freeze"
	fungibleTokenMethodGloballyUnfreeze    fungibleTokenMethod = "globally_unfreeze"
	fungibleTokenMethodSetWhitelistedLimit fungibleTokenMethod = "set_whitelisted_limit"
	// query.
	fungibleTokenMethodToken              fungibleTokenMethod = "token"
	fungibleTokenMethodFrozenBalance      fungibleTokenMethod = "frozen_balance"
	fungibleTokenMethodWhitelistedBalance fungibleTokenMethod = "whitelisted_balance"
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

	_, err = client.BroadcastTx(ctx, clientCtx, txf, msg)
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
	recipient := chain.GenAccount()
	withdrawPayload, err := json.Marshal(map[bankMethod]bankWithdrawRequest{
		withdraw: {
			Amount:    "5000",
			Denom:     chain.NetworkConfig.Denom,
			Recipient: recipient.String(),
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
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(chain.NetworkConfig.Denom, sdk.NewInt(1000))),
	}

	minGasExpected := chain.GasLimitByMsgs(&banktypes.MsgSend{}, &banktypes.MsgSend{})
	maxGasExpected := minGasExpected * 10

	clientCtx = chain.ChainContext.ClientContext.WithFromAddress(admin)
	txf = chain.ChainContext.TxFactory().WithGas(maxGasExpected)
	result, err := client.BroadcastTx(ctx, clientCtx, txf, wasmBankSend, bankSend)
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

// TestUpdateAndClearAdminOfContract runs MsgUpdateAdmin and MsgClearAdmin tx types.
func TestUpdateAndClearAdminOfContract(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	admin := chain.GenAccount()
	newAdmin := chain.GenAccount()

	requireT := require.New(t)
	requireT.NoError(chain.Faucet.FundAccounts(ctx,
		integrationtests.NewFundedAccount(admin, chain.NewCoin(sdk.NewInt(5000000000))),
	))
	requireT.NoError(chain.Faucet.FundAccountsWithOptions(ctx, newAdmin, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&wasmtypes.MsgClearAdmin{},
		},
	}))

	wasmClient := wasmtypes.NewQueryClient(chain.ClientContext)

	// deployWASMContract and init contract with the initial coins amount
	initialPayload, err := json.Marshal(struct{}{})
	requireT.NoError(err)
	contractAddr, _, err := deployAndInstantiateWASMContract(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactory().WithSimulateAndExecute(true),
		bankSendWASM,
		instantiateConfig{
			accessType: wasmtypes.AccessTypeUnspecified,
			admin:      admin,
			payload:    initialPayload,
			amount:     chain.NewCoin(sdk.NewInt(10000)),
			label:      "bank_send",
		},
	)
	requireT.NoError(err)

	// query contract info
	contractInfo, err := wasmClient.ContractInfo(ctx, &wasmtypes.QueryContractInfoRequest{
		Address: contractAddr,
	})
	requireT.NoError(err)
	requireT.EqualValues(admin.String(), contractInfo.Admin)

	// update admin
	msgUpdateAdmin := &wasmtypes.MsgUpdateAdmin{
		Sender:   admin.String(),
		NewAdmin: newAdmin.String(),
		Contract: contractAddr,
	}

	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactory().WithSimulateAndExecute(true).WithGas(chain.GasLimitByMsgs(msgUpdateAdmin)),
		msgUpdateAdmin,
	)

	requireT.NoError(err)
	requireT.NotNil(res)
	contractInfo, err = wasmClient.ContractInfo(ctx, &wasmtypes.QueryContractInfoRequest{
		Address: contractAddr,
	})
	requireT.NoError(err)
	requireT.EqualValues(newAdmin.String(), contractInfo.Admin)
	requireT.EqualValues(chain.GasLimitByMsgs(msgUpdateAdmin), res.GasUsed)

	// clear admin
	msgClearAdmin := &wasmtypes.MsgClearAdmin{
		Sender:   newAdmin.String(),
		Contract: contractAddr,
	}

	res, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(newAdmin),
		chain.TxFactory().WithSimulateAndExecute(true),
		msgClearAdmin,
	)

	requireT.NoError(err)
	requireT.NotNil(res)
	contractInfo, err = wasmClient.ContractInfo(ctx, &wasmtypes.QueryContractInfoRequest{
		Address: contractAddr,
	})
	requireT.NoError(err)
	requireT.EqualValues("", contractInfo.Admin)
	requireT.EqualValues(chain.GasLimitByMsgs(msgClearAdmin), res.GasUsed)
}

// TestWASMFungibleTokenInContract verifies that smart contract is able to execute all fungible token message and core queries.
//
//nolint:nosnakecase
func TestWASMFungibleTokenInContract(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	admin := chain.GenAccount()
	recipient := chain.GenAccount()

	requireT := require.New(t)
	requireT.NoError(chain.Faucet.FundAccounts(ctx,
		integrationtests.NewFundedAccount(admin, chain.NewCoin(sdk.NewInt(5000000000))),
	))

	clientCtx := chain.ClientContext.WithFromAddress(admin)
	txf := chain.TxFactory().
		WithSimulateAndExecute(true)
	bankClient := banktypes.NewQueryClient(clientCtx)
	ftClient := assetfttypes.NewQueryClient(clientCtx)

	// ********** Issuance **********

	burnRate := sdk.MustNewDecFromStr("0.1")
	sendCommissionRate := sdk.MustNewDecFromStr("0.2")

	issuanceAmount := sdk.NewInt(10_000)
	issuanceReq := issueFungibleTokenRequest{
		Symbol:        "symbol",
		Subunit:       "subunit",
		Precision:     6,
		InitialAmount: issuanceAmount.String(),
		Description:   "my wasm fungible token",
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_minting,
			assetfttypes.Feature_burning,
			assetfttypes.Feature_freezing,
			assetfttypes.Feature_whitelisting,
		},
		BurnRate:           burnRate.String(),
		SendCommissionRate: sendCommissionRate.String(),
	}
	issuerFTInstantiatePayload, err := json.Marshal(issuanceReq)
	requireT.NoError(err)

	// instantiate new contract
	contractAddr, _, err := deployAndInstantiateWASMContract(
		ctx,
		clientCtx,
		txf,
		fungibleTokenWASM,
		instantiateConfig{
			// we add the initial amount to let the contract issue the token on behalf of it
			amount:     chain.NewCoin(chain.NetworkConfig.AssetFTConfig.IssueFee),
			accessType: wasmtypes.AccessTypeUnspecified,
			payload:    issuerFTInstantiatePayload,
			label:      "fungible_token",
		},
	)
	requireT.NoError(err)

	denom := assetfttypes.BuildDenom(issuanceReq.Subunit, sdk.MustAccAddressFromBech32(contractAddr))

	tokenRes, err := ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{Denom: denom})
	requireT.NoError(err)

	expectedToken := assetfttypes.Token{
		Denom:          denom,
		Issuer:         contractAddr,
		Symbol:         issuanceReq.Symbol,
		Subunit:        issuanceReq.Subunit,
		Precision:      issuanceReq.Precision,
		Description:    issuanceReq.Description,
		GloballyFrozen: false,
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_minting,
			assetfttypes.Feature_burning,
			assetfttypes.Feature_freezing,
			assetfttypes.Feature_whitelisting,
		},
		BurnRate:           burnRate,
		SendCommissionRate: sendCommissionRate,
	}
	requireT.Equal(
		expectedToken, tokenRes.Token,
	)

	balanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: contractAddr,
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(issuanceReq.InitialAmount, balanceRes.Balance.Amount.String())

	// ********** Transactions **********

	// ********** Mint **********

	amountToMint := sdk.NewInt(500)
	mintPayload, err := json.Marshal(map[fungibleTokenMethod]amountBodyFungibleTokenRequest{
		fungibleTokenMethodMint: {
			Amount: amountToMint.String(),
		},
	})
	requireT.NoError(err)

	txf = txf.WithSimulateAndExecute(true)
	_, err = executeWASMContract(ctx, clientCtx, txf, contractAddr, mintPayload, sdk.Coin{})
	requireT.NoError(err)

	balanceRes, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: contractAddr,
		Denom:   denom,
	})
	requireT.NoError(err)
	newAmount := issuanceAmount.Add(amountToMint)
	requireT.Equal(newAmount.String(), balanceRes.Balance.Amount.String())

	// ********** Burn **********

	amountToBurn := sdk.NewInt(100)
	burnPayload, err := json.Marshal(map[fungibleTokenMethod]amountBodyFungibleTokenRequest{
		fungibleTokenMethodBurn: {
			Amount: amountToBurn.String(),
		},
	})
	requireT.NoError(err)

	txf = txf.WithSimulateAndExecute(true)
	_, err = executeWASMContract(ctx, clientCtx, txf, contractAddr, burnPayload, sdk.Coin{})
	requireT.NoError(err)

	balanceRes, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: contractAddr,
		Denom:   denom,
	})
	requireT.NoError(err)
	newAmount = newAmount.Sub(amountToBurn)
	requireT.Equal(newAmount.String(), balanceRes.Balance.Amount.String())

	// ********** Freeze **********

	amountToFreeze := sdk.NewInt(100)
	freezePayload, err := json.Marshal(map[fungibleTokenMethod]accountAmountBodyFungibleTokenRequest{
		fungibleTokenMethodFreeze: {
			Account: recipient.String(),
			Amount:  amountToFreeze.String(),
		},
	})
	requireT.NoError(err)

	txf = txf.WithSimulateAndExecute(true)
	_, err = executeWASMContract(ctx, clientCtx, txf, contractAddr, freezePayload, sdk.Coin{})
	requireT.NoError(err)

	frozenRes, err := ftClient.FrozenBalance(ctx, &assetfttypes.QueryFrozenBalanceRequest{
		Account: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(amountToFreeze.String(), frozenRes.Balance.Amount.String())

	// ********** Unfreeze **********

	amountToUnfreeze := sdk.NewInt(40)
	unfreezePayload, err := json.Marshal(map[fungibleTokenMethod]accountAmountBodyFungibleTokenRequest{
		fungibleTokenMethodUnfreeze: {
			Account: recipient.String(),
			Amount:  amountToUnfreeze.String(),
		},
	})
	requireT.NoError(err)

	txf = txf.WithSimulateAndExecute(true)
	_, err = executeWASMContract(ctx, clientCtx, txf, contractAddr, unfreezePayload, sdk.Coin{})
	requireT.NoError(err)

	frozenRes, err = ftClient.FrozenBalance(ctx, &assetfttypes.QueryFrozenBalanceRequest{
		Account: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(amountToFreeze.Sub(amountToUnfreeze).String(), frozenRes.Balance.Amount.String())

	// ********** GloballyFreeze **********

	globallyFreezePayload, err := json.Marshal(map[fungibleTokenMethod]struct{}{
		fungibleTokenMethodGloballyFreeze: {},
	})
	requireT.NoError(err)

	txf = txf.WithSimulateAndExecute(true)
	_, err = executeWASMContract(ctx, clientCtx, txf, contractAddr, globallyFreezePayload, sdk.Coin{})
	requireT.NoError(err)

	tokenRes, err = ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: denom,
	})
	requireT.NoError(err)
	requireT.True(tokenRes.Token.GloballyFrozen)

	// ********** GloballyUnfreeze **********

	globallyUnfreezePayload, err := json.Marshal(map[fungibleTokenMethod]struct{}{
		fungibleTokenMethodGloballyUnfreeze: {},
	})
	requireT.NoError(err)

	txf = txf.WithSimulateAndExecute(true)
	_, err = executeWASMContract(ctx, clientCtx, txf, contractAddr, globallyUnfreezePayload, sdk.Coin{})
	requireT.NoError(err)

	tokenRes, err = ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: denom,
	})
	requireT.NoError(err)
	requireT.False(tokenRes.Token.GloballyFrozen)

	// ********** Whitelisting **********

	amountToWhitelist := sdk.NewInt(100)
	whitelistPayload, err := json.Marshal(map[fungibleTokenMethod]accountAmountBodyFungibleTokenRequest{
		fungibleTokenMethodSetWhitelistedLimit: {
			Account: recipient.String(),
			Amount:  amountToWhitelist.String(),
		},
	})
	requireT.NoError(err)

	txf = txf.WithSimulateAndExecute(true)
	_, err = executeWASMContract(ctx, clientCtx, txf, contractAddr, whitelistPayload, sdk.Coin{})
	requireT.NoError(err)

	whitelistedRes, err := ftClient.WhitelistedBalance(ctx, &assetfttypes.QueryWhitelistedBalanceRequest{
		Account: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(amountToWhitelist.String(), whitelistedRes.Balance.Amount.String())

	// ********** Query **********

	// ********** Token **********

	tokenPayload, err := json.Marshal(map[fungibleTokenMethod]struct{}{
		fungibleTokenMethodToken: {},
	})
	requireT.NoError(err)
	queryOut, err := queryWASMContract(ctx, clientCtx, contractAddr, tokenPayload)
	requireT.NoError(err)
	var wasmTokenRes assetfttypes.QueryTokenResponse
	requireT.NoError(json.Unmarshal(queryOut, &wasmTokenRes))
	requireT.Equal(
		expectedToken, wasmTokenRes.Token,
	)

	// ********** FrozenBalance **********

	frozenBalancePayload, err := json.Marshal(map[fungibleTokenMethod]accountBodyFungibleTokenRequest{
		fungibleTokenMethodFrozenBalance: {
			Account: recipient.String(),
		},
	})
	requireT.NoError(err)
	queryOut, err = queryWASMContract(ctx, clientCtx, contractAddr, frozenBalancePayload)
	requireT.NoError(err)
	var wasmFrozenBalanceRes assetfttypes.QueryFrozenBalanceResponse
	requireT.NoError(json.Unmarshal(queryOut, &wasmFrozenBalanceRes))
	requireT.Equal(
		sdk.NewCoin(denom, amountToFreeze.Sub(amountToUnfreeze)).String(), wasmFrozenBalanceRes.Balance.String(),
	)

	// ********** WhitelistedBalance **********

	whitelistedBalancePayload, err := json.Marshal(map[fungibleTokenMethod]accountBodyFungibleTokenRequest{
		fungibleTokenMethodWhitelistedBalance: {
			Account: recipient.String(),
		},
	})
	requireT.NoError(err)
	queryOut, err = queryWASMContract(ctx, clientCtx, contractAddr, whitelistedBalancePayload)
	requireT.NoError(err)
	var wasmWhitelistedBalanceRes assetfttypes.QueryWhitelistedBalanceResponse
	requireT.NoError(json.Unmarshal(queryOut, &wasmWhitelistedBalanceRes))
	requireT.Equal(
		sdk.NewCoin(denom, amountToWhitelist), wasmWhitelistedBalanceRes.Balance,
	)
}

func methodToEmptyBodyPayload(methodName simpleStateMethod) (json.RawMessage, error) {
	return json.Marshal(map[simpleStateMethod]struct{}{
		methodName: {},
	})
}

func incrementAndVerify(
	ctx context.Context,
	clientCtx client.Context,
	txf client.Factory,
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
	admin      sdk.AccAddress
	accessType wasmtypes.AccessType
	payload    json.RawMessage
	amount     sdk.Coin
	label      string
	CodeID     uint64
}

// deployAndInstantiateWASMContract deploys, instantiateWASMContract the wasm contract and returns its address.
func deployAndInstantiateWASMContract(ctx context.Context, clientCtx client.Context, txf client.Factory, wasmData []byte, initConfig instantiateConfig) (string, uint64, error) {
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
func executeWASMContract(ctx context.Context, clientCtx client.Context, txf client.Factory, contractAddr string, payload json.RawMessage, fundAmt sdk.Coin) (int64, error) {
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

	res, err := client.BroadcastTx(ctx, clientCtx, txf, msg)
	if err != nil {
		return 0, err
	}
	return res.GasUsed, nil
}

// queryWASMContract queries the contract with the requested payload.
func queryWASMContract(ctx context.Context, clientCtx client.Context, contractAddr string, payload json.RawMessage) (json.RawMessage, error) {
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

// isWASMContractPinned returns true if smart contract is pinned.
func isWASMContractPinned(ctx context.Context, clientCtx client.Context, codeID uint64) (bool, error) {
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
func deployWASMContract(ctx context.Context, clientCtx client.Context, txf client.Factory, wasmData []byte) (uint64, error) {
	msgStoreCode := &wasmtypes.MsgStoreCode{
		Sender:       clientCtx.FromAddress().String(),
		WASMByteCode: wasmData,
	}

	txf = txf.
		WithGasAdjustment(gasMultiplier)

	res, err := client.BroadcastTx(ctx, clientCtx, txf, msgStoreCode)
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
func instantiateWASMContract(ctx context.Context, clientCtx client.Context, txf client.Factory, req instantiateConfig) (string, error) {
	funds := sdk.NewCoins()
	if amount := req.amount; !amount.Amount.IsNil() {
		funds = funds.Add(amount)
	}
	msg := &wasmtypes.MsgInstantiateContract{
		Sender: clientCtx.FromAddress().String(),
		Admin:  req.admin.String(),
		CodeID: req.CodeID,
		Label:  req.label,
		Msg:    wasmtypes.RawContractMessage(req.payload),
		Funds:  funds,
	}

	txf = txf.
		WithGasAdjustment(gasMultiplier)

	res, err := client.BroadcastTx(ctx, clientCtx, txf, msg)
	if err != nil {
		return "", err
	}

	contractAddr, err := event.FindStringEventAttribute(res.Events, wasmtypes.EventTypeInstantiate, wasmtypes.AttributeKeyContractAddr)
	if err != nil {
		return "", err
	}

	return contractAddr, nil
}
