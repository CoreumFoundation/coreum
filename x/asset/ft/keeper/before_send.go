package keeper

import (
	"encoding/json"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	cwasmtypes "github.com/CoreumFoundation/coreum/v4/x/wasm/types"
	wibctransfertypes "github.com/CoreumFoundation/coreum/v4/x/wibctransfer/types"
)

// extension method calls.
const (
	// the function name of the extension smart contract, which will be invoked
	// when doing the transfer.
	ExtenstionTransferMethod = "extension_transfer"
)

// sudoExtensionTransferMsg contains the fields passed to extension method call.
//
//nolint:tagliatelle // these will be exposed to rust and must be snake case.
type sudoExtensionTransferMsg struct {
	Recipient        string                       `json:"recipient,omitempty"`
	Sender           string                       `json:"sender,omitempty"`
	TransferAmount   sdkmath.Int                  `json:"transfer_amount,omitempty"`
	BurnAmount       sdkmath.Int                  `json:"burn_amount,omitempty"`
	CommissionAmount sdkmath.Int                  `json:"commission_amount,omitempty"`
	Context          sudoExtensionTransferContext `json:"context,omitempty"`
}

//nolint:tagliatelle // these will be exposed to rust and must be snake case.
type sudoExtensionTransferContext struct {
	SenderIsSmartContract    bool   `json:"sender_is_smart_contract"`
	RecipientIsSmartContract bool   `json:"recipient_is_smart_contract"`
	IBCPurpose               string `json:"ibc_purpose"`
}

func ibcPurposeToExtensionString(ctx sdk.Context) string {
	ibcPurpose, ok := wibctransfertypes.GetPurpose(ctx)
	if !ok {
		return "none"
	}
	return string(ibcPurpose)
}

// BeforeSendCoins checks that a transfer request is allowed or not.
func (k Keeper) BeforeSendCoins(ctx sdk.Context, fromAddress, toAddress sdk.AccAddress, coins sdk.Coins) error {
	return k.applyFeatures(
		ctx,
		banktypes.Input{Address: fromAddress.String(), Coins: coins},
		[]banktypes.Output{{Address: toAddress.String(), Coins: coins}},
	)
}

// BeforeInputOutputCoins extends InputOutputCoins method of the bank keeper.
func (k Keeper) BeforeInputOutputCoins(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	if len(inputs) > 1 {
		return banktypes.ErrMultipleSenders
	}
	if len(inputs) == 0 {
		return banktypes.ErrNoInputs
	}
	return k.applyFeatures(ctx, inputs[0], outputs)
}

func (k Keeper) applyFeatures(ctx sdk.Context, input banktypes.Input, outputs []banktypes.Output) error {
	outputCoinsSum := sdk.NewCoins()
	sender, err := sdk.AccAddressFromBech32(input.Address)
	if err != nil {
		return sdkerrors.Wrapf(err, "invalid address %s", input.Address)
	}
	for _, output := range outputs {
		outputCoinsSum = outputCoinsSum.Add(output.Coins...)
		recipient, err := sdk.AccAddressFromBech32(output.Address)
		if err != nil {
			return sdkerrors.Wrapf(err, "invalid address %s", input.Address)
		}
		for _, coin := range output.Coins {
			def, err := k.GetDefinition(ctx, coin.Denom)
			if types.ErrInvalidDenom.Is(err) || types.ErrTokenNotFound.Is(err) {
				// if the token is not defined in asset ft module, we assume this is different
				// type of token (e.g core, ibc, etc) and don't apply asset ft rules.
				if err := k.bankKeeper.SendCoins(ctx, sender, recipient, sdk.NewCoins(coin)); err != nil {
					return err
				}
				continue
			}

			// This check is effective when IBC transfer is acknowledged by the peer chain or timed out.
			// It happens in the following situations:
			// - when transfer succeeded
			// - when transfer has been rejected by the other chain and funds should be refunded.
			// - when transfer has timedout and funds should be refuned.
			// So, whenever it happens here, it means that funds are going to be refunded
			// back to the sender by the IBC transfer module.
			// It should succeed even if the issuer decided, for whatever reason, to freeze the escrow address.
			// It is done before cehcking for global freeze because refunding should not be blocked by this.
			// Otherwise, funds would be lost forever, being blocked on the escrow account.
			if wibctransfertypes.IsPurposeAck(ctx) || wibctransfertypes.IsPurposeTimeout(ctx) {
				if err := k.bankKeeper.SendCoins(ctx, sender, recipient, sdk.NewCoins(coin)); err != nil {
					return err
				}
				continue
			}

			burnAmount := k.CalculateRate(ctx, def.BurnRate, sender, coin)
			commissionAmount := k.CalculateRate(ctx, def.SendCommissionRate, sender, coin)

			if def.IsFeatureEnabled(types.Feature_extension) {
				if err := k.invokeAssetExtension(ctx, sender, recipient, def, coin, commissionAmount, burnAmount); err != nil {
					return err
				}
				// We will not enforce any policies(e.g whitelisting, burn rate, ) or perform bank transfers
				// if the token has extensions. It is up to the contract to enforce them as needed. As a result
				// we will skip the next operations in this for loop.
				continue
			}

			senderOrReceiverIsAdmin := def.Admin == sender.String() || def.Admin == recipient.String()

			if !senderOrReceiverIsAdmin && commissionAmount.IsPositive() {
				adminAddr := sdk.MustAccAddressFromBech32(def.Admin)
				commissionCoin := sdk.NewCoins(sdk.NewCoin(def.Denom, commissionAmount))
				if err := k.bankKeeper.SendCoins(ctx, sender, adminAddr, commissionCoin); err != nil {
					return err
				}
			}

			if !senderOrReceiverIsAdmin && burnAmount.IsPositive() {
				if err := k.burnIfSpendable(ctx, sender, def, burnAmount); err != nil {
					return err
				}
			}

			if err := k.isCoinSpendable(ctx, sender, def, coin.Amount); err != nil {
				return err
			}

			if err := k.isCoinReceivable(ctx, recipient, def, coin.Amount); err != nil {
				return err
			}

			if err := k.bankKeeper.SendCoins(ctx, sender, recipient, sdk.NewCoins(coin)); err != nil {
				return err
			}
		}
	}

	if !outputCoinsSum.IsEqual(input.Coins) {
		return banktypes.ErrInputOutputMismatch
	}
	return nil
}

// invokeAssetExtension calls the smart contract of the extension. This smart contract is
// responsible to enforce any policies and do the final tranfer. The amount attached to the call
// is the send amount plus the burn and commission amount.
func (k Keeper) invokeAssetExtension(
	ctx sdk.Context,
	sender sdk.AccAddress,
	recipient sdk.AccAddress,
	def types.Definition,
	sendAmount sdk.Coin,
	commissionAmount sdkmath.Int,
	burnAmount sdkmath.Int,
) error {
	// FIXME(milad) we need to write tests in which we check
	// 1. sending to and from smart contract.
	// 2. calling the smart contract directly and sending from it.
	// 3. calling smart contract directly/indirectly, in which smart contracts sends
	// 	  and also receives (receive can happen by invoking another contract)
	// 4. testing sending and receiving from smart contract that is not admin
	// 5. test IBC send and receives
	extensionContract, err := sdk.AccAddressFromBech32(def.ExtensionCWAddress)
	if err != nil {
		return err
	}

	// We need this if statement so we will not have an infinite loop. Otherwise
	// when we call Execute method in wasm keeper, in which we have funds transfer,
	// then we will end up in an infinite recursoin.
	if extensionContract.Equals(recipient) || extensionContract.Equals(sender) {
		return k.bankKeeper.SendCoins(ctx, sender, recipient, sdk.NewCoins(sendAmount))
	}

	attachedFunds := sdk.NewCoins(sendAmount).
		Add(sdk.NewCoin(def.Denom, commissionAmount)).
		Add(sdk.NewCoin(def.Denom, burnAmount))

	if err := k.bankKeeper.SendCoins(ctx, sender, extensionContract, attachedFunds); err != nil {
		return err
	}

	contractMsg := map[string]interface{}{
		ExtenstionTransferMethod: sudoExtensionTransferMsg{
			Sender:           sender.String(),
			Recipient:        recipient.String(),
			TransferAmount:   sendAmount.Amount,
			BurnAmount:       burnAmount,
			CommissionAmount: commissionAmount,
			Context: sudoExtensionTransferContext{
				IBCPurpose: ibcPurposeToExtensionString(ctx),
			},
		},
	}
	contractMsgBytes, err := json.Marshal(contractMsg)
	if err != nil {
		return err
	}

	_, err = k.wasmPermissionedKeeper.Sudo(
		ctx,
		extensionContract,
		contractMsgBytes,
	)
	if err != nil {
		return types.ErrExtensionCallFailed.Wrapf("was error: %s", err)
	}
	return nil
}

// CalculateRate calculates how the burn or commission amount should be calculated.
func (k Keeper) CalculateRate(
	ctx sdk.Context,
	rate sdk.Dec,
	sender sdk.AccAddress,
	amount sdk.Coin,
) sdkmath.Int {
	// We decided that rates should not be charged on incoming IBC transfers.
	// According to our current protocol, it cannot be done because sender pays the rates, meaning that escrow address
	// would be charged leading to breaking the IBC mechanics.
	if wibctransfertypes.IsPurposeIn(ctx) {
		return sdk.ZeroInt()
	}

	// Context is marked with ACK purpose in two cases:
	// - when IBC transfer succeeded on the receiving chain (positive ACK)
	// - when IBC transfer has been rejected by the other chain (negative ACK)
	// This function is called only in the negative case, when the IBC transfer must be rolled back and funds
	// must be sent back to the sender. In this case we should not charge the rates.
	if wibctransfertypes.IsPurposeAck(ctx) {
		return sdk.ZeroInt()
	}

	// Same thing as above just in case of IBC timeout.
	if wibctransfertypes.IsPurposeTimeout(ctx) {
		return sdk.ZeroInt()
	}

	if rate.IsNil() || !rate.IsPositive() {
		return sdk.ZeroInt()
	}

	// We do not apply burn and commission rate if sender is a smart contract address.
	if cwasmtypes.IsSendingSmartContract(ctx, sender.String()) {
		return sdk.ZeroInt()
	}

	return rate.MulInt(amount.Amount).Ceil().RoundInt()
}
