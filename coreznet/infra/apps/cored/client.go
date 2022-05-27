package cored

import (
	"context"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// NewClient creates new client for cored
func NewClient(chainID string, addr string) Client {
	rpcClient, err := client.NewClientFromNode("tcp://" + addr)
	must.OK(err)
	clientCtx := NewContext(chainID, rpcClient)
	return Client{
		txManager:       NewTxManager(clientCtx),
		bankQueryClient: banktypes.NewQueryClient(clientCtx),
	}
}

// Client is the client for cored blockchain
type Client struct {
	txManager       TxManager
	bankQueryClient banktypes.QueryClient
}

// QBankBalances queries for bank balances owned by wallet
func (c Client) QBankBalances(ctx context.Context, wallet Wallet) (map[string]Balance, error) {
	// FIXME (wojtek): support pagination
	resp, err := c.bankQueryClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: wallet.Key.Address()})
	if err != nil {
		return nil, err
	}

	balances := map[string]Balance{}
	for _, b := range resp.Balances {
		balances[b.Denom] = Balance{Amount: b.Amount.BigInt(), Denom: b.Denom}
	}
	return balances, nil
}

// TxBankSend sends tokens from one wallet to another
func (c Client) TxBankSend(sender, receiver Wallet, balance Balance) (string, error) {
	fromAddress, err := sdk.AccAddressFromBech32(sender.Key.Address())
	must.OK(err)
	toAddress, err := sdk.AccAddressFromBech32(receiver.Key.Address())
	must.OK(err)

	return c.txManager.Broadcast(sender.Key, banktypes.NewMsgSend(fromAddress, toAddress, sdk.Coins{
		{
			Denom:  balance.Denom,
			Amount: sdk.NewIntFromBigInt(balance.Amount),
		},
	}))
}
