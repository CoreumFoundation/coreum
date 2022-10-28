package types

// FTHasOption checks weather an option is enable on a list of token options
func FTHasOption(options []FungibleTokenOption, option FungibleTokenOption) error {
	for _, o := range options {
		if o == option {
			return nil
		}
	}
	return ErrOptionNotActive
}
