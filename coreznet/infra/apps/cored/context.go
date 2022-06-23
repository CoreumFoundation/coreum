package cored

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

// NewContext creates a context required by other cosmos-sdk types
func NewContext(chainID string, rpcClient rpcclient.Client) client.Context {
	interfaceRegistry := types.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	stakingtypes.RegisterInterfaces(interfaceRegistry)

	codec := codec.NewProtoCodec(interfaceRegistry)
	return client.Context{
		ChainID:           chainID,
		Codec:             codec,
		InterfaceRegistry: interfaceRegistry,
		Client:            rpcClient,
		TxConfig:          tx.NewTxConfig(codec, []signing.SignMode{signing.SignMode_SIGN_MODE_DIRECT}),
	}
}
