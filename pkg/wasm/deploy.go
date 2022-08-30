package wasm

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"
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
	// ArtefactPath is a filesystem path to *.wasm artefact to deploy. The blob might be gzipped.
	// If not provided, will be guessed from WorkspaceDir. Make sure that either WorkspaceDir or ArtefactPath
	// is provied and exists.
	ArtefactPath string

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
	AccessType string
	// AccessAddress is respected when AccessTypeOnlyAddress is chosen as AccessType.
	AccessAddress string
	// NeedAdmin controls the option to set admin address explicitly. If false, there will be no admin.
	NeedAdmin bool
	// AdminAddress sets the address of an admin, optional. Used if `NeedAdmin` is true.
	AdminAddress string
	// InstantiatePayload is a path to a file containing JSON-encoded contract instantiate args, or JSON-encoded body itself.
	InstantiatePayload string
	// Amount specifies Coins to send to the contract during instantiation.
	Amount types.Coin
	// Label sets the human-readable label for the contract instance.
	Label string

	instantiatePayloadBody json.RawMessage
	accessTypeParsed       wasmtypes.AccessType
	accessAddressParsed    sdk.AccAddress
	adminAddressParsed     sdk.AccAddress
}

// AccessType encodes possible values of the access type flag
type AccessType string

const (
	// AccessTypeNobody forbidden
	AccessTypeNobody AccessType = "nobody"
	// AccessTypeOnlyAddress restricted to an address
	AccessTypeOnlyAddress AccessType = "address"
	// AccessTypeEverybody unrestricted
	AccessTypeEverybody AccessType = "unrestricted"
)

// Deploy implements logic for "contracts deploy" CLI command.
func Deploy(ctx context.Context, config DeployConfig) (*DeployOutput, error) {
	if err := config.ValidateAndLoad(); err != nil {
		return nil, errors.Wrap(err, "failed to validate the deployment config")
	}

	if !checkWasmFile(config.ArtefactPath) {
		return nil, errors.New(fmt.Sprintf("WASM artefact is not valid at path:%s", config.ArtefactPath))
	}

	wasmData, codeDataHash, err := loadContractCode(config)
	if err != nil {
		return nil, err
	}

	out := &DeployOutput{
		CodeID: config.CodeID,
	}
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

	if len(config.InstantiationConfig.Label) == 0 {
		artefactBase := filepath.Base(config.ArtefactPath)
		contractName := strings.TrimSuffix(artefactBase, filepath.Ext(artefactBase))
		config.InstantiationConfig.Label = contractName
	}

	return instantiateContract(ctx, config, out)
}

func loadContractCode(config DeployConfig) (wasmData []byte, codeDataHash string, err error) {
	wasmData, err = os.ReadFile(config.ArtefactPath)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to read artefact data from the fs")
	}

	switch {
	case ioutils.IsWasm(wasmData):
		codeDataHash = hashContractCode(wasmData)
		wasmData, err = ioutils.GzipIt(wasmData)

		if err != nil {
			return nil, "", errors.Wrap(err, "failed to gzip the wasm data")
		}
	case ioutils.IsGzip(wasmData):
		srcWasmData, err := ioutils.Uncompress(wasmData, uint64(wasmtypes.MaxWasmSize))
		if err != nil {
			return nil, "", errors.Wrap(err, "failed to uncompress the gzip data")
		} else if !ioutils.IsWasm(srcWasmData) {
			err := errors.New("invalid input file. Use wasm binary or gzip of a wasm binary")
			return nil, "", err
		}

		codeDataHash = hashContractCode(srcWasmData)
	default:
		err := errors.New("invalid input file. Use wasm binary or gzip")
		return nil, "", err
	}

	return wasmData, codeDataHash, err
}

func deployCode(
	ctx context.Context,
	config DeployConfig,
	out *DeployOutput,
	codeDataHash string,
	wasmData []byte,
) (*DeployOutput, error) {
	artefactBase := filepath.Base(config.ArtefactPath)
	contractName := strings.TrimSuffix(artefactBase, filepath.Ext(artefactBase))

	deployLog := logger.Get(ctx).With(zap.String("name", contractName))
	deployLog.Info("Deploying artefact", zap.String("artefact", artefactBase), zap.String("from", config.From.Address().String()))

	var accessConfig *wasmtypes.AccessConfig
	if config.InstantiationConfig.accessTypeParsed != wasmtypes.AccessTypeUnspecified {
		accessConfig = &wasmtypes.AccessConfig{
			Permission: config.InstantiationConfig.accessTypeParsed,
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
		config.InstantiationConfig.instantiatePayloadBody,
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

// ValidateAndLoad checks the deployment config and loads it's initial state.
// TODO(dhil) it would be better not to sore the state in the config and not set in the validation.
func (c *DeployConfig) ValidateAndLoad() error {
	if len(c.ArtefactPath) == 0 {
		return errors.New("the ArtefactPath can't be empty")
	}

	if len(c.InstantiationConfig.InstantiatePayload) > 0 {
		body, err := getPayloadBody(c.InstantiationConfig.InstantiatePayload)
		if err != nil {
			return err
		}

		c.InstantiationConfig.instantiatePayloadBody = body
	}

	if c.InstantiationConfig.Amount.Amount != nil {
		if err := c.InstantiationConfig.Amount.Validate(); err != nil {
			return errors.Wrapf(err, "invalid Amount: %v", c.InstantiationConfig.Amount)
		}
	}

	switch AccessType(c.InstantiationConfig.AccessType) {
	case "":
		c.InstantiationConfig.accessTypeParsed = wasmtypes.AccessTypeUnspecified
	case AccessTypeNobody:
		c.InstantiationConfig.accessTypeParsed = wasmtypes.AccessTypeNobody
	case AccessTypeEverybody:
		c.InstantiationConfig.accessTypeParsed = wasmtypes.AccessTypeEverybody
	case AccessTypeOnlyAddress:
		addr, err := sdk.AccAddressFromBech32(c.InstantiationConfig.AccessAddress)
		if err != nil {
			return errors.Wrapf(err, "failed to parse instantiation access address from bech32: %s", c.InstantiationConfig.AccessAddress)
		}

		c.InstantiationConfig.accessAddressParsed = addr
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

func getPayloadBody(payloadPathOrBody string) (json.RawMessage, error) {
	if body := []byte(payloadPathOrBody); json.Valid(body) {
		return body, nil
	}

	payloadFilePath := payloadPathOrBody

	body, err := os.ReadFile(payloadFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "file specified for instantiate payload, but couldn't be read: %s", payloadFilePath)
	}

	if !json.Valid(body) {
		return nil, errors.Wrapf(err, "file specified for instantiate payload, but doesn't contain valid JSON: %s", payloadFilePath)
	}

	return body, nil
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

func checkWasmFile(path string) bool {
	wasmData, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	return ioutils.IsWasm(wasmData) || ioutils.IsGzip(wasmData)
}

func hashContractCode(wasmData []byte) string {
	h := sha256.Sum256(wasmData)
	return fmt.Sprintf("%X", h[:])
}
