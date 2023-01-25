//go:build integrationtests

package modules

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// TestVestingAccountCreationAndBankSend tests vesting account can be created, and it's send limits are applied.
func TestVestingAccountCreationAndBankSend(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	creator := chain.GenAccount()
	vestingAcc := chain.GenAccount()
	recipient := chain.GenAccount()

	requireT := require.New(t)
	authClient := authtypes.NewQueryClient(chain.ClientContext)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	amountToVest := sdk.NewInt(100)
	requireT.NoError(chain.Faucet.FundAccountsWithOptions(ctx, creator, integrationtests.BalancesOptions{
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

	txRes, err := tx.BroadcastTx(
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
	requireT.NoError(chain.Faucet.FundAccountsWithOptions(ctx, vestingAcc, integrationtests.BalancesOptions{
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
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(vestingAcc),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	requireT.True(sdkerrors.ErrInsufficientFunds.Is(err))

	// await vesting time to unlock the vesting coins
	<-time.After(vestingDuration)

	// fund the vesting account to pay fees one more time
	requireT.NoError(chain.Faucet.FundAccountsWithOptions(ctx, vestingAcc, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
	}))

	// try to send one more time, the coins should be unlocked at that time
	_, err = tx.BroadcastTx(
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

	ctx, chain := integrationtests.NewTestingContext(t)

	creator := chain.GenAccount()
	vestingAcc := chain.GenAccount()

	requireT := require.New(t)
	authClient := authtypes.NewQueryClient(chain.ClientContext)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	stakingClient := stakingtypes.NewQueryClient(chain.ClientContext)

	amountToVest := sdk.NewInt(100)
	requireT.NoError(chain.Faucet.FundAccountsWithOptions(ctx, creator, integrationtests.BalancesOptions{
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
		Delayed:     false,
	}

	txRes, err := tx.BroadcastTx(
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
	requireT.NoError(chain.Faucet.FundAccountsWithOptions(ctx, vestingAcc, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&stakingtypes.MsgDelegate{},
		},
	}))

	validatorsRes, err := stakingClient.Validators(ctx, &stakingtypes.QueryValidatorsRequest{
		Status: stakingtypes.BondStatusBonded,
		Pagination: &query.PageRequest{
			Limit: 1,
		},
	})
	requireT.NoError(err)

	validatorOperatorAddress := validatorsRes.Validators[0].OperatorAddress

	msgDelegate := &stakingtypes.MsgDelegate{
		DelegatorAddress: vestingAcc.String(),
		ValidatorAddress: validatorOperatorAddress,
		Amount:           vestingCoin,
	}

	// try to delegate full amount from created vesting account
	_, err = tx.BroadcastTx(
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

	ctx, chain := integrationtests.NewTestingContext(t)

	issuer := chain.GenAccount()
	vestingAcc := chain.GenAccount()
	recipient := chain.GenAccount()

	requireT := require.New(t)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assetfttypes.MsgSetWhitelistedLimit{},
				&assetfttypes.MsgSetWhitelistedLimit{},
				&vestingtypes.MsgCreateVestingAccount{},
				&banktypes.MsgSend{},
				&assetfttypes.MsgIssue{},
				&assetfttypes.MsgFreeze{},
				&assetfttypes.MsgUnfreeze{},
			},
			Amount: chain.NetworkConfig.AssetFTConfig.IssueFee,
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
			assetfttypes.Feature_burning,      //nolint:nosnakecase
			assetfttypes.Feature_freezing,     //nolint:nosnakecase
			assetfttypes.Feature_whitelisting, //nolint:nosnakecase
		},
	}

	_, err := tx.BroadcastTx(
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
	_, err = tx.BroadcastTx(
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
	_, err = tx.BroadcastTx(
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

	txRes, err := tx.BroadcastTx(
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
	requireT.NoError(chain.Faucet.FundAccountsWithOptions(ctx, vestingAcc, integrationtests.BalancesOptions{
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
	_, err = tx.BroadcastTx(
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
	_, err = tx.BroadcastTx(
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
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	requireT.NoError(err)

	// await vesting time to unlock the vesting coins
	<-time.After(vestingDuration)

	// try to burn one more time, now the coins are unlocked so can be burnt
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(vestingAcc),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.NoError(err)

	// try to send unlocker coins
	_, err = tx.BroadcastTx(
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
	_, err = tx.BroadcastTx(
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
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(unfreezeMsg)),
		unfreezeMsg,
	)
	requireT.NoError(err)

	// try to send the unfrozen coins now
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(vestingAcc),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	requireT.NoError(err)
}
