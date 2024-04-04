//go:build integrationtests

package modules

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	nfttypes "github.com/cosmos/cosmos-sdk/x/nft"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v4/integration-tests"
	moduleswasm "github.com/CoreumFoundation/coreum/v4/integration-tests/contracts/modules"
	"github.com/CoreumFoundation/coreum/v4/pkg/client"
	"github.com/CoreumFoundation/coreum/v4/testutil/event"
	"github.com/CoreumFoundation/coreum/v4/testutil/integration"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	assetnfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/nft/types"
	deterministicgastypes "github.com/CoreumFoundation/coreum/v4/x/deterministicgas/types"
)

// authz models

type authz struct {
	Granter string `json:"granter"`
}

//nolint:tagliatelle
type authzNFTOfferRequest struct {
	ClassID string   `json:"class_id"`
	ID      string   `json:"id"`
	Price   sdk.Coin `json:"price"`
}

//nolint:tagliatelle
type authzAcceptNFTOfferRequest struct {
	ClassID string `json:"class_id"`
	ID      string `json:"id"`
}

type authzNFTMethod string

const (
	offerNft       authzNFTMethod = "offer_nft"
	acceptNftOffer authzNFTMethod = "accept_nft_offer"
)

// fungible token wasm models
//
//nolint:tagliatelle
type issueFTRequest struct {
	Symbol             string                 `json:"symbol"`
	Subunit            string                 `json:"subunit"`
	Precision          uint32                 `json:"precision"`
	InitialAmount      string                 `json:"initial_amount"`
	Description        string                 `json:"description"`
	Features           []assetfttypes.Feature `json:"features"`
	BurnRate           string                 `json:"burn_rate"`
	SendCommissionRate string                 `json:"send_commission_rate"`
	URI                string                 `json:"uri"`
	URIHash            string                 `json:"uri_hash"`
}

type amountBodyFTRequest struct {
	Amount string `json:"amount"`
}

type amountRecipientBodyFTRequest struct {
	Amount    string `json:"amount"`
	Recipient string `json:"recipient"`
}

type accountAmountBodyFTRequest struct {
	Account string `json:"account"`
	Amount  string `json:"amount"`
}

type issuerBodyFTRequest struct {
	Issuer string `json:"issuer"`
}

type accountBodyFTRequest struct {
	Account string `json:"account"`
}

type ftMethod string

const (
	// tx.
	ftMethodMint                ftMethod = "mint"
	ftMethodBurn                ftMethod = "burn"
	ftMethodFreeze              ftMethod = "freeze"
	ftMethodUnfreeze            ftMethod = "unfreeze"
	ftMethodSetFrozen           ftMethod = "set_frozen"
	ftMethodGloballyFreeze      ftMethod = "globally_freeze"
	ftMethodGloballyUnfreeze    ftMethod = "globally_unfreeze"
	ftMethodSetWhitelistedLimit ftMethod = "set_whitelisted_limit"
	// query.
	ftMethodParams              ftMethod = "params"
	ftMethodTokens              ftMethod = "tokens"
	ftMethodToken               ftMethod = "token"
	ftMethodBalance             ftMethod = "balance"
	ftMethodFrozenBalance       ftMethod = "frozen_balance"
	ftMethodWhitelistedBalance  ftMethod = "whitelisted_balance"
	ftMethodFrozenBalances      ftMethod = "frozen_balances"
	ftMethodWhitelistedBalances ftMethod = "whitelisted_balances"
)

// TestContractInstantiation tests contract instantiation using two instantiation methods.
func TestContractInstantiation(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	admin := chain.GenAccount()

	requireT := require.New(t)
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(admin, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

	txf := chain.TxFactory().
		WithSimulateAndExecute(true)

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx,
		txf,
		admin,
		moduleswasm.BankSendWASM,
	)
	requireT.NoError(err)

	// instantiate

	contractAddr1, err := chain.Wasm.InstantiateWASMContract(
		ctx,
		txf,
		admin,
		integration.InstantiateConfig{
			CodeID:     codeID,
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    moduleswasm.EmptyPayload,
			Amount:     chain.NewCoin(sdkmath.NewInt(10000)),
			Label:      "bank_send",
		},
	)
	requireT.NoError(err)

	// instantiate again

	contractAddr2, err := chain.Wasm.InstantiateWASMContract(
		ctx,
		txf,
		admin,
		integration.InstantiateConfig{
			CodeID:     codeID,
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    moduleswasm.EmptyPayload,
			Amount:     chain.NewCoin(sdkmath.NewInt(10000)),
			Label:      "bank_send",
		},
	)
	requireT.NoError(err)
	requireT.NotEqual(contractAddr1, contractAddr2)

	// instantiate2 for the first time

	contractAddr3, err := chain.Wasm.InstantiateWASMContract2(
		ctx,
		txf,
		admin,
		[]byte{0x00},
		integration.InstantiateConfig{
			CodeID:     codeID,
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    moduleswasm.EmptyPayload,
			Amount:     chain.NewCoin(sdkmath.NewInt(10000)),
			Label:      "bank_send",
		},
	)
	requireT.NoError(err)

	// try instantiate2 again using the same salt - should fail

	_, err = chain.Wasm.InstantiateWASMContract2(
		ctx,
		txf,
		admin,
		[]byte{0x00},
		integration.InstantiateConfig{
			CodeID:     codeID,
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    moduleswasm.EmptyPayload,
			Amount:     chain.NewCoin(sdkmath.NewInt(10000)),
			Label:      "bank_send",
		},
	)
	requireT.ErrorContains(err, "duplicate")

	// instantiate2 with different salt - should succeed

	contractAddr4, err := chain.Wasm.InstantiateWASMContract2(
		ctx,
		txf,
		admin,
		[]byte{0x01},
		integration.InstantiateConfig{
			CodeID:     codeID,
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    moduleswasm.EmptyPayload,
			Amount:     chain.NewCoin(sdkmath.NewInt(10000)),
			Label:      "bank_send",
		},
	)
	requireT.NoError(err)
	requireT.NotEqual(contractAddr3, contractAddr4)
}

// TestWASMBankSendContract runs a contract deployment flow and tests that the contract is able to use Bank module
// to disperse the native coins.
func TestWASMBankSendContract(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	admin := chain.GenAccount()
	nativeDenom := chain.ChainSettings.Denom

	requireT := require.New(t)
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(admin, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

	clientCtx := chain.ClientContext
	txf := chain.TxFactory().
		WithSimulateAndExecute(true)
	bankClient := banktypes.NewQueryClient(clientCtx)

	// deployWASMContract and init contract with the initial coins amount
	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		txf,
		admin,
		moduleswasm.BankSendWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    moduleswasm.EmptyPayload,
			Amount:     chain.NewCoin(sdkmath.NewInt(10000)),
			Label:      "bank_send",
		},
	)
	requireT.NoError(err)

	// send additional coins to contract directly
	sdkContractAddress, err := sdk.AccAddressFromBech32(contractAddr)
	requireT.NoError(err)

	msg := &banktypes.MsgSend{
		FromAddress: admin.String(),
		ToAddress:   sdkContractAddress.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(5000))),
	}

	_, err = client.BroadcastTx(ctx, clientCtx.WithFromAddress(admin), txf, msg)
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

	// try to withdraw more than the admin has
	_, err = chain.Wasm.ExecuteWASMContract(
		ctx,
		txf.
			WithSimulateAndExecute(false).
			// the gas here is to try to execute the tx and don't fail on the gas estimation
			WithGas(uint64(getFeemodelParams(ctx, t, chain.ClientContext).MaxBlockGas)),
		admin,
		contractAddr,
		moduleswasm.BankSendExecuteWithdrawRequest(sdk.NewInt64Coin(nativeDenom, 16000), recipient),
		sdk.Coin{})
	requireT.True(cosmoserrors.ErrInsufficientFunds.Is(err))

	// send coin from the contract to test wallet
	res, err := chain.Wasm.ExecuteWASMContract(
		ctx,
		txf,
		admin,
		contractAddr,
		moduleswasm.BankSendExecuteWithdrawRequest(sdk.NewInt64Coin(nativeDenom, 5000), recipient),
		sdk.Coin{},
	)
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

	// verify deterministic gas event
	gasEvents, err := event.FindTypedEvents[*deterministicgastypes.EventGas](res.Events)
	require.NoError(t, err)
	require.Len(t, gasEvents, 1)

	msgGas, ok := chain.DeterministicGasConfig.GasRequiredByMessage(&banktypes.MsgSend{})
	require.True(t, ok)

	require.Equal(t, "cosmos.bank.v1beta1.MsgSend", gasEvents[0].MsgURL)
	require.Equal(t, msgGas, gasEvents[0].DeterministicGas)
	require.Greater(t, gasEvents[0].RealGas, uint64(0))
}

// TestWASMGasBankSendAndBankSend checks that a message containing a deterministic and a
// non-deterministic transaction takes gas within appropriate limits.
func TestWASMGasBankSendAndBankSend(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	admin := chain.GenAccount()

	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(admin, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

	// deployWASMContract and init contract with the initial coins amount
	clientCtx := chain.ClientContext
	txf := chain.TxFactory().
		WithSimulateAndExecute(true)

	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		txf,
		admin,
		moduleswasm.BankSendWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    moduleswasm.EmptyPayload,
			Amount:     chain.NewCoin(sdkmath.NewInt(10000)),
			Label:      "bank_send",
		},
	)
	requireT.NoError(err)

	// Send tokens
	recipient := chain.GenAccount()

	wasmBankSend := &wasmtypes.MsgExecuteContract{
		Sender:   admin.String(),
		Contract: contractAddr,
		Msg: wasmtypes.RawContractMessage(
			moduleswasm.BankSendExecuteWithdrawRequest(sdk.NewInt64Coin(chain.ChainSettings.Denom, 5000), recipient),
		),
		Funds: sdk.Coins{},
	}

	bankSend := &banktypes.MsgSend{
		FromAddress: admin.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(chain.ChainSettings.Denom, sdkmath.NewInt(1000))),
	}

	minGasExpected := chain.GasLimitByMsgs(&banktypes.MsgSend{}, &banktypes.MsgSend{})
	maxGasExpected := minGasExpected * 10

	txf = chain.ChainContext.TxFactory().WithGas(maxGasExpected)
	result, err := client.BroadcastTx(ctx, clientCtx.WithFromAddress(admin), txf, wasmBankSend, bankSend)
	require.NoError(t, err)

	assert.Greater(t, uint64(result.GasUsed), minGasExpected)
	assert.Less(t, uint64(result.GasUsed), maxGasExpected)
}

// TestWASMPinningAndUnpinningSmartContractUsingGovernance deploys simple smart contract, verifies that it works
// properly and then tests that pinning and unpinning through proposals works correctly. We also verify that
// pinned smart contract consumes less gas.
func TestWASMPinningAndUnpinningSmartContractUsingGovernance(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	admin := chain.GenAccount()
	proposer := chain.GenAccount()

	requireT := require.New(t)

	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	requireT.NoError(err)
	proposerBalance.Amount = proposerBalance.Amount.MulRaw(2)

	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(admin, chain.NewCoin(sdkmath.NewInt(5000000000))),
		integration.NewFundedAccount(proposer, proposerBalance),
	)

	// instantiateWASMContract the contract and set the initial counter state.
	initialPayload, err := json.Marshal(moduleswasm.SimpleState{
		Count: 1337,
	})
	requireT.NoError(err)

	txf := chain.TxFactory().
		WithSimulateAndExecute(true)

	contractAddr, codeID, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		txf,
		admin,
		moduleswasm.SimpleStateWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayload,
			Label:      "simple_state",
		},
	)
	requireT.NoError(err)

	// get the current counter state
	getCountPayload, err := moduleswasm.MethodToEmptyBodyPayload(moduleswasm.SimpleGetCount)
	requireT.NoError(err)
	queryOut, err := chain.Wasm.QueryWASMContract(ctx, contractAddr, getCountPayload)
	requireT.NoError(err)
	var response moduleswasm.SimpleState
	requireT.NoError(json.Unmarshal(queryOut, &response))
	requireT.Equal(1337, response.Count)

	// execute contract to increment the count
	gasUsedBeforePinning := moduleswasm.IncrementSimpleStateAndVerify(ctx, txf, admin, chain, contractAddr, requireT, 1338)

	// verify that smart contract is not pinned
	requireT.False(chain.Wasm.IsWASMContractPinned(ctx, codeID))

	// pin smart contract
	proposalMsg, err := chain.Governance.NewMsgSubmitProposal(
		ctx,
		proposer,
		[]sdk.Msg{
			&wasmtypes.MsgPinCodes{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				CodeIDs:   []uint64{codeID},
			},
		},
		"",
		"Pin smart contract",
		"Testing smart contract pinning",
	)
	requireT.NoError(err)

	proposalID, err := chain.Governance.Propose(ctx, t, proposalMsg)
	requireT.NoError(err)

	proposal, err := chain.Governance.GetProposal(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1.StatusVotingPeriod, proposal.Status)

	err = chain.Governance.VoteAll(ctx, govtypesv1.OptionYes, proposal.Id)
	requireT.NoError(err)

	// Wait for proposal result.
	finalStatus, err := chain.Governance.WaitForVotingToFinalize(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1.StatusPassed, finalStatus)

	requireT.True(chain.Wasm.IsWASMContractPinned(ctx, codeID))

	gasUsedAfterPinning := moduleswasm.IncrementSimpleStateAndVerify(ctx, txf, admin, chain, contractAddr, requireT, 1339)

	// unpin smart contract
	proposalMsg, err = chain.Governance.NewMsgSubmitProposal(
		ctx,
		proposer,
		[]sdk.Msg{
			&wasmtypes.MsgUnpinCodes{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				CodeIDs:   []uint64{codeID},
			},
		},
		"",
		"Unpin smart contract",
		"Testing smart contract unpinning",
	)
	requireT.NoError(err)

	requireT.NoError(err)
	proposalID, err = chain.Governance.Propose(ctx, t, proposalMsg)
	requireT.NoError(err)

	proposal, err = chain.Governance.GetProposal(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1.StatusVotingPeriod, proposal.Status)

	err = chain.Governance.VoteAll(ctx, govtypesv1.OptionYes, proposal.Id)
	requireT.NoError(err)
	finalStatus, err = chain.Governance.WaitForVotingToFinalize(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1.StatusPassed, finalStatus)

	requireT.False(chain.Wasm.IsWASMContractPinned(ctx, codeID))

	gasUsedAfterUnpinning := moduleswasm.IncrementSimpleStateAndVerify(
		ctx, txf, admin, chain, contractAddr, requireT, 1340,
	)

	t.Logf(
		"Gas saved on pinned contract, gasBeforePinning:%d, gasAfterPinning:%d",
		gasUsedBeforePinning, gasUsedAfterPinning,
	)

	assertT := assert.New(t)
	assertT.Less(gasUsedAfterPinning, gasUsedBeforePinning)
	assertT.Greater(gasUsedAfterUnpinning, gasUsedAfterPinning)
}

// TestWASMContractUpgrade deploys simple state smart contract do its upgrade and upgrades/migrates it.
func TestWASMContractUpgrade(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	admin := chain.GenAccount()
	noneAdmin := chain.GenAccount()

	requireT := require.New(t)

	wasmClient := wasmtypes.NewQueryClient(chain.ClientContext)

	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(admin, chain.NewCoin(sdkmath.NewInt(5000000))),
		integration.NewFundedAccount(noneAdmin, chain.NewCoin(sdkmath.NewInt(5000000))),
	)

	// instantiateWASMContract the contract and set the initial counter state.
	initialPayload, err := json.Marshal(moduleswasm.SimpleState{
		Count: 787,
	})
	requireT.NoError(err)

	txf := chain.TxFactory().
		WithSimulateAndExecute(true)

	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		txf,
		admin,
		moduleswasm.SimpleStateWASM,
		integration.InstantiateConfig{
			Admin:      admin,
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayload,
			Label:      "simple_state_before_upgrade",
		},
	)
	requireT.NoError(err)

	// get the current counter state before migration.
	getCountPayload, err := moduleswasm.MethodToEmptyBodyPayload(moduleswasm.SimpleGetCount)
	requireT.NoError(err)
	queryOut, err := chain.Wasm.QueryWASMContract(ctx, contractAddr, getCountPayload)
	requireT.NoError(err)
	var response moduleswasm.SimpleState
	requireT.NoError(json.Unmarshal(queryOut, &response))
	requireT.Equal(787, response.Count)

	// execute the migration.

	// deploy new version of the contract
	newCodeID, err := chain.Wasm.DeployWASMContract(
		ctx,
		txf,
		admin,
		moduleswasm.SimpleStateWASM,
	)
	requireT.NoError(err)
	// prepare migration payload.
	migrationPayload, err := json.Marshal(moduleswasm.SimpleState{
		Count: 999,
	})
	requireT.NoError(err)
	// try to migrate from non-admin.
	err = chain.Wasm.MigrateWASMContract(ctx, txf, noneAdmin, contractAddr, newCodeID, migrationPayload)
	requireT.Error(err)
	requireT.Contains(err.Error(), "unauthorized")
	// migrate from admin.
	requireT.NoError(chain.Wasm.MigrateWASMContract(ctx, txf, admin, contractAddr, newCodeID, migrationPayload))
	// check state after the migration.
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, getCountPayload)
	requireT.NoError(err)
	requireT.NoError(json.Unmarshal(queryOut, &response))
	requireT.Equal(999, response.Count)
	// check that the contract works after the migration.
	_ = moduleswasm.IncrementSimpleStateAndVerify(ctx, txf, admin, chain, contractAddr, requireT, 1000)

	contractInfoRes, err := wasmClient.ContractInfo(ctx, &wasmtypes.QueryContractInfoRequest{
		Address: contractAddr,
	})
	requireT.NoError(err)
	requireT.Equal(newCodeID, contractInfoRes.CodeID)
}

// TestUpdateAndClearAdminOfContract runs MsgUpdateAdmin and MsgClearAdmin tx types.
func TestUpdateAndClearAdminOfContract(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	admin := chain.GenAccount()
	newAdmin := chain.GenAccount()

	requireT := require.New(t)
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(admin, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)
	chain.FundAccountWithOptions(ctx, t, newAdmin, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&wasmtypes.MsgClearAdmin{},
		},
	})

	wasmClient := wasmtypes.NewQueryClient(chain.ClientContext)

	// deployWASMContract and init contract with the initial coins amount
	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		admin,
		moduleswasm.BankSendWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Admin:      admin,
			Payload:    moduleswasm.EmptyPayload,
			Amount:     chain.NewCoin(sdkmath.NewInt(10000)),
			Label:      "bank_send",
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

// TestWASMAuthzContract runs a contract deployment flow and tests that the contract is able to use the Authz module
// to send native coins in place of another one.
func TestWASMAuthzContract(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	granter := chain.GenAccount()
	receiver := chain.GenAccount()

	authzClient := authztypes.NewQueryClient(chain.ClientContext)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	nftClient := nfttypes.NewQueryClient(chain.ClientContext)

	totalAmountToSend := sdkmath.NewInt(2_000)

	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(granter, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

	// deployWASMContract and init contract with the granter.
	initialPayloadAuthzTransfer, err := json.Marshal(authz{
		Granter: granter.String(),
	})
	requireT.NoError(err)

	initialPayloadAuthzStargate, err := json.Marshal(struct{}{})
	requireT.NoError(err)

	initialPayloadAuthzNftTrade, err := json.Marshal(struct{}{})
	requireT.NoError(err)

	contractAddrAuthzTransfer, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		granter,
		moduleswasm.AuthzTransferWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayloadAuthzTransfer,
			Label:      "authzTransfer",
		},
	)
	requireT.NoError(err)

	contractAddrAuthzStargate, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		granter,
		moduleswasm.AuthzStargateWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayloadAuthzStargate,
			Label:      "authzStargate",
		},
	)
	requireT.NoError(err)

	contractAddrAuthzNftTrade, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		granter,
		moduleswasm.AuthzNftTradeWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayloadAuthzNftTrade,
			Label:      "authzNftTrade",
		},
	)
	requireT.NoError(err)

	// ********** Test sending funds with Authz **********

	// grant the bank send authorization
	grantMsg, err := authztypes.NewMsgGrant(
		granter,
		sdk.MustAccAddressFromBech32(contractAddrAuthzTransfer),
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{})),
		lo.ToPtr(time.Now().Add(time.Minute)),
	)
	require.NoError(t, err)

	txResult, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(grantMsg)),
		grantMsg,
	)
	requireT.NoError(err)
	requireT.Equal(chain.GasLimitByMsgs(grantMsg), uint64(txResult.GasUsed))
	// assert granted
	gransRes, err := authzClient.Grants(ctx, &authztypes.QueryGrantsRequest{
		Granter: granter.String(),
		Grantee: contractAddrAuthzTransfer,
	})
	requireT.NoError(err)
	requireT.Len(gransRes.Grants, 1)

	// ********** Transfer **********

	_, err = chain.Wasm.ExecuteWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		granter,
		contractAddrAuthzTransfer,
		moduleswasm.AuthZExecuteTransferRequest(receiver.String(), chain.NewCoin(totalAmountToSend)),
		sdk.Coin{},
	)
	requireT.NoError(err)

	// ********** Stargate **********

	// grant the bank send authorization
	grantMsg, err = authztypes.NewMsgGrant(
		granter,
		sdk.MustAccAddressFromBech32(contractAddrAuthzStargate),
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{})),
		lo.ToPtr(time.Now().Add(time.Minute)),
	)
	require.NoError(t, err)

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(grantMsg)),
		grantMsg,
	)
	require.NoError(t, err)

	msgSendAny, err := codectypes.NewAnyWithValue(&banktypes.MsgSend{
		FromAddress: granter.String(),
		ToAddress:   receiver.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(totalAmountToSend)),
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		granter,
		contractAddrAuthzStargate,
		moduleswasm.AuthZExecuteStargateRequest(&authztypes.MsgExec{
			Grantee: contractAddrAuthzStargate,
			Msgs: []*codectypes.Any{
				msgSendAny,
			},
		}),
		sdk.Coin{},
	)
	requireT.NoError(err)

	// check receiver balance

	receiverBalancesRes, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: receiver.String(),
	})
	requireT.NoError(err)
	requireT.Equal(chain.NewCoin(totalAmountToSend.MulRaw(2)).String(), receiverBalancesRes.Balances.String())

	// ********** Test trading an NFT for an AssetFT with Authz **********

	// Issue and mind an NFT to the sender (will offer it)

	issueMsg := &assetnfttypes.MsgIssueClass{
		Issuer:   granter.String(),
		Symbol:   "NFTClassSymbol",
		Features: []assetnfttypes.ClassFeature{},
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)

	classID := assetnfttypes.BuildClassID(issueMsg.Symbol, granter)

	mintMsg := &assetnfttypes.MsgMint{
		Sender:    granter.String(),
		Recipient: granter.String(),
		ID:        "id-1",
		ClassID:   classID,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)

	requireT.NoError(err)

	// Issue an AssetFT that will be used to buy the NFT

	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(receiver, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

	issueAssetFTMsg := &assetfttypes.MsgIssue{
		Issuer:        receiver.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdkmath.NewInt(100000),
		Features:      []assetfttypes.Feature{},
		URI:           "https://my-class-meta.valid/1",
		URIHash:       "content-hash",
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(receiver),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueAssetFTMsg)),
		issueAssetFTMsg,
	)
	requireT.NoError(err)

	denom := assetfttypes.BuildDenom(issueAssetFTMsg.Subunit, receiver)

	// grant the nft transfer authorization to the contract
	grantMsg, err = authztypes.NewMsgGrant(
		granter,
		sdk.MustAccAddressFromBech32(contractAddrAuthzNftTrade),
		assetnfttypes.NewSendAuthorization([]assetnfttypes.NFTIdentifier{
			{ClassId: classID, Id: "id-1"},
		}),
		lo.ToPtr(time.Now().Add(time.Minute)),
	)
	requireT.NoError(err)
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(grantMsg)),
		grantMsg,
	)
	requireT.NoError(err)

	// assert granted
	gransRes, err = authzClient.Grants(ctx, &authztypes.QueryGrantsRequest{
		Granter: granter.String(),
		Grantee: contractAddrAuthzNftTrade,
	})
	requireT.NoError(err)
	requireT.Len(gransRes.Grants, 1)
	updatedGrant := assetnfttypes.SendAuthorization{}
	chain.ClientContext.Codec().MustUnmarshal(gransRes.Grants[0].Authorization.Value, &updatedGrant)
	requireT.ElementsMatch([]assetnfttypes.NFTIdentifier{
		{ClassId: classID, Id: "id-1"},
	}, updatedGrant.Nfts)

	// Make the offer of the NFT for the AssetFT

	nftOfferPayload, err := json.Marshal(map[authzNFTMethod]authzNFTOfferRequest{
		offerNft: {
			ClassID: classID,
			ID:      "id-1",
			Price:   sdk.NewCoin(denom, sdkmath.NewInt(10000)),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(
		ctx, chain.TxFactory().WithSimulateAndExecute(true), granter, contractAddrAuthzNftTrade, nftOfferPayload, sdk.Coin{},
	)
	requireT.NoError(err)

	ownerResp, err := nftClient.Owner(ctx, &nfttypes.QueryOwnerRequest{
		ClassId: classID,
		Id:      "id-1",
	})
	requireT.NoError(err)
	requireT.EqualValues(ownerResp.Owner, contractAddrAuthzNftTrade)

	// Accept the offer
	acceptNftOfferPayload, err := json.Marshal(map[authzNFTMethod]authzAcceptNFTOfferRequest{
		acceptNftOffer: {
			ClassID: classID,
			ID:      "id-1",
		},
	})
	requireT.NoError(err)
	_, err = chain.Wasm.ExecuteWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		receiver,
		contractAddrAuthzNftTrade,
		acceptNftOfferPayload,
		sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(10000)},
	)
	requireT.NoError(err)

	balanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: granter.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal("10000", balanceRes.Balance.Amount.String())

	ownerResp, err = nftClient.Owner(ctx, &nfttypes.QueryOwnerRequest{
		ClassId: classID,
		Id:      "id-1",
	})
	requireT.NoError(err)
	requireT.EqualValues(ownerResp.Owner, receiver.String())
}

// TestWASMFungibleTokenInContract verifies that smart contract is able to execute all fungible token message
// and core queries.
func TestWASMFungibleTokenInContract(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	admin := chain.GenAccount()
	recipient1 := chain.GenAccount()
	recipient2 := chain.GenAccount()

	requireT := require.New(t)
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(admin, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

	clientCtx := chain.ClientContext
	txf := chain.TxFactory().
		WithSimulateAndExecute(true)
	bankClient := banktypes.NewQueryClient(clientCtx)
	ftClient := assetfttypes.NewQueryClient(clientCtx)

	// ********** Issuance **********

	burnRate := sdk.MustNewDecFromStr("0.1")
	sendCommissionRate := sdk.MustNewDecFromStr("0.2")

	issuanceAmount := sdkmath.NewInt(10_000)
	issuanceReq := issueFTRequest{
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
		URI:                "https://example.com",
		URIHash:            "1234567890abcdef",
	}
	issuerFTInstantiatePayload, err := json.Marshal(issuanceReq)
	requireT.NoError(err)

	// instantiate new contract
	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		txf,
		admin,
		moduleswasm.FTWASM,
		integration.InstantiateConfig{
			// we add the initial amount to let the contract issue the token on behalf of it
			Amount:     chain.QueryAssetFTParams(ctx, t).IssueFee,
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    issuerFTInstantiatePayload,
			Label:      "fungible_token",
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
		Version:            assetfttypes.CurrentTokenVersion, // test should work with any token version
		URI:                issuanceReq.URI,
		URIHash:            issuanceReq.URIHash,
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

	txf = txf.WithSimulateAndExecute(true)

	// ********** Transactions **********

	// ********** Mint **********

	amountToMint := sdkmath.NewInt(500)
	mintPayload, err := json.Marshal(map[ftMethod]amountBodyFTRequest{
		ftMethodMint: {
			Amount: amountToMint.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, mintPayload, sdk.Coin{})
	requireT.NoError(err)

	balanceRes, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: contractAddr,
		Denom:   denom,
	})
	requireT.NoError(err)
	newAmount := issuanceAmount.Add(amountToMint)
	requireT.Equal(newAmount.String(), balanceRes.Balance.Amount.String())

	// ********** Mint (sending to someone) **********

	amountToMint = sdkmath.NewInt(100)
	whitelistPayload, err := json.Marshal(map[ftMethod]accountAmountBodyFTRequest{
		ftMethodSetWhitelistedLimit: {
			Account: recipient2.String(),
			Amount:  amountToMint.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, whitelistPayload, sdk.Coin{})
	requireT.NoError(err)

	mintPayload, err = json.Marshal(map[ftMethod]amountRecipientBodyFTRequest{
		ftMethodMint: {
			Amount:    amountToMint.String(),
			Recipient: recipient2.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, mintPayload, sdk.Coin{})
	requireT.NoError(err)

	balanceRes, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient2.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(amountToMint.String(), balanceRes.Balance.Amount.String())

	// ********** Burn **********

	amountToBurn := sdkmath.NewInt(100)
	burnPayload, err := json.Marshal(map[ftMethod]amountBodyFTRequest{
		ftMethodBurn: {
			Amount: amountToBurn.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, burnPayload, sdk.Coin{})
	requireT.NoError(err)

	balanceRes, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: contractAddr,
		Denom:   denom,
	})
	requireT.NoError(err)
	newAmount = newAmount.Sub(amountToBurn)
	requireT.Equal(newAmount.String(), balanceRes.Balance.Amount.String())

	// ********** Freeze **********

	amountToFreeze := sdkmath.NewInt(100)
	freezePayload, err := json.Marshal(map[ftMethod]accountAmountBodyFTRequest{
		ftMethodFreeze: {
			Account: recipient1.String(),
			Amount:  amountToFreeze.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, freezePayload, sdk.Coin{})
	requireT.NoError(err)

	frozenRes, err := ftClient.FrozenBalance(ctx, &assetfttypes.QueryFrozenBalanceRequest{
		Account: recipient1.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(amountToFreeze.String(), frozenRes.Balance.Amount.String())

	// ********** Unfreeze **********

	amountToUnfreeze := sdkmath.NewInt(40)
	unfreezePayload, err := json.Marshal(map[ftMethod]accountAmountBodyFTRequest{
		ftMethodUnfreeze: {
			Account: recipient1.String(),
			Amount:  amountToUnfreeze.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, unfreezePayload, sdk.Coin{})
	requireT.NoError(err)

	frozenRes, err = ftClient.FrozenBalance(ctx, &assetfttypes.QueryFrozenBalanceRequest{
		Account: recipient1.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(amountToFreeze.Sub(amountToUnfreeze).String(), frozenRes.Balance.Amount.String())

	// ********** SetFrozen **********

	amountToSetFrozen := sdkmath.NewInt(30)
	setFrozenPayload, err := json.Marshal(map[ftMethod]accountAmountBodyFTRequest{
		ftMethodSetFrozen: {
			Account: recipient1.String(),
			Amount:  amountToSetFrozen.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, setFrozenPayload, sdk.Coin{})
	requireT.NoError(err)

	frozenRes, err = ftClient.FrozenBalance(ctx, &assetfttypes.QueryFrozenBalanceRequest{
		Account: recipient1.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(amountToSetFrozen.String(), frozenRes.Balance.Amount.String())

	// ********** GloballyFreeze **********

	globallyFreezePayload, err := json.Marshal(map[ftMethod]struct{}{
		ftMethodGloballyFreeze: {},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, globallyFreezePayload, sdk.Coin{})
	requireT.NoError(err)

	tokenRes, err = ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: denom,
	})
	requireT.NoError(err)
	requireT.True(tokenRes.Token.GloballyFrozen)

	// ********** GloballyUnfreeze **********

	globallyUnfreezePayload, err := json.Marshal(map[ftMethod]struct{}{
		ftMethodGloballyUnfreeze: {},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, globallyUnfreezePayload, sdk.Coin{})
	requireT.NoError(err)

	tokenRes, err = ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: denom,
	})
	requireT.NoError(err)
	requireT.False(tokenRes.Token.GloballyFrozen)

	// ********** Whitelisting **********

	amountToWhitelist := sdkmath.NewInt(100)
	whitelistPayload, err = json.Marshal(map[ftMethod]accountAmountBodyFTRequest{
		ftMethodSetWhitelistedLimit: {
			Account: recipient1.String(),
			Amount:  amountToWhitelist.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, whitelistPayload, sdk.Coin{})
	requireT.NoError(err)

	whitelistedRes, err := ftClient.WhitelistedBalance(ctx, &assetfttypes.QueryWhitelistedBalanceRequest{
		Account: recipient1.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(amountToWhitelist.String(), whitelistedRes.Balance.Amount.String())

	// ********** Query **********

	// ********** Params **********

	paramsPayLoad, err := json.Marshal(map[ftMethod]struct{}{
		ftMethodParams: {},
	})
	requireT.NoError(err)
	queryOut, err := chain.Wasm.QueryWASMContract(ctx, contractAddr, paramsPayLoad)
	requireT.NoError(err)
	var wasmParamsRes assetfttypes.QueryParamsResponse
	requireT.NoError(json.Unmarshal(queryOut, &wasmParamsRes))
	requireT.Equal(
		chain.QueryAssetFTParams(ctx, t).IssueFee, wasmParamsRes.Params.IssueFee,
	)

	// ********** Token **********

	tokenPayload, err := json.Marshal(map[ftMethod]struct{}{
		ftMethodToken: {},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, tokenPayload)
	requireT.NoError(err)
	var wasmTokenRes assetfttypes.QueryTokenResponse
	requireT.NoError(json.Unmarshal(queryOut, &wasmTokenRes))
	wasmTokenRes.Token.Version = expectedToken.Version // test should work with any version
	requireT.Equal(
		expectedToken, wasmTokenRes.Token,
	)

	// ********** Tokens **********

	tokensPayload, err := json.Marshal(map[ftMethod]issuerBodyFTRequest{
		ftMethodTokens: {
			Issuer: contractAddr,
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, tokensPayload)
	requireT.NoError(err)
	var wasmTokensRes assetfttypes.QueryTokensResponse
	requireT.NoError(json.Unmarshal(queryOut, &wasmTokensRes))
	wasmTokensRes.Tokens[0].Version = expectedToken.Version
	requireT.Equal(
		expectedToken, wasmTokensRes.Tokens[0],
	)

	// ********** Balance **********

	balancePayload, err := json.Marshal(map[ftMethod]accountBodyFTRequest{
		ftMethodBalance: {
			Account: contractAddr,
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, balancePayload)
	requireT.NoError(err)
	var wasmBalanceRes assetfttypes.QueryBalanceResponse
	requireT.NoError(json.Unmarshal(queryOut, &wasmBalanceRes))
	requireT.Equal(
		newAmount.String(), wasmBalanceRes.Balance.String(),
	)

	// ********** FrozenBalance **********

	frozenBalancePayload, err := json.Marshal(map[ftMethod]accountBodyFTRequest{
		ftMethodFrozenBalance: {
			Account: recipient1.String(),
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, frozenBalancePayload)
	requireT.NoError(err)
	var wasmFrozenBalanceRes assetfttypes.QueryFrozenBalanceResponse
	requireT.NoError(json.Unmarshal(queryOut, &wasmFrozenBalanceRes))
	requireT.Equal(
		sdk.NewCoin(denom, amountToSetFrozen).String(), wasmFrozenBalanceRes.Balance.String(),
	)

	// ********** FrozenBalances **********

	frozenBalancesPayload, err := json.Marshal(map[ftMethod]accountBodyFTRequest{
		ftMethodFrozenBalances: {
			Account: recipient1.String(),
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, frozenBalancesPayload)
	requireT.NoError(err)
	var wasmFrozenBalancesRes assetfttypes.QueryFrozenBalancesResponse
	requireT.NoError(json.Unmarshal(queryOut, &wasmFrozenBalancesRes))
	requireT.Equal(
		sdk.NewCoin(denom, amountToSetFrozen).String(), wasmFrozenBalancesRes.Balances[0].String(),
	)

	// ********** WhitelistedBalance **********

	whitelistedBalancePayload, err := json.Marshal(map[ftMethod]accountBodyFTRequest{
		ftMethodWhitelistedBalance: {
			Account: recipient1.String(),
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, whitelistedBalancePayload)
	requireT.NoError(err)
	var wasmWhitelistedBalanceRes assetfttypes.QueryWhitelistedBalanceResponse
	requireT.NoError(json.Unmarshal(queryOut, &wasmWhitelistedBalanceRes))
	requireT.Equal(
		sdk.NewCoin(denom, amountToWhitelist), wasmWhitelistedBalanceRes.Balance,
	)

	// ********** WhitelistedBalances **********

	whitelistedBalancesPayload, err := json.Marshal(map[ftMethod]accountBodyFTRequest{
		ftMethodWhitelistedBalances: {
			Account: recipient1.String(),
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, whitelistedBalancesPayload)
	requireT.NoError(err)
	var wasmWhitelistedBalancesRes assetfttypes.QueryWhitelistedBalancesResponse
	requireT.NoError(json.Unmarshal(queryOut, &wasmWhitelistedBalancesRes))
	requireT.Equal(
		sdk.NewCoin(denom, amountToWhitelist), wasmWhitelistedBalancesRes.Balances[0],
	)
}

// TestWASMNonFungibleTokenInContract verifies that smart contract is able to execute all non-fungible
// token message and core queries.
//
//nolint:nosnakecase
func TestWASMNonFungibleTokenInContract(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	admin := chain.GenAccount()
	recipient := chain.GenAccount()
	mintRecipient := chain.GenAccount()

	requireT := require.New(t)
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(admin, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

	clientCtx := chain.ClientContext
	txf := chain.TxFactory().
		WithSimulateAndExecute(true)
	assetNftClient := assetnfttypes.NewQueryClient(clientCtx)
	nftClient := nfttypes.NewQueryClient(clientCtx)

	// ********** Issuance **********

	royaltyRate := sdk.MustNewDecFromStr("0.1")
	data := make([]byte, 256)
	for i := 0; i < 256; i++ {
		data[i] = uint8(i)
	}
	encodedData := base64.StdEncoding.EncodeToString(data)

	issueClassReq := moduleswasm.IssueNFTRequest{
		Name:        "name",
		Symbol:      "symbol",
		Description: "description",
		URI:         "https://my-nft-class-meta.invalid/1",
		URIHash:     "hash",
		Data:        encodedData,
		Features: []assetnfttypes.ClassFeature{
			assetnfttypes.ClassFeature_burning,
			assetnfttypes.ClassFeature_freezing,
			assetnfttypes.ClassFeature_whitelisting,
			assetnfttypes.ClassFeature_disable_sending,
		},
		RoyaltyRate: royaltyRate.String(),
	}
	issuerNFTInstantiatePayload, err := json.Marshal(issueClassReq)
	requireT.NoError(err)

	// instantiate new contract
	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		txf,
		admin,
		moduleswasm.NftWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    issuerNFTInstantiatePayload,
			Label:      "non_fungible_token",
		},
	)
	requireT.NoError(err)

	classID := assetnfttypes.BuildClassID(issueClassReq.Symbol, sdk.MustAccAddressFromBech32(contractAddr))
	classRes, err := assetNftClient.Class(ctx, &assetnfttypes.QueryClassRequest{Id: classID})
	requireT.NoError(err)

	dataBytes, err := codectypes.NewAnyWithValue(&assetnfttypes.DataBytes{Data: data})
	// we need to do this, otherwise assertion fails because some private fields are set differently
	dataToCompare := &codectypes.Any{
		TypeUrl: dataBytes.TypeUrl,
		Value:   dataBytes.Value,
	}
	requireT.NoError(err)

	expectedClass := assetnfttypes.Class{
		Id:          classID,
		Issuer:      contractAddr,
		Name:        issueClassReq.Name,
		Symbol:      issueClassReq.Symbol,
		Description: issueClassReq.Description,
		URI:         issueClassReq.URI,
		URIHash:     issueClassReq.URIHash,
		Data:        dataToCompare,
		Features:    issueClassReq.Features,
		RoyaltyRate: royaltyRate,
	}
	requireT.Equal(
		expectedClass, classRes.Class,
	)

	// ********** Mint **********

	mintNFTReq1 := moduleswasm.NftMintRequest{
		ID:      "id-1",
		URI:     "https://my-nft-meta.invalid/1",
		URIHash: "hash",
		Data:    encodedData,
	}
	mintPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftMintRequest{
		moduleswasm.NftMethodMint: mintNFTReq1,
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, mintPayload, sdk.Coin{})
	requireT.NoError(err)

	nftResp, err := nftClient.NFT(ctx, &nfttypes.QueryNFTRequest{
		ClassId: classID,
		Id:      mintNFTReq1.ID,
	})
	requireT.NoError(err)

	expectedNFT1 := &nfttypes.NFT{
		ClassId: classID,
		Id:      mintNFTReq1.ID,
		Uri:     mintNFTReq1.URI,
		UriHash: mintNFTReq1.URIHash,
		Data:    dataToCompare,
	}
	requireT.Equal(
		expectedNFT1, nftResp.Nft,
	)

	nftOwner, err := nftClient.Owner(ctx, &nfttypes.QueryOwnerRequest{
		ClassId: classID,
		Id:      mintNFTReq1.ID,
	})
	requireT.NoError(err)
	requireT.Equal(nftOwner.Owner, contractAddr)

	// ********** Mint (to someone) **********

	issueClassReqNoWhitelist := moduleswasm.IssueNFTRequest{
		Name:        "name",
		Symbol:      "symbol",
		Description: "description",
		URI:         "https://my-nft-class-meta.invalid/1",
		URIHash:     "hash",
		Data:        encodedData,
		Features: []assetnfttypes.ClassFeature{
			assetnfttypes.ClassFeature_burning,
			assetnfttypes.ClassFeature_freezing,
			assetnfttypes.ClassFeature_disable_sending,
		},
		RoyaltyRate: royaltyRate.String(),
	}
	issuerNFTInstantiatePayload, err = json.Marshal(issueClassReqNoWhitelist)
	requireT.NoError(err)

	// instantiate new contract
	contractAddrNoWhitelist, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		txf,
		admin,
		moduleswasm.NftWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    issuerNFTInstantiatePayload,
			Label:      "non_fungible_token",
		},
	)
	requireT.NoError(err)

	classIDNoWhitelist := assetnfttypes.BuildClassID(
		issueClassReq.Symbol, sdk.MustAccAddressFromBech32(contractAddrNoWhitelist),
	)

	mintNFTReq1NoWhitelist := moduleswasm.NftMintRequest{
		ID:        "id-1",
		Recipient: mintRecipient.String(),
	}

	// mint
	mintPayload, err = json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftMintRequest{
		moduleswasm.NftMethodMint: mintNFTReq1NoWhitelist,
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddrNoWhitelist, mintPayload, sdk.Coin{})
	requireT.NoError(err)

	nftResp, err = nftClient.NFT(ctx, &nfttypes.QueryNFTRequest{
		ClassId: classIDNoWhitelist,
		Id:      mintNFTReq1NoWhitelist.ID,
	})
	requireT.NoError(err)

	expectedNFT1 = &nfttypes.NFT{
		ClassId: classIDNoWhitelist,
		Id:      mintNFTReq1NoWhitelist.ID,
	}
	requireT.Equal(
		expectedNFT1, nftResp.Nft,
	)

	nftOwner, err = nftClient.Owner(ctx, &nfttypes.QueryOwnerRequest{
		ClassId: classIDNoWhitelist,
		Id:      mintNFTReq1NoWhitelist.ID,
	})
	requireT.NoError(err)
	requireT.Equal(nftOwner.Owner, mintRecipient.String())

	// ********** Freeze **********

	freezePayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftIDRequest{
		moduleswasm.NftMethodFreeze: {
			ID: mintNFTReq1.ID,
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, freezePayload, sdk.Coin{})
	requireT.NoError(err)

	assertNftFrozenRes, err := assetNftClient.Frozen(ctx, &assetnfttypes.QueryFrozenRequest{
		Id:      mintNFTReq1.ID,
		ClassId: classID,
	})
	requireT.NoError(err)
	requireT.True(assertNftFrozenRes.Frozen)

	// ********** Unfreeze **********

	unfreezePayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftIDRequest{
		moduleswasm.NftMethodUnfreeze: {
			ID: mintNFTReq1.ID,
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, unfreezePayload, sdk.Coin{})
	requireT.NoError(err)

	assertNftFrozenRes, err = assetNftClient.Frozen(ctx, &assetnfttypes.QueryFrozenRequest{
		Id:      mintNFTReq1.ID,
		ClassId: classID,
	})
	requireT.NoError(err)
	requireT.False(assertNftFrozenRes.Frozen)

	// ********** ClassFreeze **********

	classFreezePayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftAccountRequest{
		moduleswasm.NftMethodClassFreeze: {
			Account: recipient.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, classFreezePayload, sdk.Coin{})
	requireT.NoError(err)

	assertNftClassFrozenRes, err := assetNftClient.ClassFrozen(ctx, &assetnfttypes.QueryClassFrozenRequest{
		ClassId: classID,
		Account: recipient.String(),
	})
	requireT.NoError(err)
	requireT.True(assertNftClassFrozenRes.Frozen)

	// ********** ClassUnFreeze **********

	classUnfreezePayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftAccountRequest{
		moduleswasm.NftMethodClassUnfreeze: {
			Account: recipient.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, classUnfreezePayload, sdk.Coin{})
	requireT.NoError(err)

	assertNftClassFrozenRes, err = assetNftClient.ClassFrozen(ctx, &assetnfttypes.QueryClassFrozenRequest{
		ClassId: classID,
		Account: recipient.String(),
	})
	requireT.NoError(err)
	requireT.False(assertNftClassFrozenRes.Frozen)

	// ********** AddToWhitelist **********

	addToWhitelistPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftIDWithAccountRequest{
		moduleswasm.NftMethodAddToWhitelist: {
			ID:      mintNFTReq1.ID,
			Account: recipient.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, addToWhitelistPayload, sdk.Coin{})
	requireT.NoError(err)

	assertNftWhitelistedRes, err := assetNftClient.Whitelisted(ctx, &assetnfttypes.QueryWhitelistedRequest{
		Id:      mintNFTReq1.ID,
		ClassId: classID,
		Account: recipient.String(),
	})
	requireT.NoError(err)
	requireT.True(assertNftWhitelistedRes.Whitelisted)

	// ********** RemoveFromWhitelist **********

	removeFromWhitelistPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftIDWithAccountRequest{
		moduleswasm.NftMethodRemoveFromWhiteList: {
			ID:      mintNFTReq1.ID,
			Account: recipient.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, removeFromWhitelistPayload, sdk.Coin{})
	requireT.NoError(err)

	assertNftWhitelistedRes, err = assetNftClient.Whitelisted(ctx, &assetnfttypes.QueryWhitelistedRequest{
		Id:      mintNFTReq1.ID,
		ClassId: classID,
		Account: recipient.String(),
	})
	requireT.NoError(err)
	requireT.False(assertNftWhitelistedRes.Whitelisted)

	// ********** AddToClassWhitelist **********

	addToClassWhitelistPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftAccountRequest{
		moduleswasm.NftMethodAddToClassWhitelist: {
			Account: recipient.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, addToClassWhitelistPayload, sdk.Coin{})
	requireT.NoError(err)

	assertNftClassWhitelistedRes, err := assetNftClient.ClassWhitelistedAccounts(
		ctx,
		&assetnfttypes.QueryClassWhitelistedAccountsRequest{
			ClassId: classID,
		})
	requireT.NoError(err)
	requireT.Contains(assertNftClassWhitelistedRes.Accounts, recipient.String())

	// ********** RemoveFromClassWhitelist **********

	removeFromClassWhitelistPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftAccountRequest{
		moduleswasm.NftMethodRemoveFromClassWhitelist: {
			Account: recipient.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, removeFromClassWhitelistPayload, sdk.Coin{})
	requireT.NoError(err)

	assertNftClassWhitelistedRes, err = assetNftClient.ClassWhitelistedAccounts(
		ctx,
		&assetnfttypes.QueryClassWhitelistedAccountsRequest{
			ClassId: classID,
		})
	requireT.NoError(err)
	requireT.NotContains(assertNftClassWhitelistedRes.Accounts, recipient.String())

	// ********** Burn **********

	burnPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftIDRequest{
		moduleswasm.NftMethodBurn: {
			ID: mintNFTReq1.ID,
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, burnPayload, sdk.Coin{})
	requireT.NoError(err)

	_, err = nftClient.NFT(ctx, &nfttypes.QueryNFTRequest{
		ClassId: classID,
		Id:      mintNFTReq1.ID,
	})
	requireT.Error(err)
	// the nft wraps the errors with the `errors` so the client doesn't decode them as sdk errors.
	requireT.Contains(err.Error(), nfttypes.ErrNFTNotExists.Error())

	// ********** Send **********

	// mint
	mintNFTReq2 := mintNFTReq1
	mintNFTReq2.ID = "id-2"
	mint2Payload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftMintRequest{
		moduleswasm.NftMethodMint: mintNFTReq2,
	})
	requireT.NoError(err)
	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, mint2Payload, sdk.Coin{})
	requireT.NoError(err)

	// addToWhitelistPayload
	addToWhitelistPayload, err = json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftIDWithAccountRequest{
		moduleswasm.NftMethodAddToWhitelist: {
			ID:      mintNFTReq2.ID,
			Account: recipient.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, addToWhitelistPayload, sdk.Coin{})
	requireT.NoError(err)

	// send
	sendPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftIDWithReceiverRequest{
		moduleswasm.NftMethodSend: {
			ID:       mintNFTReq2.ID,
			Receiver: recipient.String(),
		},
	})
	requireT.NoError(err)
	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, sendPayload, sdk.Coin{})
	requireT.NoError(err)

	// ********** Query **********

	// ********** Params **********

	paramsPayLoad, err := json.Marshal(map[moduleswasm.NftMethod]struct{}{
		moduleswasm.NftMethodParams: {},
	})
	requireT.NoError(err)
	queryOut, err := chain.Wasm.QueryWASMContract(ctx, contractAddr, paramsPayLoad)
	requireT.NoError(err)
	var wasmParamsRes assetnfttypes.QueryParamsResponse
	requireT.NoError(json.Unmarshal(queryOut, &wasmParamsRes))
	requireT.Equal(
		chain.QueryAssetNFTParams(ctx, t).MintFee, wasmParamsRes.Params.MintFee,
	)

	// ********** Class **********

	classPayload, err := json.Marshal(map[moduleswasm.NftMethod]struct{}{
		moduleswasm.NftMethodClass: {},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, classPayload)
	requireT.NoError(err)
	var classQueryRes moduleswasm.AssetnftClassResponse
	requireT.NoError(json.Unmarshal(queryOut, &classQueryRes))
	requireT.Equal(
		moduleswasm.AssetnftClass{
			ID:          expectedClass.Id,
			Issuer:      expectedClass.Issuer,
			Name:        expectedClass.Name,
			Symbol:      expectedClass.Symbol,
			Description: expectedClass.Description,
			URI:         expectedClass.URI,
			URIHash:     expectedClass.URIHash,
			Data:        encodedData,
			Features:    expectedClass.Features,
			RoyaltyRate: expectedClass.RoyaltyRate,
		}, classQueryRes.Class,
	)

	// ********** Classes **********

	classesPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftIssuerRequest{
		moduleswasm.NftMethodClasses: {
			Issuer: contractAddr,
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, classesPayload)
	requireT.NoError(err)
	var classesQueryRes assetnfttypes.QueryClassesResponse
	requireT.NoError(json.Unmarshal(queryOut, &classesQueryRes))
	requireT.Equal(
		moduleswasm.AssetnftClass{
			ID:          expectedClass.Id,
			Issuer:      expectedClass.Issuer,
			Name:        expectedClass.Name,
			Symbol:      expectedClass.Symbol,
			Description: expectedClass.Description,
			URI:         expectedClass.URI,
			URIHash:     expectedClass.URIHash,
			Data:        encodedData,
			Features:    expectedClass.Features,
			RoyaltyRate: expectedClass.RoyaltyRate,
		}, moduleswasm.AssetnftClass{
			ID:          classesQueryRes.Classes[0].Id,
			Issuer:      classesQueryRes.Classes[0].Issuer,
			Name:        classesQueryRes.Classes[0].Name,
			Symbol:      classesQueryRes.Classes[0].Symbol,
			Description: classesQueryRes.Classes[0].Description,
			URI:         classesQueryRes.Classes[0].URI,
			URIHash:     classesQueryRes.Classes[0].URIHash,
			Data:        encodedData,
			Features:    classesQueryRes.Classes[0].Features,
			RoyaltyRate: classesQueryRes.Classes[0].RoyaltyRate,
		},
	)

	// ********** Frozen **********

	freezePayload, err = json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftIDRequest{
		moduleswasm.NftMethodFreeze: {
			ID: mintNFTReq2.ID,
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, freezePayload, sdk.Coin{})
	requireT.NoError(err)

	frozenPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftIDRequest{
		moduleswasm.NftMethodFrozen: {
			ID: mintNFTReq2.ID,
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, frozenPayload)
	requireT.NoError(err)
	var frozenQueryRes assetnfttypes.QueryFrozenResponse
	requireT.NoError(json.Unmarshal(queryOut, &frozenQueryRes))
	requireT.True(frozenQueryRes.Frozen)

	// ********** ClassFrozen **********

	classFreezePayload, err = json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftAccountRequest{
		moduleswasm.NftMethodClassFreeze: {
			Account: recipient.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, classFreezePayload, sdk.Coin{})
	requireT.NoError(err)

	classFrozenPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftAccountRequest{
		moduleswasm.NftMethodClassFrozen: {
			Account: recipient.String(),
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, classFrozenPayload)
	requireT.NoError(err)
	var classFrozenQueryRes assetnfttypes.QueryClassFrozenResponse
	requireT.NoError(json.Unmarshal(queryOut, &classFrozenQueryRes))
	requireT.True(frozenQueryRes.Frozen)

	// ********** ClassFrozenAccounts **********

	classFreezePayload, err = json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftAccountRequest{
		moduleswasm.NftMethodClassFreeze: {
			Account: recipient.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, classFreezePayload, sdk.Coin{})
	requireT.NoError(err)

	classFrozenAccountsPayLoad, err := json.Marshal(map[moduleswasm.NftMethod]struct{}{
		moduleswasm.NftMethodClassFrozenAccounts: {},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, classFrozenAccountsPayLoad)
	requireT.NoError(err)
	var classFrozenAccountsQueryRes assetnfttypes.QueryClassFrozenAccountsResponse
	requireT.NoError(json.Unmarshal(queryOut, &classFrozenAccountsQueryRes))
	requireT.Contains(classFrozenAccountsQueryRes.Accounts, recipient.String())

	// ********** Whitelisted **********

	whitelistedPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftIDWithAccountRequest{
		moduleswasm.NftMethodWhitelisted: {
			ID:      mintNFTReq2.ID,
			Account: recipient.String(),
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, whitelistedPayload)
	requireT.NoError(err)
	var whitelistedQueryRes assetnfttypes.QueryWhitelistedResponse
	requireT.NoError(json.Unmarshal(queryOut, &whitelistedQueryRes))
	requireT.True(whitelistedQueryRes.Whitelisted)

	// ********** WhitelistedAccountsforNFT **********

	whitelistedAccountsForNFTPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftIDRequest{
		moduleswasm.NftMethodWhitelistedAccountsForNft: {
			ID: mintNFTReq2.ID,
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, whitelistedAccountsForNFTPayload)
	requireT.NoError(err)
	var whitelistedAccountsForNFTQueryRes assetnfttypes.QueryWhitelistedAccountsForNFTResponse
	requireT.NoError(json.Unmarshal(queryOut, &whitelistedAccountsForNFTQueryRes))
	requireT.Equal(whitelistedAccountsForNFTQueryRes.Accounts[0], recipient.String())

	// ********** ClassWhitelistedAccounts **********

	addToClassWhitelistPayload, err = json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftAccountRequest{
		moduleswasm.NftMethodAddToClassWhitelist: {
			Account: recipient.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, addToClassWhitelistPayload, sdk.Coin{})
	requireT.NoError(err)

	classWhitelistedAccountsPayLoad, err := json.Marshal(map[moduleswasm.NftMethod]struct{}{
		moduleswasm.NftMethodClassWhitelistedAccounts: {},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, classWhitelistedAccountsPayLoad)
	requireT.NoError(err)
	var classWhitelistedAccountsQueryRes assetnfttypes.QueryClassWhitelistedAccountsResponse
	requireT.NoError(json.Unmarshal(queryOut, &classWhitelistedAccountsQueryRes))
	requireT.Contains(classWhitelistedAccountsQueryRes.Accounts, recipient.String())

	// ********** BurntNFT **********

	burntNFTPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.BurntNftIDRequest{
		moduleswasm.NftMethodBurntNft: {
			NftID: "id-1",
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, burntNFTPayload)
	requireT.NoError(err)
	var burntNFTQueryRes assetnfttypes.QueryBurntNFTResponse
	requireT.NoError(json.Unmarshal(queryOut, &burntNFTQueryRes))
	requireT.True(burntNFTQueryRes.Burnt)

	// ********** BurntNFTsInClass **********

	burntNFTsInClassPayload, err := json.Marshal(map[moduleswasm.NftMethod]struct{}{
		moduleswasm.NftMethodBurntNftInClass: {},
	})

	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, burntNFTsInClassPayload)
	requireT.NoError(err)
	var burntNFTsInClassQueryRes assetnfttypes.QueryBurntNFTsInClassResponse
	requireT.NoError(json.Unmarshal(queryOut, &burntNFTsInClassQueryRes))
	requireT.Equal([]string{"id-1"}, burntNFTsInClassQueryRes.NftIds)

	// ********** Balance **********

	balancePayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftOwnerRequest{
		moduleswasm.NftMethodBalance: {
			Owner: recipient.String(),
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, balancePayload)
	requireT.NoError(err)
	var balanceQueryRes nfttypes.QueryBalanceResponse
	requireT.NoError(json.Unmarshal(queryOut, &balanceQueryRes))
	requireT.Equal(uint64(1), balanceQueryRes.Amount)

	// ********** Owner **********

	ownerPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftIDRequest{
		moduleswasm.NftMethodOwner: {
			ID: mintNFTReq2.ID,
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, ownerPayload)
	requireT.NoError(err)
	var ownerQueryRes nfttypes.QueryOwnerResponse
	requireT.NoError(json.Unmarshal(queryOut, &ownerQueryRes))
	requireT.Equal(recipient.String(), ownerQueryRes.Owner)

	// ********** Supply **********

	supplyPayload, err := json.Marshal(map[moduleswasm.NftMethod]struct{}{
		moduleswasm.NftMethodSupply: {},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, supplyPayload)
	requireT.NoError(err)
	var supplyQueryRes nfttypes.QuerySupplyResponse
	requireT.NoError(json.Unmarshal(queryOut, &supplyQueryRes))
	requireT.Equal(uint64(1), supplyQueryRes.Amount)

	// ********** NFT **********

	nftPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftIDRequest{
		moduleswasm.NftMethodNFT: {
			ID: mintNFTReq2.ID,
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, nftPayload)
	requireT.NoError(err)
	var nftQueryRes moduleswasm.NftRes
	requireT.NoError(json.Unmarshal(queryOut, &nftQueryRes))

	requireT.Equal(
		moduleswasm.NftItem{
			ClassID: classID,
			ID:      mintNFTReq2.ID,
			URI:     mintNFTReq2.URI,
			URIHash: mintNFTReq2.URIHash,
			Data:    encodedData,
		}, nftQueryRes.NFT,
	)

	// ********** NFTs **********

	nftsPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftOwnerRequest{
		moduleswasm.NftMethodNFTs: {
			Owner: recipient.String(),
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, nftsPayload)
	requireT.NoError(err)
	var nftsQueryRes moduleswasm.NftsRes
	requireT.NoError(json.Unmarshal(queryOut, &nftsQueryRes))

	requireT.Equal(
		moduleswasm.NftItem{
			ClassID: classID,
			ID:      mintNFTReq2.ID,
			URI:     mintNFTReq2.URI,
			URIHash: mintNFTReq2.URIHash,
			Data:    encodedData,
		}, nftsQueryRes.NFTs[0],
	)

	// ********** Class **********

	nftClassPayload, err := json.Marshal(map[moduleswasm.NftMethod]struct{}{
		moduleswasm.NftMethodClassNFT: {},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, nftClassPayload)
	requireT.NoError(err)
	var nftClassQueryRes moduleswasm.NftClassResponse
	requireT.NoError(json.Unmarshal(queryOut, &nftClassQueryRes))

	requireT.Equal(
		moduleswasm.NftClass{
			ID:          expectedClass.Id,
			Name:        expectedClass.Name,
			Symbol:      expectedClass.Symbol,
			Description: expectedClass.Description,
			URI:         expectedClass.URI,
			URIHash:     expectedClass.URIHash,
			Data:        encodedData,
		}, nftClassQueryRes.Class,
	)

	// ********** Classes **********

	nftClassesPayload, err := json.Marshal(map[moduleswasm.NftMethod]struct{}{
		moduleswasm.NftMethodClassesNFT: {},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, nftClassesPayload)
	requireT.NoError(err)
	var nftClassesQueryRes moduleswasm.NftClassesResponse
	requireT.NoError(json.Unmarshal(queryOut, &nftClassesQueryRes))

	requireT.Contains(nftClassesQueryRes.Classes, moduleswasm.NftClass{
		ID:          expectedClass.Id,
		Name:        expectedClass.Name,
		Symbol:      expectedClass.Symbol,
		Description: expectedClass.Description,
		URI:         expectedClass.URI,
		URIHash:     expectedClass.URIHash,
		Data:        encodedData,
	})
}

// TestWASMContractInstantiationIsNotRejectedIfAccountExists verifies that smart contract instantiation
// is rejected if account exists.
func TestWASMContractInstantiationIsNotRejectedIfAccountExists(t *testing.T) {
	t.Parallel()

	requireT := require.New(t)
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	fundAmount := chain.NewCoin(sdkmath.NewInt(5000000000))
	adminAccount := chain.GenAccount()
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(adminAccount, fundAmount),
	)

	// Deploy smart contract.

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		adminAccount,
		moduleswasm.BankSendWASM,
	)
	requireT.NoError(err)

	testCases := []struct {
		Name        string
		Amount      sdk.Coin
		AccountType string
		MsgFunc     func(adminAccount, contractAddress sdk.AccAddress, amount sdk.Coin) sdk.Msg
	}{
		{
			Name:        "BaseAccount",
			Amount:      chain.NewCoin(sdkmath.NewInt(500)),
			AccountType: "/cosmos.auth.v1beta1.BaseAccount",
			MsgFunc: func(adminAccount, contractAddress sdk.AccAddress, amount sdk.Coin) sdk.Msg {
				return &banktypes.MsgSend{
					FromAddress: adminAccount.String(),
					ToAddress:   contractAddress.String(),
					Amount:      sdk.NewCoins(amount),
				}
			},
		},
		{
			Name:        "DelayedVestingAccount",
			Amount:      chain.NewCoin(sdkmath.NewInt(600)),
			AccountType: "/cosmos.vesting.v1beta1.DelayedVestingAccount",
			MsgFunc: func(adminAccount, contractAddress sdk.AccAddress, amount sdk.Coin) sdk.Msg {
				return &vestingtypes.MsgCreateVestingAccount{
					FromAddress: adminAccount.String(),
					ToAddress:   contractAddress.String(),
					Amount:      sdk.NewCoins(amount),
					EndTime:     time.Now().Unix(),
					Delayed:     true,
				}
			},
		},
		{
			Name:        "ContinuousVestingAccount",
			Amount:      chain.NewCoin(sdkmath.NewInt(700)),
			AccountType: "/cosmos.vesting.v1beta1.ContinuousVestingAccount",
			MsgFunc: func(adminAccount, contractAddress sdk.AccAddress, amount sdk.Coin) sdk.Msg {
				return &vestingtypes.MsgCreateVestingAccount{
					FromAddress: adminAccount.String(),
					ToAddress:   contractAddress.String(),
					Amount:      sdk.NewCoins(amount),
					EndTime:     time.Now().Unix(),
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			requireT := require.New(t)

			adminAccount := chain.GenAccount()
			chain.Faucet.FundAccounts(ctx, t,
				integration.NewFundedAccount(adminAccount, fundAmount),
			)

			salt, err := chain.Wasm.GenerateSalt()
			requireT.NoError(err)

			contractAddress, err := chain.Wasm.PredictWASMContractAddress(
				ctx,
				adminAccount,
				salt,
				codeID,
			)
			requireT.NoError(err)

			msg := tc.MsgFunc(adminAccount, contractAddress, tc.Amount)
			_, err = client.BroadcastTx(
				ctx,
				chain.ClientContext.WithFromAddress(adminAccount),
				chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
				msg,
			)
			requireT.NoError(err)

			testSmartContractAccount(
				ctx,
				t,
				chain,
				codeID,
				salt,
				contractAddress,
				adminAccount,
				tc.Amount,
				tc.AccountType,
			)
		})
	}
}

// TestVestingToWASMContract verifies that smart contract instantiated on top of vesting account
// receives funds correctly.
func TestVestingToWASMContract(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	admin := chain.GenAccount()
	recipient := chain.GenAccount()
	amount := chain.NewCoin(sdkmath.NewInt(500))

	requireT := require.New(t)
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(admin, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

	txf := chain.TxFactory().
		WithSimulateAndExecute(true)

	// Deploy smart contract.

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx,
		txf,
		admin,
		moduleswasm.BankSendWASM,
	)
	requireT.NoError(err)

	// Predict the address of smart contract.

	salt, err := chain.Wasm.GenerateSalt()
	requireT.NoError(err)

	contract, err := chain.Wasm.PredictWASMContractAddress(
		ctx,
		admin,
		salt,
		codeID,
	)
	requireT.NoError(err)

	// Create vesting account using address of the smart contract.

	vestingDuration := 30 * time.Second
	createVestingAccMsg := &vestingtypes.MsgCreateVestingAccount{
		FromAddress: admin.String(),
		ToAddress:   contract.String(),
		Amount:      sdk.NewCoins(amount),
		EndTime:     time.Now().Add(vestingDuration).Unix(),
		Delayed:     true,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(createVestingAccMsg)),
		createVestingAccMsg,
	)
	requireT.NoError(err)

	// Instantiate the smart contract.

	contractAddr, err := chain.Wasm.InstantiateWASMContract2(
		ctx,
		txf,
		admin,
		salt,
		integration.InstantiateConfig{
			CodeID:     codeID,
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    moduleswasm.EmptyPayload,
			Label:      "bank_send",
		},
	)
	requireT.NoError(err)

	// Check that this is still a vesting account.

	authClient := authtypes.NewQueryClient(chain.ClientContext)
	accountRes, err := authClient.Account(ctx, &authtypes.QueryAccountRequest{
		Address: contractAddr,
	})
	requireT.NoError(err)
	requireT.Equal("/cosmos.vesting.v1beta1.DelayedVestingAccount", accountRes.Account.TypeUrl)

	// Verify that funds hasn't been vested yet.

	_, err = chain.Wasm.ExecuteWASMContract(
		ctx,
		txf,
		admin,
		contractAddr,
		moduleswasm.BankSendExecuteWithdrawRequest(amount, recipient),
		sdk.Coin{})
	requireT.ErrorContains(err, "insufficient funds")

	// Await vesting time to unlock the vesting coins

	select {
	case <-ctx.Done():
		return
	case <-time.After(vestingDuration):
	}

	// Verify funds are there.

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	qres, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: contractAddr,
		Denom:   amount.Denom,
	})
	requireT.NoError(err)
	requireT.Equal(amount.String(), qres.Balance.String())

	// Verify that funds has been vested.

	_, err = chain.Wasm.ExecuteWASMContract(
		ctx,
		txf,
		admin,
		contractAddr,
		moduleswasm.BankSendExecuteWithdrawRequest(amount, recipient),
		sdk.Coin{})
	requireT.NoError(err)
}

func testSmartContractAccount(
	ctx context.Context,
	t *testing.T,
	chain integration.CoreumChain,
	codeID uint64,
	salt []byte,
	contractAddress sdk.AccAddress,
	adminAccount sdk.AccAddress,
	expectedAmount sdk.Coin,
	expectedAccountType string,
) {
	requireT := require.New(t)
	clientCtx := chain.ClientContext
	txf := chain.TxFactory().
		WithSimulateAndExecute(true)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	authClient := authtypes.NewQueryClient(chain.ClientContext)

	// Instantiate the smart contract.

	contractAddr, err := chain.Wasm.InstantiateWASMContract2(
		ctx,
		txf,
		adminAccount,
		salt,
		integration.InstantiateConfig{
			CodeID:     codeID,
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    moduleswasm.EmptyPayload,
			Label:      "bank_send",
		},
	)
	requireT.NoError(err)
	requireT.Equal(contractAddr, contractAddress.String())

	// Await next block to ensure that funds are vested (for vesting accounts).

	requireT.NoError(client.AwaitNextBlocks(ctx, clientCtx, 1))

	// Verify balance.

	qres, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: contractAddr,
		Denom:   expectedAmount.Denom,
	})
	requireT.NoError(err)
	requireT.Equal(expectedAmount.String(), qres.Balance.String())

	// Verify account type.

	accountRes, err := authClient.Account(ctx, &authtypes.QueryAccountRequest{
		Address: contractAddr,
	})
	requireT.NoError(err)
	requireT.Equal(expectedAccountType, accountRes.Account.TypeUrl)
}
