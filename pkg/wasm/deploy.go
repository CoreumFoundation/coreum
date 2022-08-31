package wasm

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	tmtypes "github.com/tendermint/tendermint/types"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// gasMultiplier is the gas multiplier used for the wasm deployment.
const gasMultiplier = 1.5

// DeployConfig provides params for the deploying stage.
type DeployConfig struct {
	// Contract represent the contract binary decoded to bytes.
	Contract []byte

	// CodeID allows to specify existing program code ID to skip the store stage. If CodeID has been provided
	// and NeedInstantiation if false, the deployment just checks the program for existence on the chain.
	CodeID uint64

	// InstantiationConfig sets params specific to contract instantiation. If the instantiation phase is
	// skipped, make sure to have correct access type setting for the code store.
	InstantiationConfig ContractInstanceConfig

	// Network holds the chain config of the network
	Network ChainConfig

	// From specifies credentials for signing deployement / instantiation transactions.
	From types.Wallet
}

// ChainConfig encapsulates chain-specific parameters, used to communicate with daemon.
type ChainConfig struct {
	// GasPrice sets the minimum gas price required to be paid to get the transaction
	// included in a block. The real gasPrice is a dynamic value, so this option sets its minimum.
	GasPrice types.Coin
	// Client the RPC chain client
	Client client.Client
}

// ContractInstanceConfig contains params specific to contract instantiation.
type ContractInstanceConfig struct {
	// NeedInstantiation enables 2nd stage (contract instantiation) to be executed after code has been stored on chain.
	NeedInstantiation bool
	// AccessType sets the permission flag, affecting who can instantiate this contract.
	AccessType wasmtypes.AccessType
	// AccessAddress is respected when AccessTypeOnlyAddress is chosen as AccessType.
	AccessAddress string
	// NeedAdmin controls the option to set admin address explicitly. If false, there will be no admin.
	NeedAdmin bool
	// AdminAddress sets the address of an admin, optional. Used if `NeedAdmin` is true.
	AdminAddress string
	// InstantiatePayload is a path to a file containing JSON-encoded contract instantiate args, or JSON-encoded body itself.
	InstantiatePayload json.RawMessage
	// Amount specifies Coins to send to the contract during instantiation.
	Amount types.Coin
	// Label sets the human-readable label for the contract instance.
	Label string

	accessAddressParsed sdk.AccAddress
	adminAddressParsed  sdk.AccAddress
}

// Deploy implements logic for "contracts deploy" CLI command.
func Deploy(ctx context.Context, config DeployConfig) (*DeployOutput, error) {
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate the deployment config")
	}

	wasmData := config.Contract
	codeDataHash := hashContractCode(config.Contract)

	out := &DeployOutput{
		CodeID: config.CodeID,
	}

	// fixme refactor it
	var err error
	if config.CodeID == 0 {
		if out, err = deployCode(ctx, config, out, codeDataHash, wasmData); err != nil {
			return nil, err
		}

		config.CodeID = out.CodeID
	} else if out, err = checkCode(ctx, config, out, codeDataHash); err != nil {
		return nil, err
	}

	if !config.InstantiationConfig.NeedInstantiation {
		// code ID is known (stored) and 2nd stage is not needed
		return out, nil
	}

	// FIXME try not to do it
	if len(config.InstantiationConfig.Label) == 0 {
		config.InstantiationConfig.Label = codeDataHash
	}

	return instantiateContract(ctx, config, out)
}

func deployCode(
	ctx context.Context,
	config DeployConfig,
	out *DeployOutput,
	codeDataHash string,
	wasmData []byte,
) (*DeployOutput, error) {
	contractName := config.InstantiationConfig.Label

	deployLog := logger.Get(ctx).With(zap.String("name", contractName))
	deployLog.Info("Deploying artefact", zap.String("contractName", contractName), zap.String("from", config.From.Address().String()))

	var accessConfig *wasmtypes.AccessConfig
	if config.InstantiationConfig.AccessType != wasmtypes.AccessTypeUnspecified {
		accessConfig = &wasmtypes.AccessConfig{
			Permission: config.InstantiationConfig.AccessType,
			Address:    config.InstantiationConfig.accessAddressParsed.String(),
		}
	}

	codeID, storeTxHash, err := runContractStore(
		ctx,
		config.Network,
		config.From,
		wasmData,
		accessConfig,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to run contract code store")
	}

	out.CodeID = codeID
	out.StoreTxHash = storeTxHash
	out.Creator = config.From.Address().String()
	out.CodeDataHash = codeDataHash
	return out, nil
}

func checkCode(
	ctx context.Context,
	config DeployConfig,
	out *DeployOutput,
	codeDataHash string,
) (*DeployOutput, error) {
	info, err := queryContractCodeInfo(ctx, config.Network, config.CodeID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check contract code on chain")
	}

	if config.CodeID == info.CodeID {
		if codeDataHash != info.CodeDataHash {
			return nil, errors.Errorf("code hash mismatch: expected %s, chain has %s",
				codeDataHash, info.CodeDataHash,
			)
		}
	}

	out.Creator = info.Creator
	out.CodeDataHash = info.CodeDataHash

	return out, nil
}

func instantiateContract(ctx context.Context, config DeployConfig, out *DeployOutput) (*DeployOutput, error) {
	var adminAddress *sdk.AccAddress
	if config.InstantiationConfig.NeedAdmin {
		adminAddress = &config.InstantiationConfig.adminAddressParsed
	}

	contractAddr, initTxHash, err := runContractInstantiate(
		ctx,
		config.Network,
		config.From,
		config.CodeID,
		config.InstantiationConfig.InstantiatePayload,
		config.InstantiationConfig.Amount,
		config.InstantiationConfig.Label,
		adminAddress,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to run contract instantiate")
	}

	out.ContractAddr = contractAddr
	out.InitTxHash = initTxHash

	return out, nil
}

// DeployOutput has the results of running the Deploy method.
type DeployOutput struct {
	Creator      string `json:"creator"`
	CodeID       uint64 `json:"codeId"`
	ContractAddr string `json:"contractAddr,omitempty"`
	CodeDataHash string `json:"codeDataHash,omitempty"`
	StoreTxHash  string `json:"storeTxHash,omitempty"`
	InitTxHash   string `json:"initTxHash,omitempty"`
}

// Validate checks the deployment config and loads it's initial state.
func (c *DeployConfig) Validate() error {
	if len(c.Contract) == 0 {
		return errors.New("the Contract is empty")
	}

	if c.InstantiationConfig.Amount.Amount != nil {
		if err := c.InstantiationConfig.Amount.Validate(); err != nil {
			return errors.Wrapf(err, "invalid Amount: %v", c.InstantiationConfig.Amount)
		}
	}

	if c.InstantiationConfig.NeedAdmin {
		if len(c.InstantiationConfig.AdminAddress) > 0 {
			addr, err := sdk.AccAddressFromBech32(c.InstantiationConfig.AdminAddress)
			if err != nil {
				return errors.Wrapf(err, "failed to parse admin address from bech32: %s", c.InstantiationConfig.AdminAddress)
			}

			c.InstantiationConfig.adminAddressParsed = addr
		} else {
			c.InstantiationConfig.adminAddressParsed = c.From.Address()
		}
	}

	if err := c.Network.GasPrice.Validate(); err != nil {
		return errors.Wrapf(err, "invalid GasPrice: %v", c.Network.GasPrice)
	}

	return nil
}

func runContractStore(
	ctx context.Context,
	network ChainConfig,
	from types.Wallet,
	wasmData []byte,
	accessConfig *wasmtypes.AccessConfig,
) (codeID uint64, txHash string, err error) {
	log := logger.Get(ctx)
	chainClient := network.Client

	input := tx.BaseInput{
		Signer:   from,
		GasPrice: network.GasPrice,
	}

	msgStoreCode := &wasmtypes.MsgStoreCode{
		Sender:                from.Address().String(),
		WASMByteCode:          wasmData,
		InstantiatePermission: accessConfig,
	}

	gasLimit, err := chainClient.EstimateGas(ctx, input, msgStoreCode)
	if err != nil {
		return 0, "", errors.Wrap(err, "failed to estimate gas for MsgStoreCode")
	}

	log.Info("Estimated gas limit",
		zap.Int("bytecode_size", len(wasmData)),
		zap.Uint64("gas_limit", gasLimit),
	)

	input.GasLimit = uint64(float64(gasLimit) * gasMultiplier)

	signedTx, err := chainClient.Sign(ctx, input, msgStoreCode)
	if err != nil {
		return 0, "", errors.Wrapf(err, "failed to sign transaction as %s", from.Address().String())
	}

	txBytes := chainClient.Encode(signedTx)
	txHash = fmt.Sprintf("%X", tmtypes.Tx(txBytes).Hash())
	res, err := chainClient.Broadcast(ctx, txBytes)
	if err != nil {
		return 0, txHash, errors.Wrapf(err, "failed to broadcast Tx %s", txHash)
	}

	for _, ev := range res.EventLogs {
		if ev.Type == wasmtypes.EventTypeStoreCode {
			if value, ok := attrFromEvent(ev, wasmtypes.AttributeKeyCodeID); ok {
				codeID, err = strconv.ParseUint(value, 10, 64)
				if err != nil {
					return 0, txHash, errors.Wrapf(err, "failed to parse event attribute CodeID: %s as uint64", value)
				}

				break
			}

			log.With(
				zap.String("txHash", txHash),
			).Warn("contract code stored MsgStoreCode, but events don't have codeID")
		}
	}

	return codeID, txHash, nil
}

func runContractInstantiate(
	ctx context.Context,
	network ChainConfig,
	from types.Wallet,
	codeID uint64,
	initMsg json.RawMessage,
	amount types.Coin,
	label string,
	adminAcc *sdk.AccAddress,
) (contractAddr, txHash string, err error) {
	log := logger.Get(ctx)
	chainClient := network.Client

	input := tx.BaseInput{
		Signer:   from,
		GasPrice: network.GasPrice,
	}

	funds := sdk.NewCoins()
	if amount.Amount != nil {
		funds = funds.Add(sdk.NewCoin(amount.Denom, sdk.NewIntFromBigInt(amount.Amount)))
	}
	msgInstantiateContract := &wasmtypes.MsgInstantiateContract{
		Sender: from.Address().String(),
		CodeID: codeID,
		Label:  label,
		Msg:    wasmtypes.RawContractMessage(initMsg),
		Funds:  funds,
	}

	if adminAcc != nil {
		msgInstantiateContract.Admin = adminAcc.String()
	}

	gasLimit, err := chainClient.EstimateGas(ctx, input, msgInstantiateContract)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to estimate gas for MsgInstantiateContract")
	}

	log.Info("Estimated gas limit",
		zap.Int("contract_msg_size", len(initMsg)),
		zap.Uint64("gas_limit", gasLimit),
	)
	input.GasLimit = uint64(float64(gasLimit) * gasMultiplier)

	signedTx, err := chainClient.Sign(ctx, input, msgInstantiateContract)
	if err != nil {
		return "", "", errors.Wrapf(err, "failed to sign transaction as %s", from.Address().String())
	}

	txBytes := chainClient.Encode(signedTx)
	txHash = fmt.Sprintf("%X", tmtypes.Tx(txBytes).Hash())
	res, err := chainClient.Broadcast(ctx, txBytes)
	if err != nil {
		return "", txHash, errors.Wrapf(err, "failed to broadcast Tx %s", txHash)
	}

	for _, ev := range res.EventLogs {
		if ev.Type == wasmtypes.EventTypeInstantiate {
			if value, ok := attrFromEvent(ev, wasmtypes.AttributeKeyContractAddr); ok {
				contractAddr = value
				break
			}

			log.With(
				zap.String("txHash", txHash),
			).Warn("contract instantiated with MsgInstantiateContract, but events don't have _contract_address")
		}
	}

	return contractAddr, txHash, nil
}

type contractCodeInfo struct {
	CodeID       uint64
	Creator      string
	CodeDataHash string
}

func queryContractCodeInfo(
	ctx context.Context,
	network ChainConfig,
	codeID uint64,
) (info *contractCodeInfo, err error) {
	chainClient := network.Client
	resp, err := chainClient.WASMQueryClient().Code(ctx, &wasmtypes.QueryCodeRequest{
		CodeId: codeID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "code = InvalidArgument desc = not found") {
			return nil, errors.Errorf("contract codeID=%d not found on chain", codeID)
		}

		return nil, errors.Wrap(err, "WASMQueryClient failed to query the chain")
	}

	info = &contractCodeInfo{
		CodeID:       resp.CodeID,
		Creator:      resp.Creator,
		CodeDataHash: resp.DataHash.String(),
	}
	return info, nil
}

func attrFromEvent(ev sdk.StringEvent, attr string) (value string, ok bool) {
	for _, attrItem := range ev.Attributes {
		if attrItem.Key == attr {
			value = attrItem.Value
			ok = true
			return value, ok
		}
	}

	return "", false
}

func hashContractCode(wasmData []byte) string {
	h := sha256.Sum256(wasmData)
	return fmt.Sprintf("%X", h[:])
}
