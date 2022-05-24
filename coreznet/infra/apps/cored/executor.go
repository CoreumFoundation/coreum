package cored

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	osexec "os/exec"
	"strings"

	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
)

// NewExecutor returns new executor
func NewExecutor(name, binPath, homeDir, keyName string) *Executor {
	must.Any(os.Stat(binPath))

	return &Executor{
		name:    name,
		binPath: binPath,
		homeDir: homeDir,
		keyName: keyName,
	}
}

// Executor exposes methods for executing cored binary
type Executor struct {
	name    string
	binPath string
	homeDir string
	keyName string
}

// Name returns name of the chain
func (e *Executor) Name() string {
	return e.name
}

// Bin returns path to cored binary
func (e *Executor) Bin() string {
	return e.binPath
}

// Home returns path to home dir
func (e *Executor) Home() string {
	return e.homeDir
}

// AddKey adds key to the client
func (e *Executor) AddKey(ctx context.Context, name string) (string, error) {
	keyData := &bytes.Buffer{}
	addrBuf := &bytes.Buffer{}

	err := libexec.Exec(ctx,
		e.coredOut(keyData, "keys", "add", name, "--output", "json", "--keyring-backend", "test"),
		e.coredOut(addrBuf, "keys", "show", name, "-a", "--keyring-backend", "test"),
	)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(addrBuf.String(), "\n"), ioutil.WriteFile(e.homeDir+"/"+name+".json", keyData.Bytes(), 0o600)
}

// PrepareNode prepares node to start
func (e *Executor) PrepareNode(ctx context.Context, genesis *Genesis) error {
	addr, err := e.AddKey(ctx, e.keyName)
	if err != nil {
		return err
	}

	cmds := []*osexec.Cmd{
		e.cored("init", e.name, "--chain-id", e.name, "-o"),
		e.cored("add-genesis-account", addr, "500000000000000000000000core,990000000000000000000000000stake", "--keyring-backend", "test"),
	}
	for wallet, balances := range genesis.wallets {
		if len(balances) == 0 {
			continue
		}
		balancesStr := ""
		for _, balance := range balances {
			if balancesStr != "" {
				balancesStr += ","
			}
			balancesStr += balance.Amount.String() + balance.Denom
		}
		cmds = append(cmds, e.cored("add-genesis-account", wallet.Address, balancesStr, "--keyring-backend", "test"))
	}

	cmds = append(cmds,
		e.cored("gentx", e.keyName, "1000000000000000000000000stake", "--chain-id", e.name, "--keyring-backend", "test"),
		e.cored("collect-gentxs"),
	)
	return libexec.Exec(ctx, cmds...)
}

// QBankBalances queries for bank balances owned by address
func (e *Executor) QBankBalances(ctx context.Context, address string, ip net.IP) ([]byte, error) {
	balances := &bytes.Buffer{}
	if err := libexec.Exec(ctx, e.coredOut(balances, "q", "bank", "balances", address, "--chain-id", e.name, "--node", fmt.Sprintf("tcp://%s:26657", ip), "--output", "json")); err != nil {
		return nil, err
	}
	return balances.Bytes(), nil
}

// TxBankSend sends tokens from one address to another
func (e *Executor) TxBankSend(ctx context.Context, sender, address string, balance Balance, ip net.IP) ([]byte, error) {
	tx := &bytes.Buffer{}
	if err := libexec.Exec(ctx, e.coredOut(tx, "tx", "bank", "send", sender, address, balance.Amount.String()+balance.Denom, "--yes", "--chain-id", e.name, "--node", fmt.Sprintf("tcp://%s:26657", ip), "--keyring-backend", "test", "--broadcast-mode", "block", "--output", "json")); err != nil {
		return nil, err
	}
	return tx.Bytes(), nil
}

func (e *Executor) cored(args ...string) *osexec.Cmd {
	return osexec.Command(e.binPath, append([]string{"--home", e.homeDir}, args...)...)
}

func (e *Executor) coredOut(buf *bytes.Buffer, args ...string) *osexec.Cmd {
	cmd := e.cored(args...)
	cmd.Stdout = buf
	return cmd
}
