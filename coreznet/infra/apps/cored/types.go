package cored

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
)

// Wallet stores information related to wallet
type Wallet struct {
	// Name is the name of the key stored in keystore
	Name string

	// Address is the address of the wallet
	Address string
}

// Balance stores balance of denom
type Balance struct {
	// Amount is stored amount
	Amount *big.Int `json:"amount"`

	// Denom is a token symbol
	Denom string `json:"denom"`
}

// NewClient creates new client for cored
func NewClient(executor Executor, ip net.IP, rpcPort int) *Client {
	return &Client{
		executor: executor,
		ip:       ip,
		rpcPort:  rpcPort,
	}
}

// Client is the client for cored blockchain
type Client struct {
	executor Executor
	ip       net.IP
	rpcPort  int
}

// QBankBalances queries for bank balances owned by wallet
func (c *Client) QBankBalances(ctx context.Context, wallet Wallet) (map[string]Balance, error) {
	// FIXME (wojtek): support pagination
	out, err := c.executor.QBankBalances(ctx, wallet.Address, c.ip, c.rpcPort)
	if err != nil {
		return nil, err
	}
	data := struct {
		Balances []struct {
			Amount string `json:"amount"`
			Denom  string `json:"denom"`
		} `json:"balances"`
	}{}
	if err := json.Unmarshal(out, &data); err != nil {
		return nil, err
	}

	balances := map[string]Balance{}
	for _, b := range data.Balances {
		amount, ok := big.NewInt(0).SetString(b.Amount, 10)
		if !ok {
			panic(fmt.Sprintf("invalid amount %s received for denom %s on wallet %s", b.Amount, b.Denom, wallet.Address))
		}
		balances[b.Denom] = Balance{Amount: amount, Denom: b.Denom}
	}
	return balances, nil
}

// TxBankSend sends tokens from one wallet to another
func (c *Client) TxBankSend(ctx context.Context, sender, receiver Wallet, balance Balance) (string, error) {
	out, err := c.executor.TxBankSend(ctx, sender.Name, receiver.Address, balance, c.ip, c.rpcPort)
	if err != nil {
		return "", err
	}
	data := struct {
		TxHash string `json:"txhash"`
	}{}
	if err := json.Unmarshal(out, &data); err != nil {
		return "", err
	}
	return data.TxHash, nil
}
