package main

import (
	_ "embed"
	"os"
	"sort"
	"text/template"

	"github.com/samber/lo"

	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/x/deterministicgas"
)

var (
	//go:embed README.tmpl.md
	readmeTmpl string

	specialCases = []deterministicgas.MsgType{
		"/cosmos.authz.v1beta1.MsgExec",
		"/cosmos.bank.v1beta1.MsgMultiSend",
		"/cosmos.bank.v1beta1.MsgSend",
	}
)

func main() {
	type determMsg struct {
		Type deterministicgas.MsgType
		Gas  uint64
	}
	var determMsgs []determMsg

	var nonDetermMsgs []deterministicgas.MsgType

	cfg := deterministicgas.DefaultConfig()
	for msgType, gasFunc := range cfg.GasByMessageMap() {
		if lo.Contains(specialCases, msgType) {
			continue
		}

		gas, ok := gasFunc(nil)
		if ok {
			determMsgs = append(determMsgs, determMsg{msgType, gas})
		} else {
			nonDetermMsgs = append(nonDetermMsgs, msgType)
		}
	}

	sort.Slice(determMsgs, func(i, j int) bool {
		return determMsgs[i].Type < determMsgs[j].Type
	})

	sort.Strings(nonDetermMsgs)

	msgIssueGasPrice, _ := cfg.GasRequiredByMessage(&assetfttypes.MsgIssue{})

	generatorComment := `[//]: # (GENERATED DOC.)
[//]: # (DO NOT EDIT MANUALLY!!!)`

	file, err := os.OpenFile(os.Args[1], os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}

	err = template.Must(template.New("README.md").Parse(readmeTmpl)).Execute(file, struct {
		GeneratorComment  string
		SigVerifyCost     uint64
		TxSizeCostPerByte uint64
		FixedGas          uint64
		FreeBytes         uint64
		FreeSignatures    uint64

		MsgIssueGasPrice              uint64
		BankSendPerCoinGas            uint64
		BankMultiSendPerOperationsGas uint64
		AuthzExecOverhead             uint64

		DetermMsgsSpecialCases []deterministicgas.MsgType
		DetermMsgs             []determMsg
		NonDetermMsgs          []deterministicgas.MsgType

		DeterministicMessagesTable    string
		NonDeterministicMessagesTable string
	}{
		GeneratorComment:  generatorComment,
		FixedGas:          cfg.FixedGas,
		SigVerifyCost:     1000,
		TxSizeCostPerByte: 10,
		FreeBytes:         cfg.FreeBytes,
		FreeSignatures:    cfg.FreeSignatures,

		MsgIssueGasPrice:              msgIssueGasPrice,
		BankSendPerCoinGas:            deterministicgas.BankSendPerCoinGas,
		BankMultiSendPerOperationsGas: deterministicgas.BankMultiSendPerOperationsGas,
		AuthzExecOverhead:             deterministicgas.AuthzExecOverhead,

		DetermMsgsSpecialCases: specialCases,
		DetermMsgs:             determMsgs,
		NonDetermMsgs:          nonDetermMsgs,
	})
	if err != nil {
		panic(err)
	}
}
