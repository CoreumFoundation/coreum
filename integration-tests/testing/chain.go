package testing

import (
	"fmt"
	"math/rand"

	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

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

	// TODO: Migrate to keyring.Keyring after legacy types.Wallet is removed.
	Keyring keyring.Keyring
}

// RandomWallet generates a wallet for the chain with random name and
// private key and stores mnemonic in Keyring.
func (c Chain) RandomWallet() types.Wallet {
	tmp := fmt.Sprintf("tmp-%v", rand.Int())
	// we are using panics here, since we are sure it will not error out, and handling error
	// upstream is a waste of time.
	keyInfo, mnemonic, err := c.Keyring.NewMnemonic(tmp, keyring.English, "", "", hd.Secp256k1)
	if err != nil {
		panic(err)
	}
	//if err := c.Keyring.Delete(tmp); err != nil {
	//	panic(err)
	//}
	if _, err := c.Keyring.NewAccount(keyInfo.GetAddress().String(), mnemonic, "", "", hd.Secp256k1); err != nil {
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
		WithGasPrices(sdk.NewCoin(c.NetworkConfig.TokenSymbol, c.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice).String())
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
			// TODO
		}
	}

	return totalGasRequired
}
