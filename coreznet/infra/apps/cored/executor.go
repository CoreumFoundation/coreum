package cored

import (
	"bytes"
	"context"
	"net"
	"os"
	osexec "os/exec"
	"strconv"

	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
)

// NewExecutor returns new executor
func NewExecutor(chainID, binPath, homeDir string) Executor {
	must.Any(os.Stat(binPath))

	return Executor{
		chainID: chainID,
		binPath: binPath,
		homeDir: homeDir,
	}
}

// Executor exposes methods for executing cored binary
type Executor struct {
	chainID string
	binPath string
	homeDir string
}

// Bin returns path to cored binary
func (e Executor) Bin() string {
	return e.binPath
}

// Home returns path to home dir
func (e Executor) Home() string {
	return e.homeDir
}

// QBankBalances queries for bank balances owned by address
func (e Executor) QBankBalances(ctx context.Context, address string, ip net.IP, rpcPort int) ([]byte, error) {
	balances := &bytes.Buffer{}
	if err := libexec.Exec(ctx, e.coredOut(balances, "q", "bank", "balances", address, "--chain-id", e.chainID, "--node", "tcp://"+net.JoinHostPort(ip.String(), strconv.Itoa(rpcPort)), "--output", "json")); err != nil {
		return nil, err
	}
	return balances.Bytes(), nil
}

// TxBankSend sends tokens from one address to another
func (e Executor) TxBankSend(ctx context.Context, sender, address string, balance Balance, ip net.IP, rpcPort int) ([]byte, error) {
	tx := &bytes.Buffer{}
	if err := libexec.Exec(ctx, e.coredOut(tx, "tx", "bank", "send", sender, address, balance.Amount.String()+balance.Denom, "--yes", "--chain-id", e.chainID, "--node", "tcp://"+net.JoinHostPort(ip.String(), strconv.Itoa(rpcPort)), "--keyring-backend", "test", "--broadcast-mode", "block", "--output", "json")); err != nil {
		return nil, err
	}
	return tx.Bytes(), nil
}

func (e Executor) cored(args ...string) *osexec.Cmd {
	return osexec.Command(e.binPath, append([]string{"--home", e.homeDir}, args...)...)
}

func (e Executor) coredOut(buf *bytes.Buffer, args ...string) *osexec.Cmd {
	cmd := e.cored(args...)
	cmd.Stdout = buf
	return cmd
}
