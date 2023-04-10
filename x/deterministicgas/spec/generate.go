package main

import (
	"bytes"
	_ "embed"
	"fmt"
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

	specialCases = []string{
		"/cosmos.authz.v1beta1.MsgExec",
		"/cosmos.bank.v1beta1.MsgMultiSend",
		"/cosmos.bank.v1beta1.MsgSend",
	}
)

func main() {
	var determMsgs []struct {
		Type string
		Gas  uint64
	}
	var nonDetermMsgs []string

	cfg := deterministicgas.DefaultConfig()
	for msgType, gasFunc := range cfg.GasByMessageMap() {
		if lo.Contains(specialCases, msgType) {
			continue
		}

		gas, ok := gasFunc(nil)
		if ok {
			determMsgs = append(determMsgs, struct {
				Type string
				Gas  uint64
			}{msgType, gas})
		} else {
			nonDetermMsgs = append(nonDetermMsgs, msgType)
		}
	}

	sort.Slice(determMsgs, func(i, j int) bool {
		return determMsgs[i].Type < determMsgs[j].Type
	})
	sort.Strings(nonDetermMsgs)

	var determMsgsTableFormatted, nonDetermMsgsTableFormatted string

	for _, msgType := range specialCases {
		determMsgsTableFormatted += fmt.Sprintf("| `%-60v` | [special case](#special-cases) |\n", msgType)
	}
	for _, typeGas := range determMsgs {
		determMsgsTableFormatted += fmt.Sprintf("| `%-60v` | %-30v |\n", typeGas.Type, typeGas.Gas)
	}
	for _, msgType := range nonDetermMsgs {
		nonDetermMsgsTableFormatted += fmt.Sprintf("| `%-60v` |\n", msgType)
	}

	generatorComment := `[//]: # (GENERATED DOC.)
[//]: # (DO NOT EDIT MANUALLY!!!)`

	readmeBuf := new(bytes.Buffer)
	msgIssueGasPrice, _ := cfg.GasRequiredByMessage(&assetfttypes.MsgIssue{})
	err := template.Must(template.New("README.md").Parse(readmeTmpl)).Execute(readmeBuf, struct {
		GeneratorComment              string
		SigVerifyCost                 uint64
		TxSizeCostPerByte             uint64
		FixedGas                      uint64
		FreeBytes                     uint64
		FreeSignatures                uint64
		MsgIssueGasPrice              uint64
		BankSendPerCoinGas            uint64
		BankMultiSendPerOperationsGas uint64
		AuthzExecOverhead             uint64
		DeterministicMessagesTable    string
		NonDeterministicMessagesTable string
	}{
		GeneratorComment:              generatorComment,
		FixedGas:                      cfg.FixedGas,
		SigVerifyCost:                 1000,
		TxSizeCostPerByte:             10,
		FreeBytes:                     cfg.FreeBytes,
		FreeSignatures:                cfg.FreeSignatures,
		MsgIssueGasPrice:              msgIssueGasPrice,
		BankSendPerCoinGas:            deterministicgas.BankSendPerCoinGas,
		BankMultiSendPerOperationsGas: deterministicgas.BankMultiSendPerOperationsGas,
		AuthzExecOverhead:             deterministicgas.AuthzExecOverhead,
		DeterministicMessagesTable:    determMsgsTableFormatted,
		NonDeterministicMessagesTable: nonDetermMsgsTableFormatted,
	})
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(os.Args[1], readmeBuf.Bytes(), 0644); err != nil {
		panic(err)
	}
}
