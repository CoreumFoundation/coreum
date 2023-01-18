<!--
order: 0
title: Deterministic Gas Overview
parent:
  title: "Deterministic Gas"
-->

# `x/deterministicgas`

## Intro

Coreum is using a deterministic gas model for its transactions. Meaning that given a transaction type (e.g 
asset.MintFungibleToken) one can know how much gas will be used before hand, and this amount is fixed if some
preconditions are met. Of course this deterministic gas does not apply  to the type of transactions that have a complicated, nondeterministic execution path (e.g wasm smart contracts). We will provide a table with all 
deterministic gas for all our types. For a more recent data, consult 
[this file](https://github.com/CoreumFoundation/coreum/blob/master/pkg/config/deterministic_gas.go#L17)

## Formula
Here is formula for the transaction 

`
Gas = FixedGas + Sum(Gas for each message) + GasForExtraBytes + GasForExtraSignatures
`

If message type is deterministic, then the value is looked up from the table, if it is non-deterministic, then the required gas is determined after the execution.

`
GasForExtraBytes = max(0, TxByteSize-FreeBytes) * TxSizeCostPerByte
`

`
GasForExtraSignatures = max(0, NumOfSigs-FreeSigs) * SigVerifyCost
`

Currently we have values for the above variables as follows: 
- `FixedGas`: 50000
- `SigVerifyCost`: 1000  
- `TxSizeCostPerByte`: 10
- `FreeSignatures`: 1
- `FreeBytes`: 2048

As an example if the transaction has 1 signature on it and is below 
2048 bytes, the user will not pay anything extra, and if one of those values exceed those limits, the user will pay for the extra resources.

#### Full example
Let's say we have a transaction with 2 messages of type 
asset.MintNonFungibleToken inside, also there are 2
signatures and the tx size is 2050 bytes, total will be:

`
TotalGas = 50000 + 35000 * 2 + (2050-2048) * 10 + 1 * 1000
`

`
TotalGas = 121040
`

### Special Cases
There are some special cases where an extra step is introduced to the formula. 

#### Bank
1. bank.MsgSend: `DeterministicGasForMsg = SendPerEntry * NumberOfCoins`
2. bank.MsgMultiSend: `DeterministicGasForMsg = MultiSendPerEntry * Max(NumberOfInputs, NumberOfOutputs)`

Where `SendPerEntry` and `MultiSendPerEntry` are constant values defined for each of the message types.
## Deterministic Gas Table 

### Deterministic messages

| Message Type                             | Gas  |
|------------------------------------------|------|
|asset.IssueFungibleToken                  | 80000|
|asset.MintFungibleToken                   | 35000|
|asset.BurnFungibleToken                   | 35000|
|asset.FreezeFungibleToken                 | 55000|
|asset.UnfreezeFungibleToken               | 55000|
|asset.GloballyFreezeFungibleToken         | 5000 |
|asset.GloballyUnfreezeFungibleToken       | 5000 |
|asset.SetWhitelistedLimitFungibleToken    | 35000|
|asset.IssueNonFungibleTokenClass          | 20000|
|asset.MintNonFungibleToken                | 30000|
|bank.SendPerEntry                         | 22000|
|bank.MultiSendPerEntry                    | 27000|
|distribution.FundCommunityPool            | 50000|
|distribution.SetWithdrawAddress           | 50000|
|distribution.WithdrawDelegatorReward      | 120000|
|distribution.WithdrawValidatorCommission  | 50000|
|gov.SubmitProposal                        | 95000|
|gov.Vote                                  | 8000 |
|gov.VoteWeighted                          | 11000|
|gov.Deposit                               | 91000|
|slashing.Unjail                           | 25000|
|staking.Delegate                          | 51000|
|staking.Undelegate                        | 51000|
|staking.BeginRedelegate                   | 51000|
|staking.CreateValidator                   | 50000|
|staking.EditValidator                     | 50000|

### Nondeterministic messages
all the messages related to wasm. 
