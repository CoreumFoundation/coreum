package keeper

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/pkg/store"
	"github.com/CoreumFoundation/coreum/x/asset"
	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
	"github.com/CoreumFoundation/coreum/x/nft"
)

// ParamSubspace represents a subscope of methods exposed by param module to store and retrieve parameters.
type ParamSubspace interface {
	GetParamSet(ctx sdk.Context, ps paramtypes.ParamSet)
	SetParamSet(ctx sdk.Context, ps paramtypes.ParamSet)
}

// Keeper is the asset module non-fungible token nftKeeper.
type Keeper struct {
	cdc           codec.BinaryCodec
	paramSubspace ParamSubspace
	storeKey      sdk.StoreKey
	nftKeeper     types.NFTKeeper
	bankKeeper    types.BankKeeper
}

// NewKeeper creates a new instance of the Keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	paramSubspace ParamSubspace,
	storeKey sdk.StoreKey,
	nftKeeper types.NFTKeeper,
	bankKeeper types.BankKeeper,
) Keeper {
	return Keeper{
		cdc:           cdc,
		paramSubspace: paramSubspace,
		storeKey:      storeKey,
		nftKeeper:     nftKeeper,
		bankKeeper:    bankKeeper,
	}
}

// GetParams gets the parameters of the model.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var params types.Params
	k.paramSubspace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the parameters of the model.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSubspace.SetParamSet(ctx, &params)
}

// GetClass reruns the Class.
func (k Keeper) GetClass(ctx sdk.Context, classID string) (types.Class, error) {
	definition, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return types.Class{}, err
	}

	class, found := k.nftKeeper.GetClass(ctx, classID)
	if !found {
		return types.Class{}, sdkerrors.Wrapf(types.ErrClassNotFound, "nft class with ID:%s not found", classID)
	}

	return types.Class{
		Id:          class.Id,
		Issuer:      definition.Issuer,
		Name:        class.Name,
		Symbol:      class.Symbol,
		Description: class.Description,
		URI:         class.Uri,
		URIHash:     class.UriHash,
		Data:        class.Data,
		Features:    definition.Features,
		RoyaltyRate: definition.RoyaltyRate,
	}, nil
}

// IssueClass issues new non-fungible token class and returns its id.
func (k Keeper) IssueClass(ctx sdk.Context, settings types.IssueClassSettings) (string, error) {
	if err := types.ValidateClassSymbol(settings.Symbol); err != nil {
		return "", err
	}

	if err := types.ValidateRoyaltyRate(settings.RoyaltyRate); err != nil {
		return "", err
	}

	id := types.BuildClassID(settings.Symbol, settings.Issuer)
	if err := nft.ValidateClassID(id); err != nil {
		return "", sdkerrors.Wrap(types.ErrInvalidInput, err.Error())
	}

	if err := types.ValidateData(settings.Data); err != nil {
		return "", sdkerrors.Wrap(types.ErrInvalidInput, err.Error())
	}

	found := k.nftKeeper.HasClass(ctx, id)
	if found {
		return "", sdkerrors.Wrapf(
			types.ErrInvalidInput,
			"symbol %q already used for the address %q",
			settings.Symbol,
			settings.Issuer,
		)
	}

	if err := k.nftKeeper.SaveClass(ctx, nft.Class{
		Id:          id,
		Symbol:      settings.Symbol,
		Name:        settings.Name,
		Description: settings.Description,
		Uri:         settings.URI,
		UriHash:     settings.URIHash,
		Data:        settings.Data,
	}); err != nil {
		return "", sdkerrors.Wrapf(types.ErrInvalidInput, "can't save non-fungible token: %s", err)
	}

	k.SetClassDefinition(ctx, types.ClassDefinition{
		ID:          id,
		Issuer:      settings.Issuer.String(),
		Features:    settings.Features,
		RoyaltyRate: settings.RoyaltyRate,
	})

	if err := ctx.EventManager().EmitTypedEvent(&types.EventClassIssued{
		ID:          id,
		Issuer:      settings.Issuer.String(),
		Symbol:      settings.Symbol,
		Name:        settings.Name,
		Description: settings.Description,
		URI:         settings.URI,
		URIHash:     settings.URIHash,
		Features:    settings.Features,
		RoyaltyRate: settings.RoyaltyRate,
	}); err != nil {
		return "", sdkerrors.Wrapf(types.ErrInvalidInput, "can't emit event EventClassIssued: %s", err)
	}

	return id, nil
}

// GetClassDefinition reruns the ClassDefinition.
func (k Keeper) GetClassDefinition(ctx sdk.Context, classID string) (types.ClassDefinition, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.CreateClassKey(classID))
	if bz == nil {
		return types.ClassDefinition{}, sdkerrors.Wrapf(types.ErrClassNotFound, "classID: %s", classID)
	}
	var definition types.ClassDefinition
	k.cdc.MustUnmarshal(bz, &definition)

	return definition, nil
}

// GetClassDefinitions returns all non-fungible class token definitions.
func (k Keeper) GetClassDefinitions(ctx sdk.Context, pagination *query.PageRequest) ([]types.ClassDefinition, *query.PageResponse, error) {
	definitionsPointers, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		prefix.NewStore(ctx.KVStore(k.storeKey), types.NFTClassKeyPrefix),
		pagination,
		// builder
		func(key []byte, definition *types.ClassDefinition) (*types.ClassDefinition, error) {
			return definition, nil
		},
		// constructor
		func() *types.ClassDefinition {
			return &types.ClassDefinition{}
		},
	)
	if err != nil {
		return nil, nil, err
	}

	definitions := make([]types.ClassDefinition, 0, len(definitionsPointers))
	for _, definition := range definitionsPointers {
		definitions = append(definitions, *definition)
	}

	return definitions, pageRes, err
}

// SetClassDefinition stores the ClassDefinition.
func (k Keeper) SetClassDefinition(ctx sdk.Context, definition types.ClassDefinition) {
	ctx.KVStore(k.storeKey).Set(types.CreateClassKey(definition.ID), k.cdc.MustMarshal(&definition))
}

// Mint mints new non-fungible token.
func (k Keeper) Mint(ctx sdk.Context, settings types.MintSettings) error {
	if err := types.ValidateTokenID(settings.ID); err != nil {
		return sdkerrors.Wrap(types.ErrInvalidInput, err.Error())
	}

	if err := types.ValidateData(settings.Data); err != nil {
		return sdkerrors.Wrap(types.ErrInvalidInput, err.Error())
	}

	definition, err := k.GetClassDefinition(ctx, settings.ClassID)
	if err != nil {
		return err
	}

	if !definition.IsIssuer(settings.Sender) {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "address %q is unauthorized to perform the mint operation", settings.Sender.String())
	}

	if !k.nftKeeper.HasClass(ctx, settings.ClassID) {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "classID %q not found", settings.ClassID)
	}

	if k.nftKeeper.HasNFT(ctx, settings.ClassID, settings.ID) {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "ID %q already defined for the class", settings.ID)
	}

	burnt, err := k.IsBurnt(ctx, settings.ClassID, settings.ID)
	if err != nil {
		return err
	}
	if burnt {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "ID %q has been burnt for the class", settings.ID)
	}

	params := k.GetParams(ctx)
	if params.MintFee.IsPositive() {
		coinsToBurn := sdk.NewCoins(params.MintFee)
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, settings.Sender, types.ModuleName, coinsToBurn); err != nil {
			return sdkerrors.Wrapf(err, "can't send coins from account %s to module %s", settings.Sender.String(), types.ModuleName)
		}
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coinsToBurn); err != nil {
			return sdkerrors.Wrapf(err, "can't burn %s for the module %s", coinsToBurn.String(), types.ModuleName)
		}
	}

	if err := k.nftKeeper.Mint(ctx, nft.NFT{
		ClassId: settings.ClassID,
		Id:      settings.ID,
		Uri:     settings.URI,
		UriHash: settings.URIHash,
		Data:    settings.Data,
	}, settings.Sender); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "can't save non-fungible token: %s", err)
	}

	return nil
}

// Burn burns non-fungible token.
func (k Keeper) Burn(ctx sdk.Context, owner sdk.AccAddress, classID, id string) error {
	ndfd, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return err
	}

	if err = ndfd.CheckFeatureAllowed(owner, types.ClassFeature_burning); err != nil {
		return err
	}

	if !k.nftKeeper.HasNFT(ctx, classID, id) {
		return sdkerrors.Wrapf(types.ErrNFTNotFound, "nft with classID:%s and ID:%s not found", classID, id)
	}

	if k.nftKeeper.GetOwner(ctx, classID, id).String() != owner.String() {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "only owner can burn the nft")
	}

	if err := k.checkBurnable(ctx, owner, ndfd, classID, id); err != nil {
		return err
	}

	// If the token is burnt the storage needs to be cleaned up.
	// We clean freezing because it's a single record only.
	// We don't clean whitelisting because potential number of records is unlimited.
	if err := k.SetFrozen(ctx, classID, id, false); err != nil {
		return err
	}

	if err := k.nftKeeper.Burn(ctx, classID, id); err != nil {
		return err
	}

	return k.SetBurnt(ctx, classID, id)
}

func (k Keeper) checkBurnable(ctx sdk.Context, owner sdk.AccAddress, ndfd types.ClassDefinition, classID, nftID string) error {
	frozen, err := k.IsFrozen(ctx, classID, nftID)
	if err != nil && !errors.Is(err, types.ErrFeatureDisabled) {
		return err
	}

	// non issuer is not allowed to burn frozen NFT, but the issuer can
	if frozen && owner.String() != ndfd.Issuer {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "frozen token cannot be burnt")
	}

	return nil
}

// IsBurnt return whether a non-fungible token is burnt or not.
func (k Keeper) IsBurnt(ctx sdk.Context, classID, nftID string) (bool, error) {
	key, err := types.CreateBurningKey(classID, nftID)
	if err != nil {
		return false, err
	}

	return bytes.Equal(ctx.KVStore(k.storeKey).Get(key), asset.StoreTrue), nil
}

// SetBurnt marks the nft burnt, but does not make any checks
// should not be used directly outside the module except for genesis.
func (k Keeper) SetBurnt(ctx sdk.Context, classID, nftID string) error {
	if k.nftKeeper.HasNFT(ctx, classID, nftID) {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "nft with classID:%s and ID:%s exists", classID, nftID)
	}
	burnt, err := k.IsBurnt(ctx, classID, nftID)
	if err != nil {
		return err
	}
	if burnt {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "nft with classID:%s and ID:%s has been already burnt", classID, nftID)
	}

	key, err := types.CreateBurningKey(classID, nftID)
	if err != nil {
		return err
	}
	ctx.KVStore(k.storeKey).Set(key, asset.StoreTrue)
	return nil
}

// GetBurntNFTs return paginated burnt NFTs.
//
//nolint:dupl
func (k Keeper) GetBurntNFTs(ctx sdk.Context, q *query.PageRequest) (*query.PageResponse, []types.BurntNFT, error) {
	mp := make(map[string][]string, 0)
	pageRes, err := query.Paginate(prefix.NewStore(ctx.KVStore(k.storeKey), types.NFTBurningKeyPrefix),
		q, func(key, value []byte) error {
			if !bytes.Equal(value, asset.StoreTrue) {
				return errors.Errorf("value stored in burning store is not %x, value %x", asset.StoreTrue, value)
			}
			classID, nftID, err := types.ParseBurningKey(key)
			if err != nil {
				return err
			}

			mp[classID] = append(mp[classID], nftID)
			return nil
		})
	if err != nil {
		return nil, nil, err
	}

	burnt := make([]types.BurntNFT, 0, len(mp))
	for classID, nfts := range mp {
		burnt = append(burnt, types.BurntNFT{
			ClassID: classID,
			NftIDs:  nfts,
		})
	}

	return pageRes, burnt, nil
}

// Freeze freezes a non-fungible token.
func (k Keeper) Freeze(ctx sdk.Context, sender sdk.AccAddress, classID, nftID string) error {
	classDefinition, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return err
	}

	if err = classDefinition.CheckFeatureAllowed(sender, types.ClassFeature_freezing); err != nil {
		return err
	}

	if !k.nftKeeper.HasNFT(ctx, classID, nftID) {
		return sdkerrors.Wrapf(types.ErrNFTNotFound, "nft with classID:%s and ID:%s not found", classID, nftID)
	}

	if err := k.SetFrozen(ctx, classID, nftID, true); err != nil {
		return err
	}

	owner := k.nftKeeper.GetOwner(ctx, classID, nftID)
	return ctx.EventManager().EmitTypedEvent(&types.EventFrozen{
		ClassId: classID,
		Id:      nftID,
		Owner:   owner.String(),
	})
}

// Unfreeze unfreezes a non-fungible token.
func (k Keeper) Unfreeze(ctx sdk.Context, sender sdk.AccAddress, classID, nftID string) error {
	classDefinition, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return err
	}

	if err = classDefinition.CheckFeatureAllowed(sender, types.ClassFeature_freezing); err != nil {
		return err
	}

	if !k.nftKeeper.HasNFT(ctx, classID, nftID) {
		return sdkerrors.Wrapf(types.ErrNFTNotFound, "nft with classID:%s and ID:%s not found", classID, nftID)
	}

	if err := k.SetFrozen(ctx, classID, nftID, false); err != nil {
		return err
	}

	owner := k.nftKeeper.GetOwner(ctx, classID, nftID)
	return ctx.EventManager().EmitTypedEvent(&types.EventUnfrozen{
		ClassId: classID,
		Id:      nftID,
		Owner:   owner.String(),
	})
}

// SetFrozen marks the nft frozen, but does not make any checks
// should not be used directly outside the module except for genesis.
func (k Keeper) SetFrozen(ctx sdk.Context, classID, nftID string, frozen bool) error {
	key, err := types.CreateFreezingKey(classID, nftID)
	if err != nil {
		return err
	}
	s := ctx.KVStore(k.storeKey)
	if frozen {
		s.Set(key, asset.StoreTrue)
	} else {
		s.Delete(key)
	}
	return nil
}

// IsFrozen return whether a non-fungible token is frozen or not.
func (k Keeper) IsFrozen(ctx sdk.Context, classID, nftID string) (bool, error) {
	classDefinition, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return false, err
	}

	if !classDefinition.IsFeatureEnabled(types.ClassFeature_freezing) {
		return false, sdkerrors.Wrapf(types.ErrFeatureDisabled, `feature "freezing" is disabled`)
	}

	if !k.nftKeeper.HasNFT(ctx, classID, nftID) {
		return false, sdkerrors.Wrapf(types.ErrNFTNotFound, "nft with classID:%s and ID:%s not found", classID, nftID)
	}

	key, err := types.CreateFreezingKey(classID, nftID)
	if err != nil {
		return false, err
	}

	return bytes.Equal(ctx.KVStore(k.storeKey).Get(key), asset.StoreTrue), nil
}

// GetFrozenNFTs return paginated frozen NFTs.
//
//nolint:dupl
func (k Keeper) GetFrozenNFTs(ctx sdk.Context, q *query.PageRequest) (*query.PageResponse, []types.FrozenNFT, error) {
	mp := make(map[string][]string, 0)
	pageRes, err := query.Paginate(prefix.NewStore(ctx.KVStore(k.storeKey), types.NFTFreezingKeyPrefix),
		q, func(key, value []byte) error {
			if !bytes.Equal(value, asset.StoreTrue) {
				return errors.Errorf("value stored in freezing store is not %x, value %x", asset.StoreTrue, value)
			}
			classID, nftID, err := types.ParseFreezingKey(key)
			if err != nil {
				return err
			}

			mp[classID] = append(mp[classID], nftID)
			return nil
		})
	if err != nil {
		return nil, nil, err
	}

	frozen := make([]types.FrozenNFT, 0, len(mp))
	for classID, nfts := range mp {
		frozen = append(frozen, types.FrozenNFT{
			ClassID: classID,
			NftIDs:  nfts,
		})
	}

	return pageRes, frozen, nil
}

// IsWhitelisted checks to see if an account is whitelisted for an NFT.
func (k Keeper) IsWhitelisted(ctx sdk.Context, classID, nftID string, account sdk.AccAddress) (bool, error) {
	classDefinition, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return false, err
	}

	if !classDefinition.IsFeatureEnabled(types.ClassFeature_whitelisting) {
		return false, sdkerrors.Wrapf(types.ErrFeatureDisabled, `feature "whitelisting" is disabled`)
	}

	if !k.nftKeeper.HasNFT(ctx, classID, nftID) {
		return false, sdkerrors.Wrapf(types.ErrNFTNotFound, "nft with classID:%s and ID:%s not found", classID, nftID)
	}

	key, err := types.CreateWhitelistingKey(classID, nftID, account)
	if err != nil {
		return false, err
	}

	return bytes.Equal(ctx.KVStore(k.storeKey).Get(key), asset.StoreTrue), nil
}

// GetAllWhitelistedAccountsForNFT returns all whitelisted accounts for all NFTs.
func (k Keeper) GetAllWhitelistedAccountsForNFT(ctx sdk.Context, classID, nftID string, q *query.PageRequest) (*query.PageResponse, []string, error) {
	if !k.nftKeeper.HasNFT(ctx, classID, nftID) {
		return nil, nil, sdkerrors.Wrapf(types.ErrNFTNotFound, "nft with classID:%s and ID:%s not found", classID, nftID)
	}

	compositeKey, err := store.JoinKeysWithLength([]byte(classID), []byte(nftID))
	if err != nil {
		return nil, nil, err
	}
	key := store.JoinKeys(types.NFTWhitelistingKeyPrefix, compositeKey)
	accounts := []string{}
	pageRes, err := query.Paginate(prefix.NewStore(ctx.KVStore(k.storeKey), key),
		q, func(key, value []byte) error {
			if !bytes.Equal(value, asset.StoreTrue) {
				return errors.Errorf("value stored in whitelisting store is not %x, value %x", asset.StoreTrue, value)
			}

			account := sdk.AccAddress(key[1:]) // the first byte contains the length prefix
			accounts = append(accounts, account.String())
			return nil
		})
	if err != nil {
		return nil, nil, err
	}

	return pageRes, accounts, nil
}

// GetAllWhitelisted returns all whitelisted accounts for all NFTs.
func (k Keeper) GetAllWhitelisted(ctx sdk.Context, q *query.PageRequest) (*query.PageResponse, []types.WhitelistedNFTAccounts, error) {
	type nftUniqueID struct {
		classID string
		nftID   string
	}
	mp := make(map[nftUniqueID][]string, 0)
	pageRes, err := query.Paginate(prefix.NewStore(ctx.KVStore(k.storeKey), types.NFTWhitelistingKeyPrefix),
		q, func(key, value []byte) error {
			if !bytes.Equal(value, asset.StoreTrue) {
				return errors.Errorf("value stored in whitelisting store is not %x, value %x", asset.StoreTrue, value)
			}
			classID, nftID, account, err := types.ParseWhitelistingKey(key)
			if err != nil {
				return err
			}
			if !k.nftKeeper.HasNFT(ctx, classID, nftID) {
				return nil
			}

			uniqueID := nftUniqueID{
				classID: classID,
				nftID:   nftID,
			}

			accountString := account.String()
			mp[uniqueID] = append(mp[uniqueID], accountString)
			return nil
		})
	if err != nil {
		return nil, nil, err
	}

	whitelisted := make([]types.WhitelistedNFTAccounts, 0, len(mp))
	for uniqueID, accounts := range mp {
		whitelisted = append(whitelisted, types.WhitelistedNFTAccounts{
			ClassID:  uniqueID.classID,
			NftID:    uniqueID.nftID,
			Accounts: accounts,
		})
	}

	return pageRes, whitelisted, nil
}

// AddToWhitelist adds an account to the whitelisted list of accounts for the NFT.
func (k Keeper) AddToWhitelist(ctx sdk.Context, classID, nftID string, sender, account sdk.AccAddress) error {
	classDefinition, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return err
	}

	if err = classDefinition.CheckFeatureAllowed(sender, types.ClassFeature_whitelisting); err != nil {
		return err
	}

	if !k.nftKeeper.HasNFT(ctx, classID, nftID) {
		return sdkerrors.Wrapf(types.ErrNFTNotFound, "nft with classID:%s and ID:%s not found", classID, nftID)
	}

	if classDefinition.Issuer == account.String() {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "setting whitelisting for the nft class issuer is forbidden")
	}

	if err := k.SetWhitelisting(ctx, classID, nftID, account, true); err != nil {
		return err
	}

	return ctx.EventManager().EmitTypedEvent(&types.EventAddedToWhitelist{
		ClassId: classID,
		Id:      nftID,
		Account: account.String(),
	})
}

// RemoveFromWhitelist removes an account from the whitelisted list of accounts for the NFT.
func (k Keeper) RemoveFromWhitelist(ctx sdk.Context, classID, nftID string, sender, account sdk.AccAddress) error {
	classDefinition, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return err
	}

	if err = classDefinition.CheckFeatureAllowed(sender, types.ClassFeature_whitelisting); err != nil {
		return err
	}

	if !k.nftKeeper.HasNFT(ctx, classID, nftID) {
		return sdkerrors.Wrapf(types.ErrNFTNotFound, "nft with classID:%s and ID:%s not found", classID, nftID)
	}

	if classDefinition.Issuer == account.String() {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "setting whitelisting for the nft class issuer is forbidden")
	}

	if err := k.SetWhitelisting(ctx, classID, nftID, account, false); err != nil {
		return err
	}

	return ctx.EventManager().EmitTypedEvent(&types.EventRemovedFromWhitelist{
		ClassId: classID,
		Id:      nftID,
		Account: account.String(),
	})
}

// SetWhitelisting adds an account to the whitelisting of the NFT, if whitelisting is true
// and removes it, if whitelisting is false.
func (k Keeper) SetWhitelisting(ctx sdk.Context, classID, nftID string, account sdk.AccAddress, whitelisting bool) error {
	key, err := types.CreateWhitelistingKey(classID, nftID, account)
	if err != nil {
		return err
	}
	s := ctx.KVStore(k.storeKey)
	if whitelisting {
		s.Set(key, asset.StoreTrue)
	} else {
		s.Delete(key)
	}
	return nil
}

func (k Keeper) isNFTSendable(ctx sdk.Context, classID, nftID string) error {
	classDefinition, err := k.GetClassDefinition(ctx, classID)
	// we return nil here, since we want the original tests of the nft module to pass, but they
	// fail if we return errors for unregistered NFTs on asset. Also the original nft module
	// does not have access to the asset module to register the NFTs
	if types.ErrClassNotFound.Is(err) {
		return nil
	}
	if err != nil {
		return err
	}

	// always allow issuer to send NFTs issued by them.
	owner := k.nftKeeper.GetOwner(ctx, classID, nftID)
	if classDefinition.Issuer == owner.String() {
		return nil
	}

	if classDefinition.IsFeatureEnabled(types.ClassFeature_disable_sending) {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "nft with classID:%s and ID:%s has sending disabled", classID, nftID)
	}

	frozen, err := k.IsFrozen(ctx, classID, nftID)
	if err != nil {
		if errors.Is(err, types.ErrFeatureDisabled) {
			return nil
		}
		return err
	}
	if frozen {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "nft with classID:%s and ID:%s is frozen", classID, nftID)
	}
	return nil
}

func (k Keeper) isNFTReceivable(ctx sdk.Context, classID, nftID string, receiver sdk.AccAddress) error {
	classDefinition, err := k.GetClassDefinition(ctx, classID)
	// we return nil here, since we want the original tests of the nft module to pass, but they
	// fail if we return errors for unregistered NFTs on asset. Also the original nft module
	// does not have access to the asset module to register the NFTs
	if types.ErrClassNotFound.Is(err) {
		return nil
	}
	if err != nil {
		return err
	}

	if !k.nftKeeper.HasNFT(ctx, classID, nftID) {
		return sdkerrors.Wrapf(types.ErrNFTNotFound, "nft with classID:%s and ID:%s not found", classID, nftID)
	}

	// always allow issuer to receive NFTs issued by them.
	if classDefinition.Issuer == receiver.String() {
		return nil
	}

	whitelisted, err := k.IsWhitelisted(ctx, classID, nftID, receiver)
	if err != nil {
		if errors.Is(err, types.ErrFeatureDisabled) {
			return nil
		}
		return err
	}
	if !whitelisted {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "nft with classID:%s and ID:%s is not whitelisted for account %s", classID, nftID, receiver)
	}
	return nil
}
