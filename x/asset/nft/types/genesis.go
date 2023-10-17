package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesis returns the default NFT genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	for _, cd := range gs.ClassDefinitions {
		if err := cd.Validate(); err != nil {
			return err
		}
	}

	for _, frozen := range gs.FrozenNFTs {
		if err := frozen.Validate(); err != nil {
			return err
		}
	}

	for _, whitelisted := range gs.WhitelistedNFTAccounts {
		if err := whitelisted.Validate(); err != nil {
			return err
		}
	}

	for _, burnt := range gs.BurntNFTs {
		if err := burnt.Validate(); err != nil {
			return err
		}
	}

	return gs.Params.ValidateBasic()
}

// Validate performs basic validation on the fields of ClassDefinition.
func (nftd ClassDefinition) Validate() error {
	if _, _, err := DeconstructClassID(nftd.ID); err != nil {
		return err
	}

	return ValidateRoyaltyRate(nftd.RoyaltyRate)
}

// Validate performs basic validation on the fields of FrozenNFT.
func (f FrozenNFT) Validate() error {
	if _, _, err := DeconstructClassID(f.ClassID); err != nil {
		return err
	}

	for _, id := range f.NftIDs {
		if err := ValidateTokenID(id); err != nil {
			return err
		}
	}

	return nil
}

// Validate performs basic validation on the fields of WhitelistedNFTAccounts.
func (w WhitelistedNFTAccounts) Validate() error {
	if _, _, err := DeconstructClassID(w.ClassID); err != nil {
		return err
	}

	if err := ValidateTokenID(w.NftID); err != nil {
		return err
	}

	for _, acc := range w.Accounts {
		if _, err := sdk.AccAddressFromBech32(acc); err != nil {
			return err
		}
	}
	return nil
}

// Validate performs basic validation on the fields of WhitelistedNFTAccounts.
func (c ClassWhitelistedAccounts) Validate() error {
	if _, _, err := DeconstructClassID(c.ClassID); err != nil {
		return err
	}

	for _, acc := range c.Accounts {
		if _, err := sdk.AccAddressFromBech32(acc); err != nil {
			return err
		}
	}
	return nil
}

// Validate performs basic validation on the fields of BurntNFT.
func (b BurntNFT) Validate() error {
	if _, _, err := DeconstructClassID(b.ClassID); err != nil {
		return err
	}

	for _, id := range b.NftIDs {
		if err := ValidateTokenID(id); err != nil {
			return err
		}
	}

	return nil
}
