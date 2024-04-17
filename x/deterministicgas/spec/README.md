[//]: # (GENERATED DOC.)
[//]: # (DO NOT EDIT MANUALLY!!!)

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

- `FixedGas`: 65000
- `TxBaseGas`: 21480
- `SigVerifyCost`: 1000
- `TxSizeCostPerByte`: 10
- `FreeSignatures`: 1
- `FreeBytes`: 2048
- `WriteCostPerByte`: 30

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
TotalGas = 65000 +  max(0, (21480 - 1 * 1000 + 1000 * 10)) + 1 * 50000
`

#### Example 2
Let's say we have a transaction with 2 messages of type
`/coreum.asset.ft.v1.MsgIssueGasPrice` inside, also there are 2
signatures and the tx size is 2050 bytes, total will be:

`
TotalGas = 65000 +  max(0, (21480 - 2 * 1000 + 2050 * 10)) + 2 * 70000
`

## Gas Tables

### Deterministic messages

| Message Type | Gas                            |
|--------------|--------------------------------|
| `/coreum.asset.nft.v1.MsgIssueClass`                                   | [special case](#special-cases) |
| `/coreum.asset.nft.v1.MsgMint`                                         | [special case](#special-cases) |
| `/coreum.asset.nft.v1.MsgUpdateData`                                   | [special case](#special-cases) |
| `/cosmos.authz.v1beta1.MsgExec`                                        | [special case](#special-cases) |
| `/cosmos.bank.v1beta1.MsgMultiSend`                                    | [special case](#special-cases) |
| `/cosmos.bank.v1beta1.MsgSend`                                         | [special case](#special-cases) |
| `/coreum.asset.ft.v1.MsgBurn`                                          | 35000                          |
| `/coreum.asset.ft.v1.MsgClawback`                                      | 15500                          |
| `/coreum.asset.ft.v1.MsgFreeze`                                        | 8500                           |
| `/coreum.asset.ft.v1.MsgGloballyFreeze`                                | 5000                           |
| `/coreum.asset.ft.v1.MsgGloballyUnfreeze`                              | 5000                           |
| `/coreum.asset.ft.v1.MsgIssue`                                         | 70000                          |
| `/coreum.asset.ft.v1.MsgMint`                                          | 31000                          |
| `/coreum.asset.ft.v1.MsgSetFrozen`                                     | 8500                           |
| `/coreum.asset.ft.v1.MsgSetWhitelistedLimit`                           | 9000                           |
| `/coreum.asset.ft.v1.MsgTransferAdmin`                                 | 3000                           |
| `/coreum.asset.ft.v1.MsgUnfreeze`                                      | 8500                           |
| `/coreum.asset.ft.v1.MsgUpgradeTokenV1`                                | 25000                          |
| `/coreum.asset.nft.v1.MsgAddToClassWhitelist`                          | 7000                           |
| `/coreum.asset.nft.v1.MsgAddToWhitelist`                               | 7000                           |
| `/coreum.asset.nft.v1.MsgBurn`                                         | 26000                          |
| `/coreum.asset.nft.v1.MsgClassFreeze`                                  | 8000                           |
| `/coreum.asset.nft.v1.MsgClassUnfreeze`                                | 5000                           |
| `/coreum.asset.nft.v1.MsgFreeze`                                       | 8000                           |
| `/coreum.asset.nft.v1.MsgRemoveFromClassWhitelist`                     | 3500                           |
| `/coreum.asset.nft.v1.MsgRemoveFromWhitelist`                          | 3500                           |
| `/coreum.asset.nft.v1.MsgUnfreeze`                                     | 5000                           |
| `/coreum.nft.v1beta1.MsgSend`                                          | 25000                          |
| `/cosmos.authz.v1beta1.MsgGrant`                                       | 28000                          |
| `/cosmos.authz.v1beta1.MsgRevoke`                                      | 8000                           |
| `/cosmos.distribution.v1beta1.MsgFundCommunityPool`                    | 17000                          |
| `/cosmos.distribution.v1beta1.MsgSetWithdrawAddress`                   | 5000                           |
| `/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward`              | 79000                          |
| `/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission`          | 22000                          |
| `/cosmos.feegrant.v1beta1.MsgGrantAllowance`                           | 11000                          |
| `/cosmos.feegrant.v1beta1.MsgRevokeAllowance`                          | 2500                           |
| `/cosmos.gov.v1.MsgDeposit`                                            | 65000                          |
| `/cosmos.gov.v1.MsgVote`                                               | 6000                           |
| `/cosmos.gov.v1.MsgVoteWeighted`                                       | 6500                           |
| `/cosmos.gov.v1beta1.MsgDeposit`                                       | 85000                          |
| `/cosmos.gov.v1beta1.MsgVote`                                          | 6000                           |
| `/cosmos.gov.v1beta1.MsgVoteWeighted`                                  | 9000                           |
| `/cosmos.group.v1.MsgCreateGroup`                                      | 55000                          |
| `/cosmos.group.v1.MsgCreateGroupPolicy`                                | 40000                          |
| `/cosmos.group.v1.MsgCreateGroupWithPolicy`                            | 95000                          |
| `/cosmos.group.v1.MsgLeaveGroup`                                       | 17500                          |
| `/cosmos.group.v1.MsgUpdateGroupAdmin`                                 | 13500                          |
| `/cosmos.group.v1.MsgUpdateGroupMembers`                               | 17500                          |
| `/cosmos.group.v1.MsgUpdateGroupMetadata`                              | 9500                           |
| `/cosmos.group.v1.MsgUpdateGroupPolicyAdmin`                           | 20000                          |
| `/cosmos.group.v1.MsgUpdateGroupPolicyDecisionPolicy`                  | 17000                          |
| `/cosmos.group.v1.MsgUpdateGroupPolicyMetadata`                        | 15000                          |
| `/cosmos.group.v1.MsgWithdrawProposal`                                 | 22000                          |
| `/cosmos.nft.v1beta1.MsgSend`                                          | 25000                          |
| `/cosmos.slashing.v1beta1.MsgUnjail`                                   | 90000                          |
| `/cosmos.staking.v1beta1.MsgBeginRedelegate`                           | 157000                         |
| `/cosmos.staking.v1beta1.MsgCancelUnbondingDelegation`                 | 75000                          |
| `/cosmos.staking.v1beta1.MsgCreateValidator`                           | 117000                         |
| `/cosmos.staking.v1beta1.MsgDelegate`                                  | 83000                          |
| `/cosmos.staking.v1beta1.MsgEditValidator`                             | 13000                          |
| `/cosmos.staking.v1beta1.MsgUndelegate`                                | 112000                         |
| `/cosmos.vesting.v1beta1.MsgCreatePeriodicVestingAccount`              | 32000                          |
| `/cosmos.vesting.v1beta1.MsgCreatePermanentLockedAccount`              | 30000                          |
| `/cosmos.vesting.v1beta1.MsgCreateVestingAccount`                      | 30000                          |
| `/cosmwasm.wasm.v1.MsgClearAdmin`                                      | 6500                           |
| `/cosmwasm.wasm.v1.MsgUpdateAdmin`                                     | 8000                           |
| `/ibc.applications.transfer.v1.MsgTransfer`                            | 54000                          |

#### Special Cases

There are some special cases when custom logic is applied for deterministic gas calculation.
Real examples of special case tests could be found [here](https://github.com/CoreumFoundation/coreum/blob/master/x/deterministicgas/config_test.go#L168)

##### `/cosmos.bank.v1beta1.MsgSend`

`DeterministicGasForMsg = bankSendPerCoinGas * NumberOfCoins`

`bankSendPerCoinGas` is currently equal to `50000`.

##### `/cosmos.bank.v1beta1.MsgMultiSend`

`DeterministicGasForMsg = bankMultiSendPerOperationGas * (NumberOfInputs + NumberOfOutputs)`

`bankMultiSendPerOperationGas` is currently equal to `35000`.

##### `/cosmos.authz.v1beta1.MsgExec`

`DeterministicGasForMsg = authzMsgExecOverhead + Sum(DeterministicGas(ChildMsg))`

`authzMsgExecOverhead` is currently equal to `1500`.

##### `/coreum.asset.nft.v1.MsgIssueClass`

`DeterministicGasForMsg = msgGas + Len(msg.Data) * WriteCostPerByte`

`msgGas` is currently equal to `16000`.

##### `/coreum.asset.nft.v1.MsgMint`

`DeterministicGasForMsg = msgGas + Len(msg.Data) * WriteCostPerByte`

`msgGas` is currently equal to `39000`.

### Nondeterministic messages

| Message Type |
|--------------|
| `/cosmos.auth.v1beta1.MsgUpdateParams`                                 |
| `/cosmos.bank.v1beta1.MsgSetSendEnabled`                               |
| `/cosmos.bank.v1beta1.MsgUpdateParams`                                 |
| `/cosmos.consensus.v1.MsgUpdateParams`                                 |
| `/cosmos.crisis.v1beta1.MsgUpdateParams`                               |
| `/cosmos.crisis.v1beta1.MsgVerifyInvariant`                            |
| `/cosmos.distribution.v1beta1.MsgCommunityPoolSpend`                   |
| `/cosmos.distribution.v1beta1.MsgUpdateParams`                         |
| `/cosmos.evidence.v1beta1.MsgSubmitEvidence`                           |
| `/cosmos.gov.v1.MsgExecLegacyContent`                                  |
| `/cosmos.gov.v1.MsgSubmitProposal`                                     |
| `/cosmos.gov.v1.MsgUpdateParams`                                       |
| `/cosmos.gov.v1beta1.MsgSubmitProposal`                                |
| `/cosmos.group.v1.MsgExec`                                             |
| `/cosmos.group.v1.MsgSubmitProposal`                                   |
| `/cosmos.group.v1.MsgVote`                                             |
| `/cosmos.mint.v1beta1.MsgUpdateParams`                                 |
| `/cosmos.slashing.v1beta1.MsgUpdateParams`                             |
| `/cosmos.staking.v1beta1.MsgUpdateParams`                              |
| `/cosmos.upgrade.v1beta1.MsgCancelUpgrade`                             |
| `/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade`                           |
| `/cosmwasm.wasm.v1.MsgExecuteContract`                                 |
| `/cosmwasm.wasm.v1.MsgIBCCloseChannel`                                 |
| `/cosmwasm.wasm.v1.MsgIBCSend`                                         |
| `/cosmwasm.wasm.v1.MsgInstantiateContract`                             |
| `/cosmwasm.wasm.v1.MsgInstantiateContract2`                            |
| `/cosmwasm.wasm.v1.MsgMigrateContract`                                 |
| `/cosmwasm.wasm.v1.MsgPinCodes`                                        |
| `/cosmwasm.wasm.v1.MsgStoreAndInstantiateContract`                     |
| `/cosmwasm.wasm.v1.MsgStoreAndMigrateContract`                         |
| `/cosmwasm.wasm.v1.MsgStoreCode`                                       |
| `/cosmwasm.wasm.v1.MsgSudoContract`                                    |
| `/cosmwasm.wasm.v1.MsgUnpinCodes`                                      |
| `/cosmwasm.wasm.v1.MsgUpdateContractLabel`                             |
| `/cosmwasm.wasm.v1.MsgUpdateInstantiateConfig`                         |
| `/cosmwasm.wasm.v1.MsgUpdateParams`                                    |
| `/ibc.core.channel.v1.MsgAcknowledgement`                              |
| `/ibc.core.channel.v1.MsgChannelCloseConfirm`                          |
| `/ibc.core.channel.v1.MsgChannelCloseInit`                             |
| `/ibc.core.channel.v1.MsgChannelOpenAck`                               |
| `/ibc.core.channel.v1.MsgChannelOpenConfirm`                           |
| `/ibc.core.channel.v1.MsgChannelOpenInit`                              |
| `/ibc.core.channel.v1.MsgChannelOpenTry`                               |
| `/ibc.core.channel.v1.MsgRecvPacket`                                   |
| `/ibc.core.channel.v1.MsgTimeout`                                      |
| `/ibc.core.channel.v1.MsgTimeoutOnClose`                               |
| `/ibc.core.client.v1.MsgCreateClient`                                  |
| `/ibc.core.client.v1.MsgSubmitMisbehaviour`                            |
| `/ibc.core.client.v1.MsgUpdateClient`                                  |
| `/ibc.core.client.v1.MsgUpgradeClient`                                 |
| `/ibc.core.connection.v1.MsgConnectionOpenAck`                         |
| `/ibc.core.connection.v1.MsgConnectionOpenConfirm`                     |
| `/ibc.core.connection.v1.MsgConnectionOpenInit`                        |
| `/ibc.core.connection.v1.MsgConnectionOpenTry`                         |

[//]: # (GENERATED DOC.)
[//]: # (DO NOT EDIT MANUALLY!!!)
