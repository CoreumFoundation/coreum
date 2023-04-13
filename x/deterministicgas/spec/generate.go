package main

import (
	_ "embed"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"text/template"

	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/x/deterministicgas"
)

var (
	//go:embed README.tmpl.md
	readmeTmpl string
)

func main() {
	type determMsg struct {
		Type deterministicgas.MsgType
		Gas  uint64
	}

	var (
		determMsgs                 []determMsg
		nonDetermMsgTypes          []deterministicgas.MsgType
		determSpeicialCaseMsgTypes []deterministicgas.MsgType

		determMsgGasFuncNameRegexp    = regexp.MustCompile(`^deterministicgas.DefaultConfig.func[0-9]{1,2}$`)
		nonDetermMsgGasFuncNameRegexp = regexp.MustCompile(`^deterministicgas.registerNondeterministicGasFuncs.func[0-9]{1,2}$`)
	)

	cfg := deterministicgas.DefaultConfig()
	for msgType, gasFunc := range cfg.GasByMessageMap() {
		// In this loop we use reflection to get function name. And depending on the name we try to match it with regexp
		// to determine gas type.
		fnFullName := runtime.FuncForPC(reflect.ValueOf(gasFunc).Pointer()).Name()
		fnParts := strings.Split(fnFullName, "/")
		fnShortName := fnParts[len(fnParts)-1]
		fmt.Println("Function name:", fnShortName)

		if determMsgGasFuncNameRegexp.MatchString(fnShortName) {
			gas, ok := gasFunc(nil)
			determMsgs = append(determMsgs, determMsg{msgType, gas})
			if !ok {
				panic(fmt.Errorf("non-deterministic values returned from function expected to be deterministic: %v", fnShortName))
			}
		} else if nonDetermMsgGasFuncNameRegexp.MatchString(fnShortName) {
			nonDetermMsgTypes = append(nonDetermMsgTypes, msgType)
		} else {
			// NOTE: For simplicity we don't match func name with regexp here because some funcs are defined as methods
			// and others are not. So we consider all funcs not matching deterministic or non-deterministic as special cases.
			determSpeicialCaseMsgTypes = append(determSpeicialCaseMsgTypes, msgType)
		}
	}

	sort.Slice(determMsgs, func(i, j int) bool {
		return determMsgs[i].Type < determMsgs[j].Type
	})

	sort.Strings(nonDetermMsgTypes)

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

		DetermMsgsSpecialCases: determSpeicialCaseMsgTypes,
		DetermMsgs:             determMsgs,
		NonDetermMsgs:          nonDetermMsgTypes,
	})
	if err != nil {
		panic(err)
	}
}
