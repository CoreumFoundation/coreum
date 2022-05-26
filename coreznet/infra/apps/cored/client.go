package cored

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
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
	// FIXME (wojtek): support pagination
	out, err := c.executor.QBankBalances(ctx, wallet.Key.Address(), c.ip, c.rpcPort)
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
			panic(fmt.Sprintf("invalid amount %s received for denom %s on wallet %s", b.Amount, b.Denom, wallet.Key.Address()))
		}
		balances[b.Denom] = Balance{Amount: amount, Denom: b.Denom}
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
