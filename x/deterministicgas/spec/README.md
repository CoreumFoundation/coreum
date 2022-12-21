<!--
order: 0
title: Deterministic Gas Overview
parent:
  title: "Deterministic Gas"
-->

# `x/deterministicgas`

## Intro

Coreum is using a deterministic gas model for its transactions. Meaning that given a transaction type (e.g 
bank.MsgSend) one can know how much gas will be used before hand, and this amount is fixed if some
preconditions are met. Of course this deterministic gas does not apply  to the type of transactions that have a complicated, nondeterministic execution path (e.g wasm smart contracts). We will provide a table with all 
deterministic gas for all our types. For a more recent data, consult 
[this file](https://github.com/CoreumFoundation/coreum/blob/master/pkg/config/deterministic_gas.go#L17)

## Formula
Here is formula for the transaction 

`
Gas = FixedGas + Sum(DeterministicGas for each message) + GasForExtraBytes + GasForExtraSignatures
`

`
GasForExtraBytes = max(0, TxByteSize-FreeBytes) * TxSizeCostPerByte
`

`
GasForExtraSignatures = max(0, NumOfSigs-FreeSigs) * SigVerifyCostSecp256k1
`

Currently Free Signatures is 1 and FreeBytes is 2048, meaning 
that if the transaction has 1 signature on it and is below 
2048 bytes, the user will not pay any thing extra, and if one of
those values exceed those limits, the user will pay for the extra
resources.


If the transaction size is under the FreeBytes limit and has only 
has one signature on it, the

### Special Cases
There are some special cases where an extra step is introduced to the formula. 

1. bank.MsgSend: `DeterministicGasForMsg = DeterministicGasForMsgSend * NumberOfCoins`
2. bank.MsgMultiSend: `DeterministicGasForMsg = DeterministicGasForMsgMultiSend * (NumberOfInputs + NumberOfOutputs)`

## Deterministic Gas Table 


| Tx Type                                | Gas  |
|----------------------------------------|------|
|AssetIssueFungibleToken                  | 80000|
|AssetMintFungibleToken                   | 35000|
|AssetBurnFungibleToken                   | 35000|
|AssetFreezeFungibleToken                 | 55000|
|AssetUnfreezeFungibleToken               | 55000|
|AssetGloballyFreezeFungibleToken         | 5000 |
|AssetGloballyUnfreezeFungibleToken       | 5000 |
|AssetSetWhitelistedLimitFungibleToken    | 35000|
|AssetIssueNonFungibleTokenClass          | 20000|
|AssetMintNonFungibleToken                | 30000|
|BankSendPerEntry                         | 22000|
|BankMultiSendPerEntry                    | 27000|
|DistributionFundCommunityPool            | 50000|
|DistributionSetWithdrawAddress           | 50000|
|DistributionWithdrawDelegatorReward      | 120000|
|DistributionWithdrawValidatorCommission  | 50000|
|GovSubmitProposal                        | 95000|
|GovVote                                  | 8000 |
|GovVoteWeighted                          | 11000|
|GovDeposit                               | 91000|
|NFTSend                                  | 20000|
|SlashingUnjail                           | 25000|
|StakingDelegate                          | 51000|
|StakingUndelegate                        | 51000|
|StakingBeginRedelegate                   | 51000|
|StakingCreateValidator                   | 50000|
|StakingEditValidator                     | 50000|
