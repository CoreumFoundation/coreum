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
	nfttypes "cosmossdk.io/x/nft"
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
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v5/integration-tests"
	moduleswasm "github.com/CoreumFoundation/coreum/v5/integration-tests/contracts/modules"
	"github.com/CoreumFoundation/coreum/v5/pkg/client"
	"github.com/CoreumFoundation/coreum/v5/testutil/event"
	"github.com/CoreumFoundation/coreum/v5/testutil/integration"
	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	assetnfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/nft/types"
	deterministicgastypes "github.com/CoreumFoundation/coreum/v5/x/deterministicgas/types"
	dextypes "github.com/CoreumFoundation/coreum/v5/x/dex/types"
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
type ftDEXSettings struct {
	UnifiedRefAmount  string   `json:"unified_ref_amount"`
	WhitelistedDenoms []string `json:"whitelisted_denoms"`
}

// fungible token wasm models
//
//nolint:tagliatelle
type issueFTRequest struct {
	Symbol             string                               `json:"symbol"`
	Subunit            string                               `json:"subunit"`
	Precision          uint32                               `json:"precision"`
	InitialAmount      string                               `json:"initial_amount"`
	Description        string                               `json:"description"`
	Features           []assetfttypes.Feature               `json:"features"`
	BurnRate           string                               `json:"burn_rate"`
	SendCommissionRate string                               `json:"send_commission_rate"`
	URI                string                               `json:"uri"`
	URIHash            string                               `json:"uri_hash"`
	ExtensionSettings  *assetfttypes.ExtensionIssueSettings `json:"extension_settings"`
	DEXSettings        *ftDEXSettings                       `json:"dex_settings"`
}

// fungible token wasm models
//
//nolint:tagliatelle
type issueFTLegacyRequest struct {
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

type placeOrderBodyDEXRequest struct {
	Order dextypes.MsgPlaceOrder `json:"order"`
}

//nolint:tagliatelle
type cancelOrderBodyDEXRequest struct {
	OrderID string `json:"order_id"`
}

type cancelOrdersByDenomBodyDEXRequest struct {
	Account string `json:"account"`
	Denom   string `json:"denom"`
}

type updateDEXUnifiedRefAmountRequest struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

//nolint:tagliatelle
type updateDEXWhitelistedDenoms struct {
	Denom             string   `json:"denom"`
	WhitelistedDenoms []string `json:"whitelisted_denoms"`
}

//nolint:tagliatelle
type orderBodyDEXRequest struct {
	Account string `json:"acc"`
	OrderID string `json:"order_id"`
}

type ordersBodyDEXRequest struct {
	Creator string `json:"creator"`
}

//nolint:tagliatelle
type orderBookOrdersBodyDEXRequest struct {
	BaseDenom  string        `json:"base_denom"`
	QuoteDenom string        `json:"quote_denom"`
	Side       dextypes.Side `json:"side"`
}

type accountDenomOrdersCountBodyDEXRequest struct {
	Account string `json:"account"`
	Denom   string `json:"denom"`
}

type dexSettingsDEXRequest struct {
	Denom string `json:"denom"`
}

type balanceDEXRequest struct {
	Account string `json:"account"`
	Denom   string `json:"denom"`
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
	ftMethodClawback            ftMethod = "clawback"
	ftMethodTransferAdmin       ftMethod = "transfer_admin"
	ftMethodClearAdmin          ftMethod = "clear_admin"
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

type dexMethod string

const (
	// tx.
	dexMethodPlaceOrder                 dexMethod = "place_order"
	dexMethodCancelOrder                dexMethod = "cancel_order"
	dexMethodCancelOrdersByDenom        dexMethod = "cancel_orders_by_denom"
	dexMethodUpdateDEXUnifiedRefAmount  dexMethod = "update_dex_unified_ref_amount"
	dexMethodUpdateDEXWhitelistedDenoms dexMethod = "update_dex_whitelisted_denoms"
	// query.
	dexMethodParams                  dexMethod = "params"
	dexMethodOrder                   dexMethod = "order"
	dexMethodOrders                  dexMethod = "orders"
	dexMethodOrderBooks              dexMethod = "order_books"
	dexMethodOrderBookOrders         dexMethod = "order_book_orders"
	dexMethodAccountDenomOrdersCount dexMethod = "account_denom_orders_count"
	dexMethodDEXSettings             dexMethod = "dex_settings"
	dexMethodBalance                 dexMethod = "balance"
)

// TestContractInstantiation tests contract instantiation using two instantiation methods.
func TestContractInstantiation(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	admin := chain.GenAccount()

	requireT := require.New(t)
	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Amount: sdkmath.NewInt(2_000_000),
	})

	txf := chain.TxFactoryAuto()

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
	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Amount: sdkmath.NewInt(1_000_000),
	})

	clientCtx := chain.ClientContext
	txf := chain.TxFactoryAuto()
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
		chain.TxFactory().
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
	require.Positive(t, gasEvents[0].RealGas)
}

// TestWASMGasBankSendAndBankSend checks that a message containing a deterministic and a
// non-deterministic transaction takes gas within appropriate limits.
func TestWASMGasBankSendAndBankSend(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	admin := chain.GenAccount()

	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Amount: sdkmath.NewInt(1_000_000),
	})

	// deployWASMContract and init contract with the initial coins amount
	clientCtx := chain.ClientContext
	txf := chain.TxFactoryAuto()

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

	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx, false)
	requireT.NoError(err)
	proposerBalance.Amount = proposerBalance.Amount.MulRaw(2)
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(proposer, proposerBalance),
	)

	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Amount: sdkmath.NewInt(1_000_000),
	})

	// instantiateWASMContract the contract and set the initial counter state.
	initialPayload, err := json.Marshal(moduleswasm.SimpleState{
		Count: 1337,
	})
	requireT.NoError(err)

	txf := chain.TxFactoryAuto()

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
		false,
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
		false,
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

	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Amount: sdkmath.NewInt(1_000_000),
	})
	chain.FundAccountWithOptions(ctx, t, noneAdmin, integration.BalancesOptions{
		Amount: sdkmath.NewInt(500_000),
	})

	// instantiateWASMContract the contract and set the initial counter state.
	initialPayload, err := json.Marshal(moduleswasm.SimpleState{
		Count: 787,
	})
	requireT.NoError(err)

	txf := chain.TxFactoryAuto()

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
	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Amount: sdkmath.NewInt(1_000_000),
	})
	chain.FundAccountWithOptions(ctx, t, newAdmin, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&wasmtypes.MsgClearAdmin{},
		},
	})

	wasmClient := wasmtypes.NewQueryClient(chain.ClientContext)

	// deployWASMContract and init contract with the initial coins amount
	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		chain.TxFactoryAuto(),
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
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgUpdateAdmin)),
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
		chain.TxFactoryAuto(),
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

	chain.FundAccountWithOptions(ctx, t, granter, integration.BalancesOptions{
		Amount: sdkmath.NewInt(2_000_000),
	})

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
		chain.TxFactoryAuto(),
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
		chain.TxFactoryAuto(),
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
		chain.TxFactoryAuto(),
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
		chain.TxFactoryAuto(),
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
		chain.TxFactoryAuto(),
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

	chain.FundAccountWithOptions(ctx, t, receiver, integration.BalancesOptions{
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.
			Add(sdkmath.NewInt(500_000)),
	})

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
		ctx, chain.TxFactoryAuto(), granter, contractAddrAuthzNftTrade, nftOfferPayload, sdk.Coin{},
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
		chain.TxFactoryAuto(),
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

// TestWASMFungibleTokenInContract verifies that smart contract is able to execute all Coreum fungible token messages
// and queries.
func TestWASMFungibleTokenInContract(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	admin := chain.GenAccount()
	recipient1 := chain.GenAccount()
	recipient2 := chain.GenAccount()

	requireT := require.New(t)
	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.
			Add(sdkmath.NewInt(4_000_000)),
	})

	clientCtx := chain.ClientContext
	txf := chain.TxFactoryAuto()
	bankClient := banktypes.NewQueryClient(clientCtx)
	ftClient := assetfttypes.NewQueryClient(clientCtx)

	// ********** Issuance **********

	burnRate := "1000000000000000000"           // LegacyDec has 18 decimal positions, so here we are passing 1e19= 100%
	sendCommissionRate := "1000000000000000000" // LegacyDec has 18 decimal positions, so here we are passing 1e19 = 100%

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
			assetfttypes.Feature_clawback,
		},
		BurnRate:           burnRate,
		SendCommissionRate: sendCommissionRate,
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
			assetfttypes.Feature_clawback,
		},
		BurnRate:           sdkmath.LegacyMustNewDecFromStr("1"),
		SendCommissionRate: sdkmath.LegacyMustNewDecFromStr("1"),
		Version:            assetfttypes.CurrentTokenVersion, // test should work with any token version
		URI:                issuanceReq.URI,
		URIHash:            issuanceReq.URIHash,
		Admin:              contractAddr,
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

	// ********** Clawback **********

	amountToClawback := sdkmath.NewInt(10)
	clawbackPayload, err := json.Marshal(map[ftMethod]accountAmountBodyFTRequest{
		ftMethodClawback: {
			Account: recipient2.String(),
			Amount:  amountToClawback.String(),
		},
	})
	requireT.NoError(err)

	balanceBeforeClawbackRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient2.String(),
		Denom:   denom,
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, clawbackPayload, sdk.Coin{})
	requireT.NoError(err)

	balanceAfterClawbackRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient2.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	clawBackedAmount := balanceBeforeClawbackRes.Balance.Amount.Sub(balanceAfterClawbackRes.Balance.Amount)
	requireT.Equal(amountToClawback.String(), clawBackedAmount.String())

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

	// ********** Transfer Admin **********

	transferAdminPayload, err := json.Marshal(map[ftMethod]accountBodyFTRequest{
		ftMethodTransferAdmin: {
			Account: recipient1.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, transferAdminPayload, sdk.Coin{})
	requireT.NoError(err)

	tokenRes, err = ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: denom,
	})
	requireT.NoError(err)
	requireT.Equal(tokenRes.Token.Admin, recipient1.String())

	// transfer it back to the contract address to test clearing the admin
	transferAdminMsg := &assetfttypes.MsgTransferAdmin{
		Sender:  recipient1.String(),
		Account: contractAddr,
		Denom:   denom,
	}

	chain.FundAccountWithOptions(ctx, t, recipient1, integration.BalancesOptions{
		Messages: []sdk.Msg{
			transferAdminMsg,
		},
	})

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(transferAdminMsg)),
		transferAdminMsg,
	)
	requireT.NoError(err)

	tokenRes, err = ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: denom,
	})
	requireT.NoError(err)
	requireT.Equal(tokenRes.Token.Admin, contractAddr)

	// ********** Clear Admin **********

	clearAdminPayload, err := json.Marshal(map[ftMethod]struct{}{
		ftMethodClearAdmin: {},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, clearAdminPayload, sdk.Coin{})
	requireT.NoError(err)

	tokenRes, err = ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: denom,
	})
	requireT.NoError(err)
	requireT.Empty(tokenRes.Token.Admin)

	// TODO: Once we upgrade to SDK v0.50 and have the new GRPCQueries, we can test the queries here as well.
}

// TestWASMFungibleTokenInContractLegacy verifies that smart contract is able to execute all
// Coreum fungible token messages and queries using the deprecated wasm bindings/handler.
func TestWASMFungibleTokenInContractLegacy(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	admin := chain.GenAccount()
	recipient1 := chain.GenAccount()
	recipient2 := chain.GenAccount()

	requireT := require.New(t)
	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.
			Add(sdkmath.NewInt(1_000_000)),
	})

	clientCtx := chain.ClientContext
	txf := chain.TxFactoryAuto()
	bankClient := banktypes.NewQueryClient(clientCtx)
	ftClient := assetfttypes.NewQueryClient(clientCtx)

	// ********** Issuance **********

	burnRate := sdkmath.LegacyMustNewDecFromStr("0.1")
	sendCommissionRate := sdkmath.LegacyMustNewDecFromStr("0.2")

	issuanceAmount := sdkmath.NewInt(10_000)
	issuanceReq := issueFTLegacyRequest{
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
		moduleswasm.FTLegacyWASM,
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
		Admin:              contractAddr,
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

	paramsPayload, err := json.Marshal(map[ftMethod]struct{}{
		ftMethodParams: {},
	})
	requireT.NoError(err)
	queryOut, err := chain.Wasm.QueryWASMContract(ctx, contractAddr, paramsPayload)
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

// TestWASMNonFungibleTokenInContract verifies that smart contract is able to execute all non-fungible Coreum
// token messages and queries from smart contracts.
//
//nolint:nosnakecase
func TestWASMNonFungibleTokenInContract(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	admin := chain.GenAccount()
	recipient := chain.GenAccount()
	mintRecipient := chain.GenAccount()

	requireT := require.New(t)
	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Amount: sdkmath.NewInt(4_000_000),
	})

	clientCtx := chain.ClientContext
	txf := chain.TxFactoryAuto()
	assetNftClient := assetnfttypes.NewQueryClient(clientCtx)
	nftClient := nfttypes.NewQueryClient(clientCtx)

	// ********** Issuance **********

	royaltyRate := "100000000000000000" // 1e18 = 10%
	data := make([]byte, 256)
	for i := range 256 {
		data[i] = uint8(i)
	}
	encodedData := base64.StdEncoding.EncodeToString(data)

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
		RoyaltyRate: royaltyRate,
	}

	issuerNFTInstantiatePayload, err := json.Marshal(issueClassReqNoWhitelist)
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
		issueClassReqNoWhitelist.Symbol, sdk.MustAccAddressFromBech32(contractAddrNoWhitelist),
	)

	dataBytes, err := codectypes.NewAnyWithValue(&assetnfttypes.DataBytes{Data: data})
	// we need to do this, otherwise assertion fails because some private fields are set differently
	requireT.NoError(err)
	dataToCompare := &codectypes.Any{
		TypeUrl: dataBytes.TypeUrl,
		Value:   dataBytes.Value,
	}

	classID := assetnfttypes.BuildClassID(
		issueClassReqNoWhitelist.Symbol,
		sdk.MustAccAddressFromBech32(contractAddrNoWhitelist),
	)
	classRes, err := assetNftClient.Class(ctx, &assetnfttypes.QueryClassRequest{Id: classID})
	requireT.NoError(err)

	expectedClass := assetnfttypes.Class{
		Id:          classIDNoWhitelist,
		Issuer:      contractAddrNoWhitelist,
		Name:        issueClassReqNoWhitelist.Name,
		Symbol:      issueClassReqNoWhitelist.Symbol,
		Description: issueClassReqNoWhitelist.Description,
		URI:         issueClassReqNoWhitelist.URI,
		URIHash:     issueClassReqNoWhitelist.URIHash,
		Data:        dataToCompare,
		Features:    issueClassReqNoWhitelist.Features,
		RoyaltyRate: sdkmath.LegacyMustNewDecFromStr("0.1"),
	}
	requireT.Equal(
		expectedClass, classRes.Class,
	)

	// ********** Mint **********

	// Mint an immutable NFT using protos instead of using the deprecated handler
	mintImmutableNFTReq := moduleswasm.NftMintRequest{
		ID:        "id-1",
		Recipient: mintRecipient.String(),
	}
	mintImmutablePayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftMintRequest{
		moduleswasm.NftMethodMintImmutable: mintImmutableNFTReq,
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddrNoWhitelist, mintImmutablePayload, sdk.Coin{})
	requireT.NoError(err)

	nftResp, err := nftClient.NFT(ctx, &nfttypes.QueryNFTRequest{
		ClassId: classIDNoWhitelist,
		Id:      mintImmutableNFTReq.ID,
	})
	requireT.NoError(err)

	emptyData, err := codectypes.NewAnyWithValue(&assetnfttypes.DataBytes{Data: nil})
	requireT.NoError(err)
	expectedNFT := &nfttypes.NFT{
		ClassId: classIDNoWhitelist,
		Id:      mintImmutableNFTReq.ID,
		Data:    emptyData,
	}

	gotNFT := nftResp.Nft
	// encode the data `from` and `to` proto.Any to load same state as `codectypes.NewAnyWithValue` does
	var gotNFTDataBytes assetnfttypes.DataBytes
	requireT.NoError(gotNFTDataBytes.Unmarshal(gotNFT.Data.Value))
	gotNFTData, err := codectypes.NewAnyWithValue(&gotNFTDataBytes)
	requireT.NoError(err)
	gotNFT.Data = gotNFTData

	requireT.Equal(
		expectedNFT, gotNFT,
	)

	nftOwner, err := nftClient.Owner(ctx, &nfttypes.QueryOwnerRequest{
		ClassId: classIDNoWhitelist,
		Id:      mintImmutableNFTReq.ID,
	})
	requireT.NoError(err)
	requireT.Equal(nftOwner.Owner, mintRecipient.String())

	// Mint a mutable NFT with both Owner and Issuer as editors
	encodedMutableData := base64.StdEncoding.EncodeToString([]byte("mutable_data"))
	mintMutableNFTReq := moduleswasm.NftMintRequest{
		ID:        "id-mut",
		Recipient: mintRecipient.String(),
		Data:      encodedMutableData,
	}

	mintMutablePayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftMintRequest{
		moduleswasm.NftMethodMintMutable: mintMutableNFTReq,
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddrNoWhitelist, mintMutablePayload, sdk.Coin{})
	requireT.NoError(err)

	nftResp, err = nftClient.NFT(ctx, &nfttypes.QueryNFTRequest{
		ClassId: classIDNoWhitelist,
		Id:      mintMutableNFTReq.ID,
	})
	requireT.NoError(err)

	nftOwner, err = nftClient.Owner(ctx, &nfttypes.QueryOwnerRequest{
		ClassId: classIDNoWhitelist,
		Id:      mintMutableNFTReq.ID,
	})
	requireT.NoError(err)
	requireT.Equal(nftOwner.Owner, mintRecipient.String())

	// Check the data of the mutable NFT
	dataBytes, err = codectypes.NewAnyWithValue(&assetnfttypes.DataDynamic{Items: []assetnfttypes.DataDynamicItem{{
		// both admin and owner
		Editors: []assetnfttypes.DataEditor{
			assetnfttypes.DataEditor_admin,
			assetnfttypes.DataEditor_owner,
		},
		Data: []byte("mutable_data"),
	}}})

	requireT.NoError(err)
	dataToCompare = &codectypes.Any{
		TypeUrl: dataBytes.TypeUrl,
		Value:   dataBytes.Value,
	}

	requireT.Equal(dataToCompare, nftResp.Nft.Data)

	// Let's check that both admin (contract) and owner (mintRecipient) can edit the data
	// First, let's try to edit the data as the admin
	encodedEditedData := base64.StdEncoding.EncodeToString([]byte("edited_data_by_admin"))
	modifyDataReq := moduleswasm.NftModifyDataRequest{
		ID:   mintMutableNFTReq.ID,
		Data: encodedEditedData,
	}

	modifyDataPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftModifyDataRequest{
		moduleswasm.NftMethodModifyData: modifyDataReq,
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddrNoWhitelist, modifyDataPayload, sdk.Coin{})
	requireT.NoError(err)

	dataBytes, err = codectypes.NewAnyWithValue(&assetnfttypes.DataDynamic{Items: []assetnfttypes.DataDynamicItem{{
		// both admin and owner
		Editors: []assetnfttypes.DataEditor{
			assetnfttypes.DataEditor_admin,
			assetnfttypes.DataEditor_owner,
		},
		Data: []byte("edited_data_by_admin"),
	}}})

	requireT.NoError(err)
	dataToCompare = &codectypes.Any{
		TypeUrl: dataBytes.TypeUrl,
		Value:   dataBytes.Value,
	}

	nftResp, err = nftClient.NFT(ctx, &nfttypes.QueryNFTRequest{
		ClassId: classIDNoWhitelist,
		Id:      mintMutableNFTReq.ID,
	})
	requireT.NoError(err)

	requireT.Equal(dataToCompare, nftResp.Nft.Data)

	// Let's edit the data directly by the owner now and see that he can also update
	msgUpdateData := &assetnfttypes.MsgUpdateData{
		Sender:  mintRecipient.String(),
		ClassID: classIDNoWhitelist,
		ID:      mintMutableNFTReq.ID,
		Items: []assetnfttypes.DataDynamicIndexedItem{
			{
				Index: 0,
				Data:  []byte("edited_data_by_owner"),
			},
		},
	}
	chain.FundAccountWithOptions(ctx, t, mintRecipient, integration.BalancesOptions{
		Messages: []sdk.Msg{
			msgUpdateData,
		},
	})

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(mintRecipient),
		txf,
		msgUpdateData,
	)
	requireT.NoError(err)

	dataBytes, err = codectypes.NewAnyWithValue(&assetnfttypes.DataDynamic{Items: []assetnfttypes.DataDynamicItem{{
		// both admin and owner
		Editors: []assetnfttypes.DataEditor{
			assetnfttypes.DataEditor_admin,
			assetnfttypes.DataEditor_owner,
		},
		Data: []byte("edited_data_by_owner"),
	}}})

	requireT.NoError(err)
	dataToCompare = &codectypes.Any{
		TypeUrl: dataBytes.TypeUrl,
		Value:   dataBytes.Value,
	}

	nftResp, err = nftClient.NFT(ctx, &nfttypes.QueryNFTRequest{
		ClassId: classIDNoWhitelist,
		Id:      mintMutableNFTReq.ID,
	})
	requireT.NoError(err)

	requireT.Equal(dataToCompare, nftResp.Nft.Data)

	// ********** Freeze **********

	freezePayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftIDRequest{
		moduleswasm.NftMethodFreeze: {
			ID: mintImmutableNFTReq.ID,
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddrNoWhitelist, freezePayload, sdk.Coin{})
	requireT.NoError(err)

	assertNftFrozenRes, err := assetNftClient.Frozen(ctx, &assetnfttypes.QueryFrozenRequest{
		Id:      mintImmutableNFTReq.ID,
		ClassId: classID,
	})
	requireT.NoError(err)
	requireT.True(assertNftFrozenRes.Frozen)

	// ********** Unfreeze **********

	unfreezePayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftIDRequest{
		moduleswasm.NftMethodUnfreeze: {
			ID: mintImmutableNFTReq.ID,
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddrNoWhitelist, unfreezePayload, sdk.Coin{})
	requireT.NoError(err)

	assertNftFrozenRes, err = assetNftClient.Frozen(ctx, &assetnfttypes.QueryFrozenRequest{
		Id:      mintImmutableNFTReq.ID,
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

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddrNoWhitelist, classFreezePayload, sdk.Coin{})
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

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddrNoWhitelist, classUnfreezePayload, sdk.Coin{})
	requireT.NoError(err)

	assertNftClassFrozenRes, err = assetNftClient.ClassFrozen(ctx, &assetnfttypes.QueryClassFrozenRequest{
		ClassId: classID,
		Account: recipient.String(),
	})
	requireT.NoError(err)
	requireT.False(assertNftClassFrozenRes.Frozen)

	// ********** AddToWhitelist **********

	// Let's issue a class first with feature whitelist enabled and mint and NFT

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
		RoyaltyRate: royaltyRate,
	}
	issuerNFTInstantiatePayload, err = json.Marshal(issueClassReq)
	requireT.NoError(err)

	// instantiate new contract
	contractAddrWhitelist, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
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

	classID = assetnfttypes.BuildClassID(issueClassReq.Symbol, sdk.MustAccAddressFromBech32(contractAddrWhitelist))
	requireT.NoError(err)

	mintNFTReq1 := moduleswasm.NftMintRequest{
		ID:      "id-1",
		URI:     "https://my-nft-meta.invalid/1",
		URIHash: "hash",
		Data:    encodedData,
	}
	mintPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftMintRequest{
		moduleswasm.NftMethodMintImmutable: mintNFTReq1,
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddrWhitelist, mintPayload, sdk.Coin{})
	requireT.NoError(err)

	nftResp, err = nftClient.NFT(ctx, &nfttypes.QueryNFTRequest{
		ClassId: classID,
		Id:      mintNFTReq1.ID,
	})
	requireT.NoError(err)

	dataBytes, err = codectypes.NewAnyWithValue(&assetnfttypes.DataBytes{Data: data})
	requireT.NoError(err)
	dataToCompare = &codectypes.Any{
		TypeUrl: dataBytes.TypeUrl,
		Value:   dataBytes.Value,
	}

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

	nftOwner, err = nftClient.Owner(ctx, &nfttypes.QueryOwnerRequest{
		ClassId: classID,
		Id:      mintNFTReq1.ID,
	})
	requireT.NoError(err)
	requireT.Equal(nftOwner.Owner, contractAddrWhitelist)

	addToWhitelistPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftIDWithAccountRequest{
		moduleswasm.NftMethodAddToWhitelist: {
			ID:      mintNFTReq1.ID,
			Account: recipient.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddrWhitelist, addToWhitelistPayload, sdk.Coin{})
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

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddrWhitelist, removeFromWhitelistPayload, sdk.Coin{})
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

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddrWhitelist, addToClassWhitelistPayload, sdk.Coin{})
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

	_, err = chain.Wasm.ExecuteWASMContract(
		ctx,
		txf,
		admin,
		contractAddrWhitelist,
		removeFromClassWhitelistPayload,
		sdk.Coin{},
	)

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

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddrWhitelist, burnPayload, sdk.Coin{})
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
	mintNFTReq2.ID = "id-send"
	mint2Payload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftMintRequest{
		moduleswasm.NftMethodMintImmutable: mintNFTReq2,
	})
	requireT.NoError(err)
	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddrWhitelist, mint2Payload, sdk.Coin{})
	requireT.NoError(err)

	// addToWhitelistPayload
	addToWhitelistPayload, err = json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftIDWithAccountRequest{
		moduleswasm.NftMethodAddToWhitelist: {
			ID:      mintNFTReq2.ID,
			Account: recipient.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddrWhitelist, addToWhitelistPayload, sdk.Coin{})
	requireT.NoError(err)

	// send
	sendPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftIDWithReceiverRequest{
		moduleswasm.NftMethodSend: {
			ID:       mintNFTReq2.ID,
			Receiver: recipient.String(),
		},
	})
	requireT.NoError(err)
	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddrWhitelist, sendPayload, sdk.Coin{})
	requireT.NoError(err)

	// TODO: Once we upgrade to SDK v0.50 and have the new GRPCQueries, we can test the queries here as well.
}

// TestWASMNonFungibleTokenInContractLegacy verifies that smart contract is able to execute all
// non-fungible Coreum token messages and queries from the deprecated wasm bindings/handler.
//
//nolint:nosnakecase
func TestWASMNonFungibleTokenInContractLegacy(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	admin := chain.GenAccount()
	recipient := chain.GenAccount()
	mintRecipient := chain.GenAccount()

	requireT := require.New(t)
	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Amount: sdkmath.NewInt(4_000_000),
	})

	clientCtx := chain.ClientContext
	txf := chain.TxFactoryAuto()
	assetNftClient := assetnfttypes.NewQueryClient(clientCtx)
	nftClient := nfttypes.NewQueryClient(clientCtx)

	// ********** Issuance **********

	royaltyRate := sdkmath.LegacyMustNewDecFromStr("0.1")
	data := make([]byte, 256)
	for i := range 256 {
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
		moduleswasm.NFTLegacyWASM,
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
	requireT.NoError(err)
	dataToCompare := &codectypes.Any{
		TypeUrl: dataBytes.TypeUrl,
		Value:   dataBytes.Value,
	}

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
		moduleswasm.NFTLegacyWASM,
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
	mintNFTReq2.ID = "id-send"
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

	paramsPayload, err := json.Marshal(map[moduleswasm.NftMethod]struct{}{
		moduleswasm.NftMethodParams: {},
	})
	requireT.NoError(err)
	queryOut, err := chain.Wasm.QueryWASMContract(ctx, contractAddr, paramsPayload)
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

	classFrozenAccountsPayload, err := json.Marshal(map[moduleswasm.NftMethod]struct{}{
		moduleswasm.NftMethodClassFrozenAccounts: {},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, classFrozenAccountsPayload)
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

	classWhitelistedAccountsPayload, err := json.Marshal(map[moduleswasm.NftMethod]struct{}{
		moduleswasm.NftMethodClassWhitelistedAccounts: {},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, classWhitelistedAccountsPayload)
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

	// Let's issue a class and mint an NFT with DataDynamic and check that it can also be queried from the contract
	issueMsg := &assetnfttypes.MsgIssueClass{
		Issuer: admin.String(),
		Symbol: "symbol",
		Name:   "name",
		Data:   nil,
	}

	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Messages: []sdk.Msg{
			issueMsg,
		},
	})

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)

	classIDDynamic := assetnfttypes.BuildClassID(issueMsg.Symbol, admin)

	jsonData := []byte(`{"name": "Name", "description": "Description"}`)
	dataDynamic := assetnfttypes.DataDynamic{
		Items: []assetnfttypes.DataDynamicItem{
			{
				Editors: []assetnfttypes.DataEditor{
					assetnfttypes.DataEditor_owner,
				},
				Data: jsonData,
			},
		},
	}
	dataD, err := codectypes.NewAnyWithValue(&dataDynamic)
	requireT.NoError(err)

	mintMsg := &assetnfttypes.MsgMint{
		Sender:    admin.String(),
		Recipient: admin.String(),
		ID:        "id-dynamic",
		ClassID:   classIDDynamic,
		Data:      dataD,
	}

	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Messages: []sdk.Msg{
			mintMsg,
		},
	})

	txRes, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.NoError(err)
	requireT.EqualValues(txRes.GasUsed, chain.GasLimitByMsgs(mintMsg))

	externalNFTPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftClassIDWithIDRequest{
		moduleswasm.NftMethodExternalNFT: {
			ClassID: classIDDynamic,
			ID:      "id-dynamic",
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, externalNFTPayload)
	requireT.NoError(err)
	var externalNFTQueryRes moduleswasm.NftRes
	requireT.NoError(json.Unmarshal(queryOut, &externalNFTQueryRes))

	dataDynamicBytes, err := dataDynamic.Marshal()
	requireT.NoError(err)
	dataDynamicToCompare := base64.StdEncoding.EncodeToString(dataDynamicBytes)

	requireT.Equal(
		moduleswasm.NftItem{
			ClassID: classIDDynamic,
			ID:      "id-dynamic",
			Data:    dataDynamicToCompare,
		}, externalNFTQueryRes.NFT,
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

	adminAccount := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, adminAccount, integration.BalancesOptions{
		Amount: sdkmath.NewInt(1_000_000),
	})

	// Deploy smart contract.

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx,
		chain.TxFactoryAuto(),
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
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			requireT := require.New(t)

			adminAccount := chain.GenAccount()
			chain.FundAccountWithOptions(ctx, t, adminAccount, integration.BalancesOptions{
				Amount: sdkmath.NewInt(1_000_000),
			})

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
	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Amount: sdkmath.NewInt(1_000_000),
	})

	txf := chain.TxFactoryAuto()

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

// TestWASMDEXInContract verifies that smart contract is able to execute all Coreum DEX messages and queries.
func TestWASMDEXInContract(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	admin := chain.GenAccount()
	issuer := chain.GenAccount()

	requireT := require.New(t)
	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.
			Add(sdkmath.NewInt(1_000_000)),
	})

	clientCtx := chain.ClientContext
	txf := chain.TxFactoryAuto()
	bankClient := banktypes.NewQueryClient(clientCtx)

	dexParms := chain.QueryDEXParams(ctx, t)

	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	// Issue a normal fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC1",
		Subunit:       "abc1",
		Precision:     6,
		InitialAmount: sdkmath.NewInt(10_000_000),
		Description:   "ABC1 Description",
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_dex_whitelisted_denoms,
			assetfttypes.Feature_dex_order_cancellation,
			assetfttypes.Feature_dex_unified_ref_amount_change,
		},
		BurnRate:           sdkmath.LegacyNewDec(0),
		SendCommissionRate: sdkmath.LegacyNewDec(0),
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	denom1 := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)

	issuanceAmount := sdkmath.NewInt(10_000_000)
	issuanceReq := issueFTRequest{
		Symbol:        "ABC2",
		Subunit:       "abc2",
		Precision:     6,
		InitialAmount: issuanceAmount.String(),
		Description:   "ABC2 Description",
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_dex_whitelisted_denoms,
			assetfttypes.Feature_dex_order_cancellation,
			assetfttypes.Feature_dex_unified_ref_amount_change,
		},
		BurnRate:           sdkmath.LegacyMustNewDecFromStr("1").BigInt().String(),
		SendCommissionRate: sdkmath.LegacyMustNewDecFromStr("1").BigInt().String(),
		URI:                "https://example.com",
		URIHash:            "1234567890abcdef",
		DEXSettings: &ftDEXSettings{
			UnifiedRefAmount:  sdkmath.LegacyMustNewDecFromStr("150").BigInt().String(),
			WhitelistedDenoms: []string{denom1},
		},
	}
	issuerFTInstantiatePayload, err := json.Marshal(issuanceReq)
	requireT.NoError(err)

	// instantiate new contract
	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		txf,
		admin,
		moduleswasm.DEXWASM,
		integration.InstantiateConfig{
			Amount:     chain.QueryAssetFTParams(ctx, t).IssueFee,
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    issuerFTInstantiatePayload,
			Label:      "dex",
		},
	)
	requireT.NoError(err)

	denom2 := assetfttypes.BuildDenom(issuanceReq.Subunit, sdk.MustAccAddressFromBech32(contractAddr))

	// send some denom1 coin to the contract
	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   contractAddr,
		Amount:      sdk.NewCoins(sdk.NewCoin(denom1, sdkmath.NewInt(10_000_000))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	balanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: contractAddr,
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sendMsg.Amount.AmountOf(denom1).String(), balanceRes.Balance.Amount.String())

	chain.FundAccountWithOptions(ctx, t, sdk.MustAccAddressFromBech32(contractAddr), integration.BalancesOptions{
		Amount: dexParms.OrderReserve.Amount,
	})

	latestBlock, err := chain.LatestBlockHeader(ctx)
	requireT.NoError(err)

	// ********** Query params **********

	paramsPayload, err := json.Marshal(map[dexMethod]struct{}{
		dexMethodParams: {},
	})
	requireT.NoError(err)
	queryOut, err := chain.Wasm.QueryWASMContract(ctx, contractAddr, paramsPayload)
	requireT.NoError(err)
	var wasmParamsRes dextypes.QueryParamsResponse
	requireT.NoError(json.Unmarshal(queryOut, &wasmParamsRes))
	requireT.Equal(
		dexParms.OrderReserve.Amount.String(), wasmParamsRes.Params.OrderReserve.Amount.String(),
	)
	requireT.Equal(
		dexParms.MaxOrdersPerDenom, wasmParamsRes.Params.MaxOrdersPerDenom,
	)
	requireT.Equal(
		dexParms.PriceTickExponent, wasmParamsRes.Params.PriceTickExponent,
	)
	// TODO: Uncomment after proto & wasm-sdk merge.
	// requireT.Equal(
	// 	dexParms.QuantityStepExponent, wasmParamsRes.Params.QuantityStepExponent,
	// )

	// ********** Query and update asset FT DEX settings **********

	dexSettingsPayload, err := json.Marshal(map[dexMethod]dexSettingsDEXRequest{
		dexMethodDEXSettings: {
			denom2,
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, dexSettingsPayload)
	requireT.NoError(err)
	//nolint:tagliatelle
	var dexSettingsRes struct {
		DEXSettings ftDEXSettings `json:"dex_settings"`
	}
	requireT.NoError(json.Unmarshal(queryOut, &dexSettingsRes))
	requireT.Equal(*issuanceReq.DEXSettings, dexSettingsRes.DEXSettings)

	newUnifiedRefAmount := sdkmath.LegacyMustNewDecFromStr("19000")
	updateDEXUnifiedRefAmountPayload, err := json.Marshal(map[dexMethod]updateDEXUnifiedRefAmountRequest{
		dexMethodUpdateDEXUnifiedRefAmount: {
			Denom:  denom2,
			Amount: newUnifiedRefAmount.BigInt().String(),
		},
	})
	requireT.NoError(err)
	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, updateDEXUnifiedRefAmountPayload, sdk.Coin{})
	requireT.NoError(err)

	newWhitelistedDenoms := []string{denom1, chain.ChainSettings.Denom}
	updateDEXWhitelistedDenomsPayload, err := json.Marshal(map[dexMethod]updateDEXWhitelistedDenoms{
		dexMethodUpdateDEXWhitelistedDenoms: {
			Denom:             denom2,
			WhitelistedDenoms: newWhitelistedDenoms,
		},
	})
	requireT.NoError(err)
	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, updateDEXWhitelistedDenomsPayload, sdk.Coin{})
	requireT.NoError(err)

	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, dexSettingsPayload)
	requireT.NoError(err)
	requireT.NoError(json.Unmarshal(queryOut, &dexSettingsRes))
	requireT.Equal(
		ftDEXSettings{
			UnifiedRefAmount:  newUnifiedRefAmount.BigInt().String(),
			WhitelistedDenoms: newWhitelistedDenoms,
		}, dexSettingsRes.DEXSettings,
	)

	// ********** Place Order **********

	orderQuantity := sdkmath.NewInt(100)
	placeOrderPayload, err := json.Marshal(map[dexMethod]placeOrderBodyDEXRequest{
		dexMethodPlaceOrder: {
			Order: dextypes.MsgPlaceOrder{
				Sender:     contractAddr,
				Type:       dextypes.ORDER_TYPE_LIMIT,
				ID:         "id1",
				BaseDenom:  denom1,
				QuoteDenom: denom2,
				Price:      lo.ToPtr(dextypes.MustNewPriceFromString("999")),
				Quantity:   orderQuantity,
				Side:       dextypes.SIDE_SELL,
				GoodTil: &dextypes.GoodTil{
					GoodTilBlockHeight: uint64(latestBlock.Height + 500),
				},
				TimeInForce: dextypes.TIME_IN_FORCE_GTC,
			},
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, placeOrderPayload, sdk.Coin{})
	requireT.NoError(err)

	balancePayload, err := json.Marshal(map[dexMethod]balanceDEXRequest{
		dexMethodBalance: {
			Account: contractAddr,
			Denom:   denom1,
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, balancePayload)
	requireT.NoError(err)
	var ftBalanceRes assetfttypes.QueryBalanceResponse
	requireT.NoError(json.Unmarshal(queryOut, &ftBalanceRes))
	requireT.Equal(orderQuantity.String(), ftBalanceRes.LockedInDEX.String())

	// ********** Query Order **********

	orderPayload, err := json.Marshal(map[dexMethod]orderBodyDEXRequest{
		dexMethodOrder: {
			Account: contractAddr,
			OrderID: "id1",
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, orderPayload)
	requireT.NoError(err)
	var wasmOrderRes dextypes.QueryOrderResponse
	requireT.NoError(json.Unmarshal(queryOut, &wasmOrderRes))

	expectedOrder := dextypes.Order{
		Creator:    contractAddr,
		Type:       dextypes.ORDER_TYPE_LIMIT,
		ID:         "id1",
		BaseDenom:  denom1,
		QuoteDenom: denom2,
		Sequence:   wasmOrderRes.Order.Sequence,
		Price:      lo.ToPtr(dextypes.MustNewPriceFromString("999")),
		Quantity:   orderQuantity,
		Side:       dextypes.SIDE_SELL,
		GoodTil: &dextypes.GoodTil{
			GoodTilBlockHeight: uint64(latestBlock.Height + 500),
		},
		TimeInForce:               dextypes.TIME_IN_FORCE_GTC,
		RemainingBaseQuantity:     orderQuantity,
		RemainingSpendableBalance: orderQuantity,
		Reserve:                   dexParms.OrderReserve,
	}
	requireT.Equal(expectedOrder, wasmOrderRes.Order)

	// ********** Query Orders **********

	ordersPayload, err := json.Marshal(map[dexMethod]ordersBodyDEXRequest{
		dexMethodOrders: {
			Creator: contractAddr,
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, ordersPayload)
	requireT.NoError(err)
	var wasmOrdersRes dextypes.QueryOrdersResponse
	requireT.NoError(json.Unmarshal(queryOut, &wasmOrdersRes))
	requireT.Len(wasmOrdersRes.Orders, 1)
	requireT.Equal(expectedOrder, wasmOrdersRes.Orders[0])

	// ********** Query Order Books **********

	orderBooksPayload, err := json.Marshal(map[dexMethod]struct{}{
		dexMethodOrderBooks: {},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, orderBooksPayload)
	requireT.NoError(err)
	var wasmOrderBooksRes dextypes.QueryOrderBooksResponse
	requireT.NoError(json.Unmarshal(queryOut, &wasmOrderBooksRes))
	requireT.Contains(wasmOrderBooksRes.OrderBooks, dextypes.OrderBookData{
		BaseDenom:  denom1,
		QuoteDenom: denom2,
	})
	requireT.Contains(wasmOrderBooksRes.OrderBooks, dextypes.OrderBookData{
		BaseDenom:  denom2,
		QuoteDenom: denom1,
	})

	// ********** Query Order Book Orders **********

	orderBookOrdersPayload, err := json.Marshal(map[dexMethod]orderBookOrdersBodyDEXRequest{
		dexMethodOrderBookOrders: {
			BaseDenom:  denom1,
			QuoteDenom: denom2,
			Side:       dextypes.SIDE_SELL,
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, orderBookOrdersPayload)
	requireT.NoError(err)
	var wasmOrderBookOrdersRes dextypes.QueryOrderBookOrdersResponse
	requireT.NoError(json.Unmarshal(queryOut, &wasmOrderBookOrdersRes))
	requireT.Len(wasmOrderBookOrdersRes.Orders, 1)
	requireT.Equal(
		expectedOrder, wasmOrderBookOrdersRes.Orders[0],
	)

	// ********** Query Account Denom Orders Count **********

	accountDenomOrdersCountPayload, err := json.Marshal(map[dexMethod]accountDenomOrdersCountBodyDEXRequest{
		dexMethodAccountDenomOrdersCount: {
			Account: contractAddr,
			Denom:   denom1,
		},
	})
	requireT.NoError(err)
	queryOut, err = chain.Wasm.QueryWASMContract(ctx, contractAddr, accountDenomOrdersCountPayload)
	requireT.NoError(err)
	var wasmAccountDenomOrdersCountRes dextypes.QueryAccountDenomOrdersCountResponse
	requireT.NoError(json.Unmarshal(queryOut, &wasmAccountDenomOrdersCountRes))
	requireT.Equal(uint64(1), wasmAccountDenomOrdersCountRes.Count)

	// ********** Cancel Order **********

	cancelOrderPayload, err := json.Marshal(map[dexMethod]cancelOrderBodyDEXRequest{
		dexMethodCancelOrder: {
			OrderID: "id1",
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, cancelOrderPayload, sdk.Coin{})
	requireT.NoError(err)

	// ********** Cancel Orders By Denom **********

	cancelOrdersByDenomPayload, err := json.Marshal(map[dexMethod]cancelOrdersByDenomBodyDEXRequest{
		dexMethodCancelOrdersByDenom: {
			Account: contractAddr,
			Denom:   denom2,
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, cancelOrdersByDenomPayload, sdk.Coin{})
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
	txf := chain.TxFactoryAuto()
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
