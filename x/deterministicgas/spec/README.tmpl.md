{{ .GeneratorComment }}

# x/deterministicgas

## Intro

Coreum uses a deterministic gas model for its transactions. Meaning that given a transaction type (e.g
`/coreum.asset.ft.v1.MsgIssueGasPrice`) one can know how much gas will be used beforehand, and this amount is fixed if some
preconditions are met. Of course this deterministic gas does not apply to the type of transactions that have a
complicated, nondeterministic execution path (e.g `/cosmwasm.wasm.v1.MsgExecuteContract`). We provide tables with all
[deterministic gas](#deterministic-messages) & [nondeterministic gas](#nondeterministic-messages) for all our types.

## Formula

Here is formula for the transaction

`
Gas = FixedGas + Sum(Gas for each message) + GasForExtraBytes + GasForExtraSignatures
`

If message type is deterministic, then the value is looked up from the table, if it is non-deterministic, then the
required gas is determined after the execution.

`
GasForExtraBytes = max(0, TxByteSize-FreeBytes) * TxSizeCostPerByte
`

`
GasForExtraSignatures = max(0, NumOfSigs-FreeSigs) * SigVerifyCost
`

Currently, we have values for the above variables as follows:

- `FixedGas`: {{ .FixedGas }}
- `SigVerifyCost`: {{ .SigVerifyCost }}
- `TxSizeCostPerByte`: {{ .TxSizeCostPerByte }}
- `FreeSignatures`: {{ .FreeSignatures }}
- `FreeBytes`: {{ .FreeBytes }}

As an example if the transaction has 1 signature on it and is below
2048 bytes, the user will not pay anything extra, and if one of those values exceed those limits, the user will pay for
the extra resources.

### Full examples

#### Example 1
Let's say we have a transaction with 1 messages of type
`/cosmos.bank.v1beta1.MsgSend` containing single coin inside, also there is a single
signatures and the tx size is 1000 bytes, total will be:

`
TotalGas = {{ .FixedGas }} +  1 * {{ .BankSendPerCoinGas }} + 1 * {{ .SigVerifyCost }} + max(0, 1000-{{ .FreeBytes }}) * {{ .TxSizeCostPerByte }}
`

#### Example 2
Let's say we have a transaction with 2 messages of type
`/coreum.asset.ft.v1.MsgIssueGasPrice` inside, also there are 2
signatures and the tx size is 2050 bytes, total will be:

`
TotalGas = {{ .FixedGas }} + 2 * {{ .MsgIssueGasPrice }} + 2 * {{ .SigVerifyCost }} + max(0, 2050-{{ .FreeBytes }}) * {{ .TxSizeCostPerByte }}
`

## Gas Tables

### Deterministic messages

| Message Type | Gas |
|--------------|-----|
{{- range .DetermMsgsSpecialCases }}
{{ printf "| %-70v | [special case](#special-cases) |" (printf "`%v`" .) }}
{{- end -}}
{{- range .DetermMsgs }}
{{ printf "| %-70v | %-30v |" (printf "`%v`" .Type) .Gas }}
{{- end }}

#### Special Cases

There are some special cases when custom logic is applied for deterministic gas calculation.
Real examples of special case tests could be found [here](https://github.com/CoreumFoundation/coreum/blob/master/x/deterministicgas/config_test.go#L168)

##### `/cosmos.bank.v1beta1.MsgSend`

`DeterministicGasForMsg = bankSendPerCoinGas * NumberOfCoins`

`bankSendPerCoinGas` is currently equal to `{{ .BankSendPerCoinGas }}`.

##### `/cosmos.bank.v1beta1.MsgMultiSend`

`DeterministicGasForMsg = bankMultiSendPerOperationGas * (NumberOfInputs + NumberOfOutputs)`

`bankMultiSendPerOperationGas` is currently equal to `{{ .BankMultiSendPerOperationsGas }}`.

##### `/cosmos.authz.v1beta1.MsgExec`

`DeterministicGasForMsg = authzMsgExecOverhead + Sum(DeterministicGas(ChildMsg))`

`authzMsgExecOverhead` is currently equal to `{{ .AuthzExecOverhead }}`.

### Nondeterministic messages

| Message Type |
|--------------|
{{- range .NonDetermMsgs }}
{{ printf "| %-70v |" (printf "`%v`" .) }}
{{- end }}

{{ .GeneratorComment }}
