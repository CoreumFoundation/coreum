package types

import "errors"

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	for _, di := range gs.DelayedItems {
		if err := di.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Validate checks all the fields are valid.
func (di DelayedItem) Validate() error {
	if di.Id == "" {
		return errors.New("id is empty")
	}
	if di.Data == nil {
		return errors.New("data is nil")
	}
	if di.ExecutionTime.Unix() < 0 {
		return errors.New("unix timestamp of the execution time must be non-negative")
	}
	return nil
}
