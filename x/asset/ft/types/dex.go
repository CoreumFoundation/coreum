package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountToCoin is mapping of address and coin.
type AccountToCoin struct {
	Address sdk.AccAddress
	Coin    sdk.Coin
}

// CoinToSend represents a coin to be sent from one address to another.
type CoinToSend struct {
	FromAddress sdk.AccAddress
	ToAddress   sdk.AccAddress
	Coin        sdk.Coin
}

// DEXOrder is DEX order.
//
//nolint:tagliatelle
type DEXOrder struct {
	Creator    sdk.AccAddress `json:"creator"`
	Type       string         `json:"type"`
	ID         string         `json:"id"`
	Sequence   uint64         `json:"sequence"`
	BaseDenom  string         `json:"base_denom"`
	QuoteDenom string         `json:"quote_denom"`
	Price      *string        `json:"price,omitempty"` // might be nil
	Quantity   sdkmath.Int    `json:"quantity"`
	Side       string         `json:"side"`
}

// DEXActions is a set of DEX actions to be executed corresponding to one order.
type DEXActions struct {
	Order                     DEXOrder
	CreatorExpectedToSpend    sdk.Coin
	CreatorExpectedToReceive  sdk.Coin
	IncreaseLocked            []AccountToCoin
	DecreaseLocked            []AccountToCoin
	IncreaseExpectedToReceive []AccountToCoin
	DecreaseExpectedToReceive []AccountToCoin
	Send                      []CoinToSend
}

// NewDEXActions returns new instance of DEXActions.
func NewDEXActions(order DEXOrder) DEXActions {
	return DEXActions{
		Order:                     order,
		IncreaseLocked:            make([]AccountToCoin, 0),
		DecreaseLocked:            make([]AccountToCoin, 0),
		IncreaseExpectedToReceive: make([]AccountToCoin, 0),
		DecreaseExpectedToReceive: make([]AccountToCoin, 0),
		Send:                      make([]CoinToSend, 0),
	}
}

// AddCreatorExpectedToSpend adds the given coin to the CreatorExpectedToSpend field of DEXActions.
func (da *DEXActions) AddCreatorExpectedToSpend(coin sdk.Coin) {
	if da.CreatorExpectedToSpend.IsNil() {
		da.CreatorExpectedToSpend = coin
		return
	}
	da.CreatorExpectedToSpend = da.CreatorExpectedToSpend.Add(coin)
}

// AddCreatorExpectedToReceive adds the given coin to the CreatorExpectedToReceive field of DEXActions.
func (da *DEXActions) AddCreatorExpectedToReceive(coin sdk.Coin) {
	if da.CreatorExpectedToReceive.IsNil() {
		da.CreatorExpectedToReceive = coin
		return
	}
	da.CreatorExpectedToReceive = da.CreatorExpectedToReceive.Add(coin)
}

// AddIncreaseLocked adds the specified coin to the IncreaseLocked list for the given address.
func (da *DEXActions) AddIncreaseLocked(address sdk.AccAddress, coin sdk.Coin) {
	da.IncreaseLocked = appendOrAddToAccountsToCoin(da.IncreaseLocked, AccountToCoin{Address: address, Coin: coin})
}

// AddDecreaseLocked adds the specified coin to the DecreaseLocked list for the given address.
func (da *DEXActions) AddDecreaseLocked(address sdk.AccAddress, coin sdk.Coin) {
	da.DecreaseLocked = appendOrAddToAccountsToCoin(da.DecreaseLocked, AccountToCoin{Address: address, Coin: coin})
}

// AddIncreaseExpectedToReceive adds the specified coin to the IncreaseExpectedToReceive list for the given address.
func (da *DEXActions) AddIncreaseExpectedToReceive(address sdk.AccAddress, coin sdk.Coin) {
	da.IncreaseExpectedToReceive = appendOrAddToAccountsToCoin(
		da.IncreaseExpectedToReceive, AccountToCoin{Address: address, Coin: coin},
	)
}

// AddDecreaseExpectedToReceive adds the specified coin to the DecreaseExpectedToReceive list for the given address.
func (da *DEXActions) AddDecreaseExpectedToReceive(address sdk.AccAddress, coin sdk.Coin) {
	da.DecreaseExpectedToReceive = appendOrAddToAccountsToCoin(
		da.DecreaseExpectedToReceive, AccountToCoin{Address: address, Coin: coin},
	)
}

// AddSend appends a new CoinToSend to the Send list with the specified fromAddr, toAddr, and coin.
func (da *DEXActions) AddSend(fromAddr, toAddr sdk.AccAddress, coin sdk.Coin) {
	for i, send := range da.Send {
		if send.FromAddress.String() == fromAddr.String() &&
			send.ToAddress.String() == toAddr.String() &&
			send.Coin.Denom == coin.Denom {
			da.Send[i].Coin = send.Coin.Add(coin)
			return
		}
	}
	da.Send = append(da.Send, CoinToSend{FromAddress: fromAddr, ToAddress: toAddr, Coin: coin})
}

func appendOrAddToAccountsToCoin(accountsToCoin []AccountToCoin, accountToCoin AccountToCoin) []AccountToCoin {
	for i, item := range accountsToCoin {
		if item.Address.String() == accountToCoin.Address.String() &&
			item.Coin.Denom == accountToCoin.Coin.Denom {
			accountsToCoin[i].Coin = item.Coin.Add(accountToCoin.Coin)
			return accountsToCoin
		}
	}
	// append if not found
	accountsToCoin = append(accountsToCoin, accountToCoin)
	return accountsToCoin
}
