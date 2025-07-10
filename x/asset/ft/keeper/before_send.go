package keeper

import (
	"encoding/json"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CoreumFoundation/coreum/v6/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v6/x/wasm"
	cwasmtypes "github.com/CoreumFoundation/coreum/v6/x/wasm/types"
	wibctransfertypes "github.com/CoreumFoundation/coreum/v6/x/wibctransfer/types"
)

// ExtensionTransferMethod the function name of the extension smart contract, which will be invoked
// when doing the transfer.
const ExtensionTransferMethod = "extension_transfer"

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
func (k Keeper) BeforeInputOutputCoins(ctx sdk.Context, input banktypes.Input, outputs []banktypes.Output) error {
	return k.applyFeatures(ctx, input, outputs)
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
			return sdkerrors.Wrapf(err, "invalid address %s", output.Address)
		}
		for _, coin := range output.Coins {
			def, err := k.getDefinitionOrNil(ctx, coin.Denom)
			if err != nil {
				return err
			}
			if def == nil {
				// if the token doesn't have the definition we validate DEX locking rule only.
				if err := k.validateCoinIsNotLockedByDEXAndBank(ctx, sender, coin); err != nil {
					return err
				}

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
			// It is done before checking for global freeze because refunding should not be blocked by this.
			// Otherwise, funds would be lost forever, being blocked on the escrow account.
			if wibctransfertypes.IsPurposeAck(ctx) || wibctransfertypes.IsPurposeTimeout(ctx) {
				if err := k.bankKeeper.SendCoins(ctx, sender, recipient, sdk.NewCoins(coin)); err != nil {
					return err
				}
				continue
			}

			burnAmount := k.CalculateRate(ctx, def.BurnRate, sender, coin)
			commissionAmount := k.CalculateRate(ctx, def.SendCommissionRate, sender, coin)

			senderOrReceiverIsAdmin := def.Admin == sender.String() || def.Admin == recipient.String()

			if !senderOrReceiverIsAdmin && !def.IsFeatureEnabled(types.Feature_extension) {
				if err := k.applyCommissionAndBurnRate(ctx, sender, def, commissionAmount, burnAmount); err != nil {
					return err
				}
			}

			if err := k.validateCoinSpendable(ctx, sender, *def, coin.Amount); err != nil {
				return err
			}

			if err := k.validateCoinReceivable(ctx, recipient, *def, coin.Amount); err != nil {
				return err
			}

			if def.IsFeatureEnabled(types.Feature_extension) {
				if err := k.invokeAssetExtensionExtensionTransferMethod(
					ctx, sender, recipient, *def, coin, commissionAmount, burnAmount,
				); err != nil {
					return err
				}
				continue
			}

			if err := k.bankKeeper.SendCoins(ctx, sender, recipient, sdk.NewCoins(coin)); err != nil {
				return err
			}
		}
	}

	if !outputCoinsSum.Equal(input.Coins) {
		return banktypes.ErrInputOutputMismatch
	}
	return nil
}

func (k Keeper) applyCommissionAndBurnRate(
	ctx sdk.Context,
	sender sdk.AccAddress,
	def *types.Definition,
	commissionAmount, burnAmount sdkmath.Int,
) error {
	if commissionAmount.IsPositive() {
		adminAddr, err := sdk.AccAddressFromBech32(def.Admin)
		if err != nil {
			return err
		}
		commissionCoin := sdk.NewCoins(sdk.NewCoin(def.Denom, commissionAmount))
		if err := k.bankKeeper.SendCoins(ctx, sender, adminAddr, commissionCoin); err != nil {
			return err
		}
	}

	if burnAmount.IsPositive() {
		if err := k.burnIfSpendable(ctx, sender, *def, burnAmount); err != nil {
			return err
		}
	}

	return nil
}

// invokeAssetExtensionExtensionTransferMethod calls the smart contract of the extension. This smart contract is
// responsible to enforce any policies and do the final transfer. The amount attached to the call
// is the send amount plus the burn and commission amount.
func (k Keeper) invokeAssetExtensionExtensionTransferMethod(
	ctx sdk.Context,
	sender sdk.AccAddress,
	recipient sdk.AccAddress,
	def types.Definition,
	sendAmount sdk.Coin,
	commissionAmount sdkmath.Int,
	burnAmount sdkmath.Int,
) error {
	extensionContract, err := sdk.AccAddressFromBech32(def.ExtensionCWAddress)
	if err != nil {
		return err
	}

	// We need this if statement so we will not have an infinite loop. Otherwise
	// when we call Execute method in wasm keeper, in which we have funds transfer,
	// then we will end up in an infinite recursoin.
	if extensionContract.Equals(sender) {
		return k.bankKeeper.SendCoins(ctx, sender, recipient, sdk.NewCoins(sendAmount))
	}

	attachedFunds := sdk.NewCoins(sendAmount).
		Add(sdk.NewCoin(def.Denom, commissionAmount)).
		Add(sdk.NewCoin(def.Denom, burnAmount))

	if err := k.bankKeeper.SendCoins(ctx, sender, extensionContract, attachedFunds); err != nil {
		return err
	}

	senderIsSmartContract := cwasmtypes.IsSendingSmartContract(ctx, sender.String()) ||
		wasm.IsSmartContract(ctx, sender, k.wasmKeeper)
	recipientIsSmartContract := cwasmtypes.IsReceivingSmartContract(ctx, recipient.String()) ||
		wasm.IsSmartContract(ctx, recipient, k.wasmKeeper)

	contractMsg := map[string]interface{}{
		ExtensionTransferMethod: sudoExtensionTransferMsg{
			Sender:           sender.String(),
			Recipient:        recipient.String(),
			TransferAmount:   sendAmount.Amount,
			BurnAmount:       burnAmount,
			CommissionAmount: commissionAmount,
			Context: sudoExtensionTransferContext{
				SenderIsSmartContract:    senderIsSmartContract,
				RecipientIsSmartContract: recipientIsSmartContract,
				IBCPurpose:               ibcPurposeToExtensionString(ctx),
			},
		},
	}
	contractMsgBytes, err := json.Marshal(contractMsg)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to marshal contract msg")
	}

	_, err = k.wasmPermissionedKeeper.Sudo(
		ctx,
		extensionContract,
		contractMsgBytes,
	)
	if err != nil {
		return types.ErrExtensionCallFailed.Wrapf("wasm error: %s", err)
	}
	return nil
}

// CalculateRate calculates how the burn or commission amount should be calculated.
func (k Keeper) CalculateRate(
	ctx sdk.Context,
	rate sdkmath.LegacyDec,
	sender sdk.AccAddress,
	amount sdk.Coin,
) sdkmath.Int {
	// We decided that rates should not be charged on incoming IBC transfers.
	// According to our current protocol, it cannot be done because sender pays the rates, meaning that escrow address
	// would be charged leading to breaking the IBC mechanics.
	if wibctransfertypes.IsPurposeIn(ctx) {
		return sdkmath.ZeroInt()
	}

	// Context is marked with ACK purpose in two cases:
	// - when IBC transfer succeeded on the receiving chain (positive ACK)
	// - when IBC transfer has been rejected by the other chain (negative ACK)
	// This function is called only in the negative case, when the IBC transfer must be rolled back and funds
	// must be sent back to the sender. In this case we should not charge the rates.
	if wibctransfertypes.IsPurposeAck(ctx) {
		return sdkmath.ZeroInt()
	}

	// Same thing as above just in case of IBC timeout.
	if wibctransfertypes.IsPurposeTimeout(ctx) {
		return sdkmath.ZeroInt()
	}

	if rate.IsNil() || !rate.IsPositive() {
		return sdkmath.ZeroInt()
	}

	// We do not apply burn and commission rate if sender is a smart contract address.
	if cwasmtypes.IsSendingSmartContract(ctx, sender.String()) {
		return sdkmath.ZeroInt()
	}

	return rate.MulInt(amount.Amount).Ceil().RoundInt()
}
