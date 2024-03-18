package types

// DenomOffered returns the offered denom.
func (o *OrderLimit) DenomOffered() string {
	return o.OfferedAmount.Denom
}

// DenomRequested returns the requested denom.
func (o *OrderLimit) DenomRequested() string {
	return o.SellPrice.Denom
}
