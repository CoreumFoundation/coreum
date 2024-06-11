use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::Uint128;

#[cw_serde]
pub struct InstantiateMsg {
    pub denom: String,
    pub issuance_msg: IssuanceMsg,
}

#[cw_serde]
pub struct IssuanceMsg {
    pub extra_data: Option<String>,
}

#[cw_serde]
pub enum ExecuteMsg {}

#[cw_serde]
pub enum SudoMsg {
    ExtensionTransfer {
        recipient: String,
        sender: String,
        transfer_amount: Uint128,
        commission_amount: Uint128,
        burn_amount: Uint128,
        context: TransferContext,
    },
}

#[cw_serde]
pub struct TransferContext {
    pub sender_is_smart_contract: bool,
    pub recipient_is_smart_contract: bool,
    pub ibc_purpose: IBCPurpose,
}

#[cw_serde]
pub enum IBCPurpose {
    None,
    Out,
    In,
    Ack,
    Timeout,
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(QueryIssuanceMsgResponse)]
    QueryIssuanceMsg {},
}

#[cw_serde]
pub struct QueryIssuanceMsgResponse {
    pub test: Option<String>,
}
