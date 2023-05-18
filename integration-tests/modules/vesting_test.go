//go:build integrationtests

package modules

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
	customparamstypes "github.com/CoreumFoundation/coreum/x/customparams/types"
)

// TestVestingAccountCreationAndBankSend tests vesting account can be created, and it's send limits are applied.
func TestVestingAccountCreationAndBankSend(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t, false)

	creator := chain.GenAccount()
	vestingAcc := chain.GenAccount()
	recipient := chain.GenAccount()

	requireT := require.New(t)
	authClient := authtypes.NewQueryClient(chain.ClientContext)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	amountToVest := sdk.NewInt(100)
	requireT.NoError(chain.FundAccountsWithOptions(ctx, creator, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&vestingtypes.MsgCreateVestingAccount{}},
		Amount:   amountToVest,
	}))

	vestingDuration := 10 * time.Second
	vestingCoin := chain.NewCoin(amountToVest)
	createAccMsg := &vestingtypes.MsgCreateVestingAccount{
		FromAddress: creator.String(),
		ToAddress:   vestingAcc.String(),
		Amount:      sdk.NewCoins(vestingCoin),
		EndTime:     time.Now().Add(vestingDuration).Unix(),
		Delayed:     true,
	}

	txRes, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(creator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(createAccMsg)),
		createAccMsg,
	)
	requireT.NoError(err)
	requireT.Equal(uint64(txRes.GasUsed), chain.GasLimitByMsgs(createAccMsg))

	// check account is created and it's vesting
	accountRes, err := authClient.Account(ctx, &authtypes.QueryAccountRequest{
		Address: vestingAcc.String(),
	})
	requireT.NoError(err)
	requireT.Equal("/cosmos.vesting.v1beta1.DelayedVestingAccount", accountRes.Account.TypeUrl)
	// check the balance is full
	balanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: vestingAcc.String(),
		Denom:   vestingCoin.Denom,
	})
	requireT.NoError(err)
	requireT.Equal(vestingCoin.String(), balanceRes.Balance.String())

	// fund the vesting account to pay fees
	requireT.NoError(chain.FundAccountsWithOptions(ctx, vestingAcc, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	}))

	msgSend := &banktypes.MsgSend{
		FromAddress: vestingAcc.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(vestingCoin),
	}

	// try to send full amount from vesting account before delay is ended
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(vestingAcc),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	requireT.True(sdkerrors.ErrInsufficientFunds.Is(err))

	// await vesting time to unlock the vesting coins
	select {
	case <-ctx.Done():
		return
	case <-time.After(vestingDuration):
	}

	// fund the vesting account to pay fees one more time
	requireT.NoError(chain.FundAccountsWithOptions(ctx, vestingAcc, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
	}))

	// try to send one more time, the coins should be unlocked at that time
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(vestingAcc),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	requireT.NoError(err)
}

// TestVestingAccountStaking tests the vesting account can delegate coins.
func TestVestingAccountStaking(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t, false)

	creator := chain.GenAccount()
	vestingAcc := chain.GenAccount()

	requireT := require.New(t)
	authClient := authtypes.NewQueryClient(chain.ClientContext)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	customParamsClient := customparamstypes.NewQueryClient(chain.ClientContext)

	amountToVest := sdk.NewInt(100)
	requireT.NoError(chain.FundAccountsWithOptions(ctx, creator, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&vestingtypes.MsgCreateVestingAccount{}},
		Amount:   amountToVest,
	}))

	// create new validator
	customStakingParams, err := customParamsClient.StakingParams(ctx, &customparamstypes.QueryStakingParamsRequest{})
	require.NoError(t, err)
	validatorStakingAmount := customStakingParams.Params.MinSelfDelegation
	_, validatorAddress, deactivateValidator, err := chain.CreateValidator(ctx, validatorStakingAmount, validatorStakingAmount)
	require.NoError(t, err)
	defer func() {
		err := deactivateValidator()
		require.NoError(t, err)
	}()

	vestingDuration := time.Hour
	vestingCoin := chain.NewCoin(amountToVest)
	createAccMsg := &vestingtypes.MsgCreateVestingAccount{
		FromAddress: creator.String(),
		ToAddress:   vestingAcc.String(),
		Amount:      sdk.NewCoins(vestingCoin),
		EndTime:     time.Now().Add(vestingDuration).Unix(),
		Delayed:     false,
	}

	txRes, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(creator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(createAccMsg)),
		createAccMsg,
	)
	requireT.NoError(err)
	requireT.Equal(uint64(txRes.GasUsed), chain.GasLimitByMsgs(createAccMsg))

	// check that account is created and it is vesting account
	accountRes, err := authClient.Account(ctx, &authtypes.QueryAccountRequest{
		Address: vestingAcc.String(),
	})
	requireT.NoError(err)
	requireT.Equal("/cosmos.vesting.v1beta1.ContinuousVestingAccount", accountRes.Account.TypeUrl)

	// check that the balance is full
	balanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: vestingAcc.String(),
		Denom:   vestingCoin.Denom,
	})
	requireT.NoError(err)
	requireT.Equal(vestingCoin.String(), balanceRes.Balance.String())

	// fund the vesting account to pay fees for the staking
	requireT.NoError(chain.FundAccountsWithOptions(ctx, vestingAcc, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&stakingtypes.MsgDelegate{},
		},
	}))

	msgDelegate := &stakingtypes.MsgDelegate{
		DelegatorAddress: vestingAcc.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           vestingCoin,
	}

	// try to delegate full amount from created vesting account
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(vestingAcc),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgDelegate)),
		msgDelegate,
	)
	requireT.NoError(err)
}

// TestVestingAccountWithFTInteraction tests that vesting accounts correctly work with the ft assets.
func TestVestingAccountWithFTInteraction(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t, false)

	issuer := chain.GenAccount()
	vestingAcc := chain.GenAccount()
	recipient := chain.GenAccount()

	requireT := require.New(t)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	requireT.NoError(
		chain.FundAccountsWithOptions(ctx, issuer, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assetfttypes.MsgSetWhitelistedLimit{},
				&assetfttypes.MsgSetWhitelistedLimit{},
				&vestingtypes.MsgCreateVestingAccount{},
				&banktypes.MsgSend{},
				&assetfttypes.MsgIssue{},
				&assetfttypes.MsgFreeze{},
				&assetfttypes.MsgUnfreeze{},
			},
			Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
		}))

	// issue a fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "symbol",
		Subunit:       "subunit",
		Precision:     6,
		Description:   "description",
		InitialAmount: sdk.NewInt(10_000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_burning,
			assetfttypes.Feature_freezing,
			assetfttypes.Feature_whitelisting,
		},
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)

	vestingCoin := sdk.NewCoin(denom, sdk.NewInt(100))

	// whitelist the vestingAcc
	msgSetWhitelistedLimit := &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: vestingAcc.String(),
		Coin:    vestingCoin,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSetWhitelistedLimit)),
		msgSetWhitelistedLimit,
	)
	requireT.NoError(err)

	// whitelist the recipient to let the vesting account send coins to it
	msgSetWhitelistedLimit = &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    vestingCoin,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSetWhitelistedLimit)),
		msgSetWhitelistedLimit,
	)
	requireT.NoError(err)

	vestingDuration := 10 * time.Second
	createAccMsg := &vestingtypes.MsgCreateVestingAccount{
		FromAddress: issuer.String(),
		ToAddress:   vestingAcc.String(),
		Amount:      sdk.NewCoins(vestingCoin),
		EndTime:     time.Now().Add(vestingDuration).Unix(),
		Delayed:     true,
	}

	txRes, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(createAccMsg)),
		createAccMsg,
	)
	requireT.NoError(err)
	requireT.Equal(uint64(txRes.GasUsed), chain.GasLimitByMsgs(createAccMsg))

	// check that the balance is received
	balanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: vestingAcc.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(vestingCoin.String(), balanceRes.Balance.String())

	// fund the vesting account to pay fees
	requireT.NoError(chain.FundAccountsWithOptions(ctx, vestingAcc, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgBurn{},
			&assetfttypes.MsgBurn{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
		},
	}))

	// try to burn vesting locked coins
	burnMsg := &assetfttypes.MsgBurn{
		Sender: vestingAcc.String(),
		Coin:   sdk.NewInt64Coin(denom, 10),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(vestingAcc),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.True(sdkerrors.ErrInsufficientFunds.Is(err))

	// try to send vesting locker coins
	msgSend := &banktypes.MsgSend{
		FromAddress: vestingAcc.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(denom, 10)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(vestingAcc),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	requireT.True(sdkerrors.ErrInsufficientFunds.Is(err))

	// freeze coins, it should work even for the vested coins
	freezeMsg := &assetfttypes.MsgFreeze{
		Sender:  issuer.String(),
		Account: vestingAcc.String(),
		Coin:    sdk.NewInt64Coin(denom, 50),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	requireT.NoError(err)

	// await vesting time to unlock the vesting coins
	select {
	case <-ctx.Done():
		return
	case <-time.After(vestingDuration):
	}

	// try to burn one more time, now the coins are unlocked so can be burnt
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(vestingAcc),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.NoError(err)

	// try to send unlocked coins
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(vestingAcc),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	requireT.NoError(err)

	// try to send the partially frozen now
	msgSend = &banktypes.MsgSend{
		FromAddress: vestingAcc.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(denom, 60)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(vestingAcc),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	requireT.True(sdkerrors.ErrInsufficientFunds.Is(err))

	// unfreeze coins, to let prev vesting account tx pass
	unfreezeMsg := &assetfttypes.MsgUnfreeze{
		Sender:  issuer.String(),
		Account: vestingAcc.String(),
		Coin:    sdk.NewInt64Coin(denom, 50),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(unfreezeMsg)),
		unfreezeMsg,
	)
	requireT.NoError(err)

	// try to send the unfrozen coins now
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(vestingAcc),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	requireT.NoError(err)
}
