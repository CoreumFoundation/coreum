<!-- This file is auto-generated. Please do not modify it yourself. -->
<!-- markdown-link-check-disable -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [coreum/asset/ft/v1/event.proto](#coreum/asset/ft/v1/event.proto)
    - [EventFrozenAmountChanged](#coreum.asset.ft.v1.EventFrozenAmountChanged)
    - [EventIssued](#coreum.asset.ft.v1.EventIssued)
    - [EventWhitelistedAmountChanged](#coreum.asset.ft.v1.EventWhitelistedAmountChanged)
  
- [coreum/asset/ft/v1/genesis.proto](#coreum/asset/ft/v1/genesis.proto)
    - [Balance](#coreum.asset.ft.v1.Balance)
    - [GenesisState](#coreum.asset.ft.v1.GenesisState)
    - [PendingTokenUpgrade](#coreum.asset.ft.v1.PendingTokenUpgrade)
  
- [coreum/asset/ft/v1/params.proto](#coreum/asset/ft/v1/params.proto)
    - [Params](#coreum.asset.ft.v1.Params)
  
- [coreum/asset/ft/v1/query.proto](#coreum/asset/ft/v1/query.proto)
    - [QueryBalanceRequest](#coreum.asset.ft.v1.QueryBalanceRequest)
    - [QueryBalanceResponse](#coreum.asset.ft.v1.QueryBalanceResponse)
    - [QueryFrozenBalanceRequest](#coreum.asset.ft.v1.QueryFrozenBalanceRequest)
    - [QueryFrozenBalanceResponse](#coreum.asset.ft.v1.QueryFrozenBalanceResponse)
    - [QueryFrozenBalancesRequest](#coreum.asset.ft.v1.QueryFrozenBalancesRequest)
    - [QueryFrozenBalancesResponse](#coreum.asset.ft.v1.QueryFrozenBalancesResponse)
    - [QueryParamsRequest](#coreum.asset.ft.v1.QueryParamsRequest)
    - [QueryParamsResponse](#coreum.asset.ft.v1.QueryParamsResponse)
    - [QueryTokenRequest](#coreum.asset.ft.v1.QueryTokenRequest)
    - [QueryTokenResponse](#coreum.asset.ft.v1.QueryTokenResponse)
    - [QueryTokenUpgradeStatusesRequest](#coreum.asset.ft.v1.QueryTokenUpgradeStatusesRequest)
    - [QueryTokenUpgradeStatusesResponse](#coreum.asset.ft.v1.QueryTokenUpgradeStatusesResponse)
    - [QueryTokensRequest](#coreum.asset.ft.v1.QueryTokensRequest)
    - [QueryTokensResponse](#coreum.asset.ft.v1.QueryTokensResponse)
    - [QueryWhitelistedBalanceRequest](#coreum.asset.ft.v1.QueryWhitelistedBalanceRequest)
    - [QueryWhitelistedBalanceResponse](#coreum.asset.ft.v1.QueryWhitelistedBalanceResponse)
    - [QueryWhitelistedBalancesRequest](#coreum.asset.ft.v1.QueryWhitelistedBalancesRequest)
    - [QueryWhitelistedBalancesResponse](#coreum.asset.ft.v1.QueryWhitelistedBalancesResponse)
  
    - [Query](#coreum.asset.ft.v1.Query)
  
- [coreum/asset/ft/v1/token.proto](#coreum/asset/ft/v1/token.proto)
    - [Definition](#coreum.asset.ft.v1.Definition)
    - [DelayedTokenUpgradeV1](#coreum.asset.ft.v1.DelayedTokenUpgradeV1)
    - [Token](#coreum.asset.ft.v1.Token)
    - [TokenUpgradeStatuses](#coreum.asset.ft.v1.TokenUpgradeStatuses)
    - [TokenUpgradeV1Status](#coreum.asset.ft.v1.TokenUpgradeV1Status)
  
    - [Feature](#coreum.asset.ft.v1.Feature)
  
- [coreum/asset/ft/v1/tx.proto](#coreum/asset/ft/v1/tx.proto)
    - [EmptyResponse](#coreum.asset.ft.v1.EmptyResponse)
    - [MsgBurn](#coreum.asset.ft.v1.MsgBurn)
    - [MsgFreeze](#coreum.asset.ft.v1.MsgFreeze)
    - [MsgGloballyFreeze](#coreum.asset.ft.v1.MsgGloballyFreeze)
    - [MsgGloballyUnfreeze](#coreum.asset.ft.v1.MsgGloballyUnfreeze)
    - [MsgIssue](#coreum.asset.ft.v1.MsgIssue)
    - [MsgMint](#coreum.asset.ft.v1.MsgMint)
    - [MsgSetWhitelistedLimit](#coreum.asset.ft.v1.MsgSetWhitelistedLimit)
    - [MsgUnfreeze](#coreum.asset.ft.v1.MsgUnfreeze)
    - [MsgUpdateParams](#coreum.asset.ft.v1.MsgUpdateParams)
    - [MsgUpgradeTokenV1](#coreum.asset.ft.v1.MsgUpgradeTokenV1)
  
    - [Msg](#coreum.asset.ft.v1.Msg)
  
- [coreum/asset/nft/v1/event.proto](#coreum/asset/nft/v1/event.proto)
    - [EventAddedToWhitelist](#coreum.asset.nft.v1.EventAddedToWhitelist)
    - [EventClassIssued](#coreum.asset.nft.v1.EventClassIssued)
    - [EventFrozen](#coreum.asset.nft.v1.EventFrozen)
    - [EventRemovedFromWhitelist](#coreum.asset.nft.v1.EventRemovedFromWhitelist)
    - [EventUnfrozen](#coreum.asset.nft.v1.EventUnfrozen)
  
- [coreum/asset/nft/v1/genesis.proto](#coreum/asset/nft/v1/genesis.proto)
    - [BurntNFT](#coreum.asset.nft.v1.BurntNFT)
    - [FrozenNFT](#coreum.asset.nft.v1.FrozenNFT)
    - [GenesisState](#coreum.asset.nft.v1.GenesisState)
    - [WhitelistedNFTAccounts](#coreum.asset.nft.v1.WhitelistedNFTAccounts)
  
- [coreum/asset/nft/v1/nft.proto](#coreum/asset/nft/v1/nft.proto)
    - [Class](#coreum.asset.nft.v1.Class)
    - [ClassDefinition](#coreum.asset.nft.v1.ClassDefinition)
  
    - [ClassFeature](#coreum.asset.nft.v1.ClassFeature)
  
- [coreum/asset/nft/v1/params.proto](#coreum/asset/nft/v1/params.proto)
    - [Params](#coreum.asset.nft.v1.Params)
  
- [coreum/asset/nft/v1/query.proto](#coreum/asset/nft/v1/query.proto)
    - [QueryBurntNFTRequest](#coreum.asset.nft.v1.QueryBurntNFTRequest)
    - [QueryBurntNFTResponse](#coreum.asset.nft.v1.QueryBurntNFTResponse)
    - [QueryBurntNFTsInClassRequest](#coreum.asset.nft.v1.QueryBurntNFTsInClassRequest)
    - [QueryBurntNFTsInClassResponse](#coreum.asset.nft.v1.QueryBurntNFTsInClassResponse)
    - [QueryClassRequest](#coreum.asset.nft.v1.QueryClassRequest)
    - [QueryClassResponse](#coreum.asset.nft.v1.QueryClassResponse)
    - [QueryClassesRequest](#coreum.asset.nft.v1.QueryClassesRequest)
    - [QueryClassesResponse](#coreum.asset.nft.v1.QueryClassesResponse)
    - [QueryFrozenRequest](#coreum.asset.nft.v1.QueryFrozenRequest)
    - [QueryFrozenResponse](#coreum.asset.nft.v1.QueryFrozenResponse)
    - [QueryParamsRequest](#coreum.asset.nft.v1.QueryParamsRequest)
    - [QueryParamsResponse](#coreum.asset.nft.v1.QueryParamsResponse)
    - [QueryWhitelistedAccountsForNFTRequest](#coreum.asset.nft.v1.QueryWhitelistedAccountsForNFTRequest)
    - [QueryWhitelistedAccountsForNFTResponse](#coreum.asset.nft.v1.QueryWhitelistedAccountsForNFTResponse)
    - [QueryWhitelistedRequest](#coreum.asset.nft.v1.QueryWhitelistedRequest)
    - [QueryWhitelistedResponse](#coreum.asset.nft.v1.QueryWhitelistedResponse)
  
    - [Query](#coreum.asset.nft.v1.Query)
  
- [coreum/asset/nft/v1/tx.proto](#coreum/asset/nft/v1/tx.proto)
    - [EmptyResponse](#coreum.asset.nft.v1.EmptyResponse)
    - [MsgAddToWhitelist](#coreum.asset.nft.v1.MsgAddToWhitelist)
    - [MsgBurn](#coreum.asset.nft.v1.MsgBurn)
    - [MsgFreeze](#coreum.asset.nft.v1.MsgFreeze)
    - [MsgIssueClass](#coreum.asset.nft.v1.MsgIssueClass)
    - [MsgMint](#coreum.asset.nft.v1.MsgMint)
    - [MsgRemoveFromWhitelist](#coreum.asset.nft.v1.MsgRemoveFromWhitelist)
    - [MsgUnfreeze](#coreum.asset.nft.v1.MsgUnfreeze)
    - [MsgUpdateParams](#coreum.asset.nft.v1.MsgUpdateParams)
  
    - [Msg](#coreum.asset.nft.v1.Msg)
  
- [coreum/asset/nft/v1/types.proto](#coreum/asset/nft/v1/types.proto)
    - [DataBytes](#coreum.asset.nft.v1.DataBytes)
  
- [coreum/customparams/v1/genesis.proto](#coreum/customparams/v1/genesis.proto)
    - [GenesisState](#coreum.customparams.v1.GenesisState)
  
- [coreum/customparams/v1/params.proto](#coreum/customparams/v1/params.proto)
    - [StakingParams](#coreum.customparams.v1.StakingParams)
  
- [coreum/customparams/v1/query.proto](#coreum/customparams/v1/query.proto)
    - [QueryStakingParamsRequest](#coreum.customparams.v1.QueryStakingParamsRequest)
    - [QueryStakingParamsResponse](#coreum.customparams.v1.QueryStakingParamsResponse)
  
    - [Query](#coreum.customparams.v1.Query)
  
- [coreum/customparams/v1/tx.proto](#coreum/customparams/v1/tx.proto)
    - [EmptyResponse](#coreum.customparams.v1.EmptyResponse)
    - [MsgUpdateStakingParams](#coreum.customparams.v1.MsgUpdateStakingParams)
  
    - [Msg](#coreum.customparams.v1.Msg)
  
- [coreum/delay/v1/genesis.proto](#coreum/delay/v1/genesis.proto)
    - [DelayedItem](#coreum.delay.v1.DelayedItem)
    - [GenesisState](#coreum.delay.v1.GenesisState)
  
- [coreum/feemodel/v1/genesis.proto](#coreum/feemodel/v1/genesis.proto)
    - [GenesisState](#coreum.feemodel.v1.GenesisState)
  
- [coreum/feemodel/v1/params.proto](#coreum/feemodel/v1/params.proto)
    - [ModelParams](#coreum.feemodel.v1.ModelParams)
    - [Params](#coreum.feemodel.v1.Params)
  
- [coreum/feemodel/v1/query.proto](#coreum/feemodel/v1/query.proto)
    - [QueryMinGasPriceRequest](#coreum.feemodel.v1.QueryMinGasPriceRequest)
    - [QueryMinGasPriceResponse](#coreum.feemodel.v1.QueryMinGasPriceResponse)
    - [QueryParamsRequest](#coreum.feemodel.v1.QueryParamsRequest)
    - [QueryParamsResponse](#coreum.feemodel.v1.QueryParamsResponse)
    - [QueryRecommendedGasPriceRequest](#coreum.feemodel.v1.QueryRecommendedGasPriceRequest)
    - [QueryRecommendedGasPriceResponse](#coreum.feemodel.v1.QueryRecommendedGasPriceResponse)
  
    - [Query](#coreum.feemodel.v1.Query)
  
- [coreum/feemodel/v1/tx.proto](#coreum/feemodel/v1/tx.proto)
    - [EmptyResponse](#coreum.feemodel.v1.EmptyResponse)
    - [MsgUpdateParams](#coreum.feemodel.v1.MsgUpdateParams)
  
    - [Msg](#coreum.feemodel.v1.Msg)
  
- [coreum/nft/v1beta1/event.proto](#coreum/nft/v1beta1/event.proto)
    - [EventBurn](#coreum.nft.v1beta1.EventBurn)
    - [EventMint](#coreum.nft.v1beta1.EventMint)
    - [EventSend](#coreum.nft.v1beta1.EventSend)
  
- [coreum/nft/v1beta1/genesis.proto](#coreum/nft/v1beta1/genesis.proto)
    - [Entry](#coreum.nft.v1beta1.Entry)
    - [GenesisState](#coreum.nft.v1beta1.GenesisState)
  
- [coreum/nft/v1beta1/nft.proto](#coreum/nft/v1beta1/nft.proto)
    - [Class](#coreum.nft.v1beta1.Class)
    - [NFT](#coreum.nft.v1beta1.NFT)
  
- [coreum/nft/v1beta1/query.proto](#coreum/nft/v1beta1/query.proto)
    - [QueryBalanceRequest](#coreum.nft.v1beta1.QueryBalanceRequest)
    - [QueryBalanceResponse](#coreum.nft.v1beta1.QueryBalanceResponse)
    - [QueryClassRequest](#coreum.nft.v1beta1.QueryClassRequest)
    - [QueryClassResponse](#coreum.nft.v1beta1.QueryClassResponse)
    - [QueryClassesRequest](#coreum.nft.v1beta1.QueryClassesRequest)
    - [QueryClassesResponse](#coreum.nft.v1beta1.QueryClassesResponse)
    - [QueryNFTRequest](#coreum.nft.v1beta1.QueryNFTRequest)
    - [QueryNFTResponse](#coreum.nft.v1beta1.QueryNFTResponse)
    - [QueryNFTsRequest](#coreum.nft.v1beta1.QueryNFTsRequest)
    - [QueryNFTsResponse](#coreum.nft.v1beta1.QueryNFTsResponse)
    - [QueryOwnerRequest](#coreum.nft.v1beta1.QueryOwnerRequest)
    - [QueryOwnerResponse](#coreum.nft.v1beta1.QueryOwnerResponse)
    - [QuerySupplyRequest](#coreum.nft.v1beta1.QuerySupplyRequest)
    - [QuerySupplyResponse](#coreum.nft.v1beta1.QuerySupplyResponse)
  
    - [Query](#coreum.nft.v1beta1.Query)
  
- [coreum/nft/v1beta1/tx.proto](#coreum/nft/v1beta1/tx.proto)
    - [MsgSend](#coreum.nft.v1beta1.MsgSend)
    - [MsgSendResponse](#coreum.nft.v1beta1.MsgSendResponse)
  
    - [Msg](#coreum.nft.v1beta1.Msg)
  
- [cosmwasm/wasm/v1/authz.proto](#cosmwasm/wasm/v1/authz.proto)
    - [AcceptedMessageKeysFilter](#cosmwasm.wasm.v1.AcceptedMessageKeysFilter)
    - [AcceptedMessagesFilter](#cosmwasm.wasm.v1.AcceptedMessagesFilter)
    - [AllowAllMessagesFilter](#cosmwasm.wasm.v1.AllowAllMessagesFilter)
    - [CombinedLimit](#cosmwasm.wasm.v1.CombinedLimit)
    - [ContractExecutionAuthorization](#cosmwasm.wasm.v1.ContractExecutionAuthorization)
    - [ContractGrant](#cosmwasm.wasm.v1.ContractGrant)
    - [ContractMigrationAuthorization](#cosmwasm.wasm.v1.ContractMigrationAuthorization)
    - [MaxCallsLimit](#cosmwasm.wasm.v1.MaxCallsLimit)
    - [MaxFundsLimit](#cosmwasm.wasm.v1.MaxFundsLimit)
  
- [cosmwasm/wasm/v1/genesis.proto](#cosmwasm/wasm/v1/genesis.proto)
    - [Code](#cosmwasm.wasm.v1.Code)
    - [Contract](#cosmwasm.wasm.v1.Contract)
    - [GenesisState](#cosmwasm.wasm.v1.GenesisState)
    - [Sequence](#cosmwasm.wasm.v1.Sequence)
  
- [cosmwasm/wasm/v1/ibc.proto](#cosmwasm/wasm/v1/ibc.proto)
    - [MsgIBCCloseChannel](#cosmwasm.wasm.v1.MsgIBCCloseChannel)
    - [MsgIBCSend](#cosmwasm.wasm.v1.MsgIBCSend)
    - [MsgIBCSendResponse](#cosmwasm.wasm.v1.MsgIBCSendResponse)
  
- [cosmwasm/wasm/v1/proposal.proto](#cosmwasm/wasm/v1/proposal.proto)
    - [AccessConfigUpdate](#cosmwasm.wasm.v1.AccessConfigUpdate)
    - [ClearAdminProposal](#cosmwasm.wasm.v1.ClearAdminProposal)
    - [ExecuteContractProposal](#cosmwasm.wasm.v1.ExecuteContractProposal)
    - [InstantiateContract2Proposal](#cosmwasm.wasm.v1.InstantiateContract2Proposal)
    - [InstantiateContractProposal](#cosmwasm.wasm.v1.InstantiateContractProposal)
    - [MigrateContractProposal](#cosmwasm.wasm.v1.MigrateContractProposal)
    - [PinCodesProposal](#cosmwasm.wasm.v1.PinCodesProposal)
    - [StoreAndInstantiateContractProposal](#cosmwasm.wasm.v1.StoreAndInstantiateContractProposal)
    - [StoreCodeProposal](#cosmwasm.wasm.v1.StoreCodeProposal)
    - [SudoContractProposal](#cosmwasm.wasm.v1.SudoContractProposal)
    - [UnpinCodesProposal](#cosmwasm.wasm.v1.UnpinCodesProposal)
    - [UpdateAdminProposal](#cosmwasm.wasm.v1.UpdateAdminProposal)
    - [UpdateInstantiateConfigProposal](#cosmwasm.wasm.v1.UpdateInstantiateConfigProposal)
  
- [cosmwasm/wasm/v1/query.proto](#cosmwasm/wasm/v1/query.proto)
    - [CodeInfoResponse](#cosmwasm.wasm.v1.CodeInfoResponse)
    - [QueryAllContractStateRequest](#cosmwasm.wasm.v1.QueryAllContractStateRequest)
    - [QueryAllContractStateResponse](#cosmwasm.wasm.v1.QueryAllContractStateResponse)
    - [QueryCodeRequest](#cosmwasm.wasm.v1.QueryCodeRequest)
    - [QueryCodeResponse](#cosmwasm.wasm.v1.QueryCodeResponse)
    - [QueryCodesRequest](#cosmwasm.wasm.v1.QueryCodesRequest)
    - [QueryCodesResponse](#cosmwasm.wasm.v1.QueryCodesResponse)
    - [QueryContractHistoryRequest](#cosmwasm.wasm.v1.QueryContractHistoryRequest)
    - [QueryContractHistoryResponse](#cosmwasm.wasm.v1.QueryContractHistoryResponse)
    - [QueryContractInfoRequest](#cosmwasm.wasm.v1.QueryContractInfoRequest)
    - [QueryContractInfoResponse](#cosmwasm.wasm.v1.QueryContractInfoResponse)
    - [QueryContractsByCodeRequest](#cosmwasm.wasm.v1.QueryContractsByCodeRequest)
    - [QueryContractsByCodeResponse](#cosmwasm.wasm.v1.QueryContractsByCodeResponse)
    - [QueryContractsByCreatorRequest](#cosmwasm.wasm.v1.QueryContractsByCreatorRequest)
    - [QueryContractsByCreatorResponse](#cosmwasm.wasm.v1.QueryContractsByCreatorResponse)
    - [QueryParamsRequest](#cosmwasm.wasm.v1.QueryParamsRequest)
    - [QueryParamsResponse](#cosmwasm.wasm.v1.QueryParamsResponse)
    - [QueryPinnedCodesRequest](#cosmwasm.wasm.v1.QueryPinnedCodesRequest)
    - [QueryPinnedCodesResponse](#cosmwasm.wasm.v1.QueryPinnedCodesResponse)
    - [QueryRawContractStateRequest](#cosmwasm.wasm.v1.QueryRawContractStateRequest)
    - [QueryRawContractStateResponse](#cosmwasm.wasm.v1.QueryRawContractStateResponse)
    - [QuerySmartContractStateRequest](#cosmwasm.wasm.v1.QuerySmartContractStateRequest)
    - [QuerySmartContractStateResponse](#cosmwasm.wasm.v1.QuerySmartContractStateResponse)
  
    - [Query](#cosmwasm.wasm.v1.Query)
  
- [cosmwasm/wasm/v1/tx.proto](#cosmwasm/wasm/v1/tx.proto)
    - [MsgAddCodeUploadParamsAddresses](#cosmwasm.wasm.v1.MsgAddCodeUploadParamsAddresses)
    - [MsgAddCodeUploadParamsAddressesResponse](#cosmwasm.wasm.v1.MsgAddCodeUploadParamsAddressesResponse)
    - [MsgClearAdmin](#cosmwasm.wasm.v1.MsgClearAdmin)
    - [MsgClearAdminResponse](#cosmwasm.wasm.v1.MsgClearAdminResponse)
    - [MsgExecuteContract](#cosmwasm.wasm.v1.MsgExecuteContract)
    - [MsgExecuteContractResponse](#cosmwasm.wasm.v1.MsgExecuteContractResponse)
    - [MsgInstantiateContract](#cosmwasm.wasm.v1.MsgInstantiateContract)
    - [MsgInstantiateContract2](#cosmwasm.wasm.v1.MsgInstantiateContract2)
    - [MsgInstantiateContract2Response](#cosmwasm.wasm.v1.MsgInstantiateContract2Response)
    - [MsgInstantiateContractResponse](#cosmwasm.wasm.v1.MsgInstantiateContractResponse)
    - [MsgMigrateContract](#cosmwasm.wasm.v1.MsgMigrateContract)
    - [MsgMigrateContractResponse](#cosmwasm.wasm.v1.MsgMigrateContractResponse)
    - [MsgPinCodes](#cosmwasm.wasm.v1.MsgPinCodes)
    - [MsgPinCodesResponse](#cosmwasm.wasm.v1.MsgPinCodesResponse)
    - [MsgRemoveCodeUploadParamsAddresses](#cosmwasm.wasm.v1.MsgRemoveCodeUploadParamsAddresses)
    - [MsgRemoveCodeUploadParamsAddressesResponse](#cosmwasm.wasm.v1.MsgRemoveCodeUploadParamsAddressesResponse)
    - [MsgStoreAndInstantiateContract](#cosmwasm.wasm.v1.MsgStoreAndInstantiateContract)
    - [MsgStoreAndInstantiateContractResponse](#cosmwasm.wasm.v1.MsgStoreAndInstantiateContractResponse)
    - [MsgStoreCode](#cosmwasm.wasm.v1.MsgStoreCode)
    - [MsgStoreCodeResponse](#cosmwasm.wasm.v1.MsgStoreCodeResponse)
    - [MsgSudoContract](#cosmwasm.wasm.v1.MsgSudoContract)
    - [MsgSudoContractResponse](#cosmwasm.wasm.v1.MsgSudoContractResponse)
    - [MsgUnpinCodes](#cosmwasm.wasm.v1.MsgUnpinCodes)
    - [MsgUnpinCodesResponse](#cosmwasm.wasm.v1.MsgUnpinCodesResponse)
    - [MsgUpdateAdmin](#cosmwasm.wasm.v1.MsgUpdateAdmin)
    - [MsgUpdateAdminResponse](#cosmwasm.wasm.v1.MsgUpdateAdminResponse)
    - [MsgUpdateInstantiateConfig](#cosmwasm.wasm.v1.MsgUpdateInstantiateConfig)
    - [MsgUpdateInstantiateConfigResponse](#cosmwasm.wasm.v1.MsgUpdateInstantiateConfigResponse)
    - [MsgUpdateParams](#cosmwasm.wasm.v1.MsgUpdateParams)
    - [MsgUpdateParamsResponse](#cosmwasm.wasm.v1.MsgUpdateParamsResponse)
  
    - [Msg](#cosmwasm.wasm.v1.Msg)
  
- [cosmwasm/wasm/v1/types.proto](#cosmwasm/wasm/v1/types.proto)
    - [AbsoluteTxPosition](#cosmwasm.wasm.v1.AbsoluteTxPosition)
    - [AccessConfig](#cosmwasm.wasm.v1.AccessConfig)
    - [AccessTypeParam](#cosmwasm.wasm.v1.AccessTypeParam)
    - [CodeInfo](#cosmwasm.wasm.v1.CodeInfo)
    - [ContractCodeHistoryEntry](#cosmwasm.wasm.v1.ContractCodeHistoryEntry)
    - [ContractInfo](#cosmwasm.wasm.v1.ContractInfo)
    - [Model](#cosmwasm.wasm.v1.Model)
    - [Params](#cosmwasm.wasm.v1.Params)
  
    - [AccessType](#cosmwasm.wasm.v1.AccessType)
    - [ContractCodeHistoryOperationType](#cosmwasm.wasm.v1.ContractCodeHistoryOperationType)
  
- [Scalar Value Types](#scalar-value-types)



<a name="coreum/asset/ft/v1/event.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/asset/ft/v1/event.proto



<a name="coreum.asset.ft.v1.EventFrozenAmountChanged"></a>

### EventFrozenAmountChanged



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `previous_amount` | [string](#string) |  |  |
| `current_amount` | [string](#string) |  |  |






<a name="coreum.asset.ft.v1.EventIssued"></a>

### EventIssued
EventIssued is emitted on MsgIssue.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `issuer` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `subunit` | [string](#string) |  |  |
| `precision` | [uint32](#uint32) |  |  |
| `initial_amount` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `features` | [Feature](#coreum.asset.ft.v1.Feature) | repeated |  |
| `burn_rate` | [string](#string) |  |  |
| `send_commission_rate` | [string](#string) |  |  |






<a name="coreum.asset.ft.v1.EventWhitelistedAmountChanged"></a>

### EventWhitelistedAmountChanged



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `previous_amount` | [string](#string) |  |  |
| `current_amount` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="coreum/asset/ft/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/asset/ft/v1/genesis.proto



<a name="coreum.asset.ft.v1.Balance"></a>

### Balance
Balance defines an account address and balance pair used module genesis genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the address of the balance holder. |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | coins defines the different coins this balance holds. |






<a name="coreum.asset.ft.v1.GenesisState"></a>

### GenesisState
GenesisState defines the module genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#coreum.asset.ft.v1.Params) |  | params defines all the parameters of the module. |
| `tokens` | [Token](#coreum.asset.ft.v1.Token) | repeated | tokens keep the fungible token state |
| `frozen_balances` | [Balance](#coreum.asset.ft.v1.Balance) | repeated | frozen_balances contains the frozen balances on all of the accounts |
| `whitelisted_balances` | [Balance](#coreum.asset.ft.v1.Balance) | repeated | whitelisted_balances contains the whitelisted balances on all of the accounts |
| `pending_token_upgrades` | [PendingTokenUpgrade](#coreum.asset.ft.v1.PendingTokenUpgrade) | repeated | pending_token_upgrades contains pending token upgrades. |






<a name="coreum.asset.ft.v1.PendingTokenUpgrade"></a>

### PendingTokenUpgrade
PendingTokenUpgrade stores the version of pending token upgrade.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `version` | [uint32](#uint32) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="coreum/asset/ft/v1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/asset/ft/v1/params.proto



<a name="coreum.asset.ft.v1.Params"></a>

### Params
Params store gov manageable parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `issue_fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | issue_fee is the fee burnt each time new token is issued. |
| `token_upgrade_decision_timeout` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | token_upgrade_decision_timeout defines the end of the decision period for upgrading the token. |
| `token_upgrade_grace_period` | [google.protobuf.Duration](#google.protobuf.Duration) |  | token_upgrade_grace_period the period after which the token upgrade is executed effectively. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="coreum/asset/ft/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/asset/ft/v1/query.proto



<a name="coreum.asset.ft.v1.QueryBalanceRequest"></a>

### QueryBalanceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | account specifies the account onto which we query balances |
| `denom` | [string](#string) |  | denom specifies balances on a specific denom |






<a name="coreum.asset.ft.v1.QueryBalanceResponse"></a>

### QueryBalanceResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `balance` | [string](#string) |  | balance contains the balance with the queried account and denom |
| `whitelisted` | [string](#string) |  | whitelisted is the whitelisted amount of the denom on the account. |
| `frozen` | [string](#string) |  | frozen is the frozen amount of the denom on the account. |
| `locked` | [string](#string) |  | locked is the balance locked by vesting. |






<a name="coreum.asset.ft.v1.QueryFrozenBalanceRequest"></a>

### QueryFrozenBalanceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | account specifies the account onto which we query frozen balances |
| `denom` | [string](#string) |  | denom specifies frozen balances on a specific denom |






<a name="coreum.asset.ft.v1.QueryFrozenBalanceResponse"></a>

### QueryFrozenBalanceResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `balance` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | balance contains the frozen balance with the queried account and denom |






<a name="coreum.asset.ft.v1.QueryFrozenBalancesRequest"></a>

### QueryFrozenBalancesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |
| `account` | [string](#string) |  | account specifies the account onto which we query frozen balances |






<a name="coreum.asset.ft.v1.QueryFrozenBalancesResponse"></a>

### QueryFrozenBalancesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |
| `balances` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | balances contains the frozen balances on the queried account |






<a name="coreum.asset.ft.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest defines the request type for querying x/asset/ft parameters.






<a name="coreum.asset.ft.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse defines the response type for querying x/asset/ft parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#coreum.asset.ft.v1.Params) |  |  |






<a name="coreum.asset.ft.v1.QueryTokenRequest"></a>

### QueryTokenRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |






<a name="coreum.asset.ft.v1.QueryTokenResponse"></a>

### QueryTokenResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `token` | [Token](#coreum.asset.ft.v1.Token) |  |  |






<a name="coreum.asset.ft.v1.QueryTokenUpgradeStatusesRequest"></a>

### QueryTokenUpgradeStatusesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |






<a name="coreum.asset.ft.v1.QueryTokenUpgradeStatusesResponse"></a>

### QueryTokenUpgradeStatusesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `statuses` | [TokenUpgradeStatuses](#coreum.asset.ft.v1.TokenUpgradeStatuses) |  |  |






<a name="coreum.asset.ft.v1.QueryTokensRequest"></a>

### QueryTokensRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |
| `issuer` | [string](#string) |  |  |






<a name="coreum.asset.ft.v1.QueryTokensResponse"></a>

### QueryTokensResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |
| `tokens` | [Token](#coreum.asset.ft.v1.Token) | repeated |  |






<a name="coreum.asset.ft.v1.QueryWhitelistedBalanceRequest"></a>

### QueryWhitelistedBalanceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | account specifies the account onto which we query whitelisted balances |
| `denom` | [string](#string) |  | denom specifies whitelisted balances on a specific denom |






<a name="coreum.asset.ft.v1.QueryWhitelistedBalanceResponse"></a>

### QueryWhitelistedBalanceResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `balance` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | balance contains the whitelisted balance with the queried account and denom |






<a name="coreum.asset.ft.v1.QueryWhitelistedBalancesRequest"></a>

### QueryWhitelistedBalancesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |
| `account` | [string](#string) |  | account specifies the account onto which we query whitelisted balances |






<a name="coreum.asset.ft.v1.QueryWhitelistedBalancesResponse"></a>

### QueryWhitelistedBalancesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |
| `balances` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | balances contains the whitelisted balances on the queried account |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="coreum.asset.ft.v1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#coreum.asset.ft.v1.QueryParamsRequest) | [QueryParamsResponse](#coreum.asset.ft.v1.QueryParamsResponse) | Params queries the parameters of x/asset/ft module. | GET|/coreum/asset/ft/v1/params|
| `Tokens` | [QueryTokensRequest](#coreum.asset.ft.v1.QueryTokensRequest) | [QueryTokensResponse](#coreum.asset.ft.v1.QueryTokensResponse) | Tokens queries the fungible tokens of the module. | GET|/coreum/asset/ft/v1/tokens|
| `Token` | [QueryTokenRequest](#coreum.asset.ft.v1.QueryTokenRequest) | [QueryTokenResponse](#coreum.asset.ft.v1.QueryTokenResponse) | Token queries the fungible token of the module. | GET|/coreum/asset/ft/v1/tokens/{denom}|
| `TokenUpgradeStatuses` | [QueryTokenUpgradeStatusesRequest](#coreum.asset.ft.v1.QueryTokenUpgradeStatusesRequest) | [QueryTokenUpgradeStatusesResponse](#coreum.asset.ft.v1.QueryTokenUpgradeStatusesResponse) | TokenUpgradeStatuses returns token upgrades info. | GET|/coreum/asset/ft/v1/tokens/{denom}/upgrade-statuses|
| `Balance` | [QueryBalanceRequest](#coreum.asset.ft.v1.QueryBalanceRequest) | [QueryBalanceResponse](#coreum.asset.ft.v1.QueryBalanceResponse) | Balance returns balance of the denom for the account. | GET|/coreum/asset/ft/v1/accounts/{account}/balances/summary/{denom}|
| `FrozenBalances` | [QueryFrozenBalancesRequest](#coreum.asset.ft.v1.QueryFrozenBalancesRequest) | [QueryFrozenBalancesResponse](#coreum.asset.ft.v1.QueryFrozenBalancesResponse) | FrozenBalances returns all the frozen balances for the account. | GET|/coreum/asset/ft/v1/accounts/{account}/balances/frozen|
| `FrozenBalance` | [QueryFrozenBalanceRequest](#coreum.asset.ft.v1.QueryFrozenBalanceRequest) | [QueryFrozenBalanceResponse](#coreum.asset.ft.v1.QueryFrozenBalanceResponse) | FrozenBalance returns frozen balance of the denom for the account. | GET|/coreum/asset/ft/v1/accounts/{account}/balances/frozen/{denom}|
| `WhitelistedBalances` | [QueryWhitelistedBalancesRequest](#coreum.asset.ft.v1.QueryWhitelistedBalancesRequest) | [QueryWhitelistedBalancesResponse](#coreum.asset.ft.v1.QueryWhitelistedBalancesResponse) | WhitelistedBalances returns all the whitelisted balances for the account. | GET|/coreum/asset/ft/v1/accounts/{account}/balances/whitelisted|
| `WhitelistedBalance` | [QueryWhitelistedBalanceRequest](#coreum.asset.ft.v1.QueryWhitelistedBalanceRequest) | [QueryWhitelistedBalanceResponse](#coreum.asset.ft.v1.QueryWhitelistedBalanceResponse) | WhitelistedBalance returns whitelisted balance of the denom for the account. | GET|/coreum/asset/ft/v1/accounts/{account}/balances/whitelisted/{denom}|

 <!-- end services -->



<a name="coreum/asset/ft/v1/token.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/asset/ft/v1/token.proto



<a name="coreum.asset.ft.v1.Definition"></a>

### Definition
Definition defines the fungible token settings to store.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `issuer` | [string](#string) |  |  |
| `features` | [Feature](#coreum.asset.ft.v1.Feature) | repeated |  |
| `burn_rate` | [string](#string) |  | burn_rate is a number between 0 and 1 which will be multiplied by send amount to determine burn_amount. This value will be burnt on top of the send amount. |
| `send_commission_rate` | [string](#string) |  | send_commission_rate is a number between 0 and 1 which will be multiplied by send amount to determine amount sent to the token issuer account. |
| `version` | [uint32](#uint32) |  |  |






<a name="coreum.asset.ft.v1.DelayedTokenUpgradeV1"></a>

### DelayedTokenUpgradeV1
DelayedTokenUpgradeV1 is executed by the delay module when it's time to enable IBC.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |






<a name="coreum.asset.ft.v1.Token"></a>

### Token
Token is a full representation of the fungible token.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `issuer` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `subunit` | [string](#string) |  |  |
| `precision` | [uint32](#uint32) |  |  |
| `description` | [string](#string) |  |  |
| `globally_frozen` | [bool](#bool) |  |  |
| `features` | [Feature](#coreum.asset.ft.v1.Feature) | repeated |  |
| `burn_rate` | [string](#string) |  | burn_rate is a number between 0 and 1 which will be multiplied by send amount to determine burn_amount. This value will be burnt on top of the send amount. |
| `send_commission_rate` | [string](#string) |  | send_commission_rate is a number between 0 and 1 which will be multiplied by send amount to determine amount sent to the token issuer account. |
| `version` | [uint32](#uint32) |  |  |






<a name="coreum.asset.ft.v1.TokenUpgradeStatuses"></a>

### TokenUpgradeStatuses
TokenUpgradeStatuses defines all statuses of the token migrations.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `v1` | [TokenUpgradeV1Status](#coreum.asset.ft.v1.TokenUpgradeV1Status) |  |  |






<a name="coreum.asset.ft.v1.TokenUpgradeV1Status"></a>

### TokenUpgradeV1Status
TokenUpgradeV1Status defines the current status of the v1 token migration.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ibc_enabled` | [bool](#bool) |  |  |
| `start_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `end_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |





 <!-- end messages -->


<a name="coreum.asset.ft.v1.Feature"></a>

### Feature
Feature defines possible features of fungible token.

| Name | Number | Description |
| ---- | ------ | ----------- |
| minting | 0 |  |
| burning | 1 |  |
| freezing | 2 |  |
| whitelisting | 3 |  |
| ibc | 4 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="coreum/asset/ft/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/asset/ft/v1/tx.proto



<a name="coreum.asset.ft.v1.EmptyResponse"></a>

### EmptyResponse







<a name="coreum.asset.ft.v1.MsgBurn"></a>

### MsgBurn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `coin` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="coreum.asset.ft.v1.MsgFreeze"></a>

### MsgFreeze



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `account` | [string](#string) |  |  |
| `coin` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="coreum.asset.ft.v1.MsgGloballyFreeze"></a>

### MsgGloballyFreeze



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |






<a name="coreum.asset.ft.v1.MsgGloballyUnfreeze"></a>

### MsgGloballyUnfreeze



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |






<a name="coreum.asset.ft.v1.MsgIssue"></a>

### MsgIssue
MsgIssue defines message to issue new fungible token.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `issuer` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `subunit` | [string](#string) |  |  |
| `precision` | [uint32](#uint32) |  |  |
| `initial_amount` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `features` | [Feature](#coreum.asset.ft.v1.Feature) | repeated |  |
| `burn_rate` | [string](#string) |  | burn_rate is a number between 0 and 1 which will be multiplied by send amount to determine burn_amount. This value will be burnt on top of the send amount. |
| `send_commission_rate` | [string](#string) |  | send_commission_rate is a number between 0 and 1 which will be multiplied by send amount to determine amount sent to the token issuer account. |






<a name="coreum.asset.ft.v1.MsgMint"></a>

### MsgMint



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `coin` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="coreum.asset.ft.v1.MsgSetWhitelistedLimit"></a>

### MsgSetWhitelistedLimit



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `account` | [string](#string) |  |  |
| `coin` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="coreum.asset.ft.v1.MsgUnfreeze"></a>

### MsgUnfreeze



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `account` | [string](#string) |  |  |
| `coin` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="coreum.asset.ft.v1.MsgUpdateParams"></a>

### MsgUpdateParams



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  |  |
| `params` | [Params](#coreum.asset.ft.v1.Params) |  |  |






<a name="coreum.asset.ft.v1.MsgUpgradeTokenV1"></a>

### MsgUpgradeTokenV1
MsgUpgradeTokenV1 is the message upgrading token to V1.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `ibc_enabled` | [bool](#bool) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="coreum.asset.ft.v1.Msg"></a>

### Msg
Msg defines the Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Issue` | [MsgIssue](#coreum.asset.ft.v1.MsgIssue) | [EmptyResponse](#coreum.asset.ft.v1.EmptyResponse) | Issue defines a method to issue a new fungible token. | |
| `Mint` | [MsgMint](#coreum.asset.ft.v1.MsgMint) | [EmptyResponse](#coreum.asset.ft.v1.EmptyResponse) | Mint mints new fungible tokens. | |
| `Burn` | [MsgBurn](#coreum.asset.ft.v1.MsgBurn) | [EmptyResponse](#coreum.asset.ft.v1.EmptyResponse) | Burn burns the specified fungible tokens from senders balance if the sender has enough balance. | |
| `Freeze` | [MsgFreeze](#coreum.asset.ft.v1.MsgFreeze) | [EmptyResponse](#coreum.asset.ft.v1.EmptyResponse) | Freeze freezes a part of the fungible tokens in an account, only if the freezable feature is enabled on that token. | |
| `Unfreeze` | [MsgUnfreeze](#coreum.asset.ft.v1.MsgUnfreeze) | [EmptyResponse](#coreum.asset.ft.v1.EmptyResponse) | Unfreeze unfreezes a part of the frozen fungible tokens in an account, only if there are such frozen tokens on that account. | |
| `GloballyFreeze` | [MsgGloballyFreeze](#coreum.asset.ft.v1.MsgGloballyFreeze) | [EmptyResponse](#coreum.asset.ft.v1.EmptyResponse) | GloballyFreeze freezes fungible token so no operations are allowed with it before unfrozen. This operation is idempotent so global freeze of already frozen token does nothing. | |
| `GloballyUnfreeze` | [MsgGloballyUnfreeze](#coreum.asset.ft.v1.MsgGloballyUnfreeze) | [EmptyResponse](#coreum.asset.ft.v1.EmptyResponse) | GloballyUnfreeze unfreezes fungible token and unblocks basic operations on it. This operation is idempotent so global unfreezing of non-frozen token does nothing. | |
| `SetWhitelistedLimit` | [MsgSetWhitelistedLimit](#coreum.asset.ft.v1.MsgSetWhitelistedLimit) | [EmptyResponse](#coreum.asset.ft.v1.EmptyResponse) | SetWhitelistedLimit sets the limit of how many tokens a specific account may hold. | |
| `UpgradeTokenV1` | [MsgUpgradeTokenV1](#coreum.asset.ft.v1.MsgUpgradeTokenV1) | [EmptyResponse](#coreum.asset.ft.v1.EmptyResponse) | TokenUpgradeV1 upgrades token to version V1. | |
| `UpdateParams` | [MsgUpdateParams](#coreum.asset.ft.v1.MsgUpdateParams) | [EmptyResponse](#coreum.asset.ft.v1.EmptyResponse) | UpdateParams is a governance operation to modify the parameters of the module. NOTE: all parameters must be provided. | |

 <!-- end services -->



<a name="coreum/asset/nft/v1/event.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/asset/nft/v1/event.proto



<a name="coreum.asset.nft.v1.EventAddedToWhitelist"></a>

### EventAddedToWhitelist



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class_id` | [string](#string) |  |  |
| `id` | [string](#string) |  |  |
| `account` | [string](#string) |  |  |






<a name="coreum.asset.nft.v1.EventClassIssued"></a>

### EventClassIssued
EventClassIssued is emitted on MsgIssueClass.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `issuer` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `name` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `uri` | [string](#string) |  |  |
| `uri_hash` | [string](#string) |  |  |
| `features` | [ClassFeature](#coreum.asset.nft.v1.ClassFeature) | repeated |  |
| `royalty_rate` | [string](#string) |  |  |






<a name="coreum.asset.nft.v1.EventFrozen"></a>

### EventFrozen



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class_id` | [string](#string) |  |  |
| `id` | [string](#string) |  |  |
| `owner` | [string](#string) |  |  |






<a name="coreum.asset.nft.v1.EventRemovedFromWhitelist"></a>

### EventRemovedFromWhitelist



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class_id` | [string](#string) |  |  |
| `id` | [string](#string) |  |  |
| `account` | [string](#string) |  |  |






<a name="coreum.asset.nft.v1.EventUnfrozen"></a>

### EventUnfrozen



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class_id` | [string](#string) |  |  |
| `id` | [string](#string) |  |  |
| `owner` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="coreum/asset/nft/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/asset/nft/v1/genesis.proto



<a name="coreum.asset.nft.v1.BurntNFT"></a>

### BurntNFT



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `classID` | [string](#string) |  |  |
| `nftIDs` | [string](#string) | repeated |  |






<a name="coreum.asset.nft.v1.FrozenNFT"></a>

### FrozenNFT



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `classID` | [string](#string) |  |  |
| `nftIDs` | [string](#string) | repeated |  |






<a name="coreum.asset.nft.v1.GenesisState"></a>

### GenesisState
GenesisState defines the nftasset module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#coreum.asset.nft.v1.Params) |  | params defines all the parameters of the module. |
| `class_definitions` | [ClassDefinition](#coreum.asset.nft.v1.ClassDefinition) | repeated | class_definitions keep the non-fungible token class definitions state |
| `frozen_nfts` | [FrozenNFT](#coreum.asset.nft.v1.FrozenNFT) | repeated |  |
| `whitelisted_nft_accounts` | [WhitelistedNFTAccounts](#coreum.asset.nft.v1.WhitelistedNFTAccounts) | repeated |  |
| `burnt_nfts` | [BurntNFT](#coreum.asset.nft.v1.BurntNFT) | repeated |  |






<a name="coreum.asset.nft.v1.WhitelistedNFTAccounts"></a>

### WhitelistedNFTAccounts



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `classID` | [string](#string) |  |  |
| `nftID` | [string](#string) |  |  |
| `accounts` | [string](#string) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="coreum/asset/nft/v1/nft.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/asset/nft/v1/nft.proto



<a name="coreum.asset.nft.v1.Class"></a>

### Class
Class is a full representation of the non-fungible token class.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `issuer` | [string](#string) |  |  |
| `name` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `uri` | [string](#string) |  |  |
| `uri_hash` | [string](#string) |  |  |
| `data` | [google.protobuf.Any](#google.protobuf.Any) |  |  |
| `features` | [ClassFeature](#coreum.asset.nft.v1.ClassFeature) | repeated |  |
| `royalty_rate` | [string](#string) |  | royalty_rate is a number between 0 and 1,which will be used in coreum native Dex. whenever an NFT this class is traded on the Dex, the traded amount will be multiplied by this value that will be transferred to the issuer of the NFT. |






<a name="coreum.asset.nft.v1.ClassDefinition"></a>

### ClassDefinition
ClassDefinition defines the non-fungible token class settings to store.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `issuer` | [string](#string) |  |  |
| `features` | [ClassFeature](#coreum.asset.nft.v1.ClassFeature) | repeated |  |
| `royalty_rate` | [string](#string) |  | royalty_rate is a number between 0 and 1,which will be used in coreum native Dex. whenever an NFT this class is traded on the Dex, the traded amount will be multiplied by this value that will be transferred to the issuer of the NFT. |





 <!-- end messages -->


<a name="coreum.asset.nft.v1.ClassFeature"></a>

### ClassFeature
ClassFeature defines possible features of non-fungible token class.

| Name | Number | Description |
| ---- | ------ | ----------- |
| burning | 0 |  |
| freezing | 1 |  |
| whitelisting | 2 |  |
| disable_sending | 3 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="coreum/asset/nft/v1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/asset/nft/v1/params.proto



<a name="coreum.asset.nft.v1.Params"></a>

### Params
Params store gov manageable parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `mint_fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | mint_fee is the fee burnt each time new NFT is minted |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="coreum/asset/nft/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/asset/nft/v1/query.proto



<a name="coreum.asset.nft.v1.QueryBurntNFTRequest"></a>

### QueryBurntNFTRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class_id` | [string](#string) |  |  |
| `nft_id` | [string](#string) |  |  |






<a name="coreum.asset.nft.v1.QueryBurntNFTResponse"></a>

### QueryBurntNFTResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `burnt` | [bool](#bool) |  |  |






<a name="coreum.asset.nft.v1.QueryBurntNFTsInClassRequest"></a>

### QueryBurntNFTsInClassRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |
| `class_id` | [string](#string) |  |  |






<a name="coreum.asset.nft.v1.QueryBurntNFTsInClassResponse"></a>

### QueryBurntNFTsInClassResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |
| `nft_ids` | [string](#string) | repeated |  |






<a name="coreum.asset.nft.v1.QueryClassRequest"></a>

### QueryClassRequest
QueryTokenRequest is request type for the Query/Class RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  | we don't use the gogoproto.customname here since the google.api.http ignores it and generates invalid code. |






<a name="coreum.asset.nft.v1.QueryClassResponse"></a>

### QueryClassResponse
QueryClassResponse is response type for the Query/Class RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class` | [Class](#coreum.asset.nft.v1.Class) |  |  |






<a name="coreum.asset.nft.v1.QueryClassesRequest"></a>

### QueryClassesRequest
QueryTokenRequest is request type for the Query/Classes RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |
| `issuer` | [string](#string) |  |  |






<a name="coreum.asset.nft.v1.QueryClassesResponse"></a>

### QueryClassesResponse
QueryClassResponse is response type for the Query/Classes RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |
| `classes` | [Class](#coreum.asset.nft.v1.Class) | repeated |  |






<a name="coreum.asset.nft.v1.QueryFrozenRequest"></a>

### QueryFrozenRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `class_id` | [string](#string) |  |  |






<a name="coreum.asset.nft.v1.QueryFrozenResponse"></a>

### QueryFrozenResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `frozen` | [bool](#bool) |  |  |






<a name="coreum.asset.nft.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest defines the request type for querying x/asset/nft parameters.






<a name="coreum.asset.nft.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse defines the response type for querying x/asset/nft parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#coreum.asset.nft.v1.Params) |  |  |






<a name="coreum.asset.nft.v1.QueryWhitelistedAccountsForNFTRequest"></a>

### QueryWhitelistedAccountsForNFTRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |
| `id` | [string](#string) |  |  |
| `class_id` | [string](#string) |  |  |






<a name="coreum.asset.nft.v1.QueryWhitelistedAccountsForNFTResponse"></a>

### QueryWhitelistedAccountsForNFTResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |
| `accounts` | [string](#string) | repeated |  |






<a name="coreum.asset.nft.v1.QueryWhitelistedRequest"></a>

### QueryWhitelistedRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `class_id` | [string](#string) |  |  |
| `account` | [string](#string) |  |  |






<a name="coreum.asset.nft.v1.QueryWhitelistedResponse"></a>

### QueryWhitelistedResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `whitelisted` | [bool](#bool) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="coreum.asset.nft.v1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#coreum.asset.nft.v1.QueryParamsRequest) | [QueryParamsResponse](#coreum.asset.nft.v1.QueryParamsResponse) | Params queries the parameters of x/asset/nft module. | GET|/coreum/asset/nft/v1/params|
| `Class` | [QueryClassRequest](#coreum.asset.nft.v1.QueryClassRequest) | [QueryClassResponse](#coreum.asset.nft.v1.QueryClassResponse) | Class queries the non-fungible token class of the module. | GET|/coreum/asset/nft/v1/classes/{id}|
| `Classes` | [QueryClassesRequest](#coreum.asset.nft.v1.QueryClassesRequest) | [QueryClassesResponse](#coreum.asset.nft.v1.QueryClassesResponse) | Classes queries the non-fungible token classes of the module. | GET|/coreum/asset/nft/v1/classes|
| `Frozen` | [QueryFrozenRequest](#coreum.asset.nft.v1.QueryFrozenRequest) | [QueryFrozenResponse](#coreum.asset.nft.v1.QueryFrozenResponse) | Frozen queries to check if an NFT is frozen or not. | GET|/coreum/asset/nft/v1/classes/{class_id}/nfts/{id}/frozen|
| `Whitelisted` | [QueryWhitelistedRequest](#coreum.asset.nft.v1.QueryWhitelistedRequest) | [QueryWhitelistedResponse](#coreum.asset.nft.v1.QueryWhitelistedResponse) | Whitelisted queries to check if an account is whitelited to hold an NFT or not. | GET|/coreum/asset/nft/v1/classes/{class_id}/nfts/{id}/whitelisted/{account}|
| `WhitelistedAccountsForNFT` | [QueryWhitelistedAccountsForNFTRequest](#coreum.asset.nft.v1.QueryWhitelistedAccountsForNFTRequest) | [QueryWhitelistedAccountsForNFTResponse](#coreum.asset.nft.v1.QueryWhitelistedAccountsForNFTResponse) | WhitelistedAccountsForNFT returns the list of accounts which are whitelisted to hold this NFT. | GET|/coreum/asset/nft/v1/classes/{class_id}/nfts/{id}/whitelisted|
| `BurntNFT` | [QueryBurntNFTRequest](#coreum.asset.nft.v1.QueryBurntNFTRequest) | [QueryBurntNFTResponse](#coreum.asset.nft.v1.QueryBurntNFTResponse) | BurntNFTsInClass checks if an nft if is in burnt NFTs list. | GET|/coreum/asset/nft/v1/classes/{class_id}/burnt/{nft_id}|
| `BurntNFTsInClass` | [QueryBurntNFTsInClassRequest](#coreum.asset.nft.v1.QueryBurntNFTsInClassRequest) | [QueryBurntNFTsInClassResponse](#coreum.asset.nft.v1.QueryBurntNFTsInClassResponse) | BurntNFTsInClass returns the list of burnt nfts in a class. | GET|/coreum/asset/nft/v1/classes/{class_id}/burnt|

 <!-- end services -->



<a name="coreum/asset/nft/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/asset/nft/v1/tx.proto



<a name="coreum.asset.nft.v1.EmptyResponse"></a>

### EmptyResponse







<a name="coreum.asset.nft.v1.MsgAddToWhitelist"></a>

### MsgAddToWhitelist



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `class_id` | [string](#string) |  |  |
| `id` | [string](#string) |  |  |
| `account` | [string](#string) |  |  |






<a name="coreum.asset.nft.v1.MsgBurn"></a>

### MsgBurn
MsgBurn defines message for the Burn method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `class_id` | [string](#string) |  |  |
| `id` | [string](#string) |  |  |






<a name="coreum.asset.nft.v1.MsgFreeze"></a>

### MsgFreeze



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `class_id` | [string](#string) |  |  |
| `id` | [string](#string) |  |  |






<a name="coreum.asset.nft.v1.MsgIssueClass"></a>

### MsgIssueClass
MsgIssueClass defines message for the IssueClass method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `issuer` | [string](#string) |  |  |
| `symbol` | [string](#string) |  |  |
| `name` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `uri` | [string](#string) |  |  |
| `uri_hash` | [string](#string) |  |  |
| `data` | [google.protobuf.Any](#google.protobuf.Any) |  |  |
| `features` | [ClassFeature](#coreum.asset.nft.v1.ClassFeature) | repeated |  |
| `royalty_rate` | [string](#string) |  |  |






<a name="coreum.asset.nft.v1.MsgMint"></a>

### MsgMint
MsgMint defines message for the Mint method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `class_id` | [string](#string) |  |  |
| `id` | [string](#string) |  |  |
| `uri` | [string](#string) |  |  |
| `uri_hash` | [string](#string) |  |  |
| `data` | [google.protobuf.Any](#google.protobuf.Any) |  |  |






<a name="coreum.asset.nft.v1.MsgRemoveFromWhitelist"></a>

### MsgRemoveFromWhitelist



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `class_id` | [string](#string) |  |  |
| `id` | [string](#string) |  |  |
| `account` | [string](#string) |  |  |






<a name="coreum.asset.nft.v1.MsgUnfreeze"></a>

### MsgUnfreeze



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `class_id` | [string](#string) |  |  |
| `id` | [string](#string) |  |  |






<a name="coreum.asset.nft.v1.MsgUpdateParams"></a>

### MsgUpdateParams



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  |  |
| `params` | [Params](#coreum.asset.nft.v1.Params) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="coreum.asset.nft.v1.Msg"></a>

### Msg
Msg defines the Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `IssueClass` | [MsgIssueClass](#coreum.asset.nft.v1.MsgIssueClass) | [EmptyResponse](#coreum.asset.nft.v1.EmptyResponse) | IssueClass creates new non-fungible token class. | |
| `Mint` | [MsgMint](#coreum.asset.nft.v1.MsgMint) | [EmptyResponse](#coreum.asset.nft.v1.EmptyResponse) | Mint mints new non-fungible token in the class. | |
| `Burn` | [MsgBurn](#coreum.asset.nft.v1.MsgBurn) | [EmptyResponse](#coreum.asset.nft.v1.EmptyResponse) | Burn burns the existing non-fungible token in the class. | |
| `Freeze` | [MsgFreeze](#coreum.asset.nft.v1.MsgFreeze) | [EmptyResponse](#coreum.asset.nft.v1.EmptyResponse) | Freeze freezes an NFT | |
| `Unfreeze` | [MsgUnfreeze](#coreum.asset.nft.v1.MsgUnfreeze) | [EmptyResponse](#coreum.asset.nft.v1.EmptyResponse) | Unfreeze removes the freeze effect already put on an NFT | |
| `AddToWhitelist` | [MsgAddToWhitelist](#coreum.asset.nft.v1.MsgAddToWhitelist) | [EmptyResponse](#coreum.asset.nft.v1.EmptyResponse) | AddToWhitelist sets the account as whitelisted to hold the NFT | |
| `RemoveFromWhitelist` | [MsgRemoveFromWhitelist](#coreum.asset.nft.v1.MsgRemoveFromWhitelist) | [EmptyResponse](#coreum.asset.nft.v1.EmptyResponse) | RemoveFromWhitelist removes an account from whitelisted list of the NFT | |
| `UpdateParams` | [MsgUpdateParams](#coreum.asset.nft.v1.MsgUpdateParams) | [EmptyResponse](#coreum.asset.nft.v1.EmptyResponse) | UpdateParams is a governance operation that sets the parameters of the module. NOTE: all parameters must be provided. | |

 <!-- end services -->



<a name="coreum/asset/nft/v1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/asset/nft/v1/types.proto



<a name="coreum.asset.nft.v1.DataBytes"></a>

### DataBytes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `Data` | [bytes](#bytes) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="coreum/customparams/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/customparams/v1/genesis.proto



<a name="coreum.customparams.v1.GenesisState"></a>

### GenesisState
GenesisState defines the module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `staking_params` | [StakingParams](#coreum.customparams.v1.StakingParams) |  | staking_params defines staking parameters of the module. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="coreum/customparams/v1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/customparams/v1/params.proto



<a name="coreum.customparams.v1.StakingParams"></a>

### StakingParams
StakingParams defines the set of additional staking params for the staking module wrapper.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `min_self_delegation` | [string](#string) |  | min_self_delegation is the validators global self declared minimum for delegation. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="coreum/customparams/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/customparams/v1/query.proto



<a name="coreum.customparams.v1.QueryStakingParamsRequest"></a>

### QueryStakingParamsRequest
QueryStakingParamsRequest defines the request type for querying x/customparams staking parameters.






<a name="coreum.customparams.v1.QueryStakingParamsResponse"></a>

### QueryStakingParamsResponse
QueryStakingParamsResponse defines the response type for querying x/customparams staking parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [StakingParams](#coreum.customparams.v1.StakingParams) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="coreum.customparams.v1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `StakingParams` | [QueryStakingParamsRequest](#coreum.customparams.v1.QueryStakingParamsRequest) | [QueryStakingParamsResponse](#coreum.customparams.v1.QueryStakingParamsResponse) | StakingParams queries the staking parameters of the module. | GET|/coreum/customparams/v1/stakingparams|

 <!-- end services -->



<a name="coreum/customparams/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/customparams/v1/tx.proto



<a name="coreum.customparams.v1.EmptyResponse"></a>

### EmptyResponse







<a name="coreum.customparams.v1.MsgUpdateStakingParams"></a>

### MsgUpdateStakingParams



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  |  |
| `staking_params` | [StakingParams](#coreum.customparams.v1.StakingParams) |  | staking_params holds the parameters related to the staking module. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="coreum.customparams.v1.Msg"></a>

### Msg
Msg defines the Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `UpdateStakingParams` | [MsgUpdateStakingParams](#coreum.customparams.v1.MsgUpdateStakingParams) | [EmptyResponse](#coreum.customparams.v1.EmptyResponse) | UpdateStakingParams is a governance operation that sets the staking parameter. NOTE: all parameters must be provided. | |

 <!-- end services -->



<a name="coreum/delay/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/delay/v1/genesis.proto



<a name="coreum.delay.v1.DelayedItem"></a>

### DelayedItem



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |
| `execution_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `data` | [google.protobuf.Any](#google.protobuf.Any) |  |  |






<a name="coreum.delay.v1.GenesisState"></a>

### GenesisState
GenesisState defines the module genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `delayed_items` | [DelayedItem](#coreum.delay.v1.DelayedItem) | repeated | tokens keep the fungible token state |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="coreum/feemodel/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/feemodel/v1/genesis.proto



<a name="coreum.feemodel.v1.GenesisState"></a>

### GenesisState
GenesisState defines the module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#coreum.feemodel.v1.Params) |  | params defines all the parameters of the module. |
| `min_gas_price` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  | min_gas_price is the current minimum gas price required by the chain. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="coreum/feemodel/v1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/feemodel/v1/params.proto



<a name="coreum.feemodel.v1.ModelParams"></a>

### ModelParams
ModelParams define fee model params.
There are four regions on the fee model curve
- between 0 and "long average block gas" where gas price goes down exponentially from InitialGasPrice to gas price with maximum discount (InitialGasPrice * (1 - MaxDiscount))
- between "long average block gas" and EscalationStartBlockGas (EscalationStartBlockGas = MaxBlockGas * EscalationStartFraction) where we offer gas price with maximum discount all the time
- between EscalationStartBlockGas (EscalationStartBlockGas = MaxBlockGas * EscalationStartFraction) and MaxBlockGas where price goes up rapidly (being an output of a power function) from gas price with maximum discount to MaxGasPrice  (MaxGasPrice = InitialGasPrice * MaxGasMultiplier)
- above MaxBlockGas (if it happens for any reason) where price is equal to MaxGasPrice (MaxGasPrice = InitialGasPrice * MaxGasMultiplier)

The input (x value) for that function is calculated by taking short block gas average.
Price (y value) being an output of the fee model is used as the minimum gas price for next block.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `initial_gas_price` | [string](#string) |  | initial_gas_price is used when block gas short average is 0. It happens when there are no transactions being broadcasted. This value is also used to initialize gas price on brand-new chain. |
| `max_gas_price_multiplier` | [string](#string) |  | max_gas_price_multiplier is used to compute max_gas_price (max_gas_price = initial_gas_price * max_gas_price_multiplier). Max gas price is charged when block gas short average is greater than or equal to MaxBlockGas. This value is used to limit gas price escalation to avoid having possible infinity GasPrice value otherwise. |
| `max_discount` | [string](#string) |  | max_discount is th maximum discount we offer on top of initial gas price if short average block gas is between long average block gas and escalation start block gas. |
| `escalation_start_fraction` | [string](#string) |  | escalation_start_fraction defines fraction of max block gas usage where gas price escalation starts if short average block gas is higher than this value. |
| `max_block_gas` | [int64](#int64) |  | max_block_gas sets the maximum capacity of block. This is enforced on tendermint level in genesis configuration. Once short average block gas goes above this value, gas price is a flat line equal to MaxGasPrice. |
| `short_ema_block_length` | [uint32](#uint32) |  | short_ema_block_length defines inertia for short average long gas in EMA model. The equation is: NewAverage = ((ShortAverageBlockLength - 1)*PreviousAverage + GasUsedByCurrentBlock) / ShortAverageBlockLength The value might be interpreted as the number of blocks which are taken to calculate the average. It would be exactly like that in SMA model, in EMA this is an approximation. |
| `long_ema_block_length` | [uint32](#uint32) |  | long_ema_block_length defines inertia for long average block gas in EMA model. The equation is: NewAverage = ((LongAverageBlockLength - 1)*PreviousAverage + GasUsedByCurrentBlock) / LongAverageBlockLength The value might be interpreted as the number of blocks which are taken to calculate the average. It would be exactly like that in SMA model, in EMA this is an approximation. |






<a name="coreum.feemodel.v1.Params"></a>

### Params
Params store gov manageable feemodel parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `model` | [ModelParams](#coreum.feemodel.v1.ModelParams) |  | model is a fee model params. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="coreum/feemodel/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/feemodel/v1/query.proto



<a name="coreum.feemodel.v1.QueryMinGasPriceRequest"></a>

### QueryMinGasPriceRequest
QueryMinGasPriceRequest is the request type for the Query/MinGasPrice RPC method.






<a name="coreum.feemodel.v1.QueryMinGasPriceResponse"></a>

### QueryMinGasPriceResponse
QueryMinGasPriceResponse is the response type for the Query/MinGasPrice RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `min_gas_price` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  | min_gas_price is the current minimum gas price required by the network. |






<a name="coreum.feemodel.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest defines the request type for querying x/feemodel parameters.






<a name="coreum.feemodel.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse defines the response type for querying x/feemodel parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#coreum.feemodel.v1.Params) |  |  |






<a name="coreum.feemodel.v1.QueryRecommendedGasPriceRequest"></a>

### QueryRecommendedGasPriceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `after_blocks` | [uint32](#uint32) |  |  |






<a name="coreum.feemodel.v1.QueryRecommendedGasPriceResponse"></a>

### QueryRecommendedGasPriceResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `low` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  |  |
| `med` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  |  |
| `high` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="coreum.feemodel.v1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `MinGasPrice` | [QueryMinGasPriceRequest](#coreum.feemodel.v1.QueryMinGasPriceRequest) | [QueryMinGasPriceResponse](#coreum.feemodel.v1.QueryMinGasPriceResponse) | MinGasPrice queries the current minimum gas price required by the network. | GET|/coreum/feemodel/v1/min_gas_price|
| `RecommendedGasPrice` | [QueryRecommendedGasPriceRequest](#coreum.feemodel.v1.QueryRecommendedGasPriceRequest) | [QueryRecommendedGasPriceResponse](#coreum.feemodel.v1.QueryRecommendedGasPriceResponse) | RecommendedGasPrice queries the recommended gas price for the next n blocks. | GET|/coreum/feemodel/v1/recommended_gas_price|
| `Params` | [QueryParamsRequest](#coreum.feemodel.v1.QueryParamsRequest) | [QueryParamsResponse](#coreum.feemodel.v1.QueryParamsResponse) | Params queries the parameters of x/feemodel module. | GET|/coreum/feemodel/v1/params|

 <!-- end services -->



<a name="coreum/feemodel/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/feemodel/v1/tx.proto



<a name="coreum.feemodel.v1.EmptyResponse"></a>

### EmptyResponse







<a name="coreum.feemodel.v1.MsgUpdateParams"></a>

### MsgUpdateParams



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  |  |
| `params` | [Params](#coreum.feemodel.v1.Params) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="coreum.feemodel.v1.Msg"></a>

### Msg
Msg defines the Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `UpdateParams` | [MsgUpdateParams](#coreum.feemodel.v1.MsgUpdateParams) | [EmptyResponse](#coreum.feemodel.v1.EmptyResponse) | UpdateParams is a governance operation which allows fee models params to be modified. NOTE: All parmas must be provided. | |

 <!-- end services -->



<a name="coreum/nft/v1beta1/event.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/nft/v1beta1/event.proto



<a name="coreum.nft.v1beta1.EventBurn"></a>

### EventBurn
EventBurn is emitted on Burn


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class_id` | [string](#string) |  |  |
| `id` | [string](#string) |  |  |
| `owner` | [string](#string) |  |  |






<a name="coreum.nft.v1beta1.EventMint"></a>

### EventMint
EventMint is emitted on Mint


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class_id` | [string](#string) |  |  |
| `id` | [string](#string) |  |  |
| `owner` | [string](#string) |  |  |






<a name="coreum.nft.v1beta1.EventSend"></a>

### EventSend
EventSend is emitted on Msg/Send


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class_id` | [string](#string) |  |  |
| `id` | [string](#string) |  |  |
| `sender` | [string](#string) |  |  |
| `receiver` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="coreum/nft/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/nft/v1beta1/genesis.proto



<a name="coreum.nft.v1beta1.Entry"></a>

### Entry
Entry Defines all nft owned by a person


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | owner is the owner address of the following nft |
| `nfts` | [NFT](#coreum.nft.v1beta1.NFT) | repeated | nfts is a group of nfts of the same owner |






<a name="coreum.nft.v1beta1.GenesisState"></a>

### GenesisState
GenesisState defines the nft module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `classes` | [Class](#coreum.nft.v1beta1.Class) | repeated | class defines the class of the nft type. |
| `entries` | [Entry](#coreum.nft.v1beta1.Entry) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="coreum/nft/v1beta1/nft.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/nft/v1beta1/nft.proto



<a name="coreum.nft.v1beta1.Class"></a>

### Class
Class defines the class of the nft type.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  | id defines the unique identifier of the NFT classification, similar to the contract address of ERC721 |
| `name` | [string](#string) |  | name defines the human-readable name of the NFT classification. Optional |
| `symbol` | [string](#string) |  | symbol is an abbreviated name for nft classification. Optional |
| `description` | [string](#string) |  | description is a brief description of nft classification. Optional |
| `uri` | [string](#string) |  | uri for the class metadata stored off chain. It can define schema for Class and NFT `Data` attributes. Optional |
| `uri_hash` | [string](#string) |  | uri_hash is a hash of the document pointed by uri. Optional |
| `data` | [google.protobuf.Any](#google.protobuf.Any) |  | data is the app specific metadata of the NFT class. Optional |






<a name="coreum.nft.v1beta1.NFT"></a>

### NFT
NFT defines the NFT.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class_id` | [string](#string) |  | class_id associated with the NFT, similar to the contract address of ERC721 |
| `id` | [string](#string) |  | id is a unique identifier of the NFT |
| `uri` | [string](#string) |  | uri for the NFT metadata stored off chain |
| `uri_hash` | [string](#string) |  | uri_hash is a hash of the document pointed by uri |
| `data` | [google.protobuf.Any](#google.protobuf.Any) |  | data is an app specific data of the NFT. Optional |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="coreum/nft/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/nft/v1beta1/query.proto



<a name="coreum.nft.v1beta1.QueryBalanceRequest"></a>

### QueryBalanceRequest
QueryBalanceRequest is the request type for the Query/Balance RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class_id` | [string](#string) |  |  |
| `owner` | [string](#string) |  |  |






<a name="coreum.nft.v1beta1.QueryBalanceResponse"></a>

### QueryBalanceResponse
QueryBalanceResponse is the response type for the Query/Balance RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [uint64](#uint64) |  |  |






<a name="coreum.nft.v1beta1.QueryClassRequest"></a>

### QueryClassRequest
QueryClassRequest is the request type for the Query/Class RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class_id` | [string](#string) |  |  |






<a name="coreum.nft.v1beta1.QueryClassResponse"></a>

### QueryClassResponse
QueryClassResponse is the response type for the Query/Class RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class` | [Class](#coreum.nft.v1beta1.Class) |  |  |






<a name="coreum.nft.v1beta1.QueryClassesRequest"></a>

### QueryClassesRequest
QueryClassesRequest is the request type for the Query/Classes RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="coreum.nft.v1beta1.QueryClassesResponse"></a>

### QueryClassesResponse
QueryClassesResponse is the response type for the Query/Classes RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `classes` | [Class](#coreum.nft.v1beta1.Class) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |






<a name="coreum.nft.v1beta1.QueryNFTRequest"></a>

### QueryNFTRequest
QueryNFTRequest is the request type for the Query/NFT RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class_id` | [string](#string) |  |  |
| `id` | [string](#string) |  |  |






<a name="coreum.nft.v1beta1.QueryNFTResponse"></a>

### QueryNFTResponse
QueryNFTResponse is the response type for the Query/NFT RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `nft` | [NFT](#coreum.nft.v1beta1.NFT) |  |  |






<a name="coreum.nft.v1beta1.QueryNFTsRequest"></a>

### QueryNFTsRequest
QueryNFTstRequest is the request type for the Query/NFTs RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class_id` | [string](#string) |  |  |
| `owner` | [string](#string) |  |  |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |






<a name="coreum.nft.v1beta1.QueryNFTsResponse"></a>

### QueryNFTsResponse
QueryNFTsResponse is the response type for the Query/NFTs RPC methods


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `nfts` | [NFT](#coreum.nft.v1beta1.NFT) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |






<a name="coreum.nft.v1beta1.QueryOwnerRequest"></a>

### QueryOwnerRequest
QueryOwnerRequest is the request type for the Query/Owner RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class_id` | [string](#string) |  |  |
| `id` | [string](#string) |  |  |






<a name="coreum.nft.v1beta1.QueryOwnerResponse"></a>

### QueryOwnerResponse
QueryOwnerResponse is the response type for the Query/Owner RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |






<a name="coreum.nft.v1beta1.QuerySupplyRequest"></a>

### QuerySupplyRequest
QuerySupplyRequest is the request type for the Query/Supply RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class_id` | [string](#string) |  |  |






<a name="coreum.nft.v1beta1.QuerySupplyResponse"></a>

### QuerySupplyResponse
QuerySupplyResponse is the response type for the Query/Supply RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [uint64](#uint64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="coreum.nft.v1beta1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Balance` | [QueryBalanceRequest](#coreum.nft.v1beta1.QueryBalanceRequest) | [QueryBalanceResponse](#coreum.nft.v1beta1.QueryBalanceResponse) | Balance queries the number of NFTs of a given class owned by the owner, same as balanceOf in ERC721 | GET|/coreum/nft/v1beta1/balance/{owner}/{class_id}|
| `Owner` | [QueryOwnerRequest](#coreum.nft.v1beta1.QueryOwnerRequest) | [QueryOwnerResponse](#coreum.nft.v1beta1.QueryOwnerResponse) | Owner queries the owner of the NFT based on its class and id, same as ownerOf in ERC721 | GET|/coreum/nft/v1beta1/owner/{class_id}/{id}|
| `Supply` | [QuerySupplyRequest](#coreum.nft.v1beta1.QuerySupplyRequest) | [QuerySupplyResponse](#coreum.nft.v1beta1.QuerySupplyResponse) | Supply queries the number of NFTs from the given class, same as totalSupply of ERC721. | GET|/coreum/nft/v1beta1/supply/{class_id}|
| `NFTs` | [QueryNFTsRequest](#coreum.nft.v1beta1.QueryNFTsRequest) | [QueryNFTsResponse](#coreum.nft.v1beta1.QueryNFTsResponse) | NFTs queries all NFTs of a given class or owner,choose at least one of the two, similar to tokenByIndex in ERC721Enumerable | GET|/coreum/nft/v1beta1/nfts|
| `NFT` | [QueryNFTRequest](#coreum.nft.v1beta1.QueryNFTRequest) | [QueryNFTResponse](#coreum.nft.v1beta1.QueryNFTResponse) | NFT queries an NFT based on its class and id. | GET|/coreum/nft/v1beta1/nfts/{class_id}/{id}|
| `Class` | [QueryClassRequest](#coreum.nft.v1beta1.QueryClassRequest) | [QueryClassResponse](#coreum.nft.v1beta1.QueryClassResponse) | Class queries an NFT class based on its id | GET|/coreum/nft/v1beta1/classes/{class_id}|
| `Classes` | [QueryClassesRequest](#coreum.nft.v1beta1.QueryClassesRequest) | [QueryClassesResponse](#coreum.nft.v1beta1.QueryClassesResponse) | Classes queries all NFT classes | GET|/coreum/nft/v1beta1/classes|

 <!-- end services -->



<a name="coreum/nft/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## coreum/nft/v1beta1/tx.proto



<a name="coreum.nft.v1beta1.MsgSend"></a>

### MsgSend
MsgSend represents a message to send a nft from one account to another account.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class_id` | [string](#string) |  | class_id defines the unique identifier of the nft classification, similar to the contract address of ERC721 |
| `id` | [string](#string) |  | id defines the unique identification of nft |
| `sender` | [string](#string) |  | sender is the address of the owner of nft |
| `receiver` | [string](#string) |  | receiver is the receiver address of nft |






<a name="coreum.nft.v1beta1.MsgSendResponse"></a>

### MsgSendResponse
MsgSendResponse defines the Msg/Send response type.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="coreum.nft.v1beta1.Msg"></a>

### Msg
Msg defines the nft Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Send` | [MsgSend](#coreum.nft.v1beta1.MsgSend) | [MsgSendResponse](#coreum.nft.v1beta1.MsgSendResponse) | Send defines a method to send a nft from one account to another account. | |

 <!-- end services -->



<a name="cosmwasm/wasm/v1/authz.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cosmwasm/wasm/v1/authz.proto



<a name="cosmwasm.wasm.v1.AcceptedMessageKeysFilter"></a>

### AcceptedMessageKeysFilter
AcceptedMessageKeysFilter accept only the specific contract message keys in
the json object to be executed.
Since: wasmd 0.30


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `keys` | [string](#string) | repeated | Messages is the list of unique keys |






<a name="cosmwasm.wasm.v1.AcceptedMessagesFilter"></a>

### AcceptedMessagesFilter
AcceptedMessagesFilter accept only the specific raw contract messages to be
executed.
Since: wasmd 0.30


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `messages` | [bytes](#bytes) | repeated | Messages is the list of raw contract messages |






<a name="cosmwasm.wasm.v1.AllowAllMessagesFilter"></a>

### AllowAllMessagesFilter
AllowAllMessagesFilter is a wildcard to allow any type of contract payload
message.
Since: wasmd 0.30






<a name="cosmwasm.wasm.v1.CombinedLimit"></a>

### CombinedLimit
CombinedLimit defines the maximal amounts that can be sent to a contract and
the maximal number of calls executable. Both need to remain >0 to be valid.
Since: wasmd 0.30


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `calls_remaining` | [uint64](#uint64) |  | Remaining number that is decremented on each execution |
| `amounts` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | Amounts is the maximal amount of tokens transferable to the contract. |






<a name="cosmwasm.wasm.v1.ContractExecutionAuthorization"></a>

### ContractExecutionAuthorization
ContractExecutionAuthorization defines authorization for wasm execute.
Since: wasmd 0.30


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `grants` | [ContractGrant](#cosmwasm.wasm.v1.ContractGrant) | repeated | Grants for contract executions |






<a name="cosmwasm.wasm.v1.ContractGrant"></a>

### ContractGrant
ContractGrant a granted permission for a single contract
Since: wasmd 0.30


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract` | [string](#string) |  | Contract is the bech32 address of the smart contract |
| `limit` | [google.protobuf.Any](#google.protobuf.Any) |  | Limit defines execution limits that are enforced and updated when the grant is applied. When the limit lapsed the grant is removed. |
| `filter` | [google.protobuf.Any](#google.protobuf.Any) |  | Filter define more fine-grained control on the message payload passed to the contract in the operation. When no filter applies on execution, the operation is prohibited. |






<a name="cosmwasm.wasm.v1.ContractMigrationAuthorization"></a>

### ContractMigrationAuthorization
ContractMigrationAuthorization defines authorization for wasm contract
migration. Since: wasmd 0.30


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `grants` | [ContractGrant](#cosmwasm.wasm.v1.ContractGrant) | repeated | Grants for contract migrations |






<a name="cosmwasm.wasm.v1.MaxCallsLimit"></a>

### MaxCallsLimit
MaxCallsLimit limited number of calls to the contract. No funds transferable.
Since: wasmd 0.30


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `remaining` | [uint64](#uint64) |  | Remaining number that is decremented on each execution |






<a name="cosmwasm.wasm.v1.MaxFundsLimit"></a>

### MaxFundsLimit
MaxFundsLimit defines the maximal amounts that can be sent to the contract.
Since: wasmd 0.30


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amounts` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | Amounts is the maximal amount of tokens transferable to the contract. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="cosmwasm/wasm/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cosmwasm/wasm/v1/genesis.proto



<a name="cosmwasm.wasm.v1.Code"></a>

### Code
Code struct encompasses CodeInfo and CodeBytes


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `code_id` | [uint64](#uint64) |  |  |
| `code_info` | [CodeInfo](#cosmwasm.wasm.v1.CodeInfo) |  |  |
| `code_bytes` | [bytes](#bytes) |  |  |
| `pinned` | [bool](#bool) |  | Pinned to wasmvm cache |






<a name="cosmwasm.wasm.v1.Contract"></a>

### Contract
Contract struct encompasses ContractAddress, ContractInfo, and ContractState


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  |  |
| `contract_info` | [ContractInfo](#cosmwasm.wasm.v1.ContractInfo) |  |  |
| `contract_state` | [Model](#cosmwasm.wasm.v1.Model) | repeated |  |
| `contract_code_history` | [ContractCodeHistoryEntry](#cosmwasm.wasm.v1.ContractCodeHistoryEntry) | repeated |  |






<a name="cosmwasm.wasm.v1.GenesisState"></a>

### GenesisState
GenesisState - genesis state of x/wasm


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#cosmwasm.wasm.v1.Params) |  |  |
| `codes` | [Code](#cosmwasm.wasm.v1.Code) | repeated |  |
| `contracts` | [Contract](#cosmwasm.wasm.v1.Contract) | repeated |  |
| `sequences` | [Sequence](#cosmwasm.wasm.v1.Sequence) | repeated |  |






<a name="cosmwasm.wasm.v1.Sequence"></a>

### Sequence
Sequence key and value of an id generation counter


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id_key` | [bytes](#bytes) |  |  |
| `value` | [uint64](#uint64) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="cosmwasm/wasm/v1/ibc.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cosmwasm/wasm/v1/ibc.proto



<a name="cosmwasm.wasm.v1.MsgIBCCloseChannel"></a>

### MsgIBCCloseChannel
MsgIBCCloseChannel port and channel need to be owned by the contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `channel` | [string](#string) |  |  |






<a name="cosmwasm.wasm.v1.MsgIBCSend"></a>

### MsgIBCSend
MsgIBCSend


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `channel` | [string](#string) |  | the channel by which the packet will be sent |
| `timeout_height` | [uint64](#uint64) |  | Timeout height relative to the current block height. The timeout is disabled when set to 0. |
| `timeout_timestamp` | [uint64](#uint64) |  | Timeout timestamp (in nanoseconds) relative to the current block timestamp. The timeout is disabled when set to 0. |
| `data` | [bytes](#bytes) |  | Data is the payload to transfer. We must not make assumption what format or content is in here. |






<a name="cosmwasm.wasm.v1.MsgIBCSendResponse"></a>

### MsgIBCSendResponse
MsgIBCSendResponse


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sequence` | [uint64](#uint64) |  | Sequence number of the IBC packet sent |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="cosmwasm/wasm/v1/proposal.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cosmwasm/wasm/v1/proposal.proto



<a name="cosmwasm.wasm.v1.AccessConfigUpdate"></a>

### AccessConfigUpdate
AccessConfigUpdate contains the code id and the access config to be
applied.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `code_id` | [uint64](#uint64) |  | CodeID is the reference to the stored WASM code to be updated |
| `instantiate_permission` | [AccessConfig](#cosmwasm.wasm.v1.AccessConfig) |  | InstantiatePermission to apply to the set of code ids |






<a name="cosmwasm.wasm.v1.ClearAdminProposal"></a>

### ClearAdminProposal
Deprecated: Do not use. Since wasmd v0.40, there is no longer a need for
an explicit ClearAdminProposal. To clear the admin of a contract,
a simple MsgClearAdmin can be invoked from the x/gov module via
a v1 governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | Title is a short summary |
| `description` | [string](#string) |  | Description is a human readable text |
| `contract` | [string](#string) |  | Contract is the address of the smart contract |






<a name="cosmwasm.wasm.v1.ExecuteContractProposal"></a>

### ExecuteContractProposal
Deprecated: Do not use. Since wasmd v0.40, there is no longer a need for
an explicit ExecuteContractProposal. To call execute on a contract,
a simple MsgExecuteContract can be invoked from the x/gov module via
a v1 governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | Title is a short summary |
| `description` | [string](#string) |  | Description is a human readable text |
| `run_as` | [string](#string) |  | RunAs is the address that is passed to the contract's environment as sender |
| `contract` | [string](#string) |  | Contract is the address of the smart contract |
| `msg` | [bytes](#bytes) |  | Msg json encoded message to be passed to the contract as execute |
| `funds` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | Funds coins that are transferred to the contract on instantiation |






<a name="cosmwasm.wasm.v1.InstantiateContract2Proposal"></a>

### InstantiateContract2Proposal
Deprecated: Do not use. Since wasmd v0.40, there is no longer a need for
an explicit InstantiateContract2Proposal. To instantiate contract 2,
a simple MsgInstantiateContract2 can be invoked from the x/gov module via
a v1 governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | Title is a short summary |
| `description` | [string](#string) |  | Description is a human readable text |
| `run_as` | [string](#string) |  | RunAs is the address that is passed to the contract's enviroment as sender |
| `admin` | [string](#string) |  | Admin is an optional address that can execute migrations |
| `code_id` | [uint64](#uint64) |  | CodeID is the reference to the stored WASM code |
| `label` | [string](#string) |  | Label is optional metadata to be stored with a constract instance. |
| `msg` | [bytes](#bytes) |  | Msg json encode message to be passed to the contract on instantiation |
| `funds` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | Funds coins that are transferred to the contract on instantiation |
| `salt` | [bytes](#bytes) |  | Salt is an arbitrary value provided by the sender. Size can be 1 to 64. |
| `fix_msg` | [bool](#bool) |  | FixMsg include the msg value into the hash for the predictable address. Default is false |






<a name="cosmwasm.wasm.v1.InstantiateContractProposal"></a>

### InstantiateContractProposal
Deprecated: Do not use. Since wasmd v0.40, there is no longer a need for
an explicit InstantiateContractProposal. To instantiate a contract,
a simple MsgInstantiateContract can be invoked from the x/gov module via
a v1 governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | Title is a short summary |
| `description` | [string](#string) |  | Description is a human readable text |
| `run_as` | [string](#string) |  | RunAs is the address that is passed to the contract's environment as sender |
| `admin` | [string](#string) |  | Admin is an optional address that can execute migrations |
| `code_id` | [uint64](#uint64) |  | CodeID is the reference to the stored WASM code |
| `label` | [string](#string) |  | Label is optional metadata to be stored with a constract instance. |
| `msg` | [bytes](#bytes) |  | Msg json encoded message to be passed to the contract on instantiation |
| `funds` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | Funds coins that are transferred to the contract on instantiation |






<a name="cosmwasm.wasm.v1.MigrateContractProposal"></a>

### MigrateContractProposal
Deprecated: Do not use. Since wasmd v0.40, there is no longer a need for
an explicit MigrateContractProposal. To migrate a contract,
a simple MsgMigrateContract can be invoked from the x/gov module via
a v1 governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | Title is a short summary |
| `description` | [string](#string) |  | Description is a human readable text

Note: skipping 3 as this was previously used for unneeded run_as |
| `contract` | [string](#string) |  | Contract is the address of the smart contract |
| `code_id` | [uint64](#uint64) |  | CodeID references the new WASM code |
| `msg` | [bytes](#bytes) |  | Msg json encoded message to be passed to the contract on migration |






<a name="cosmwasm.wasm.v1.PinCodesProposal"></a>

### PinCodesProposal
Deprecated: Do not use. Since wasmd v0.40, there is no longer a need for
an explicit PinCodesProposal. To pin a set of code ids in the wasmvm
cache, a simple MsgPinCodes can be invoked from the x/gov module via
a v1 governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | Title is a short summary |
| `description` | [string](#string) |  | Description is a human readable text |
| `code_ids` | [uint64](#uint64) | repeated | CodeIDs references the new WASM codes |






<a name="cosmwasm.wasm.v1.StoreAndInstantiateContractProposal"></a>

### StoreAndInstantiateContractProposal
Deprecated: Do not use. Since wasmd v0.40, there is no longer a need for
an explicit StoreAndInstantiateContractProposal. To store and instantiate
the contract, a simple MsgStoreAndInstantiateContract can be invoked from
the x/gov module via a v1 governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | Title is a short summary |
| `description` | [string](#string) |  | Description is a human readable text |
| `run_as` | [string](#string) |  | RunAs is the address that is passed to the contract's environment as sender |
| `wasm_byte_code` | [bytes](#bytes) |  | WASMByteCode can be raw or gzip compressed |
| `instantiate_permission` | [AccessConfig](#cosmwasm.wasm.v1.AccessConfig) |  | InstantiatePermission to apply on contract creation, optional |
| `unpin_code` | [bool](#bool) |  | UnpinCode code on upload, optional |
| `admin` | [string](#string) |  | Admin is an optional address that can execute migrations |
| `label` | [string](#string) |  | Label is optional metadata to be stored with a constract instance. |
| `msg` | [bytes](#bytes) |  | Msg json encoded message to be passed to the contract on instantiation |
| `funds` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | Funds coins that are transferred to the contract on instantiation |
| `source` | [string](#string) |  | Source is the URL where the code is hosted |
| `builder` | [string](#string) |  | Builder is the docker image used to build the code deterministically, used for smart contract verification |
| `code_hash` | [bytes](#bytes) |  | CodeHash is the SHA256 sum of the code outputted by builder, used for smart contract verification |






<a name="cosmwasm.wasm.v1.StoreCodeProposal"></a>

### StoreCodeProposal
Deprecated: Do not use. Since wasmd v0.40, there is no longer a need for
an explicit StoreCodeProposal. To submit WASM code to the system,
a simple MsgStoreCode can be invoked from the x/gov module via
a v1 governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | Title is a short summary |
| `description` | [string](#string) |  | Description is a human readable text |
| `run_as` | [string](#string) |  | RunAs is the address that is passed to the contract's environment as sender |
| `wasm_byte_code` | [bytes](#bytes) |  | WASMByteCode can be raw or gzip compressed |
| `instantiate_permission` | [AccessConfig](#cosmwasm.wasm.v1.AccessConfig) |  | InstantiatePermission to apply on contract creation, optional |
| `unpin_code` | [bool](#bool) |  | UnpinCode code on upload, optional |
| `source` | [string](#string) |  | Source is the URL where the code is hosted |
| `builder` | [string](#string) |  | Builder is the docker image used to build the code deterministically, used for smart contract verification |
| `code_hash` | [bytes](#bytes) |  | CodeHash is the SHA256 sum of the code outputted by builder, used for smart contract verification |






<a name="cosmwasm.wasm.v1.SudoContractProposal"></a>

### SudoContractProposal
Deprecated: Do not use. Since wasmd v0.40, there is no longer a need for
an explicit SudoContractProposal. To call sudo on a contract,
a simple MsgSudoContract can be invoked from the x/gov module via
a v1 governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | Title is a short summary |
| `description` | [string](#string) |  | Description is a human readable text |
| `contract` | [string](#string) |  | Contract is the address of the smart contract |
| `msg` | [bytes](#bytes) |  | Msg json encoded message to be passed to the contract as sudo |






<a name="cosmwasm.wasm.v1.UnpinCodesProposal"></a>

### UnpinCodesProposal
Deprecated: Do not use. Since wasmd v0.40, there is no longer a need for
an explicit UnpinCodesProposal. To unpin a set of code ids in the wasmvm
cache, a simple MsgUnpinCodes can be invoked from the x/gov module via
a v1 governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | Title is a short summary |
| `description` | [string](#string) |  | Description is a human readable text |
| `code_ids` | [uint64](#uint64) | repeated | CodeIDs references the WASM codes |






<a name="cosmwasm.wasm.v1.UpdateAdminProposal"></a>

### UpdateAdminProposal
Deprecated: Do not use. Since wasmd v0.40, there is no longer a need for
an explicit UpdateAdminProposal. To set an admin for a contract,
a simple MsgUpdateAdmin can be invoked from the x/gov module via
a v1 governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | Title is a short summary |
| `description` | [string](#string) |  | Description is a human readable text |
| `new_admin` | [string](#string) |  | NewAdmin address to be set |
| `contract` | [string](#string) |  | Contract is the address of the smart contract |






<a name="cosmwasm.wasm.v1.UpdateInstantiateConfigProposal"></a>

### UpdateInstantiateConfigProposal
Deprecated: Do not use. Since wasmd v0.40, there is no longer a need for
an explicit UpdateInstantiateConfigProposal. To update instantiate config
to a set of code ids, a simple MsgUpdateInstantiateConfig can be invoked from
the x/gov module via a v1 governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | Title is a short summary |
| `description` | [string](#string) |  | Description is a human readable text |
| `access_config_updates` | [AccessConfigUpdate](#cosmwasm.wasm.v1.AccessConfigUpdate) | repeated | AccessConfigUpdate contains the list of code ids and the access config to be applied. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="cosmwasm/wasm/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cosmwasm/wasm/v1/query.proto



<a name="cosmwasm.wasm.v1.CodeInfoResponse"></a>

### CodeInfoResponse
CodeInfoResponse contains code meta data from CodeInfo


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `code_id` | [uint64](#uint64) |  | id for legacy support |
| `creator` | [string](#string) |  |  |
| `data_hash` | [bytes](#bytes) |  |  |
| `instantiate_permission` | [AccessConfig](#cosmwasm.wasm.v1.AccessConfig) |  |  |






<a name="cosmwasm.wasm.v1.QueryAllContractStateRequest"></a>

### QueryAllContractStateRequest
QueryAllContractStateRequest is the request type for the
Query/AllContractState RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the address of the contract |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="cosmwasm.wasm.v1.QueryAllContractStateResponse"></a>

### QueryAllContractStateResponse
QueryAllContractStateResponse is the response type for the
Query/AllContractState RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `models` | [Model](#cosmwasm.wasm.v1.Model) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |






<a name="cosmwasm.wasm.v1.QueryCodeRequest"></a>

### QueryCodeRequest
QueryCodeRequest is the request type for the Query/Code RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `code_id` | [uint64](#uint64) |  | grpc-gateway_out does not support Go style CodID |






<a name="cosmwasm.wasm.v1.QueryCodeResponse"></a>

### QueryCodeResponse
QueryCodeResponse is the response type for the Query/Code RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `code_info` | [CodeInfoResponse](#cosmwasm.wasm.v1.CodeInfoResponse) |  |  |
| `data` | [bytes](#bytes) |  |  |






<a name="cosmwasm.wasm.v1.QueryCodesRequest"></a>

### QueryCodesRequest
QueryCodesRequest is the request type for the Query/Codes RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="cosmwasm.wasm.v1.QueryCodesResponse"></a>

### QueryCodesResponse
QueryCodesResponse is the response type for the Query/Codes RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `code_infos` | [CodeInfoResponse](#cosmwasm.wasm.v1.CodeInfoResponse) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |






<a name="cosmwasm.wasm.v1.QueryContractHistoryRequest"></a>

### QueryContractHistoryRequest
QueryContractHistoryRequest is the request type for the Query/ContractHistory
RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the address of the contract to query |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="cosmwasm.wasm.v1.QueryContractHistoryResponse"></a>

### QueryContractHistoryResponse
QueryContractHistoryResponse is the response type for the
Query/ContractHistory RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `entries` | [ContractCodeHistoryEntry](#cosmwasm.wasm.v1.ContractCodeHistoryEntry) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |






<a name="cosmwasm.wasm.v1.QueryContractInfoRequest"></a>

### QueryContractInfoRequest
QueryContractInfoRequest is the request type for the Query/ContractInfo RPC
method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the address of the contract to query |






<a name="cosmwasm.wasm.v1.QueryContractInfoResponse"></a>

### QueryContractInfoResponse
QueryContractInfoResponse is the response type for the Query/ContractInfo RPC
method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the address of the contract |
| `contract_info` | [ContractInfo](#cosmwasm.wasm.v1.ContractInfo) |  |  |






<a name="cosmwasm.wasm.v1.QueryContractsByCodeRequest"></a>

### QueryContractsByCodeRequest
QueryContractsByCodeRequest is the request type for the Query/ContractsByCode
RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `code_id` | [uint64](#uint64) |  | grpc-gateway_out does not support Go style CodID |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="cosmwasm.wasm.v1.QueryContractsByCodeResponse"></a>

### QueryContractsByCodeResponse
QueryContractsByCodeResponse is the response type for the
Query/ContractsByCode RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contracts` | [string](#string) | repeated | contracts are a set of contract addresses |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |






<a name="cosmwasm.wasm.v1.QueryContractsByCreatorRequest"></a>

### QueryContractsByCreatorRequest
QueryContractsByCreatorRequest is the request type for the
Query/ContractsByCreator RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `creator_address` | [string](#string) |  | CreatorAddress is the address of contract creator |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | Pagination defines an optional pagination for the request. |






<a name="cosmwasm.wasm.v1.QueryContractsByCreatorResponse"></a>

### QueryContractsByCreatorResponse
QueryContractsByCreatorResponse is the response type for the
Query/ContractsByCreator RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_addresses` | [string](#string) | repeated | ContractAddresses result set |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | Pagination defines the pagination in the response. |






<a name="cosmwasm.wasm.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="cosmwasm.wasm.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#cosmwasm.wasm.v1.Params) |  | params defines the parameters of the module. |






<a name="cosmwasm.wasm.v1.QueryPinnedCodesRequest"></a>

### QueryPinnedCodesRequest
QueryPinnedCodesRequest is the request type for the Query/PinnedCodes
RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="cosmwasm.wasm.v1.QueryPinnedCodesResponse"></a>

### QueryPinnedCodesResponse
QueryPinnedCodesResponse is the response type for the
Query/PinnedCodes RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `code_ids` | [uint64](#uint64) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |






<a name="cosmwasm.wasm.v1.QueryRawContractStateRequest"></a>

### QueryRawContractStateRequest
QueryRawContractStateRequest is the request type for the
Query/RawContractState RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the address of the contract |
| `query_data` | [bytes](#bytes) |  |  |






<a name="cosmwasm.wasm.v1.QueryRawContractStateResponse"></a>

### QueryRawContractStateResponse
QueryRawContractStateResponse is the response type for the
Query/RawContractState RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  | Data contains the raw store data |






<a name="cosmwasm.wasm.v1.QuerySmartContractStateRequest"></a>

### QuerySmartContractStateRequest
QuerySmartContractStateRequest is the request type for the
Query/SmartContractState RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the address of the contract |
| `query_data` | [bytes](#bytes) |  | QueryData contains the query data passed to the contract |






<a name="cosmwasm.wasm.v1.QuerySmartContractStateResponse"></a>

### QuerySmartContractStateResponse
QuerySmartContractStateResponse is the response type for the
Query/SmartContractState RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  | Data contains the json data returned from the smart contract |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="cosmwasm.wasm.v1.Query"></a>

### Query
Query provides defines the gRPC querier service

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `ContractInfo` | [QueryContractInfoRequest](#cosmwasm.wasm.v1.QueryContractInfoRequest) | [QueryContractInfoResponse](#cosmwasm.wasm.v1.QueryContractInfoResponse) | ContractInfo gets the contract meta data | GET|/cosmwasm/wasm/v1/contract/{address}|
| `ContractHistory` | [QueryContractHistoryRequest](#cosmwasm.wasm.v1.QueryContractHistoryRequest) | [QueryContractHistoryResponse](#cosmwasm.wasm.v1.QueryContractHistoryResponse) | ContractHistory gets the contract code history | GET|/cosmwasm/wasm/v1/contract/{address}/history|
| `ContractsByCode` | [QueryContractsByCodeRequest](#cosmwasm.wasm.v1.QueryContractsByCodeRequest) | [QueryContractsByCodeResponse](#cosmwasm.wasm.v1.QueryContractsByCodeResponse) | ContractsByCode lists all smart contracts for a code id | GET|/cosmwasm/wasm/v1/code/{code_id}/contracts|
| `AllContractState` | [QueryAllContractStateRequest](#cosmwasm.wasm.v1.QueryAllContractStateRequest) | [QueryAllContractStateResponse](#cosmwasm.wasm.v1.QueryAllContractStateResponse) | AllContractState gets all raw store data for a single contract | GET|/cosmwasm/wasm/v1/contract/{address}/state|
| `RawContractState` | [QueryRawContractStateRequest](#cosmwasm.wasm.v1.QueryRawContractStateRequest) | [QueryRawContractStateResponse](#cosmwasm.wasm.v1.QueryRawContractStateResponse) | RawContractState gets single key from the raw store data of a contract | GET|/cosmwasm/wasm/v1/contract/{address}/raw/{query_data}|
| `SmartContractState` | [QuerySmartContractStateRequest](#cosmwasm.wasm.v1.QuerySmartContractStateRequest) | [QuerySmartContractStateResponse](#cosmwasm.wasm.v1.QuerySmartContractStateResponse) | SmartContractState get smart query result from the contract | GET|/cosmwasm/wasm/v1/contract/{address}/smart/{query_data}|
| `Code` | [QueryCodeRequest](#cosmwasm.wasm.v1.QueryCodeRequest) | [QueryCodeResponse](#cosmwasm.wasm.v1.QueryCodeResponse) | Code gets the binary code and metadata for a singe wasm code | GET|/cosmwasm/wasm/v1/code/{code_id}|
| `Codes` | [QueryCodesRequest](#cosmwasm.wasm.v1.QueryCodesRequest) | [QueryCodesResponse](#cosmwasm.wasm.v1.QueryCodesResponse) | Codes gets the metadata for all stored wasm codes | GET|/cosmwasm/wasm/v1/code|
| `PinnedCodes` | [QueryPinnedCodesRequest](#cosmwasm.wasm.v1.QueryPinnedCodesRequest) | [QueryPinnedCodesResponse](#cosmwasm.wasm.v1.QueryPinnedCodesResponse) | PinnedCodes gets the pinned code ids | GET|/cosmwasm/wasm/v1/codes/pinned|
| `Params` | [QueryParamsRequest](#cosmwasm.wasm.v1.QueryParamsRequest) | [QueryParamsResponse](#cosmwasm.wasm.v1.QueryParamsResponse) | Params gets the module params | GET|/cosmwasm/wasm/v1/codes/params|
| `ContractsByCreator` | [QueryContractsByCreatorRequest](#cosmwasm.wasm.v1.QueryContractsByCreatorRequest) | [QueryContractsByCreatorResponse](#cosmwasm.wasm.v1.QueryContractsByCreatorResponse) | ContractsByCreator gets the contracts by creator | GET|/cosmwasm/wasm/v1/contracts/creator/{creator_address}|

 <!-- end services -->



<a name="cosmwasm/wasm/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cosmwasm/wasm/v1/tx.proto



<a name="cosmwasm.wasm.v1.MsgAddCodeUploadParamsAddresses"></a>

### MsgAddCodeUploadParamsAddresses
MsgAddCodeUploadParamsAddresses is the
MsgAddCodeUploadParamsAddresses request type.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | Authority is the address of the governance account. |
| `addresses` | [string](#string) | repeated |  |






<a name="cosmwasm.wasm.v1.MsgAddCodeUploadParamsAddressesResponse"></a>

### MsgAddCodeUploadParamsAddressesResponse
MsgAddCodeUploadParamsAddressesResponse defines the response
structure for executing a MsgAddCodeUploadParamsAddresses message.






<a name="cosmwasm.wasm.v1.MsgClearAdmin"></a>

### MsgClearAdmin
MsgClearAdmin removes any admin stored for a smart contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  | Sender is the actor that signed the messages |
| `contract` | [string](#string) |  | Contract is the address of the smart contract |






<a name="cosmwasm.wasm.v1.MsgClearAdminResponse"></a>

### MsgClearAdminResponse
MsgClearAdminResponse returns empty data






<a name="cosmwasm.wasm.v1.MsgExecuteContract"></a>

### MsgExecuteContract
MsgExecuteContract submits the given message data to a smart contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  | Sender is the that actor that signed the messages |
| `contract` | [string](#string) |  | Contract is the address of the smart contract |
| `msg` | [bytes](#bytes) |  | Msg json encoded message to be passed to the contract |
| `funds` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | Funds coins that are transferred to the contract on execution |






<a name="cosmwasm.wasm.v1.MsgExecuteContractResponse"></a>

### MsgExecuteContractResponse
MsgExecuteContractResponse returns execution result data.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  | Data contains bytes to returned from the contract |






<a name="cosmwasm.wasm.v1.MsgInstantiateContract"></a>

### MsgInstantiateContract
MsgInstantiateContract create a new smart contract instance for the given
code id.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  | Sender is the that actor that signed the messages |
| `admin` | [string](#string) |  | Admin is an optional address that can execute migrations |
| `code_id` | [uint64](#uint64) |  | CodeID is the reference to the stored WASM code |
| `label` | [string](#string) |  | Label is optional metadata to be stored with a contract instance. |
| `msg` | [bytes](#bytes) |  | Msg json encoded message to be passed to the contract on instantiation |
| `funds` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | Funds coins that are transferred to the contract on instantiation |






<a name="cosmwasm.wasm.v1.MsgInstantiateContract2"></a>

### MsgInstantiateContract2
MsgInstantiateContract2 create a new smart contract instance for the given
code id with a predicable address.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  | Sender is the that actor that signed the messages |
| `admin` | [string](#string) |  | Admin is an optional address that can execute migrations |
| `code_id` | [uint64](#uint64) |  | CodeID is the reference to the stored WASM code |
| `label` | [string](#string) |  | Label is optional metadata to be stored with a contract instance. |
| `msg` | [bytes](#bytes) |  | Msg json encoded message to be passed to the contract on instantiation |
| `funds` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | Funds coins that are transferred to the contract on instantiation |
| `salt` | [bytes](#bytes) |  | Salt is an arbitrary value provided by the sender. Size can be 1 to 64. |
| `fix_msg` | [bool](#bool) |  | FixMsg include the msg value into the hash for the predictable address. Default is false |






<a name="cosmwasm.wasm.v1.MsgInstantiateContract2Response"></a>

### MsgInstantiateContract2Response
MsgInstantiateContract2Response return instantiation result data


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | Address is the bech32 address of the new contract instance. |
| `data` | [bytes](#bytes) |  | Data contains bytes to returned from the contract |






<a name="cosmwasm.wasm.v1.MsgInstantiateContractResponse"></a>

### MsgInstantiateContractResponse
MsgInstantiateContractResponse return instantiation result data


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | Address is the bech32 address of the new contract instance. |
| `data` | [bytes](#bytes) |  | Data contains bytes to returned from the contract |






<a name="cosmwasm.wasm.v1.MsgMigrateContract"></a>

### MsgMigrateContract
MsgMigrateContract runs a code upgrade/ downgrade for a smart contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  | Sender is the that actor that signed the messages |
| `contract` | [string](#string) |  | Contract is the address of the smart contract |
| `code_id` | [uint64](#uint64) |  | CodeID references the new WASM code |
| `msg` | [bytes](#bytes) |  | Msg json encoded message to be passed to the contract on migration |






<a name="cosmwasm.wasm.v1.MsgMigrateContractResponse"></a>

### MsgMigrateContractResponse
MsgMigrateContractResponse returns contract migration result data.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  | Data contains same raw bytes returned as data from the wasm contract. (May be empty) |






<a name="cosmwasm.wasm.v1.MsgPinCodes"></a>

### MsgPinCodes
MsgPinCodes is the MsgPinCodes request type.

Since: 0.40


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | Authority is the address of the governance account. |
| `code_ids` | [uint64](#uint64) | repeated | CodeIDs references the new WASM codes |






<a name="cosmwasm.wasm.v1.MsgPinCodesResponse"></a>

### MsgPinCodesResponse
MsgPinCodesResponse defines the response structure for executing a
MsgPinCodes message.

Since: 0.40






<a name="cosmwasm.wasm.v1.MsgRemoveCodeUploadParamsAddresses"></a>

### MsgRemoveCodeUploadParamsAddresses
MsgRemoveCodeUploadParamsAddresses is the
MsgRemoveCodeUploadParamsAddresses request type.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | Authority is the address of the governance account. |
| `addresses` | [string](#string) | repeated |  |






<a name="cosmwasm.wasm.v1.MsgRemoveCodeUploadParamsAddressesResponse"></a>

### MsgRemoveCodeUploadParamsAddressesResponse
MsgRemoveCodeUploadParamsAddressesResponse defines the response
structure for executing a MsgRemoveCodeUploadParamsAddresses message.






<a name="cosmwasm.wasm.v1.MsgStoreAndInstantiateContract"></a>

### MsgStoreAndInstantiateContract
MsgStoreAndInstantiateContract is the MsgStoreAndInstantiateContract
request type.

Since: 0.40


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | Authority is the address of the governance account. |
| `wasm_byte_code` | [bytes](#bytes) |  | WASMByteCode can be raw or gzip compressed |
| `instantiate_permission` | [AccessConfig](#cosmwasm.wasm.v1.AccessConfig) |  | InstantiatePermission to apply on contract creation, optional |
| `unpin_code` | [bool](#bool) |  | UnpinCode code on upload, optional. As default the uploaded contract is pinned to cache. |
| `admin` | [string](#string) |  | Admin is an optional address that can execute migrations |
| `label` | [string](#string) |  | Label is optional metadata to be stored with a constract instance. |
| `msg` | [bytes](#bytes) |  | Msg json encoded message to be passed to the contract on instantiation |
| `funds` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | Funds coins that are transferred from the authority account to the contract on instantiation |
| `source` | [string](#string) |  | Source is the URL where the code is hosted |
| `builder` | [string](#string) |  | Builder is the docker image used to build the code deterministically, used for smart contract verification |
| `code_hash` | [bytes](#bytes) |  | CodeHash is the SHA256 sum of the code outputted by builder, used for smart contract verification |






<a name="cosmwasm.wasm.v1.MsgStoreAndInstantiateContractResponse"></a>

### MsgStoreAndInstantiateContractResponse
MsgStoreAndInstantiateContractResponse defines the response structure
for executing a MsgStoreAndInstantiateContract message.

Since: 0.40


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | Address is the bech32 address of the new contract instance. |
| `data` | [bytes](#bytes) |  | Data contains bytes to returned from the contract |






<a name="cosmwasm.wasm.v1.MsgStoreCode"></a>

### MsgStoreCode
MsgStoreCode submit Wasm code to the system


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  | Sender is the actor that signed the messages |
| `wasm_byte_code` | [bytes](#bytes) |  | WASMByteCode can be raw or gzip compressed |
| `instantiate_permission` | [AccessConfig](#cosmwasm.wasm.v1.AccessConfig) |  | InstantiatePermission access control to apply on contract creation, optional |






<a name="cosmwasm.wasm.v1.MsgStoreCodeResponse"></a>

### MsgStoreCodeResponse
MsgStoreCodeResponse returns store result data.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `code_id` | [uint64](#uint64) |  | CodeID is the reference to the stored WASM code |
| `checksum` | [bytes](#bytes) |  | Checksum is the sha256 hash of the stored code |






<a name="cosmwasm.wasm.v1.MsgSudoContract"></a>

### MsgSudoContract
MsgSudoContract is the MsgSudoContract request type.

Since: 0.40


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | Authority is the address of the governance account. |
| `contract` | [string](#string) |  | Contract is the address of the smart contract |
| `msg` | [bytes](#bytes) |  | Msg json encoded message to be passed to the contract as sudo |






<a name="cosmwasm.wasm.v1.MsgSudoContractResponse"></a>

### MsgSudoContractResponse
MsgSudoContractResponse defines the response structure for executing a
MsgSudoContract message.

Since: 0.40


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  | Data contains bytes to returned from the contract |






<a name="cosmwasm.wasm.v1.MsgUnpinCodes"></a>

### MsgUnpinCodes
MsgUnpinCodes is the MsgUnpinCodes request type.

Since: 0.40


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | Authority is the address of the governance account. |
| `code_ids` | [uint64](#uint64) | repeated | CodeIDs references the WASM codes |






<a name="cosmwasm.wasm.v1.MsgUnpinCodesResponse"></a>

### MsgUnpinCodesResponse
MsgUnpinCodesResponse defines the response structure for executing a
MsgUnpinCodes message.

Since: 0.40






<a name="cosmwasm.wasm.v1.MsgUpdateAdmin"></a>

### MsgUpdateAdmin
MsgUpdateAdmin sets a new admin for a smart contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  | Sender is the that actor that signed the messages |
| `new_admin` | [string](#string) |  | NewAdmin address to be set |
| `contract` | [string](#string) |  | Contract is the address of the smart contract |






<a name="cosmwasm.wasm.v1.MsgUpdateAdminResponse"></a>

### MsgUpdateAdminResponse
MsgUpdateAdminResponse returns empty data






<a name="cosmwasm.wasm.v1.MsgUpdateInstantiateConfig"></a>

### MsgUpdateInstantiateConfig
MsgUpdateInstantiateConfig updates instantiate config for a smart contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  | Sender is the that actor that signed the messages |
| `code_id` | [uint64](#uint64) |  | CodeID references the stored WASM code |
| `new_instantiate_permission` | [AccessConfig](#cosmwasm.wasm.v1.AccessConfig) |  | NewInstantiatePermission is the new access control |






<a name="cosmwasm.wasm.v1.MsgUpdateInstantiateConfigResponse"></a>

### MsgUpdateInstantiateConfigResponse
MsgUpdateInstantiateConfigResponse returns empty data






<a name="cosmwasm.wasm.v1.MsgUpdateParams"></a>

### MsgUpdateParams
MsgUpdateParams is the MsgUpdateParams request type.

Since: 0.40


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | Authority is the address of the governance account. |
| `params` | [Params](#cosmwasm.wasm.v1.Params) |  | params defines the x/wasm parameters to update.

NOTE: All parameters must be supplied. |






<a name="cosmwasm.wasm.v1.MsgUpdateParamsResponse"></a>

### MsgUpdateParamsResponse
MsgUpdateParamsResponse defines the response structure for executing a
MsgUpdateParams message.

Since: 0.40





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="cosmwasm.wasm.v1.Msg"></a>

### Msg
Msg defines the wasm Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `StoreCode` | [MsgStoreCode](#cosmwasm.wasm.v1.MsgStoreCode) | [MsgStoreCodeResponse](#cosmwasm.wasm.v1.MsgStoreCodeResponse) | StoreCode to submit Wasm code to the system | |
| `InstantiateContract` | [MsgInstantiateContract](#cosmwasm.wasm.v1.MsgInstantiateContract) | [MsgInstantiateContractResponse](#cosmwasm.wasm.v1.MsgInstantiateContractResponse) | InstantiateContract creates a new smart contract instance for the given code id. | |
| `InstantiateContract2` | [MsgInstantiateContract2](#cosmwasm.wasm.v1.MsgInstantiateContract2) | [MsgInstantiateContract2Response](#cosmwasm.wasm.v1.MsgInstantiateContract2Response) | InstantiateContract2 creates a new smart contract instance for the given code id with a predictable address | |
| `ExecuteContract` | [MsgExecuteContract](#cosmwasm.wasm.v1.MsgExecuteContract) | [MsgExecuteContractResponse](#cosmwasm.wasm.v1.MsgExecuteContractResponse) | Execute submits the given message data to a smart contract | |
| `MigrateContract` | [MsgMigrateContract](#cosmwasm.wasm.v1.MsgMigrateContract) | [MsgMigrateContractResponse](#cosmwasm.wasm.v1.MsgMigrateContractResponse) | Migrate runs a code upgrade/ downgrade for a smart contract | |
| `UpdateAdmin` | [MsgUpdateAdmin](#cosmwasm.wasm.v1.MsgUpdateAdmin) | [MsgUpdateAdminResponse](#cosmwasm.wasm.v1.MsgUpdateAdminResponse) | UpdateAdmin sets a new admin for a smart contract | |
| `ClearAdmin` | [MsgClearAdmin](#cosmwasm.wasm.v1.MsgClearAdmin) | [MsgClearAdminResponse](#cosmwasm.wasm.v1.MsgClearAdminResponse) | ClearAdmin removes any admin stored for a smart contract | |
| `UpdateInstantiateConfig` | [MsgUpdateInstantiateConfig](#cosmwasm.wasm.v1.MsgUpdateInstantiateConfig) | [MsgUpdateInstantiateConfigResponse](#cosmwasm.wasm.v1.MsgUpdateInstantiateConfigResponse) | UpdateInstantiateConfig updates instantiate config for a smart contract | |
| `UpdateParams` | [MsgUpdateParams](#cosmwasm.wasm.v1.MsgUpdateParams) | [MsgUpdateParamsResponse](#cosmwasm.wasm.v1.MsgUpdateParamsResponse) | UpdateParams defines a governance operation for updating the x/wasm module parameters. The authority is defined in the keeper.

Since: 0.40 | |
| `SudoContract` | [MsgSudoContract](#cosmwasm.wasm.v1.MsgSudoContract) | [MsgSudoContractResponse](#cosmwasm.wasm.v1.MsgSudoContractResponse) | SudoContract defines a governance operation for calling sudo on a contract. The authority is defined in the keeper.

Since: 0.40 | |
| `PinCodes` | [MsgPinCodes](#cosmwasm.wasm.v1.MsgPinCodes) | [MsgPinCodesResponse](#cosmwasm.wasm.v1.MsgPinCodesResponse) | PinCodes defines a governance operation for pinning a set of code ids in the wasmvm cache. The authority is defined in the keeper.

Since: 0.40 | |
| `UnpinCodes` | [MsgUnpinCodes](#cosmwasm.wasm.v1.MsgUnpinCodes) | [MsgUnpinCodesResponse](#cosmwasm.wasm.v1.MsgUnpinCodesResponse) | UnpinCodes defines a governance operation for unpinning a set of code ids in the wasmvm cache. The authority is defined in the keeper.

Since: 0.40 | |
| `StoreAndInstantiateContract` | [MsgStoreAndInstantiateContract](#cosmwasm.wasm.v1.MsgStoreAndInstantiateContract) | [MsgStoreAndInstantiateContractResponse](#cosmwasm.wasm.v1.MsgStoreAndInstantiateContractResponse) | StoreAndInstantiateContract defines a governance operation for storing and instantiating the contract. The authority is defined in the keeper.

Since: 0.40 | |
| `RemoveCodeUploadParamsAddresses` | [MsgRemoveCodeUploadParamsAddresses](#cosmwasm.wasm.v1.MsgRemoveCodeUploadParamsAddresses) | [MsgRemoveCodeUploadParamsAddressesResponse](#cosmwasm.wasm.v1.MsgRemoveCodeUploadParamsAddressesResponse) | RemoveCodeUploadParamsAddresses defines a governance operation for removing addresses from code upload params. The authority is defined in the keeper. | |
| `AddCodeUploadParamsAddresses` | [MsgAddCodeUploadParamsAddresses](#cosmwasm.wasm.v1.MsgAddCodeUploadParamsAddresses) | [MsgAddCodeUploadParamsAddressesResponse](#cosmwasm.wasm.v1.MsgAddCodeUploadParamsAddressesResponse) | AddCodeUploadParamsAddresses defines a governance operation for adding addresses to code upload params. The authority is defined in the keeper. | |

 <!-- end services -->



<a name="cosmwasm/wasm/v1/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cosmwasm/wasm/v1/types.proto



<a name="cosmwasm.wasm.v1.AbsoluteTxPosition"></a>

### AbsoluteTxPosition
AbsoluteTxPosition is a unique transaction position that allows for global
ordering of transactions.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `block_height` | [uint64](#uint64) |  | BlockHeight is the block the contract was created at |
| `tx_index` | [uint64](#uint64) |  | TxIndex is a monotonic counter within the block (actual transaction index, or gas consumed) |






<a name="cosmwasm.wasm.v1.AccessConfig"></a>

### AccessConfig
AccessConfig access control type.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `permission` | [AccessType](#cosmwasm.wasm.v1.AccessType) |  |  |
| `addresses` | [string](#string) | repeated |  |






<a name="cosmwasm.wasm.v1.AccessTypeParam"></a>

### AccessTypeParam
AccessTypeParam


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [AccessType](#cosmwasm.wasm.v1.AccessType) |  |  |






<a name="cosmwasm.wasm.v1.CodeInfo"></a>

### CodeInfo
CodeInfo is data for the uploaded contract WASM code


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `code_hash` | [bytes](#bytes) |  | CodeHash is the unique identifier created by wasmvm |
| `creator` | [string](#string) |  | Creator address who initially stored the code |
| `instantiate_config` | [AccessConfig](#cosmwasm.wasm.v1.AccessConfig) |  | InstantiateConfig access control to apply on contract creation, optional |






<a name="cosmwasm.wasm.v1.ContractCodeHistoryEntry"></a>

### ContractCodeHistoryEntry
ContractCodeHistoryEntry metadata to a contract.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `operation` | [ContractCodeHistoryOperationType](#cosmwasm.wasm.v1.ContractCodeHistoryOperationType) |  |  |
| `code_id` | [uint64](#uint64) |  | CodeID is the reference to the stored WASM code |
| `updated` | [AbsoluteTxPosition](#cosmwasm.wasm.v1.AbsoluteTxPosition) |  | Updated Tx position when the operation was executed. |
| `msg` | [bytes](#bytes) |  |  |






<a name="cosmwasm.wasm.v1.ContractInfo"></a>

### ContractInfo
ContractInfo stores a WASM contract instance


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `code_id` | [uint64](#uint64) |  | CodeID is the reference to the stored Wasm code |
| `creator` | [string](#string) |  | Creator address who initially instantiated the contract |
| `admin` | [string](#string) |  | Admin is an optional address that can execute migrations |
| `label` | [string](#string) |  | Label is optional metadata to be stored with a contract instance. |
| `created` | [AbsoluteTxPosition](#cosmwasm.wasm.v1.AbsoluteTxPosition) |  | Created Tx position when the contract was instantiated. |
| `ibc_port_id` | [string](#string) |  |  |
| `extension` | [google.protobuf.Any](#google.protobuf.Any) |  | Extension is an extension point to store custom metadata within the persistence model. |






<a name="cosmwasm.wasm.v1.Model"></a>

### Model
Model is a struct that holds a KV pair


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [bytes](#bytes) |  | hex-encode key to read it better (this is often ascii) |
| `value` | [bytes](#bytes) |  | base64-encode raw value |






<a name="cosmwasm.wasm.v1.Params"></a>

### Params
Params defines the set of wasm parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `code_upload_access` | [AccessConfig](#cosmwasm.wasm.v1.AccessConfig) |  |  |
| `instantiate_default_permission` | [AccessType](#cosmwasm.wasm.v1.AccessType) |  |  |





 <!-- end messages -->


<a name="cosmwasm.wasm.v1.AccessType"></a>

### AccessType
AccessType permission types

| Name | Number | Description |
| ---- | ------ | ----------- |
| ACCESS_TYPE_UNSPECIFIED | 0 | AccessTypeUnspecified placeholder for empty value |
| ACCESS_TYPE_NOBODY | 1 | AccessTypeNobody forbidden |
| ACCESS_TYPE_EVERYBODY | 3 | AccessTypeEverybody unrestricted |
| ACCESS_TYPE_ANY_OF_ADDRESSES | 4 | AccessTypeAnyOfAddresses allow any of the addresses |



<a name="cosmwasm.wasm.v1.ContractCodeHistoryOperationType"></a>

### ContractCodeHistoryOperationType
ContractCodeHistoryOperationType actions that caused a code change

| Name | Number | Description |
| ---- | ------ | ----------- |
| CONTRACT_CODE_HISTORY_OPERATION_TYPE_UNSPECIFIED | 0 | ContractCodeHistoryOperationTypeUnspecified placeholder for empty value |
| CONTRACT_CODE_HISTORY_OPERATION_TYPE_INIT | 1 | ContractCodeHistoryOperationTypeInit on chain contract instantiation |
| CONTRACT_CODE_HISTORY_OPERATION_TYPE_MIGRATE | 2 | ContractCodeHistoryOperationTypeMigrate code migration |
| CONTRACT_CODE_HISTORY_OPERATION_TYPE_GENESIS | 3 | ContractCodeHistoryOperationTypeGenesis based on genesis data |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |


<!-- markdown-link-check-enable -->
