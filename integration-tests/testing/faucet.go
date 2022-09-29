package testing

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// FundedAccount represents a requirement of a test to get some funds for an account
type FundedAccount struct {
	Wallet types.Wallet
	Amount sdk.Coin
}

// NewFundedAccount is the constructor of FundedAccount
func NewFundedAccount(wallet types.Wallet, amount sdk.Coin) FundedAccount {
	return FundedAccount{
		Wallet: wallet,
		Amount: amount,
	}
}

// Faucet is the test chain faucet.
type Faucet struct {
	client        client.Client
	networkConfig config.NetworkConfig

	// muCh is used to serve the same purpose as `sync.Mutex` to protect `fundingWallet` against being used
	// to broadcast many transactions in parallel by different integration tests. The difference between this and `sync.Mutex`
	// is that test may exit immediately when `ctx` is canceled, without waiting for mutex to be unlocked.
	muCh          chan struct{}
	fundingWallet types.Wallet
}

// NewFaucet creates a new instance of the Faucet.
func NewFaucet(client client.Client, networkConfig config.NetworkConfig, fundingPrivKey types.Secp256k1PrivateKey) Faucet {
	fundingWallet := types.Wallet{Key: fundingPrivKey}
	faucet := Faucet{
		client:        client,
		networkConfig: networkConfig,
		muCh:          make(chan struct{}, 1),
		fundingWallet: fundingWallet,
	}
	faucet.muCh <- struct{}{}

	return faucet
}

// FundAccounts funds the list of the received wallets.
func (f Faucet) FundAccounts(ctx context.Context, accountsToFund ...FundedAccount) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-f.muCh:
		defer func() {
			f.muCh <- struct{}{}
		}()
	}

	gasPrice := sdk.NewDecCoinFromDec(f.networkConfig.TokenSymbol, f.networkConfig.Fee.FeeModel.Params().InitialGasPrice)

	log := logger.Get(ctx)
	log.Info("Funding accounts for test, it might take a while...")
	gasLimit := f.networkConfig.Fee.DeterministicGas.BankSend + f.networkConfig.Fee.DeterministicGas.FixedGas
	for _, toFund := range accountsToFund {
		// FIXME (wojtek): Fund all accounts in single tx once new "client" is ready
		encodedTx, err := f.client.PrepareTxBankSend(ctx, client.TxBankSendInput{
			Base: tx.BaseInput{
				Signer:   f.fundingWallet,
				GasLimit: gasLimit,
				GasPrice: gasPrice,
			},
			Sender:   f.fundingWallet,
			Receiver: toFund.Wallet,
			Amount:   toFund.Amount,
		})
		if err != nil {
			return err
		}
		if _, err := f.client.Broadcast(ctx, encodedTx); err != nil {
			return err
		}
		f.fundingWallet.AccountSequence++
	}
	log.Info("Test accounts funded")

	return nil
}
