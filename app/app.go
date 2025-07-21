package app

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/evidence"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/nft"
	nftkeeper "cosmossdk.io/x/nft/keeper"
	"cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	abci "github.com/cometbft/cometbft/abci/types"
	tmos "github.com/cometbft/cometbft/libs/os"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server"
	serverapi "github.com/cosmos/cosmos-sdk/server/api"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	tx "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/keeper"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"
	ibchooks "github.com/cosmos/ibc-apps/modules/ibc-hooks/v8"
	ibchookskeeper "github.com/cosmos/ibc-apps/modules/ibc-hooks/v8/keeper"
	ibchookstypes "github.com/cosmos/ibc-apps/modules/ibc-hooks/v8/types"
	"github.com/cosmos/ibc-go/modules/capability"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"
	icacontroller "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	ibcporttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	ibclocalhost "github.com/cosmos/ibc-go/v8/modules/light-clients/09-localhost"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/CoreumFoundation/coreum/v6/app/openapi"
	appupgrade "github.com/CoreumFoundation/coreum/v6/app/upgrade"
	appupgradev6 "github.com/CoreumFoundation/coreum/v6/app/upgrade/v6"
	"github.com/CoreumFoundation/coreum/v6/docs"
	"github.com/CoreumFoundation/coreum/v6/pkg/config"
	"github.com/CoreumFoundation/coreum/v6/pkg/config/constant"
	assetft "github.com/CoreumFoundation/coreum/v6/x/asset/ft"
	assetftkeeper "github.com/CoreumFoundation/coreum/v6/x/asset/ft/keeper"
	assetfttypes "github.com/CoreumFoundation/coreum/v6/x/asset/ft/types"
	assetnft "github.com/CoreumFoundation/coreum/v6/x/asset/nft"
	assetnftkeeper "github.com/CoreumFoundation/coreum/v6/x/asset/nft/keeper"
	assetnfttypes "github.com/CoreumFoundation/coreum/v6/x/asset/nft/types"
	"github.com/CoreumFoundation/coreum/v6/x/auth/ante"
	"github.com/CoreumFoundation/coreum/v6/x/customparams"
	customparamskeeper "github.com/CoreumFoundation/coreum/v6/x/customparams/keeper"
	customparamstypes "github.com/CoreumFoundation/coreum/v6/x/customparams/types"
	"github.com/CoreumFoundation/coreum/v6/x/delay"
	delaykeeper "github.com/CoreumFoundation/coreum/v6/x/delay/keeper"
	delaytypes "github.com/CoreumFoundation/coreum/v6/x/delay/types"
	"github.com/CoreumFoundation/coreum/v6/x/deterministicgas"
	deterministicgastypes "github.com/CoreumFoundation/coreum/v6/x/deterministicgas/types"
	"github.com/CoreumFoundation/coreum/v6/x/dex"
	dexkeeper "github.com/CoreumFoundation/coreum/v6/x/dex/keeper"
	dextypes "github.com/CoreumFoundation/coreum/v6/x/dex/types"
	"github.com/CoreumFoundation/coreum/v6/x/feemodel"
	feemodelkeeper "github.com/CoreumFoundation/coreum/v6/x/feemodel/keeper"
	feemodeltypes "github.com/CoreumFoundation/coreum/v6/x/feemodel/types"
	wasmcustomhandler "github.com/CoreumFoundation/coreum/v6/x/wasm/handler"
	cwasmtypes "github.com/CoreumFoundation/coreum/v6/x/wasm/types"
	"github.com/CoreumFoundation/coreum/v6/x/wbank"
	wbankkeeper "github.com/CoreumFoundation/coreum/v6/x/wbank/keeper"
	"github.com/CoreumFoundation/coreum/v6/x/wibctransfer"
	wibctransferkeeper "github.com/CoreumFoundation/coreum/v6/x/wibctransfer/keeper"
	"github.com/CoreumFoundation/coreum/v6/x/wnft"
	wnftkeeper "github.com/CoreumFoundation/coreum/v6/x/wnft/keeper"
	"github.com/CoreumFoundation/coreum/v6/x/wstaking"
)

const (
	// Name is the blockchain name.
	Name = "core"

	// DefaultChainID is the default chain id of the network.
	DefaultChainID = constant.ChainIDMain
)

// ChosenNetwork is a hacky solution to pass network config
// from cmd package to app.
var ChosenNetwork config.NetworkConfig

var (
	// DefaultNodeHome default home directories for the application daemon.
	DefaultNodeHome string

	// module account permissions.
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		icatypes.ModuleName:            nil,
		wasmtypes.ModuleName:           {authtypes.Burner},
		assetfttypes.ModuleName:        {authtypes.Minter, authtypes.Burner},
		assetnfttypes.ModuleName:       {authtypes.Burner},
		// the line is required by the nft module to have the module account stored in the account keeper
		nft.ModuleName: {},
	}
)

var (
	_ runtime.AppI            = (*App)(nil)
	_ servertypes.Application = (*App)(nil)
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+Name)
}

// mergeProtos used to make sure proto files are merged only once.
var mergeProtos sync.Once

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry types.InterfaceRegistry

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper    authkeeper.AccountKeeper
	AuthzKeeper      authzkeeper.Keeper
	CapabilityKeeper *capabilitykeeper.Keeper
	StakingKeeper    *stakingkeeper.Keeper
	SlashingKeeper   slashingkeeper.Keeper
	MintKeeper       mintkeeper.Keeper
	DistrKeeper      distrkeeper.Keeper
	GovKeeper        govkeeper.Keeper
	CrisisKeeper     *crisiskeeper.Keeper
	UpgradeKeeper    *upgradekeeper.Keeper
	ParamsKeeper     paramskeeper.Keeper
	// IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	IBCKeeper              *ibckeeper.Keeper
	IBCHooksKeeper         ibchookskeeper.Keeper
	PacketForwardKeeper    *packetforwardkeeper.Keeper
	ICAHostKeeper          icahostkeeper.Keeper
	ICAControllerKeeper    icacontrollerkeeper.Keeper
	TransferKeeper         wibctransferkeeper.TransferKeeperWrapper
	EvidenceKeeper         evidencekeeper.Keeper
	FeeGrantKeeper         feegrantkeeper.Keeper
	ConsensusParamsKeeper  consensusparamkeeper.Keeper
	WasmKeeper             wasmkeeper.Keeper
	WasmPermissionedKeeper *wasmkeeper.PermissionedKeeper
	GroupKeeper            groupkeeper.Keeper

	AssetFTKeeper      assetftkeeper.Keeper
	AssetNFTKeeper     assetnftkeeper.Keeper
	FeeModelKeeper     feemodelkeeper.Keeper
	BankKeeper         wbankkeeper.BaseKeeperWrapper
	NFTKeeper          wnftkeeper.Wrapper
	CustomParamsKeeper customparamskeeper.Keeper
	DelayKeeper        delaykeeper.Keeper
	DEXKeeper          dexkeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper           capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper      capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper       capabilitykeeper.ScopedKeeper
	ScopedICAControllerKeeper capabilitykeeper.ScopedKeeper
	ScopedWASMKeeper          capabilitykeeper.ScopedKeeper

	// ModuleManager is the module manager
	ModuleManager      *module.Manager
	BasicModuleManager module.BasicManager

	// sm is the simulation manager
	sm *module.SimulationManager

	configurator module.Configurator

	// IBC Hooks.
	Ics20WasmHooks   *ibchooks.WasmHooks
	HooksICS4Wrapper ibchooks.ICS4Middleware
}

// New returns a reference to an initialized blockchain app.
//
//nolint:funlen // Disable linting for code generated by Cosmos SDK
func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	deterministicGasConfig := deterministicgas.DefaultConfig()
	addressPrefix := ChosenNetwork.Provider.GetAddressPrefix()
	interfaceRegistry, _ := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec: address.Bech32Codec{
				Bech32Prefix: addressPrefix,
			},
			ValidatorAddressCodec: address.Bech32Codec{
				Bech32Prefix: config.ValPrefixFromAddressPrefix(addressPrefix),
			},
		},
	})
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()
	txConfig := authtx.NewTxConfig(appCodec, authtx.DefaultSignModes)

	std.RegisterLegacyAminoCodec(legacyAmino)
	std.RegisterInterfaces(interfaceRegistry)
	// Since 0.47 all ibc clients must be registered explicitly and are not registered automatically.
	// we need to register localhost client since we use cosmos relayer in our integration tests and it
	// relies on localhost client to be registered.
	ibclocalhost.RegisterInterfaces(interfaceRegistry)

	baseAppOptions = append(baseAppOptions, baseapp.SetOptimisticExecution())

	bApp := baseapp.NewBaseApp(Name, logger, db, txConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, authz.ModuleName, banktypes.StoreKey,
		stakingtypes.StoreKey, crisistypes.StoreKey, minttypes.StoreKey,
		distrtypes.StoreKey, slashingtypes.StoreKey, govtypes.StoreKey,
		paramstypes.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey,
		evidencetypes.StoreKey, capabilitytypes.StoreKey, consensusparamtypes.StoreKey,
		wasmtypes.StoreKey, feemodeltypes.StoreKey, assetfttypes.StoreKey,
		assetnfttypes.StoreKey, nftkeeper.StoreKey, ibcexported.StoreKey,
		ibctransfertypes.StoreKey, ibchookstypes.StoreKey, packetforwardtypes.StoreKey,
		icahosttypes.StoreKey, icacontrollertypes.StoreKey, delaytypes.StoreKey,
		customparamstypes.StoreKey, group.StoreKey, dextypes.StoreKey,
	)
	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey, feemodeltypes.TransientStoreKey)
	memKeys := storetypes.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	app := &App{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		txConfig:          txConfig,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	app.ParamsKeeper = initParamsKeeper(
		appCodec,
		legacyAmino,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)

	// set the BaseApp's parameter store
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		runtime.EventService{},
	)
	bApp.SetParamStore(&app.ConsensusParamsKeeper.ParamsStore)

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)

	// grant capabilities for the ibc and ibc-transfer modules
	app.ScopedIBCKeeper = app.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	app.ScopedICAHostKeeper = app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	app.ScopedICAControllerKeeper = app.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	app.ScopedTransferKeeper = app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	app.ScopedWASMKeeper = app.CapabilityKeeper.ScopeToModule(wasmtypes.ModuleName)
	app.CapabilityKeeper.Seal()

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		interfaceRegistry.SigningContext().AddressCodec(),
		addressPrefix,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.AuthzKeeper = authzkeeper.NewKeeper(
		runtime.NewKVStoreService(keys[authzkeeper.StoreKey]),
		appCodec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
	)

	delayRouter := delaytypes.NewRouter()
	app.DelayKeeper = delaykeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[delaytypes.StoreKey]),
		delayRouter,
		app.interfaceRegistry,
	)

	originalBankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		app.AccountKeeper,
		app.ModuleAccountAddrs(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		logger,
	)
	app.WasmPermissionedKeeper = wasmkeeper.NewGovPermissionKeeper(&app.WasmKeeper)
	if err := delayRouter.RegisterHandler(
		&assetfttypes.DelayedTokenUpgradeV1{},
		assetftkeeper.NewDelayTokenUpgradeV1Handler(app.AssetFTKeeper),
	); err != nil {
		panic(err)
	}

	app.BankKeeper = wbankkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		app.AccountKeeper,
		// pointer is used here because there is cycle in keeper dependencies:
		// AssetFTKeeper -> WasmKeeper -> BankKeeper -> AssetFTKeeper
		&app.WasmKeeper,
		app.ModuleAccountAddrs(),
		// pointer is used here because there is cycle in keeper dependencies
		&app.AssetFTKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		logger,
	)

	enabledSignModes := make([]signingtypes.SignMode, 0)
	enabledSignModes = append(enabledSignModes, authtx.DefaultSignModes...)
	enabledSignModes = append(enabledSignModes, signingtypes.SignMode_SIGN_MODE_TEXTUAL)
	txConfigOpts := authtx.ConfigOptions{
		EnabledSignModes:           enabledSignModes,
		TextualCoinMetadataQueryFn: tx.NewBankKeeperCoinMetadataQueryFn(app.BankKeeper),
	}
	var err error
	txConfig, err = authtx.NewTxConfigWithOptions(
		appCodec,
		txConfigOpts,
	)
	if err != nil {
		panic(err.Error())
	}

	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		interfaceRegistry.SigningContext().ValidatorAddressCodec(),
		address.NewBech32Codec(config.ConsPrefixFromAddressPrefix(addressPrefix)),
	)

	app.AssetFTKeeper = assetftkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[assetfttypes.StoreKey]),
		// for the assetft we use the clear bank keeper without the assets integration to prevent cycling calls.
		originalBankKeeper,
		app.DelayKeeper,
		app.StakingKeeper,
		// pointer is used here because there is cycle in keeper dependencies:
		// AssetFTKeeper -> WasmKeeper -> BankKeeper -> AssetFTKeeper
		&app.WasmKeeper,
		app.WasmPermissionedKeeper,
		&app.AccountKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[minttypes.StoreKey]),
		app.StakingKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[distrtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		runtime.NewKVStoreService(keys[slashingtypes.StoreKey]),
		app.StakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	invCheckPeriod := cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod))
	app.CrisisKeeper = crisiskeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[crisistypes.StoreKey]),
		invCheckPeriod,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.AccountKeeper.AddressCodec(),
	)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[feegrant.StoreKey]),
		app.AccountKeeper,
	)

	// get skipUpgradeHeights from the app options
	skipUpgradeHeights := map[int64]bool{}
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}
	homePath := cast.ToString(appOpts.Get(flags.FlagHome))

	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
		appCodec,
		homePath,
		app.BaseApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks()),
	)

	app.FeeModelKeeper = feemodelkeeper.NewKeeper(
		runtime.NewKVStoreService(keys[feemodeltypes.StoreKey]),
		runtime.NewTransientStoreService(tkeys[feemodeltypes.TransientStoreKey]),
		appCodec,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.CustomParamsKeeper = customparamskeeper.NewKeeper(
		runtime.NewKVStoreService(keys[customparamstypes.StoreKey]),
		appCodec,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		keys[ibcexported.StoreKey],
		app.GetSubspace(ibcexported.ModuleName),
		app.StakingKeeper,
		app.UpgradeKeeper,
		app.ScopedIBCKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	nftKeeper := nftkeeper.NewKeeper(
		runtime.NewKVStoreService(keys[nftkeeper.StoreKey]),
		appCodec,
		app.AccountKeeper,
		app.BankKeeper,
	)
	app.AssetNFTKeeper = assetnftkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[assetnfttypes.StoreKey]),
		nftKeeper,
		// for the assetnft we use the clear bank keeper without the assets integration
		// because it interacts only with native token.
		originalBankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.NFTKeeper = wnftkeeper.NewWrappedNFTKeeper(nftKeeper, app.AssetNFTKeeper)

	// IBC Hooks.
	// The contract WASM keeper needs to be set later since it depends on WASM hooks.
	wasmHooks := ibchooks.NewWasmHooks(&app.IBCHooksKeeper, nil, addressPrefix)
	app.Ics20WasmHooks = &wasmHooks
	app.HooksICS4Wrapper = ibchooks.NewICS4Middleware(
		app.IBCKeeper.ChannelKeeper,
		app.Ics20WasmHooks,
	)

	// Packet Forward Middleware.
	app.PacketForwardKeeper = packetforwardkeeper.NewKeeper(
		appCodec,
		keys[packetforwardtypes.StoreKey],
		nil, // will be zero-value here, reference is set later on with SetTransferKeeper.
		app.IBCKeeper.ChannelKeeper,
		app.BankKeeper,
		app.HooksICS4Wrapper, // Wrap IBC hooks with PFM.
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Create Transfer Keepers
	app.TransferKeeper = wibctransferkeeper.NewTransferKeeperWrapper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.PacketForwardKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		app.ScopedTransferKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.PacketForwardKeeper.SetTransferKeeper(app.TransferKeeper)

	app.IBCHooksKeeper = ibchookskeeper.NewKeeper(
		app.keys[ibchookstypes.StoreKey],
	)

	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec,
		app.keys[icahosttypes.StoreKey],
		app.GetSubspace(icahosttypes.SubModuleName),
		app.HooksICS4Wrapper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.ScopedICAHostKeeper,
		bApp.MsgServiceRouter(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.ICAHostKeeper.WithQueryRouter(app.GRPCQueryRouter())

	app.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
		appCodec,
		app.keys[icacontrollertypes.StoreKey],
		app.GetSubspace(icacontrollertypes.SubModuleName),
		app.HooksICS4Wrapper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.ScopedICAControllerKeeper,
		bApp.MsgServiceRouter(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Create evidence Keeper for to register the IBC light client misbehaviour evidence route
	govConfig := govtypes.DefaultConfig()
	/*
		Example of setting gov params:
		govConfig.MaxMetadataLen = 10000
	*/
	govKeeper := govkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[govtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.DistrKeeper,
		app.MsgServiceRouter(),
		govConfig,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.GovKeeper = *govKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
		// register the governance hooks
		),
	)

	groupConfig := group.DefaultConfig()
	/*
		Example of setting group params:
		groupConfig.MaxExecutionPeriod = 2 * time.Hour * 24 // 2 days
		groupConfig.MaxMetadataLen = 1000
	*/
	app.GroupKeeper = groupkeeper.NewKeeper(
		keys[group.StoreKey],
		appCodec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
		groupConfig,
	)

	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[evidencetypes.StoreKey]),
		app.StakingKeeper,
		app.SlashingKeeper,
		app.AccountKeeper.AddressCodec(),
		runtime.ProvideCometInfoService(),
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	wasmDir := filepath.Join(homePath, "wasm-data")
	wasmNodeConfig, err := wasm.ReadNodeConfig(appOpts)
	if err != nil {
		panic(errors.Wrapf(err, "error while reading wasm node config"))
	}
	wasmVMConfig := wasmtypes.VMConfig{}

	wasmOpts := []wasmkeeper.Option{
		wasmkeeper.WithAcceptedAccountTypesOnContractInstantiation(
			&authtypes.BaseAccount{},
			&vestingtypes.ContinuousVestingAccount{},
			&vestingtypes.DelayedVestingAccount{},
			&vestingtypes.PeriodicVestingAccount{},
			&vestingtypes.BaseVestingAccount{},
			&vestingtypes.PermanentLockedAccount{},
		),
		wasmkeeper.WithAccountPruner(cwasmtypes.AccountPruner{}),
		wasmkeeper.WithCoinTransferrer(cwasmtypes.NewBankCoinTransferrer(app.BankKeeper)),
		wasmkeeper.WithMessageHandler(wasmcustomhandler.NewMessengerWrapper(wasmkeeper.NewDefaultMessageHandler(
			&app.WasmKeeper,
			app.MsgServiceRouter(),
			app.IBCKeeper.ChannelKeeper,
			app.IBCKeeper.ChannelKeeper,
			app.ScopedWASMKeeper,
			app.BankKeeper,
			appCodec,
			&app.TransferKeeper,
			wasmcustomhandler.NewCoreumMsgHandler(),
		))),
		wasmkeeper.WithQueryPlugins(wasmcustomhandler.NewCoreumQueryHandler(
			assetftkeeper.NewQueryService(app.AssetFTKeeper, app.BankKeeper),
			assetnftkeeper.NewQueryService(app.AssetNFTKeeper),
			app.NFTKeeper, app.GRPCQueryRouter(), appCodec,
		)),
	}

	if cast.ToBool(appOpts.Get("telemetry.enabled")) {
		wasmOpts = append(wasmOpts, wasmkeeper.WithVMCacheMetrics(prometheus.DefaultRegisterer))
	}

	app.WasmKeeper = wasmkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[wasmtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		distrkeeper.NewQuerier(app.DistrKeeper),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.ScopedWASMKeeper,
		&app.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		wasmDir,
		wasmNodeConfig,
		wasmVMConfig,
		wasmkeeper.BuiltInCapabilities(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		wasmOpts...,
	)

	// Set WASM keeper in WASM hooks.
	app.Ics20WasmHooks.ContractKeeper = &app.WasmKeeper

	// IBC transfer stack contains (from top to bottom):
	// - wibctransfer
	// - packetforward
	// - ibchooks
	// - ibctransfer
	var ibcTransferStack ibcporttypes.IBCModule
	ibcTransferStack = transfer.NewIBCModule(app.TransferKeeper.Keeper)
	ibcTransferStack = ibchooks.NewIBCMiddleware(ibcTransferStack, &app.HooksICS4Wrapper)
	ibcTransferStack = packetforward.NewIBCMiddleware(
		ibcTransferStack,
		app.PacketForwardKeeper,
		0,
		packetforwardkeeper.DefaultForwardTransferPacketTimeoutTimestamp,
	)
	ibcTransferStack = wibctransfer.NewPurposeMiddleware(ibcTransferStack)

	// Create ICAHost Stack
	icaHostStack := icahost.NewIBCModule(app.ICAHostKeeper)

	// Create Interchain Accounts Controller Stack
	icaControllerStack := icacontroller.NewIBCMiddleware(nil, app.ICAControllerKeeper)

	ibcWasmStack := wasm.NewIBCHandler(app.WasmKeeper, app.IBCKeeper.ChannelKeeper, app.IBCKeeper.ChannelKeeper)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := ibcporttypes.NewRouter().
		AddRoute(ibctransfertypes.ModuleName, ibcTransferStack).
		AddRoute(icahosttypes.SubModuleName, icaHostStack).
		AddRoute(icacontrollertypes.SubModuleName, icaControllerStack).
		AddRoute(wasmtypes.ModuleName, ibcWasmStack)
	app.IBCKeeper.SetRouter(ibcRouter)

	app.DEXKeeper = dexkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[dextypes.StoreKey]),
		app.AccountKeeper,
		authkeeper.NewQueryServer(app.AccountKeeper),
		app.AssetFTKeeper,
		app.DelayKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	if err := delayRouter.RegisterHandler(
		&dextypes.CancelGoodTil{},
		dexkeeper.NewDelayCancelOrderHandler(app.DEXKeeper),
	); err != nil {
		panic(err)
	}

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	assetFTModule := assetft.NewAppModule(
		appCodec,
		app.AssetFTKeeper,
		app.AccountKeeper,
		app.BankKeeper.BaseKeeper,
		app.ParamsKeeper,
	)
	assetNFTModule := assetnft.NewAppModule(
		appCodec,
		app.AssetNFTKeeper,
		app.NFTKeeper.Keeper,
		app.WasmKeeper,
		app.ParamsKeeper,
	)
	feeModule := feemodel.NewAppModule(app.FeeModelKeeper, app.ParamsKeeper)

	wnftModule := wnft.NewAppModule(appCodec, app.NFTKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry)

	customParamsModule := customparams.NewAppModule(app.CustomParamsKeeper, app.ParamsKeeper)
	wstakingModule := wstaking.NewAppModule(
		appCodec,
		app.StakingKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		app.GetSubspace(stakingtypes.ModuleName),
		app.CustomParamsKeeper,
	)

	delayModule := delay.NewAppModule(app.DelayKeeper)
	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	app.ModuleManager = module.NewManager(
		genutil.NewAppModule(app.AccountKeeper, app.StakingKeeper, app, txConfig),
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		wbank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, &app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil, app.GetSubspace(minttypes.ModuleName)),
		slashing.NewAppModule(
			appCodec,
			app.SlashingKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.StakingKeeper,
			app.GetSubspace(slashingtypes.ModuleName),
			app.interfaceRegistry,
		),
		distr.NewAppModule(
			appCodec, app.DistrKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.StakingKeeper,
			app.GetSubspace(distrtypes.ModuleName),
		),
		wstakingModule,
		upgrade.NewAppModule(app.UpgradeKeeper, app.AccountKeeper.AddressCodec()),
		evidence.NewAppModule(app.EvidenceKeeper),
		params.NewAppModule(app.ParamsKeeper),
		wasm.NewAppModule(
			appCodec,
			&app.WasmKeeper,
			app.StakingKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.MsgServiceRouter(),
			app.GetSubspace(wasmtypes.ModuleName),
		),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		feeModule,
		assetFTModule,
		assetNFTModule,
		wnftModule,
		customParamsModule,
		delayModule,
		dex.NewAppModule(appCodec, app.DEXKeeper, app.AccountKeeper),

		// IBC modules
		ibc.NewAppModule(app.IBCKeeper),
		wibctransfer.NewAppModule(app.TransferKeeper),
		packetforward.NewAppModule(app.PacketForwardKeeper, app.GetSubspace(packetforwardtypes.ModuleName)),
		ica.NewAppModule(&app.ICAControllerKeeper, &app.ICAHostKeeper),
		ibctm.NewAppModule(),
		ibchooks.NewAppModule(app.AccountKeeper),

		// always be last to make sure that it checks for all invariants and not only part of them
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
	)

	// BasicModuleManager defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration and genesis verification.
	// By default it is composed of all the module from the module manager.
	// Additionally, app module basics can be overwritten by passing them as argument.
	app.BasicModuleManager = module.NewBasicManagerFromManager(
		app.ModuleManager,
		map[string]module.AppModuleBasic{
			genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
			govtypes.ModuleName: gov.NewAppModuleBasic(
				[]govclient.ProposalHandler{
					paramsclient.ProposalHandler,
				},
			),
		})
	app.BasicModuleManager.RegisterLegacyAminoCodec(legacyAmino)
	app.BasicModuleManager.RegisterInterfaces(interfaceRegistry)

	// NOTE: upgrade module is required to be prioritized
	app.ModuleManager.SetOrderPreBlockers(
		upgradetypes.ModuleName,
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.ModuleManager.SetOrderBeginBlockers(
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		customparamstypes.ModuleName,
		stakingtypes.ModuleName,
		vestingtypes.ModuleName,
		ibcexported.ModuleName,
		ibctransfertypes.ModuleName,
		authtypes.ModuleName,
		authz.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		consensusparamtypes.ModuleName,
		ibchookstypes.ModuleName,
		packetforwardtypes.ModuleName,
		icatypes.ModuleName,
		wasmtypes.ModuleName,
		feemodeltypes.ModuleName,
		assetfttypes.ModuleName,
		assetnfttypes.ModuleName,
		nft.ModuleName,
		delaytypes.ModuleName,
		dextypes.ModuleName,
		// should be last
		genutiltypes.ModuleName,
	)

	app.ModuleManager.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		customparamstypes.ModuleName,
		stakingtypes.ModuleName,
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		authz.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		vestingtypes.ModuleName,
		minttypes.ModuleName,
		evidencetypes.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		ibcexported.ModuleName,
		ibctransfertypes.ModuleName,
		consensusparamtypes.ModuleName,
		ibchookstypes.ModuleName,
		packetforwardtypes.ModuleName,
		icatypes.ModuleName,
		wasmtypes.ModuleName,
		feemodeltypes.ModuleName,
		assetfttypes.ModuleName,
		assetnfttypes.ModuleName,
		nft.ModuleName,
		delaytypes.ModuleName,
		dextypes.ModuleName,
		// should be last
		genutiltypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: The genutils module must also occur after auth so that it can access the params from auth.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	genesisModuleOrder := []string{
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		authz.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		customparamstypes.ModuleName,
		stakingtypes.ModuleName,
		vestingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		ibcexported.ModuleName,
		evidencetypes.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		ibctransfertypes.ModuleName,
		packetforwardtypes.ModuleName,
		icatypes.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		consensusparamtypes.ModuleName,
		ibchookstypes.ModuleName,
		wasmtypes.ModuleName,
		feemodeltypes.ModuleName,
		nft.ModuleName,
		assetfttypes.ModuleName,
		assetnfttypes.ModuleName,
		delaytypes.ModuleName,
		// dex depends on auth(account) module
		dextypes.ModuleName,
		// should be last
		genutiltypes.ModuleName,
	}

	app.ModuleManager.SetOrderInitGenesis(genesisModuleOrder...)
	app.ModuleManager.SetOrderExportGenesis(genesisModuleOrder...)

	app.ModuleManager.RegisterInvariants(app.CrisisKeeper)

	app.configurator = module.NewConfigurator(
		app.appCodec,
		deterministicgastypes.NewDeterministicMsgServer(
			app.MsgServiceRouter(),
			deterministicGasConfig,
			app.AssetFTKeeper,
		),
		app.GRPCQueryRouter(),
	)
	if err = app.ModuleManager.RegisterServices(app.configurator); err != nil {
		panic(err)
	}

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.ModuleManager.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	// add test gRPC service for testing gRPC queries in isolation
	// testdata_pulsar.RegisterQueryServer(app.GRPCQueryRouter(), testdata_pulsar.QueryImpl{})

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing
	// transactions
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(
			app.appCodec,
			app.AccountKeeper,
			authsims.RandomGenesisAccounts,
			app.GetSubspace(authtypes.ModuleName),
		),
	}

	// exclude the nft simulation since it is incompatible with the asset nft
	simModules := excludeModules(app.ModuleManager.Modules, []string{nft.ModuleName})
	app.sm = module.NewSimulationManagerFromAppModules(simModules, overrideModules)
	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			HandlerOptions: authante.HandlerOptions{
				AccountKeeper:   app.AccountKeeper,
				BankKeeper:      app.BankKeeper,
				SignModeHandler: txConfig.SignModeHandler(),
				FeegrantKeeper:  app.FeeGrantKeeper,
				SigGasConsumer:  authante.DefaultSigVerificationGasConsumer,
			},
			DeterministicGasConfig: deterministicGasConfig,
			IBCKeeper:              app.IBCKeeper,
			GovKeeper:              &app.GovKeeper,
			FeeModelKeeper:         app.FeeModelKeeper,
			WasmTXCounterStoreKey:  runtime.NewKVStoreService(keys[wasmtypes.StoreKey]),
			WasmConfig:             wasmNodeConfig,
		},
	)
	if err != nil {
		panic(err)
	}

	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetPreBlocker(app.PreBlocker)
	app.SetAnteHandler(anteHandler)
	app.SetEndBlocker(app.EndBlocker)

	// In v0.46, the SDK introduces _postHandlers_. PostHandlers are like
	// antehandlers, but are run _after_ the `runMsgs` execution. They are also
	// defined as a chain, and have the same signature as antehandlers.
	//
	// In baseapp, postHandlers are run in the same store branch as `runMsgs`,
	// meaning that both `runMsgs` and `postHandler` state will be committed if
	// both are successful, and both will be reverted if any of the two fails.
	//
	// The SDK exposes a default postHandlers chain, which comprises of only
	// one decorator: the Transaction Tips decorator. However, some chains do
	// not need it by default, so feel free to comment the next line if you do
	// not need tips.
	// To read more about tips:
	// https://docs.cosmos.network/main/core/tips.html
	//
	// Please note that changing any of the anteHandler or postHandler chain is
	// likely to be a state-machine breaking change, which needs a coordinated
	// upgrade.
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		panic(err)
	}

	app.SetPostHandler(postHandler)

	// must be before Loading version
	// requires the snapshot store to be created and registered as a BaseAppOption
	if manager := app.SnapshotManager(); manager != nil {
		err := manager.RegisterExtensions(
			wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.WasmKeeper),
		)
		if err != nil {
			panic(errors.Wrapf(err, "failed to register wasm snapshot extension"))
		}
	}

	/**** Upgrades ****/
	upgrades := []appupgrade.Upgrade{
		appupgradev6.New(
			app.ModuleManager,
			app.configurator,
			app.BankKeeper,
		),
	}

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(errors.Errorf("failed to read upgrade info from disk %s", err))
	}

	isSkipHeight := app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height)

	// register the upgrades
	for _, upgradeItem := range upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			upgradeItem.Name,
			upgradeItem.Upgrade,
		)

		if upgradeInfo.Name == upgradeItem.Name && !isSkipHeight {
			// The line below is essential. If `&upgradeItem.StoreUpgrades` is passed to `UpgradeStoreLoader`
			// directly, then due to how `for` loop works in go, the `StoreUpgrades` of the last defined upgrade plan is
			// always used. To overcome this, here we make a copy of the store upgrades before taking a pointer.
			storeUpgrades := upgradeItem.StoreUpgrades
			// configure store loader that checks if version == upgradeHeight and applies store upgrades
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
		}
	}

	// At startup, after all modules have been registered, check that all proto
	// annotations are correct.
	protoFiles, err := proto.MergedRegistry()
	if err != nil {
		logger.Error("failed to merge proto files", "err", err)
	}
	err = msgservice.ValidateProtoAnnotations(protoFiles)
	if err != nil {
		// Once we switch to using protoreflect-based antehandlers, we might
		// want to panic here instead of logging a warning.
		logger.Error("failed to validate proto annotations", "err", err)
	}

	mergeProtos.Do(func() {
		protoregistry.GlobalFiles = protoFiles
	})

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}

		ctx := app.NewUncachedContext(true, tmproto.Header{})

		// Initialize pinned codes in wasmvm as they are not persisted there
		if err := app.WasmKeeper.InitializePinnedCodes(ctx); err != nil {
			tmos.Exit(errors.Wrapf(err, "failed initialize wasmp pinned codes").Error())
		}
	}

	return app
}

// Name returns the name of the App.
func (app *App) Name() string { return app.BaseApp.Name() }

// GetBaseApp returns the base app of the application.
func (app *App) GetBaseApp() *baseapp.BaseApp { return app.BaseApp }

// PreBlocker application updates every pre block.
func (app *App) PreBlocker(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.ModuleManager.PreBlock(ctx)
}

// BeginBlocker application updates every begin block.
func (app *App) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.ModuleManager.BeginBlock(ctx)
}

// EndBlocker application updates every end block.
func (app *App) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.ModuleManager.EndBlock(ctx)
}

// Configurator returns the app Configurator.
func (app *App) Configurator() module.Configurator {
	return app.configurator
}

// InitChainer application update at chain initialization.
func (app *App) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		return nil, err
	}
	if err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap()); err != nil {
		return nil, err
	}
	return app.ModuleManager.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height.
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns an app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns an InterfaceRegistry.
func (app *App) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// TxConfig returns TxConfig.
func (app *App) TxConfig() client.TxConfig {
	return app.txConfig
}

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (app *App) DefaultGenesis() map[string]json.RawMessage {
	return app.BasicModuleManager.DefaultGenesis(app.appCodec)
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *App) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// SimulationManager implements the SimulationApp interface.
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *serverapi.Server, _ serverconfig.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register new tendermint queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	app.BasicModuleManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register app's OpenAPI routes.
	apiSvr.Router.Handle("/static/openapi.json", http.FileServer(http.FS(docs.Docs)))
	apiSvr.Router.HandleFunc("/", openapi.Handler(Name, "/static/openapi.json"))
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.GRPCQueryRouter(), clientCtx, app.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

// RegisterNodeService registers the app node service.
func (app *App) RegisterNodeService(clientCtx client.Context, cfg serverconfig.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// AutoCliOpts returns the autocli options for the app.
func (app *App) AutoCliOpts() autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule, 0)
	for _, m := range app.ModuleManager.Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	return autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.ModuleManager.Modules),
		AddressCodec:          authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	}
}

// initParamsKeeper init params keeper and its subspaces.
func initParamsKeeper(
	appCodec codec.BinaryCodec,
	legacyAmino *codec.LegacyAmino,
	key, tkey storetypes.StoreKey,
) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	// TODO(v6): Remove after ibc is migrated to the param management system.
	// register the key tables for legacy param subspaces
	keyTable := ibcclienttypes.ParamKeyTable()
	keyTable.RegisterParamSet(&ibcconnectiontypes.Params{})
	paramsKeeper.Subspace(ibcexported.ModuleName).WithKeyTable(keyTable)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName).WithKeyTable(ibctransfertypes.ParamKeyTable())
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName).WithKeyTable(icacontrollertypes.ParamKeyTable())
	paramsKeeper.Subspace(icahosttypes.SubModuleName).WithKeyTable(icahosttypes.ParamKeyTable())

	return paramsKeeper
}

func excludeModules(modules map[string]interface{}, modulesToExclude []string) map[string]interface{} {
	filteredModules := make(map[string]interface{}, 0)
	modulesToExcludeMap := lo.SliceToMap(modulesToExclude, func(k string) (string, struct{}) {
		return k, struct{}{}
	})
	for n, m := range modules {
		if _, ok := modulesToExcludeMap[n]; ok {
			continue
		}
		filteredModules[n] = m
	}

	return filteredModules
}
