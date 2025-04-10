package types

import (
	"context"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
)

// AccountKeeper defines the expected account keeper interface.
type AccountKeeper interface {
	//nolint:inamedparam // the sdk interface
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI
}

// BankKeeper defines the expected bank interface.
type BankKeeper interface {
	GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool)
	SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata)
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amounts sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amounts sdk.Coins) error
	SendCoinsFromModuleToAccount(
		ctx context.Context,
		senderModule string,
		recipientAddr sdk.AccAddress,
		amt sdk.Coins,
	) error
	SendCoinsFromAccountToModule(
		ctx context.Context,
		senderAddr sdk.AccAddress,
		recipientModule string,
		amt sdk.Coins,
	) error
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	LockedCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	HasSupply(ctx context.Context, denom string) bool
}

// DelayKeeper defines methods required from the delay keeper.
type DelayKeeper interface {
	DelayExecution(ctx sdk.Context, id string, data proto.Message, delay time.Duration) error
}

// StakingKeeper defines the expected staking interface.
type StakingKeeper interface {
	GetParams(ctx context.Context) (params stakingtypes.Params, err error)
}

// WasmPermissionedKeeper defines methods required from the WASM permissioned keeper.
type WasmPermissionedKeeper interface {
	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
	Instantiate2(
		ctx sdk.Context,
		codeID uint64,
		creator, admin sdk.AccAddress,
		initMsg []byte,
		label string,
		deposit sdk.Coins,
		salt []byte,
		fixMsg bool,
	) (sdk.AccAddress, []byte, error)
	Create(
		ctx sdk.Context,
		creator sdk.AccAddress,
		wasmCode []byte,
		instantiateAccess *wasmtypes.AccessConfig,
	) (codeID uint64, checksum []byte, err error)
}
