package testing

import (
	"context"
	"fmt"

	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// CreateValidator creates a new validator on the chain and returns the staker addresses, validator addresses and callback function to deactivate it.
func CreateValidator(ctx context.Context, chain Chain, stakingAmount sdk.Int, selfDelegationAmount sdk.Int) (sdk.AccAddress, sdk.ValAddress, func() error, error) {
	stakingClient := stakingtypes.NewQueryClient(chain.ClientContext)
	staker := chain.GenAccount()

	if err := chain.Faucet.FundAccountsWithOptions(ctx, staker, BalancesOptions{
		Messages: []sdk.Msg{&stakingtypes.MsgCreateValidator{}, &stakingtypes.MsgUndelegate{}},
		Amount:   stakingAmount,
	}); err != nil {
		return nil, nil, nil, err
	}

	// Create staker
	validatorAddr := sdk.ValAddress(staker)
	msg, err := stakingtypes.NewMsgCreateValidator(
		validatorAddr,
		cosmosed25519.GenPrivKey().PubKey(),
		chain.NewCoin(stakingAmount),
		stakingtypes.Description{Moniker: fmt.Sprintf("testing-staker-%s", staker)},
		stakingtypes.NewCommissionRates(sdk.MustNewDecFromStr("0.1"), sdk.MustNewDecFromStr("0.1"), sdk.MustNewDecFromStr("0.1")),
		selfDelegationAmount,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	result, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(staker),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	logger.Get(ctx).Info("Validator creation executed", zap.String("txHash", result.TxHash))

	// Make sure staker has been created
	resp, err := stakingClient.Validator(ctx, &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: validatorAddr.String(),
	})
	if err != nil {
		return nil, nil, nil, errors.WithStack(err)
	}
	if stakingAmount.String() != resp.Validator.Tokens.String() {
		return nil, nil, nil, errors.Errorf("unexpected validator %q tokens after creation: %s", validatorAddr, resp.Validator.Tokens)
	}
	if stakingtypes.Bonded != resp.Validator.Status {
		return nil, nil, nil, errors.Errorf("unexpected validator %q status after creation: %s", validatorAddr, resp.Validator.Status)
	}

	return staker, validatorAddr, func() error {
		// Undelegate coins, i.e. deactivate staker
		undelegateMsg := stakingtypes.NewMsgUndelegate(staker, validatorAddr, chain.NewCoin(stakingAmount))
		_, err = tx.BroadcastTx(
			ctx,
			chain.ClientContext.WithFromAddress(staker),
			chain.TxFactory().WithGas(chain.GasLimitByMsgs(undelegateMsg)),
			undelegateMsg,
		)
		if err != nil {
			return err
		}

		// make sure the validator isn't bonded now
		resp, err := stakingClient.Validator(ctx, &stakingtypes.QueryValidatorRequest{
			ValidatorAddr: validatorAddr.String(),
		})
		if err != nil {
			return errors.WithStack(err)
		}

		if stakingtypes.Bonded == resp.Validator.Status {
			return errors.Errorf("unexpected validator %q status after removal: %s", validatorAddr, resp.Validator.Status)
		}

		return nil
	}, nil
}
