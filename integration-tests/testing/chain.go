package testing

import (
	"reflect"

	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// Chain holds network and client for the blockchain
type Chain struct {
	Client        client.Client
	ClientContext cosmosclient.Context

	NetworkConfig app.NetworkConfig
	Faucet        Faucet

	Keyring keyring.Keyring
}

// RandomWallet generates a wallet for the chain with random name and
// private key and stores mnemonic in Keyring.
func (c Chain) RandomWallet() types.Wallet {
	// Generate and store a new mnemonic using temporary keyring
	keyInfo, mnemonic, err := keyring.NewInMemory().NewMnemonic("tmp", keyring.English, "", "", hd.Secp256k1)
	// we are using panics here, since we are sure it will not error out, and handling error
	// upstream is a waste of time.
	if err != nil {
		panic(err)
	}

	// Store generated mnemonic using account address as UID
	if _, err = c.Keyring.NewAccount(keyInfo.GetAddress().String(), mnemonic, "", "", hd.Secp256k1); err != nil {
		panic(err)
	}

	return types.Wallet{Name: keyInfo.GetAddress().String()}
}

// TxFactory returns factory with present values for the Chain.
func (c Chain) TxFactory() tx.Factory {
	return tx.Factory{}.
		WithKeybase(c.Keyring).
		WithChainID(string(c.NetworkConfig.ChainID)).
		WithTxConfig(c.ClientContext.TxConfig).
		WithGasPrices(c.NewCoin(c.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice).String())
}

// NewCoin helper function to initialize sdk.Coin by passing just amount.
// TODO: Use NewCoin instead of sdk.NewCoin & testing.MustNewCoin everywhere
func (c Chain) NewCoin(amount sdk.Int) sdk.Coin {
	return sdk.NewCoin(c.NetworkConfig.TokenSymbol, amount)
}

func (c Chain) GasLimitByMsgs(msgs ...sdk.Msg) uint64 {
	deterministicGas := NetworkConfig.Fee.DeterministicGas

	var totalGasRequired uint64 = 0
	for _, msg := range msgs {
		switch msg.(type) {
		case *banktypes.MsgSend:
			totalGasRequired += deterministicGas.BankSend
		default:
			panic(errors.Errorf("unsuported message type for deterministic gas: %v", reflect.TypeOf(msg).String()))
		}
	}

	return totalGasRequired
}
