package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

// Keeper is the asset module keeper.
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   sdk.StoreKey
	bankKeeper types.BankKeeper
}

// NewKeeper creates a new instance of the Keeper.
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, bankKeeper types.BankKeeper) Keeper {
	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		bankKeeper: bankKeeper,
	}
}

// IssueFungibleToken issues new fungible token and returns it's denom.
func (k Keeper) IssueFungibleToken(ctx sdk.Context, settings types.IssueFungibleTokenSettings) (string, error) {
	denom := types.BuildFungibleTokenDenom(settings.Symbol, settings.Issuer)
	if _, found := k.bankKeeper.GetDenomMetaData(ctx, denom); found {
		return "", sdkerrors.Wrapf(
			types.ErrInvalidFungibleToken,
			"symbol %s already registered for the address %s",
			settings.Symbol,
			settings.Issuer.String(),
		)
	}

	k.setFungibleTokenDenomMetadata(ctx, settings.Symbol, denom, settings.Description)

	if err := k.mintFungibleToken(ctx, denom, settings.InitialAmount, settings.Recipient); err != nil {
		return "", err
	}

	store := ctx.KVStore(k.storeKey)
	definition := types.FungibleTokenDefinition{
		Denom:  denom,
		Issuer: settings.Issuer.String(),
	}
	store.Set(types.GetFungibleTokenKey(denom), k.cdc.MustMarshal(&definition))

	if err := ctx.EventManager().EmitTypedEvent(&types.EventFungibleTokenIssued{
		Denom:         denom,
		Issuer:        settings.Issuer.String(),
		Symbol:        settings.Symbol,
		Recipient:     settings.Recipient.String(),
		Description:   settings.Description,
		InitialAmount: settings.InitialAmount,
	}); err != nil {
		return "", sdkerrors.Wrap(err, "can't emit EventFungibleTokenIssued event")
	}

	k.Logger(ctx).Debug("issued new fungible token with denom %d", denom)

	return denom, nil
}

// GetFungibleToken return the fungible token by its denom.
func (k Keeper) GetFungibleToken(ctx sdk.Context, denom string) (types.FungibleToken, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetFungibleTokenKey(denom))
	if bz == nil {
		return types.FungibleToken{}, sdkerrors.Wrapf(types.ErrFungibleTokenNotFound, "denom: %s", denom)
	}
	var definition types.FungibleTokenDefinition
	k.cdc.MustUnmarshal(bz, &definition)

	metadata, found := k.bankKeeper.GetDenomMetaData(ctx, denom)
	if !found {
		return types.FungibleToken{}, sdkerrors.Wrapf(types.ErrFungibleTokenNotFound, "metadata for %s denom not found", denom)
	}

	return types.FungibleToken{
		Denom:       definition.Denom,
		Issuer:      definition.Issuer,
		Symbol:      metadata.Symbol,
		Description: metadata.Description,
	}, nil
}

// GetLockedCoins returns the locked coins for the account.
func (k Keeper) GetLockedCoins(ctx sdk.Context, address sdk.AccAddress) sdk.Coins {
	return sdk.NewCoins()
}

// Logger returns the Keeper logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) setFungibleTokenDenomMetadata(ctx sdk.Context, symbol, denom, description string) {
	denomMetadata := banktypes.Metadata{
		Name:        denom,
		Symbol:      symbol,
		Description: description,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    denom,
				Exponent: uint32(0),
			},
		},
		Base:    denom,
		Display: denom,
	}

	k.bankKeeper.SetDenomMetaData(ctx, denomMetadata)
}

func (k Keeper) mintFungibleToken(ctx sdk.Context, denom string, amount sdk.Int, recipient sdk.AccAddress) error {
	if !amount.IsPositive() {
		return nil
	}
	coinsToMint := sdk.NewCoins(sdk.NewCoin(denom, amount))
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coinsToMint); err != nil {
		return sdkerrors.Wrapf(err, "can't mint %s for the module %s", coinsToMint.String(), types.ModuleName)
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, coinsToMint); err != nil {
		return sdkerrors.Wrapf(err, "can't send minted coins from module %s to account %s", types.ModuleName, recipient.String())
	}

	return nil
}
