package main

import (
	"bytes"
	_ "embed"
	"fmt"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/x/deterministicgas"
	"github.com/samber/lo"
	"os"
	"sort"
	"text/template"
	"time"
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

	generatorComment := fmt.Sprintf(`[//]: # (Doc generate at: %v)
[//]: # (DO NOT EDIT MANUALLY)`, time.Now().Format(time.DateTime))

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

	readmeBuf := new(bytes.Buffer)
	msgMintGasPrice, _ := cfg.GasRequiredByMessage(&assetfttypes.MsgMint{})
	err := template.Must(template.New("README.md").Parse(readmeTmpl)).Execute(readmeBuf, struct {
		GeneratorComment              string
		SigVerifyCost                 uint64
		TxSizeCostPerByte             uint64
		FixedGas                      uint64
		FreeBytes                     uint64
		FreeSignatures                uint64
		DeterministicMessagesTable    string
		NonDeterministicMessagesTable string
		MsgMintGasPrice               uint64
		BankSendPerCoinGas            uint64
		BankMultiSendPerOperationsGas uint64
		AuthzExecOverhead             uint64
	}{
		GeneratorComment:              generatorComment,
		FixedGas:                      cfg.FixedGas,
		SigVerifyCost:                 1000,
		TxSizeCostPerByte:             10,
		FreeBytes:                     cfg.FreeBytes,
		FreeSignatures:                cfg.FreeSignatures,
		MsgMintGasPrice:               msgMintGasPrice,
		BankSendPerCoinGas:            deterministicgas.BankSendPerCoinGas,
		BankMultiSendPerOperationsGas: deterministicgas.BankMultiSendPerOperationsGas,
		AuthzExecOverhead:             deterministicgas.AuthzExecOverhead,
		DeterministicMessagesTable:    determMsgsTableFormatted,
		NonDeterministicMessagesTable: nonDetermMsgsTableFormatted,
	})
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("x/deterministicgas/spec/README.md", readmeBuf.Bytes(), 0644); err != nil {
		panic(err)
	}
}
