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
Gas = FixedGas + max((GasForBytes + GasForSignatures - TxBaseGas), 0) + Sum(Gas for each message)
`

If message type is deterministic, then the value is looked up from the table, if it is non-deterministic, then the
required gas is determined after the execution.

`
GasForBytes = TxByteSize * TxSizeCostPerByte
`

`
GasForSignatures = NumOfSigs * SigVerifyCost
`

Currently, we have values for the above variables as follows:

- `FixedGas`: {{ .FixedGas }}
- `TxBaseGas`: {{ .TxBaseGas }}
- `SigVerifyCost`: {{ .SigVerifyCost }}
- `TxSizeCostPerByte`: {{ .TxSizeCostPerByte }}
- `FreeSignatures`: {{ .FreeSignatures }}
- `FreeBytes`: {{ .FreeBytes }}
- `WriteCostPerByte`: {{ .WriteCostPerByte }}

To summarize user pays FixedGas as long as `GasForBytes + GasForSignatures <= TxBaseGas`.
If `GasForBytes + GasForSignatures > TxBaseGas` user will have to pay anything above `TxBaseGas` on top of `FixedGas`. 

As an example if the transaction has 1 signature on it and size is below
2048 bytes, the user will not pay anything extra. Or user can have multiple signatures but fewer bytes then nothing extra should be paid.


### Full examples

#### Example 1
Let's say we have a transaction with 1 messages of type
`/cosmos.bank.v1beta1.MsgSend` containing single coin inside, also there is a single
signatures and the tx size is 1000 bytes, total will be:

`
TotalGas = {{ .FixedGas }} +  max(0, ({{ .TxBaseGas }} - 1 * {{ .SigVerifyCost }} + 1000 * {{ .TxSizeCostPerByte }})) + 1 * {{ .BankSendPerCoinGas }}
`

#### Example 2
Let's say we have a transaction with 2 messages of type
`/coreum.asset.ft.v1.MsgIssue` inside, also there are 2
signatures and the tx size is 2050 bytes, total will be:

`
TotalGas = {{ .FixedGas }} +  max(0, ({{ .TxBaseGas }} - 2 * {{ .SigVerifyCost }} + 2050 * {{ .TxSizeCostPerByte }})) + 2 * {{ .MsgIssueGasPrice }}
`

## Extensions
If one of the follwoing messages contains a token which have the extension feature enabled, it will not be considered deterministic any more . The reason is that extensions invlove smart contract calls which are nondeterministic in nature. 

 - `/ibc.applications.transfer.v1.MsgTransfer`
 - `/coreum.asset.ft.v1.MsgIssue`
 - `/cosmos.bank.v1beta1.MsgSend`
 - `/cosmos.bank.v1beta1.MsgMultiSend`
 - `/cosmos.distribution.v1beta1.MsgCommunityPoolSpend`
 - `/cosmos.distribution.v1beta1.MsgFundCommunityPool`
 - `/cosmos.vesting.v1beta1.MsgCreatePeriodicVestingAccount`
 - `/cosmos.vesting.v1beta1.MsgCreatePermanentLockedAccount`	
 - `/cosmos.vesting.v1beta1.MsgCreateVestingAccount`        		

It should also be mentioned that this rule applies for all the messages inside `/cosmos.authz.v1beta1.MsgExec`

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

##### `/cosmos.authz.v1beta1.MsgGrant`
MsgGrant is deterministic with gas value of `{{ .GrantBaseGas}}`, but if the authorization type is
one of the following, then it gets an overhead for every byte of the authorization.
The authorization types with overhead are:
- `/coreum.assert.nft.SendAuthorization`
- `/coreum.assert.ft.MintAuthorization`
- `/coreum.assert.ft.BurnAuthorization`

and the formula for them is
`DeterministicGas = GrantBaseGas + Size(Authorization) * WriteCostPerByte `


##### `/coreum.asset.nft.v1.MsgIssueClass`

`DeterministicGasForMsg = msgGas + Len(msg.Data) * WriteCostPerByte`

`msgGas` is currently equal to `{{ .NFTMsgIssueClassCost }}`.

##### `/coreum.asset.nft.v1.MsgMint`

`DeterministicGasForMsg = msgGas + Len(msg.Data) * WriteCostPerByte`

`msgGas` is currently equal to `{{ .NFTMsgMintCost }}`.


##### `/coreum.asset.ft.v1.MsgUpdateDEXWhitelistedDenoms`

`DeterministicGasForMsg = DEXUpdateWhitelistedDenomBaseGas + DEXWhitelistedPerDenomGas * NumberOfDenom`

`DEXWhitelistedPerDenomGas` is currently equal to `{{ .DEXWhitelistedPerDenomGas }}`.
`DEXUpdateWhitelistedDenomBaseGas` is currently equal to `{{ .DEXUpdateWhitelistedDenomBaseGas }}`.

### Nondeterministic messages

| Message Type |
|--------------|
{{- range .NonDetermMsgs }}
{{ printf "| %-70v |" (printf "`%v`" .) }}
{{- end }}

{{ .GeneratorComment }}
