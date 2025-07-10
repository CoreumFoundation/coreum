package integration

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	sdkmath "cosmossdk.io/math"
	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v6/pkg/client"
	assetfttypes "github.com/CoreumFoundation/coreum/v6/x/asset/ft/types"
	assetnfttypes "github.com/CoreumFoundation/coreum/v6/x/asset/nft/types"
	"github.com/CoreumFoundation/coreum/v6/x/deterministicgas"
	dextypes "github.com/CoreumFoundation/coreum/v6/x/dex/types"
)

// CoreumChain is configured coreum chain.
type CoreumChain struct {
	Chain
	Governance             Governance
	DeterministicGasConfig deterministicgas.Config
}

// NewCoreumChain returns a new instance of the CoreumChain.
func NewCoreumChain(chain Chain, stakerMnemonics []string) CoreumChain {
	gov := NewGovernance(chain.ChainContext, stakerMnemonics, chain.Faucet)
	return CoreumChain{
		Chain:                  chain,
		Governance:             gov,
		DeterministicGasConfig: deterministicgas.DefaultConfig(),
	}
}

// BalancesOptions is the input type for the ComputeNeededBalanceFromOptions.
type BalancesOptions struct {
	Messages                    []sdk.Msg
	NondeterministicMessagesGas uint64
	GasPrice                    sdkmath.LegacyDec
	Amount                      sdkmath.Int
}

// AccWithBalancesOptions is the input type for the FundAccountWithOptionsFast.
type AccWithBalancesOptions struct {
	Acc     sdk.AccAddress
	Options BalancesOptions
}

// GasLimitByMsgs calculates sum of gas limits required for message types passed.
// It panics if unsupported message type specified.
func (c CoreumChain) GasLimitByMsgs(msgs ...sdk.Msg) uint64 {
	var totalGasRequired uint64
	for _, msg := range msgs {
		msgGas, exists := c.DeterministicGasConfig.GasRequiredByMessage(msg)
		if !exists {
			panic(errors.Errorf("unsupported message type for deterministic gas: %v", reflect.TypeOf(msg).String()))
		}
		totalGasRequired += msgGas + c.DeterministicGasConfig.FixedGas
	}

	return totalGasRequired
}

// GasLimitForMultiMsgTx calculates sum of gas limits required for message types passed and includes the FixedGas once.
// It panics if unsupported message type specified.
func (c CoreumChain) GasLimitForMultiMsgTx(msgs ...sdk.Msg) uint64 {
	var totalGasRequired uint64
	for _, msg := range msgs {
		msgGas, exists := c.DeterministicGasConfig.GasRequiredByMessage(msg)
		if !exists {
			panic(errors.Errorf("unsupported message type for deterministic gas: %v", reflect.TypeOf(msg).String()))
		}
		totalGasRequired += msgGas
	}

	return totalGasRequired + c.DeterministicGasConfig.FixedGas
}

// ComputeNeededBalanceFromOptions computes the required balance based on the input options.
func (c CoreumChain) ComputeNeededBalanceFromOptions(options BalancesOptions) sdkmath.Int {
	if options.GasPrice.IsNil() {
		options.GasPrice = c.ChainSettings.GasPrice
	}

	if options.Amount.IsNil() {
		options.Amount = sdkmath.ZeroInt()
	}

	// NOTE: we assume that each message goes to one transaction, which is not
	// very accurate and may cause some over funding in cases that there are multiple
	// messages in a single transaction
	totalAmount := sdkmath.ZeroInt()
	for _, msg := range options.Messages {
		gas := c.GasLimitByMsgs(msg)
		// Ceil().RoundInt() is here to be compatible with the sdk's TxFactory
		// https://github.com/cosmos/cosmos-sdk/blob/ff416ee63d32da5d520a8b2d16b00da762416146/client/tx/factory.go#L223
		amt := options.GasPrice.Mul(sdkmath.LegacyNewDec(int64(gas))).Ceil().RoundInt()
		totalAmount = totalAmount.Add(amt)
	}

	return totalAmount.
		Add(options.GasPrice.
			Mul(sdkmath.LegacyNewDec(int64(options.NondeterministicMessagesGas))).
			Ceil().
			RoundInt()).
		Add(options.Amount)
}

// FundAccountWithOptions computes the needed balances and fund account with it.
func (c CoreumChain) FundAccountWithOptions(
	ctx context.Context,
	t *testing.T,
	address sdk.AccAddress,
	options BalancesOptions,
) {
	t.Helper()

	amount := c.ComputeNeededBalanceFromOptions(options)
	c.Faucet.FundAccounts(ctx, t, FundedAccount{
		Address: address,
		Amount:  c.NewCoin(amount),
	})
}

// FundAccountsWithOptions computes the needed balances and fund accounts with it.
func (c CoreumChain) FundAccountsWithOptions(
	ctx context.Context,
	t *testing.T,
	accWithBalancesOptions []AccWithBalancesOptions,
) {
	t.Helper()

	fundedAccounts := make([]FundedAccount, len(accWithBalancesOptions))
	for i, accWithBalancesOption := range accWithBalancesOptions {
		amount := c.ComputeNeededBalanceFromOptions(accWithBalancesOption.Options)
		fundedAccounts[i] = FundedAccount{
			Address: accWithBalancesOption.Acc,
			Amount:  c.NewCoin(amount),
		}
	}

	c.Faucet.FundAccounts(ctx, t, fundedAccounts...)
}

// CreateValidator creates a new validator on the chain and returns the staker addresses,
// validator addresses and callback function to deactivate it.
func (c CoreumChain) CreateValidator(
	ctx context.Context,
	t *testing.T,
	stakingAmount,
	selfDelegationAmount sdkmath.Int,
) (sdk.AccAddress, sdk.ValAddress, func(), error) {
	t.Helper()
	SkipUnsafe(ctx, t)

	stakingClient := stakingtypes.NewQueryClient(c.ClientContext)
	staker := c.GenAccount()

	c.FundAccountWithOptions(ctx, t, staker, BalancesOptions{
		Messages: []sdk.Msg{&stakingtypes.MsgCreateValidator{}, &stakingtypes.MsgUndelegate{}},
		Amount:   stakingAmount,
	})

	// Create staker
	validatorAddr := sdk.ValAddress(staker)
	msg, err := stakingtypes.NewMsgCreateValidator(
		validatorAddr.String(),
		cosmosed25519.GenPrivKey().PubKey(),
		c.NewCoin(stakingAmount),
		stakingtypes.Description{Moniker: fmt.Sprintf("testing-staker-%s", staker)},
		stakingtypes.NewCommissionRates(
			sdkmath.LegacyMustNewDecFromStr("0.1"),
			sdkmath.LegacyMustNewDecFromStr("0.1"),
			sdkmath.LegacyMustNewDecFromStr("0.1"),
		),
		selfDelegationAmount,
	)
	require.NoError(t, err)

	result, err := client.BroadcastTx(
		ctx,
		c.ClientContext.WithFromAddress(staker),
		c.TxFactory().WithGas(c.GasLimitByMsgs(msg)),
		msg,
	)
	if err != nil {
		// we still need that error to be returned since we assert it depending on the test
		return nil, nil, nil, err
	}

	t.Logf("Validator creation executed, txHash:%s", result.TxHash)

	// Make sure staker has been created
	resp, err := stakingClient.Validator(ctx, &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: validatorAddr.String(),
	})
	require.NoError(t, err)
	if stakingAmount.String() != resp.Validator.Tokens.String() {
		t.Fatalf("unexpected validator %q tokens after creation: %s", validatorAddr, resp.Validator.Tokens)
	}
	if stakingtypes.Bonded != resp.Validator.Status {
		t.Fatalf("unexpected validator %q status after creation: %s", validatorAddr, resp.Validator.Status)
	}

	return staker, validatorAddr, func() {
		// Undelegate coins, i.e. deactivate staker
		undelegateMsg := stakingtypes.NewMsgUndelegate(staker.String(), validatorAddr.String(), c.NewCoin(stakingAmount))
		_, err = client.BroadcastTx(
			ctx,
			c.ClientContext.WithFromAddress(staker),
			c.TxFactory().WithGas(c.GasLimitByMsgs(&stakingtypes.MsgUndelegate{})),
			undelegateMsg,
		)
		require.NoError(t, err)

		// make sure the validator isn't bonded now
		resp, err := stakingClient.Validator(ctx, &stakingtypes.QueryValidatorRequest{
			ValidatorAddr: validatorAddr.String(),
		})
		require.NoError(t, err)

		if stakingtypes.Bonded == resp.Validator.Status {
			t.Fatalf("unexpected validator %q status after removal: %s", validatorAddr, resp.Validator.Status)
		}
	}, nil
}

// QueryAssetFTParams queries for asset/ft module params and returns them.
func (c CoreumChain) QueryAssetFTParams(ctx context.Context, t *testing.T) assetfttypes.Params {
	t.Helper()

	queryClient := assetfttypes.NewQueryClient(c.ClientContext)
	resp, err := queryClient.Params(ctx, &assetfttypes.QueryParamsRequest{})
	require.NoError(t, err)

	return resp.Params
}

// QueryAssetNFTParams queries for asset/nft module params and returns them.
func (c CoreumChain) QueryAssetNFTParams(ctx context.Context, t *testing.T) assetnfttypes.Params {
	t.Helper()

	queryClient := assetnfttypes.NewQueryClient(c.ClientContext)
	resp, err := queryClient.Params(ctx, &assetnfttypes.QueryParamsRequest{})
	require.NoError(t, err)

	return resp.Params
}

// QueryDEXParams queries for dex module params and returns them.
func (c CoreumChain) QueryDEXParams(ctx context.Context, t *testing.T) dextypes.Params {
	t.Helper()

	queryClient := dextypes.NewQueryClient(c.ClientContext)
	resp, err := queryClient.Params(ctx, &dextypes.QueryParamsRequest{})
	require.NoError(t, err)

	return resp.Params
}
