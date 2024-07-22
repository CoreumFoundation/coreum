package main

import (
	_ "embed"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"text/template"

	storetypes "cosmossdk.io/store/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"

	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v4/x/deterministicgas"
)

//go:generate go run . ./README.md

//go:embed README.tmpl.md
var readmeTmpl string

//nolint:funlen
func main() {
	type determMsg struct {
		Type deterministicgas.MsgURL
		Gas  uint64
	}

	var (
		determMsgs                []determMsg
		nonDetermMsgURLs          []deterministicgas.MsgURL
		determSpeicialCaseMsgURLs []deterministicgas.MsgURL
	)

	cfg := deterministicgas.DefaultConfig()
	for msgURL, gasFunc := range cfg.GasByMessageMap() {
		fnFullName := runtime.FuncForPC(reflect.ValueOf(gasFunc).Pointer()).Name()
		fnParts := strings.Split(fnFullName, "/")
		fnShortName := fnParts[len(fnParts)-1]

		if fnShortName == "deterministicgas.nondeterministicGasFunc" {
			nonDetermMsgURLs = append(nonDetermMsgURLs, msgURL)
			continue
		}

		gas, ok := gasFunc(nil)
		// gasFunc returns ok equal to true only for deterministic messages which are not special cases.
		// For special cases it returns false because type-casting to a specific message type inside function fails.
		if ok {
			determMsgs = append(determMsgs, determMsg{msgURL, gas})
		} else {
			determSpeicialCaseMsgURLs = append(determSpeicialCaseMsgURLs, msgURL)
		}
	}

	sort.Slice(determMsgs, func(i, j int) bool {
		return determMsgs[i].Type < determMsgs[j].Type
	})
	sort.Slice(determSpeicialCaseMsgURLs, func(i, j int) bool {
		return determSpeicialCaseMsgURLs[i] < determSpeicialCaseMsgURLs[j]
	})
	sort.Slice(nonDetermMsgURLs, func(i, j int) bool {
		return nonDetermMsgURLs[i] < nonDetermMsgURLs[j]
	})

	msgIssueGasPrice, _ := cfg.GasRequiredByMessage(&assetfttypes.MsgIssue{})

	generatorComment := `[//]: # (GENERATED DOC.)
[//]: # (DO NOT EDIT MANUALLY!!!)`

	file, err := os.OpenFile(os.Args[1], os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}

	authParams := auth.DefaultParams()
	storeConfig := storetypes.KVGasConfig()
	err = template.Must(template.New("README.md").Parse(readmeTmpl)).Execute(file, struct {
		GeneratorComment  string
		SigVerifyCost     uint64
		TxSizeCostPerByte uint64
		FixedGas          uint64
		TxBaseGas         uint64
		FreeBytes         uint64
		FreeSignatures    uint64
		WriteCostPerByte  uint64

		MsgIssueGasPrice              uint64
		BankSendPerCoinGas            uint64
		BankMultiSendPerOperationsGas uint64
		AuthzExecOverhead             uint64
		NFTMsgIssueClassCost          uint64
		NFTMsgMintCost                uint64

		DetermMsgsSpecialCases []deterministicgas.MsgURL
		DetermMsgs             []determMsg
		NonDetermMsgs          []deterministicgas.MsgURL

		DeterministicMessagesTable    string
		NonDeterministicMessagesTable string
	}{
		GeneratorComment:  generatorComment,
		FixedGas:          cfg.FixedGas,
		TxBaseGas:         cfg.TxBaseGas(authParams),
		SigVerifyCost:     authParams.SigVerifyCostSecp256k1,
		TxSizeCostPerByte: authParams.TxSizeCostPerByte,
		FreeBytes:         cfg.FreeBytes,
		FreeSignatures:    cfg.FreeSignatures,
		WriteCostPerByte:  storeConfig.WriteCostPerByte,

		MsgIssueGasPrice:              msgIssueGasPrice,
		BankSendPerCoinGas:            deterministicgas.BankSendPerCoinGas,
		BankMultiSendPerOperationsGas: deterministicgas.BankMultiSendPerOperationsGas,
		AuthzExecOverhead:             deterministicgas.AuthzExecOverhead,
		NFTMsgIssueClassCost:          deterministicgas.NFTIssueClassBaseGas,
		NFTMsgMintCost:                deterministicgas.NFTMintBaseGas,

		DetermMsgsSpecialCases: determSpeicialCaseMsgURLs,
		DetermMsgs:             determMsgs,
		NonDetermMsgs:          nonDetermMsgURLs,
	})
	if err != nil {
		panic(err)
	}
}
