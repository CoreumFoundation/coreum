package cored

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// NewClient creates new client for cored
func NewClient(executor Executor, chainID string, ip net.IP, rpcPort int) *Client {
	marshaler := NewCodec()
	return &Client{
		executor:         executor,
		chainID:          chainID,
		ip:               ip,
		rpcPort:          rpcPort,
		marshaler:        marshaler,
		txConfig:         NewTxConfig(marshaler),
		txBuilder:        NewTxBuilder(chainID),
		accountRetriever: authtypes.AccountRetriever{},
	}
}

// Client is the client for cored blockchain
type Client struct {
	executor Executor

	chainID          string
	ip               net.IP
	rpcPort          int
	marshaler        *codec.ProtoCodec
	txConfig         client.TxConfig
	txBuilder        *TxBuilder
	accountRetriever client.AccountRetriever
}

// QBankBalances queries for bank balances owned by wallet
func (c *Client) QBankBalances(ctx context.Context, wallet Wallet) (map[string]Balance, error) {
	cl, err := client.NewClientFromNode("tcp://" + net.JoinHostPort(c.ip.String(), strconv.Itoa(c.rpcPort)))
	must.OK(err)
	clientCtx := client.Context{
		InterfaceRegistry: c.marshaler.InterfaceRegistry(),
		Client:            cl,
	}
	qClient := banktypes.NewQueryClient(clientCtx)

	// FIXME (wojtek): support pagination
	resp, err := qClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: wallet.Key.Address()})
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
func (c *Client) TxBankSend(ctx context.Context, sender, receiver Wallet, balance Balance) (string, error) {
	fromAddress, err := sdk.AccAddressFromBech32(sender.Key.Address())
	must.OK(err)
	toAddress, err := sdk.AccAddressFromBech32(receiver.Key.Address())
	must.OK(err)
	msg := banktypes.NewMsgSend(fromAddress, toAddress, sdk.Coins{
		{
			Denom:  balance.Denom,
			Amount: sdk.NewIntFromBigInt(balance.Amount),
		},
	})

	cl, err := client.NewClientFromNode("tcp://" + net.JoinHostPort(c.ip.String(), strconv.Itoa(c.rpcPort)))
	must.OK(err)
	clientCtx := client.Context{
		InterfaceRegistry: c.marshaler.InterfaceRegistry(),
		Client:            cl,
	}

	accNum, accSeq, err := c.accountRetriever.GetAccountNumberSequence(clientCtx, fromAddress)
	if err != nil {
		return "", err
	}

	txResp, err := clientCtx.BroadcastTxCommit(must.Bytes(c.txConfig.TxEncoder()(c.txBuilder.Sign(sender.Key, accNum, accSeq, msg))))
	if err != nil {
		return "", err
	}

	if txResp.Code != 0 {
		return "", fmt.Errorf("trasaction failed: %s", txResp.RawLog)
	}
	return txResp.TxHash, nil
}
